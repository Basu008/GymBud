package middleware

import (
	"net/http"
	"time"

	"github.com/Basu008/GymBud/server/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
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
	}
	event.
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int("status", rw.statusCode).
		Dur("duration", duration).
		Msg("request")
}
