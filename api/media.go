package api

import (
	"net/http"

	appmedia "github.com/Basu008/GymBud/app/media"
	"github.com/Basu008/GymBud/server/handler"
)

func (a *API) uploadImage(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(appmedia.MaxMultipartImageMemory()); err != nil {
		handler.BadRequest(w, "invalid multipart form")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		handler.BadRequest(w, "image is required")
		return
	}
	defer file.Close()

	input := &appmedia.UploadImageInput{
		OwnerID:    ctx.UserClaim.UserID,
		EntityType: r.FormValue("entity_type"),
		File:       file,
	}

	response, err := a.App.MediaService.UploadImage(r.Context(), input)
	if err != nil {
		handler.BadRequest(w, err.Error())
		return
	}

	handler.Created(w, response)
}
