package handler

import (
	"encoding/json"
	"net/http"
)

type AppResponse struct {
	Success bool `json:"success"`
	Payload any  `json:"payload"`
}

func OK(w http.ResponseWriter, payload any) {
	Success(w, http.StatusOK, payload)
}

func Success(w http.ResponseWriter, statusCode int, payload any) {
	writeJSON(w, statusCode, AppResponse{Success: true, Payload: payload})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
