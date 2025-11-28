package internal

import (
	"fmt"
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
			Verified: false,
			CreatedAt: time.Now(),
		}

		err = gorm.G[dao.User](db).Create(r.Context(), &user)

		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w)
			return
		}

		sender, err := platform.GetEmailSenderManager()

		if err != nil {
			log.Error(err)
			tools.ServiceUnavailableErrorHandler(w)
			return
		}

		token := tools.GenerateSecureToken(3)

		to := []string{account.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", token))

		_, err = sender.SendEmail(to, msg)

		if err != nil {
			log.Warning("Couldn't send the email: ", err)
			tools.ServiceUnavailableErrorHandler(w)
			return
		}

		resp := tools.Message {
			Message: "Token sended to the given email, please verify the account",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Post("/verify-account", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Chance the verify flag in database
	})

	router.Post("/resend-code", func(w http.ResponseWriter, r *http.Request){
		// TODO: Create a new token to resend to the requested email
	})

	router.Get("/login", func(w http.ResponseWriter, r *http.Request){
		
		w.Header().Set("Content-Type", "application/json")
  	err := json.NewEncoder(w).Encode("auth get")

		if err != nil {
			log.Error(err)
		}
	})
}

