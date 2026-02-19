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

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
	mysql "github.com/go-sql-driver/mysql"
)

func AuthRouter(router chi.Router) {
	router.Post("/create-account", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		account := requests.CreateAccount{}
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		token, err := db.CreateAccount(account, r.Context())
		if err != nil  {
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

		to := []string{account.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", token))
		if err := platform.SendEmail(to, msg); err != nil {
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
		if err := db.FindMagicLinkForUser(verify.Token, verify.Email, &magic); err != nil {
			tools.UnprocessableContent(w, "The token or email are invalid")
			return
		}

		if magic.ExpirationDate.Before(time.Now().UTC()) {
			msg := "Expired Token"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}
	
		if err := db.VerifyUserAccount(magic, r.Context()); err != nil {
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
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		to := []string{resend.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", magic.Token))
		if err := platform.SendEmail(to, msg); err != nil {
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

	router.Post("/login", func(w http.ResponseWriter, r *http.Request){		
		w.Header().Set("Content-Type", "application/json")

		var login requests.Login
		if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
			log.Error(err)
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		var user dao.User
		if err := db.FetchUser(login.Email, &user, r.Context()); err != nil {
			msg := "The email or token are invalid"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		if !user.Verified {
			msg := "The email or token are invalid"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		var magic dao.Magic
		if err := db.FetchMagicLink(user.Uuid, login.Token, &magic, r.Context()); err != nil {
			msg := "The email or token are invalid"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		if magic.ExpirationDate.Before(time.Now().UTC()) {
			msg := "The email or token are invalid"
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		token, err := tools.GenerateJWT(
			time.Now().UTC().Add(120 * time.Hour), 
			map[string]any{ "uuid": user.Uuid, "name": user.Name })

		if err != nil {
			log.Error(err)
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if err := db.DeleteMagicLink(user.Uuid, r.Context()); err != nil {
			log.Error(err)
		}

		resp := tools.Message {
			Message: token,
			Data: "success",
		}

		resp.WriteMessage(w)
	})
}

