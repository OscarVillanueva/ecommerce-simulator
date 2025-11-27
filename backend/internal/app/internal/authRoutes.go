package internal

import (
	"time"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"gorm.io/gorm"
	"github.com/google/uuid"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func AuthRouter(router chi.Router) {
	router.Post("/create-account", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		account := requests.CreateAccount{}

		err := json.NewDecoder(r.Body).Decode(&account)

		if err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		db := platform.GetInstance()

		if db == nil {
			tools.InternalServerErrorHandler(w)
			return
		}

		user := dao.User{
			Uuid: uuid.New().String(),
			Name: account.Name,
			Email: account.Email,
			CreatedAt: time.Now(),
		}

		err = gorm.G[dao.User](db).Create(r.Context(), &user)

		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w)
			return
		}

		err = json.NewEncoder(w).Encode(user)

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

