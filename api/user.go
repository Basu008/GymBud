package api

import (
	"errors"
	"net/http"

	appsocial "github.com/Basu008/GymBud/app/social"
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

func (a *API) login(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.LoginUserBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.UserService.LoginUser(r.Context(), &body)
	if err != nil {
		if errors.Is(err, appuser.ErrInvalidCredentials) {
			handler.Unauthorized(w, err.Error())
			return
		}
		if errors.Is(err, appuser.ErrUserInactive) {
			handler.Unauthorized(w, err.Error())
			return
		}
		handler.InternalServerError(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) logout(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	if err := a.App.UserService.Logout(r.Context(), ctx.UserClaim.SessionID); err != nil {
		handler.InternalServerError(w, err.Error())
		return
	}

	handler.OK(w, true)
}

func (a *API) updatePrivacy(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.UpdatePrivacyBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	response, err := a.App.UserService.UpdatePrivacy(r.Context(), ctx.UserClaim.UserID, body.IsPrivate)
	if err != nil {
		if errors.Is(err, appuser.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) followUser(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	userID := pathID(r)
	response, err := a.App.SocialService.FollowUser(r.Context(), ctx.UserClaim.UserID, userID)
	if err != nil {
		if errors.Is(err, appsocial.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) unfollowUser(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	userID := pathID(r)
	response, err := a.App.SocialService.UnfollowUser(r.Context(), ctx.UserClaim.UserID, userID)
	if err != nil {
		if errors.Is(err, appsocial.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) acceptFollowRequest(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	userID := pathID(r)
	response, err := a.App.SocialService.AcceptFollowRequest(r.Context(), ctx.UserClaim.UserID, userID)
	if err != nil {
		switch {
		case errors.Is(err, appsocial.ErrUserNotFound), errors.Is(err, appsocial.ErrFollowRequestNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) rejectFollowRequest(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	userID := pathID(r)
	response, err := a.App.SocialService.RejectFollowRequest(r.Context(), ctx.UserClaim.UserID, userID)
	if err != nil {
		switch {
		case errors.Is(err, appsocial.ErrUserNotFound), errors.Is(err, appsocial.ErrFollowRequestNotFound):
			handler.NotFound(w, err.Error())
		default:
			handler.BadRequest(w, err.Error())
		}
		return
	}

	handler.OK(w, response)
}

func (a *API) updateActive(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.UpdateActiveBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	response, err := a.App.UserService.UpdateActive(r.Context(), ctx.UserClaim.UserID, body.IsActive)
	if err != nil {
		if errors.Is(err, appuser.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	if !body.IsActive {
		if err := a.App.UserService.Logout(r.Context(), ctx.UserClaim.SessionID); err != nil {
			handler.InternalServerError(w, err.Error())
			return
		}
	}

	handler.OK(w, response)
}

func (a *API) getCurrentUser(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	response, err := a.App.UserService.GetUserByID(r.Context(), ctx.UserClaim.UserID)
	if err != nil {
		if errors.Is(err, appuser.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.InternalServerError(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) getUserByID(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	userID := pathID(r)
	response, err := a.App.UserService.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, appuser.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) updateUser(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.UpdateUserBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	response, err := a.App.UserService.UpdateUser(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		if errors.Is(err, appuser.ErrUserNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}

func (a *API) createBodyMetrics(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	var body schema.CreateBodyMetricsBody
	if err := handler.BindJSON(r, &body); err != nil {
		handler.BadRequest(w, err.Error())
		return
	}
	if !a.validateBody(w, &body) {
		return
	}

	response, err := a.App.UserService.CreateBodyMetrics(r.Context(), ctx.UserClaim.UserID, &body)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.Created(w, response)
}

func (a *API) deleteBodyMetrics(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	metricsID := pathID(r)
	response, err := a.App.UserService.DeleteBodyMetrics(r.Context(), ctx.UserClaim.UserID, metricsID)
	if err != nil {
		if errors.Is(err, appuser.ErrBodyMetricsNotFound) {
			handler.NotFound(w, err.Error())
			return
		}
		handler.BadRequest(w, err.Error())
		return
	}

	handler.OK(w, response)
}
