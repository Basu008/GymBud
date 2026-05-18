package api

import (
	"net/http"

	"github.com/Basu008/GymBud/server/handler"
)

type appVersionResponse struct {
	MinimumSupportedVersion string `json:"minimum_supported_version"`
	LatestVersion           string `json:"latest_version"`
	UpdateURL               string `json:"update_url"`
	Title                   string `json:"title"`
	Message                 string `json:"message"`
}

func (a *API) getAppVersion(ctx *handler.RequestCtx, w http.ResponseWriter, r *http.Request) {
	cfg := a.Config.AdditionalConfig
	handler.OK(w, appVersionResponse{
		MinimumSupportedVersion: cfg.MinSupportedVersion,
		LatestVersion:           cfg.LatestVersion,
		UpdateURL:               cfg.UpdateURL,
		Title:                   cfg.Title,
		Message:                 cfg.Message,
	})
}
