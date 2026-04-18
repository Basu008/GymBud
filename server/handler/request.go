package handler

import "net/http"

type TokenAuth interface {
}

type RequestCtx struct {
	RequestID string
	Path      string
	UserClaim interface{}
}

type Request struct {
	HandlerFunc func(*RequestCtx, http.ResponseWriter, *http.Request)
	AuthFunc    TokenAuth
	IsLoggedIn  bool
	IsPremium   bool
	IsAdmin     bool
}

func (rh *Request) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RequestCtx{
		Path: r.URL.Path,
	}
	if id := GetRequestID(r); id != "" {
		ctx.RequestID = id
	}
	rh.HandlerFunc(ctx, w, r)
}
