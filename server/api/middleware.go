package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// requestIDCtxKey is the context key under which the request id is stored.
type ctxKey string

const requestIDCtxKey = ctxKey("request-id")

// withRequestID ensures every request carries an id. A client-provided
// X-Request-Id is reused; otherwise a random one is generated. The id is
// echoed in the response header and stored in the request context.
func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			buf := make([]byte, 16)
			if _, err := rand.Read(buf); err != nil {
				id = "unknown"
			} else {
				id = hex.EncodeToString(buf)
			}
		}
		w.Header().Set("X-Request-Id", id)
		ctx := context.WithValue(r.Context(), requestIDCtxKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestID returns the request id stored by withRequestID, or "".
func requestID(r *http.Request) string {
	if id, ok := r.Context().Value(requestIDCtxKey).(string); ok {
		return id
	}
	return ""
}

// statusRecorder captures the response status code for request logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withRequestLog logs method, path, status and latency for every request.
func (s *Server) withRequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Printf("http method=%s path=%s status=%d latency_ms=%d request_id=%s", r.Method, r.URL.Path, rec.status, time.Since(start).Milliseconds(), requestID(r))
	})
}

// withRecovery converts panics into 500 responses and logs the stack.
func (s *Server) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				writeError(w, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// withSecurityHeaders sets conservative security headers on every response.
func (s *Server) withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
