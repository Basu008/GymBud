package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Basu008/GymBud/server/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode >= 400 {
		rw.body.Write(b)
	}
	return rw.ResponseWriter.Write(b)
}

// extractErrors parses the error field from error response bodies.
// Handles both single string ("error": "msg") and array ("error": ["e1","e2"]).
func extractErrors(body []byte) []string {
	var raw struct {
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(body, &raw); err != nil || raw.Error == nil {
		return nil
	}
	var single string
	if err := json.Unmarshal(raw.Error, &single); err == nil {
		return []string{single}
	}
	var multi []string
	if err := json.Unmarshal(raw.Error, &multi); err == nil {
		return multi
	}
	return nil
}

type RequestLogger struct {
	log *zerolog.Logger
}

func NewRequestLoggerWithLogger(log *zerolog.Logger) *RequestLogger {
	return &RequestLogger{log: log}
}

func (l *RequestLogger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = uuid.New().String()
	}
	w.Header().Set("X-Request-ID", reqID)
	r = handler.SetContextValue(r, handler.KeyRequestID, reqID)

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	start := time.Now()
	next(rw, r)
	duration := time.Since(start)

	event := l.log.Info()
	if rw.statusCode >= 400 {
		event = l.log.Error()
		if errs := extractErrors(rw.body.Bytes()); len(errs) > 0 {
			event = event.Strs("errors", errs)
		}
	}
	event.
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", rw.statusCode).
		Dur("duration", duration).
		Msg("request")
}
