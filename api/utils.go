package api

import (
	"net/http"

	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) healthCheck(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	handler.OK(w, true)
}
