package handlers

import (
	"net/http"
	"encoding/json"
	"time"
	"fmt"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Router(r *chi.Mux) {

	r.Use(chimiddle.StripSlashes)

	r.Route("/ping", func(router chi.Router){
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			response := struct {
				Message string `json:"message"`
			}{
				Message: fmt.Sprintf("Current time: %v", time.Now()),
			}

			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(response)

			if err != nil {
				log.Error(err)
			}
		})
	})

}
