package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

type logEntry struct {
	RequestID  string          `json:"request_id"`
	Method     string          `json:"method"`
	Endpoint   string          `json:"endpoint"`
	StatusCode int             `json:"status_code"`
	DurationMs int64           `json:"duration_ms"`
	UserID     *string         `json:"user_id"`
	Timestamp  string          `json:"timestamp"`
	Body       json.RawMessage `json:"body,omitempty"`
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		w.Header().Set("X-Request-Id", requestID)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		var rawBody json.RawMessage
		if isMutating(r.Method) {
			body, err := io.ReadAll(r.Body)
			if err == nil {
				rawBody = maskSensitive(body)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		}

		next.ServeHTTP(rw, r)

		entry := logEntry{
			RequestID:  requestID,
			Method:     r.Method,
			Endpoint:   r.URL.Path,
			StatusCode: rw.statusCode,
			DurationMs: time.Since(start).Milliseconds(),
			UserID:     nil,
			Timestamp:  start.UTC().Format(time.RFC3339),
			Body:       rawBody,
		}

		data, _ := json.Marshal(entry)
		slog.Info(string(data))
	})
}

func isMutating(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}

func maskSensitive(body []byte) json.RawMessage {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return body
	}

	sensitiveFields := []string{"password", "secret", "token"}
	for _, field := range sensitiveFields {
		if _, ok := m[field]; ok {
			m[field] = json.RawMessage(`"***"`)
		}
	}

	masked, err := json.Marshal(m)
	if err != nil {
		return body
	}
	return masked
}
