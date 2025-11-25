package internal

import (
	"net/http"
	"encoding/json"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func AuthRouter(router chi.Router) {
	router.Get("/login", func(w http.ResponseWriter, r *http.Request){
		
		w.Header().Set("Content-Type", "application/json")
  	err := json.NewEncoder(w).Encode("auth get")

		if err != nil {
			log.Error(err)
		}
	})
}

