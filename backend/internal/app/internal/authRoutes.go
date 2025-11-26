package internal

import (
	"fmt"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/app/models"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func AuthRouter(router chi.Router) {
	router.Post("/create-account", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		account := models.CreateAccount{}

		err := json.NewDecoder(r.Body).Decode(&account)

		if err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
		}

		err = json.NewEncoder(w).Encode(fmt.Sprintf("account: %v", account.Email))

		if err != nil {
			log.Error(err)
		}
	})

	router.Get("/login", func(w http.ResponseWriter, r *http.Request){
		
		w.Header().Set("Content-Type", "application/json")
  	err := json.NewEncoder(w).Encode("auth get")

		if err != nil {
			log.Error(err)
		}
	})
}

