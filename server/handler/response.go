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

type MultiErrorResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error"`
}

func OK(w http.ResponseWriter, payload any) {
	Success(w, http.StatusOK, payload)
}

func Created(w http.ResponseWriter, payload any) {
	Success(w, http.StatusCreated, payload)
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

func BadRequestMulti(w http.ResponseWriter, errs []error) {
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	Errors(w, http.StatusBadRequest, messages)
}

func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message)
}

func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, message)
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

func Errors(w http.ResponseWriter, statusCode int, messages []string) {
	writeJSON(w, statusCode, MultiErrorResponse{Success: false, Errors: messages})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
