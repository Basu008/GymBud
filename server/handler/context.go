package handler

import (
	"context"
	"net/http"
)

type contextKey string

const (
	KeyUserClaim contextKey = "user_claim"
	KeyRequestID contextKey = "request_id"
)

func GetRequestID(r *http.Request) string {
	v, _ := r.Context().Value(KeyRequestID).(string)
	return v
}

func SetContextValue(r *http.Request, key contextKey, val interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}
