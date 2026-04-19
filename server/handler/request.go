package handler

import (
	"net/http"
	"strings"

	"github.com/Basu008/GymBud/server/auth"
)

type TokenAuth interface {
	AuthenticateRequest(token string) (*auth.AuthUser, error)
}

type RequestCtx struct {
	RequestID string
	Path      string
	UserClaim *auth.AuthUser
}

type Request struct {
	HandlerFunc func(*RequestCtx, http.ResponseWriter, *http.Request)
	AuthFunc    TokenAuth
	IsLoggedIn  bool
}

func (rh *Request) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &RequestCtx{
		Path: r.URL.Path,
	}

	if id := GetRequestID(r); id != "" {
		ctx.RequestID = id
	}

	if rh.IsLoggedIn {
		if rh.AuthFunc == nil {
			http.Error(w, "auth not configured", http.StatusInternalServerError)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		user, err := rh.AuthFunc.AuthenticateRequest(parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx.UserClaim = user
	}

	rh.HandlerFunc(ctx, w, r)
}
