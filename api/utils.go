package api

import (
	"errors"
	"net/http"
	"strconv"
	"unicode"

	"github.com/Basu008/GymBud/server/handler"
	"github.com/gorilla/mux"
)

func (a *API) healthCheck(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, true)
}

func pathID(r *http.Request) string {
	return mux.Vars(r)["id"]
}

var errInvalidPagination = errors.New("invalid pagination params")

func (a *API) paginationParams(r *http.Request) (int, int, error) {
	page := 1
	limit := a.Config.AdditionalConfig.WorkoutPaginationLimit
	if limit <= 0 {
		limit = 20
	}

	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		parsedPage, err := strconv.Atoi(pageParam)
		if err != nil || parsedPage <= 0 {
			return 0, 0, errInvalidPagination
		}
		page = parsedPage
	}

	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		parsedLimit, err := strconv.Atoi(limitParam)
		if err != nil || parsedLimit <= 0 {
			return 0, 0, errInvalidPagination
		}
		if parsedLimit > 100 {
			parsedLimit = 100
		}
		limit = parsedLimit
	}
	return page, limit, nil
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
