package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
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
var errInvalidDateRange = errors.New("invalid date range params")

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

func (a *API) workoutAnalyticsDateRange(r *http.Request) (*time.Time, *time.Time, error) {
	startParam := strings.TrimSpace(r.URL.Query().Get("start_date"))
	endParam := strings.TrimSpace(r.URL.Query().Get("end_date"))

	if startParam == "" && endParam == "" {
		return nil, nil, nil
	}
	if startParam == "" || endParam == "" {
		return nil, nil, errInvalidDateRange
	}

	startDate, err := time.Parse("2006-01-02", startParam)
	if err != nil {
		return nil, nil, errInvalidDateRange
	}
	endDate, err := time.Parse("2006-01-02", endParam)
	if err != nil {
		return nil, nil, errInvalidDateRange
	}
	if endDate.Before(startDate) {
		return nil, nil, errInvalidDateRange
	}

	startedAtGTE := startDate.UTC()
	startedAtLT := endDate.AddDate(0, 0, 1).UTC()
	return &startedAtGTE, &startedAtLT, nil
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
