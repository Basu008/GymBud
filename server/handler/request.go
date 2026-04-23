package handler

import (
	"errors"
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
			Unauthorized(w, auth.ErrLoginRequired.Error())
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			Unauthorized(w, auth.ErrInvalidAccessToken.Error())
			return
		}

		user, err := rh.AuthFunc.AuthenticateRequest(parts[1])
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrSessionExpired):
				Unauthorized(w, auth.ErrSessionExpired.Error())
			case errors.Is(err, auth.ErrLoginRequired):
				Unauthorized(w, auth.ErrLoginRequired.Error())
			default:
				Unauthorized(w, auth.ErrInvalidAccessToken.Error())
			}
			return
		}

		ctx.UserClaim = user
	}

	rh.HandlerFunc(ctx, w, r)
}
