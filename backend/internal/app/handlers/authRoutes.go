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

		var verify requests.VerifyAccount
		if err := json.NewDecoder(r.Body).Decode(&verify); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		var magic dao.Magic
		if err := db.FindToken(verify.Token, verify.Email, &magic, &dao.User{}); err != nil {
			tools.UnprocessableContent(w, "The token or email are invalid")
			return
		}

		if magic.ExpirationDate.Before(time.Now().UTC()) {
			msg := "Expired Token"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}
	
		sharedDB := platform.GetInstance()

		err := sharedDB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {

			userError := tx.Model(&dao.User{}).Where("uuid = ?", magic.BelongsTo).Update("verified", true).Error
			if userError != nil {
				return userError
			}

			if magicError := tx.Where("token = ?", magic.Token).Delete(&magic).Error; magicError != nil {
				return magicError
			}

			return nil
		})

		if err != nil {
			log.Error("We couldnt delete the token")
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "Verified account",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Post("/resend-code", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		var resend requests.ResendCode
		if err := json.NewDecoder(r.Body).Decode(&resend); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		sender, err := platform.GetEmailSenderManager()

		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		var user dao.User
		if err := db.FetchUser(resend.Email, &user, r.Context()); err != nil {
			resp := tools.Message {
				Message: "If an account exists with this email, a code has been sent",
				Data: "success",
			}

			resp.WriteMessage(w)
			return
		}

		var magic dao.Magic
		if err := db.RegenerateMagicLink(r.Context(), user.Uuid, &magic); err != nil {

		}

		to := []string{resend.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", magic.Token))

		if ok, err := sender.SendEmail(to, msg); !ok {
			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		resp := tools.Message {
			Message: "If an account exists with this email, a code has been sent",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Get("/login", func(w http.ResponseWriter, r *http.Request){
		
		w.Header().Set("Content-Type", "application/json")
  	err := json.NewEncoder(w).Encode("auth get")

		if err != nil {
			log.Error(err)
		}
	})
}

