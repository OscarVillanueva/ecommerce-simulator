package handlers

import (
	"net/http"
	"encoding/json"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	chimiddle "github.com/go-chi/chi/middleware"
)

func Router(r *chi.Mux) {

	r.Use(chimiddle.StripSlashes)

	r.Route("/ping", func(router chi.Router){
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			response := struct {
				message string
			}{
				message: "Its alive . . .",
			}

			err := json.NewEncoder(w).Encode(response)

			if err != nil {
				log.Error(err)
			}
		})
	})

}
