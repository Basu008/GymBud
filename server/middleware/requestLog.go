package middleware

import (
	"net/http"
	"time"

	"github.com/Basu008/GymBud/server/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

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

	start := time.Now()
	next(w, r)
	duration := time.Since(start)

	l.log.Info().
		Str("request_id", reqID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Dur("duration", duration).
		Msg("request")
}
