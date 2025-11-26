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
)
