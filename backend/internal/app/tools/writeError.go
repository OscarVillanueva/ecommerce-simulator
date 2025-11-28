package tools

import (
	"net/http"
	"encoding/json"
)

type Error struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, message string, code int){
	resp := Error {
		Code: code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	BadRequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	UnprocessableContent = func(w http.ResponseWriter, message string) {
		writeError(w, message, http.StatusUnprocessableEntity)
	}
	InternalServerErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An unexpected expected error occured", http.StatusInternalServerError)
	}
	ServiceUnavailableErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "The server is not ready to handle the request", http.StatusServiceUnavailable)
	}
)
