package handler

import (
	"encoding/json"
	"net/http"
)

type AppResponse struct {
	Success bool `json:"success"`
	Payload any  `json:"payload"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func OK(w http.ResponseWriter, payload any) {
	Success(w, http.StatusOK, payload)
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, message)
}

func Success(w http.ResponseWriter, statusCode int, payload any) {
	writeJSON(w, statusCode, AppResponse{Success: true, Payload: payload})
}

func Error(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, ErrorResponse{Success: false, Error: message})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
