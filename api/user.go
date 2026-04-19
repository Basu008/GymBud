package api

import (
	"net/http"

	"github.com/Basu008/GymBud/schema"
	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) signUp(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var s schema.SignUpUserBody
	if err := handler.BindJSON(r, &s); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &s) {
		return
	}
	if err := a.App.UserService.SignUpUser(r.Context(), &s); err != nil {
		handler.InternalServerError(w, err.Error())
		return
	}
	handler.Created(w, true)
}
