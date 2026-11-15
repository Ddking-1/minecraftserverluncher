// Async skin-processing queue. Uploads are accepted with HTTP 202 and a
// task_id; a small worker pool processes them off the request path so the
// backend stays responsive under load.
package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"
)

// SkinJob is a unit of work in the skin processing queue.
type SkinJob struct {
	ID        string
	AccountID int64
	Variant   string
	Data      []byte
	Created   time.Time
}

// JobResult records the outcome of a processed skin job, retained for polling.
type JobResult struct {
	ID        string
	AccountID int64
	Status    string // "queued", "processing", "done", "failed"
	Err       string
	DoneAt    time.Time
}

const (
	// workerCount bounds the pool size.
	workerCount = 4
	// maxJobResults bounds retained polling results to stop unbounded growth
	// when clients never poll (or poll after the fact).
	maxJobResults = 512
	// resultTTL is how long a finished job result is kept.
	resultTTL = 15 * time.Minute
)

// skinQueue is the in-memory task queue and result store.
type skinQueue struct {
	jobs      chan SkinJob
	mu        sync.Mutex
	results   map[string]*JobResult
	nextID    uint64
	process   func(SkinJob) error
	closeOnce sync.Once
}

// newSkinQueue creates and starts the worker pool. process is the persistence
// step executed for every job (it may be nil in tests).
func newSkinQueue(process func(SkinJob) error) *skinQueue {
	q := &skinQueue{
		jobs:    make(chan SkinJob, 64),
		results: make(map[string]*JobResult),
		process: process,
	}
	for i := 0; i < workerCount; i++ {
		go q.worker()
	}
	return q
}

// Close shuts down the worker pool. Workers finish any already-queued jobs and
// then exit. It is safe to call Close more than once.
func (q *skinQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.jobs)
	})
}

// worker processes jobs from the queue until it is closed.
func (q *skinQueue) worker() {
	for job := range q.jobs {
		q.mu.Lock()
		if r, ok := q.results[job.ID]; ok {
			r.Status = "processing"
		}
		q.mu.Unlock()

		var err error
		if q.process != nil {
			err = q.process(job)
		}

		q.mu.Lock()
		if r, ok := q.results[job.ID]; ok {
			if err != nil {
				r.Status = "failed"
				r.Err = err.Error()
			} else {
				r.Status = "done"
			}
			r.DoneAt = time.Now().UTC()
		}
		q.mu.Unlock()
	}
}

// enqueue adds a job and returns its id, or an error if the queue is full.
func (q *skinQueue) enqueue(accountID int64, variant string, data []byte) (string, error) {
	q.mu.Lock()
	q.nextID++
	id := q.nextID
	q.mu.Unlock()

	token := make([]byte, 8)
	_, _ = rand.Read(token)
	jobID := hex.EncodeToString(token) + "-" + strconv.FormatUint(id, 10)

	q.mu.Lock()
	q.pruneLocked()
	q.results[jobID] = &JobResult{ID: jobID, AccountID: accountID, Status: "queued"}
	q.mu.Unlock()

	job := SkinJob{ID: jobID, AccountID: accountID, Variant: variant, Data: data, Created: time.Now().UTC()}
	select {
	case q.jobs <- job:
		return jobID, nil
	default:
		q.mu.Lock()
		delete(q.results, jobID)
		q.mu.Unlock()
		return "", errors.New("skin queue is full")
	}
}

// result returns a copy of the current status for a job id.
func (q *skinQueue) result(id string) (*JobResult, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.pruneLocked()
	r, ok := q.results[id]
	if !ok {
		return nil, false
	}
	copyR := *r
	return &copyR, true
}

// pruneLocked drops terminal results past their TTL and, if the map is still
// over the cap, the oldest finished entries. The caller must hold q.mu.
func (q *skinQueue) pruneLocked() {
	cutoff := time.Now().Add(-resultTTL)
	for id, r := range q.results {
		if (r.Status == "done" || r.Status == "failed") && r.DoneAt.Before(cutoff) {
			delete(q.results, id)
		}
	}
	if len(q.results) <= maxJobResults {
		return
	}

	// Still too many: evict the oldest finished entries first.
	type candidate struct {
		id   string
		done time.Time
	}
	var finished []candidate
	for id, r := range q.results {
		if r.Status == "done" || r.Status == "failed" {
			finished = append(finished, candidate{id: id, done: r.DoneAt})
		}
	}
	if len(finished) == 0 {
		return
	}
	oldest := finished[0]
	for _, c := range finished[1:] {
		if c.done.Before(oldest.done) {
			oldest = c
		}
	}
	delete(q.results, oldest.id)
}
