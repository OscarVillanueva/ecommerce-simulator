package handlers

import (
	"fmt"
	"time"
	"errors"
	"net/http"
	"encoding/json"

	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/internal/db"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"gorm.io/gorm"
	"github.com/google/uuid"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	mysql "github.com/go-sql-driver/mysql"
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
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		user := dao.User{
			Uuid: uuid.New().String(),
			Name: account.Name,
			Email: account.Email,
			Verified: false,
			CreatedAt: time.Now().UTC(),
		}

		err = gorm.G[dao.User](db).Create(r.Context(), &user)

		if err != nil {
			var mysqlErr *mysql.MySQLError

			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				log.Error(err)
				tools.UnprocessableContent(w, "The email already exists")
				return
			}

			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		sender, err := platform.GetEmailSenderManager()

		if err != nil {
			log.Error(err)
			tools.ServiceUnavailableErrorHandler(w)
			return
		}

		token := tools.GenerateSecureToken(3)

		magic := dao.Magic {
			Token: token,
			ExpirationDate: time.Now().UTC().Add(15 * time.Minute),
			BelongsTo: user.Uuid,
		}

		err = gorm.G[dao.Magic](db).Create(r.Context(), &magic)

		if err != nil {
			log.Error(err)
			msg := "Account creatd but we couldn't create the verification code, request another account"
			tools.InternalServerErrorHandler(w, &msg)
			return
		}

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
		w.Header().Set("Content-Type", "application/json")

		verify := requests.VerifyAccount{}

		err := json.NewDecoder(r.Body).Decode(&verify)

		if err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		user := dao.User{}
		magic := dao.Magic{}

		err = db.FindToken(verify.Token, verify.Email, &magic, &user)

		if err != nil {
			log.Error("Provided data do not exists")
			tools.UnprocessableContent(w, "The token or email are invalid")
			return
		}

		current := time.Now().UTC()

		if magic.ExpirationDate.Before(current) {
			log.Error("The token do not belong to the user")

			msg := "Expired Token"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}
	
		// Actualizar el usuario -> sync
		db := platform.GetInstance()

		err = db.Model(&user).Where("email = ?", verify.Email).Update("verified", true).Error

		if err != nil {
			log.Error("We couldnt verify the user")
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		_, err = gorm.G[dao.Magic](db).Where("token = ? AND belongs_to = ?", magic.Token, magic.BelongsTo).Delete(r.Context())

		if err != nil {
			log.Error("We couldnt delete the token")
		}

		resp := tools.Message {
			Message: "Verified account",
			Data: "success",
		}

		resp.WriteMessage(w)
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

