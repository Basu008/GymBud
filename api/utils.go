package api

import (
	"net/http"
	"unicode"

	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) healthCheck(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, true)
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasLower bool
	var hasUpper bool
	var hasDigit bool
	var hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	return hasLower && hasUpper && hasDigit && hasSpecial
}
