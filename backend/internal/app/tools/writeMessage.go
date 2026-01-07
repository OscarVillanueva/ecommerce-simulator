package tools

import (
	"net/http"
	"encoding/json"
)

type Message struct {
	Message string `json:"message"`
	Data any `json:"data"`
}

type Interactions interface {
	WriteMessage(w http.ResponseWriter)
}

func (m Message) WriteMessage(w http.ResponseWriter){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(m)
}
