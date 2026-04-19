package api

import (
	"errors"
	"net/http"

	appuser "github.com/Basu008/GymBud/app/user"
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
	if !isStrongPassword(s.Password) {
		handler.BadRequest(w, "password must be at least 8 characters and include one uppercase letter, one lowercase letter, one number, and one special character")
		return
	}
	if err := a.App.UserService.SignUpUser(r.Context(), &s); err != nil {
		if errors.Is(err, appuser.ErrUsernameAlreadyExists) {
			handler.Conflict(w, err.Error())
			return
		}
		handler.InternalServerError(w, err.Error())
		return
	}
	handler.Created(w, true)
}
