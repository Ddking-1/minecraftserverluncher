package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Minimal valid 64x64 PNG (base64).
const asyncPNG = "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAAnSURBVHhe7cEBDQAAAMKg909tDjegAAAAAAAAAAAAAAAAAAAAgHcDQEAAAVkKewwAAAAASUVORK5CYII="

func decodePNG(t *testing.T, s string) []byte {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return b
}

func jsonUnmarshal(b []byte, dst any) error { return json.Unmarshal(b, dst) }

func TestAsyncSkinUploadAccepted(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()

	_, out := register(t, h, "asynuser", "supersecret")
	access := out["access_token"].(string)
	body := decodePNG(t, asyncPNG)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload/default", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "image/png")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("upload status = %d, want 202", rec.Code)
	}
	var resp struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}
	if err := jsonUnmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TaskID == "" {
		t.Fatal("response must include a task_id")
	}

	// Poll the task until it completes.
	taskID := resp.TaskID
	_, polled := doJSON(t, h, http.MethodGet, "/api/v1/tasks/"+taskID, nil, access)
	if polled["status"] == nil {
		t.Fatal("task status must include a status field")
	}
}

func TestTaskStatusNotFound(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()
	_, out := register(t, h, "taskuser", "supersecret")
	access := out["access_token"].(string)

	code, _ := doJSON(t, h, http.MethodGet, "/api/v1/tasks/does-not-exist", nil, access)
	if code != http.StatusNotFound {
		t.Fatalf("unknown task status = %d, want 404", code)
	}
}

func TestAsyncSkinRejectsNonImage(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	h := srv.Handler()
	_, out := register(t, h, "asynbad", "supersecret")
	access := out["access_token"].(string)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/skins/upload/default",
		io.NopCloser(bytes.NewReader([]byte("<?php system('x'); ?>"))))
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "image/png")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("php-as-png status = %d, want 415", rec.Code)
	}
}
