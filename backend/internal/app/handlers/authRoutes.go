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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/go-chi/chi/v5"
	mysql "github.com/go-sql-driver/mysql"
)

const AuthRouterName = "auth-router"

func AuthRouter(router chi.Router) {
	router.Post("/create-account", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s/create-account", AuthRouterName))
		defer span.End()

		account := requests.CreateAccount{}
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := account.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		token, err := db.CreateAccount(account, ctx)
		if err != nil  {
			var mysqlErr *mysql.MySQLError

			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				span.SetAttributes(
					attribute.String("Email", account.Email),
					attribute.String("Name", account.Name),
					attribute.String("Message", "The email already exists"),
				)
				tools.UnprocessableContent(w, "The email already exists")
				return
			}

			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		to := []string{account.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", token))
		if err := platform.SendEmail(to, msg); err != nil {
			span.SetAttributes(
				attribute.String("Email", account.Email),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.ServiceUnavailableErrorHandler(w)
			return
		}

		span.SetStatus(codes.Ok, "Created Account")

		resp := tools.Message {
			Message: "Token sended to the given email, please verify the account",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Post("/verify-account", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s/verify-account", AuthRouterName))
		defer span.End()

		var verify requests.VerifyAccount
		if err := json.NewDecoder(r.Body).Decode(&verify); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := verify.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		var magic dao.Magic
		if err := db.FindMagicLinkForUser(verify.Token, verify.Email, &magic, ctx); err != nil {
			span.SetAttributes(
				attribute.String("Token", verify.Token),
				attribute.String("Email", verify.Email),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnprocessableContent(w, "The token or email are invalid")
			return
		}

		if magic.ExpirationDate.Before(time.Now().UTC()) {
			msg := "Expired Token"
			span.SetAttributes(
				attribute.String("Token", verify.Token),
			)
			span.RecordError(errors.New(msg))
			span.SetStatus(codes.Error, msg)
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}
	
		if err := db.VerifyUserAccount(magic, ctx); err != nil {
			span.SetAttributes(
				attribute.String("Token", magic.Token),
				attribute.String("BelongsTo", magic.BelongsTo),
				attribute.String("ExpirationDate", magic.ExpirationDate.String()),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		span.SetStatus(codes.Ok, "Verified Account")

		resp := tools.Message {
			Message: "Verified account",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Post("/resend-code", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s/resend-code", AuthRouterName))
		defer span.End()

		var resend requests.ResendCode
		if err := json.NewDecoder(r.Body).Decode(&resend); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := resend.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		var user dao.User
		if err := db.FetchUser(resend.Email, &user, ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			resp := tools.Message {
				Message: "If an account exists with this email, a code has been sent",
				Data: "success",
			}

			resp.WriteMessage(w)
			return
		}

		var magic dao.Magic
		if err := db.RegenerateMagicLink(ctx, user.Uuid, &magic); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		to := []string{resend.Email}
		msg := []byte(fmt.Sprintf("Use this token to verify your account: %s", magic.Token))
		if err := platform.SendEmail(to, msg); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}


		span.SetStatus(codes.Ok, "ResendCode success")

		resp := tools.Message {
			Message: "If an account exists with this email, a code has been sent",
			Data: "success",
		}

		resp.WriteMessage(w)
	})

	router.Post("/login", func(w http.ResponseWriter, r *http.Request){		
		w.Header().Set("Content-Type", "application/json")

		tr := otel.Tracer(AuthRouterName)
		ctx, span := tr.Start(r.Context(), fmt.Sprintf("%s/login", AuthRouterName))
		defer span.End()

		var login requests.Login
		if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, errors.New("Invalid body request"))
			return
		}

		if err := login.Validate(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.BadRequestErrorHandler(w, err)
			return
		}

		var user dao.User
		if err := db.FetchUser(login.Email, &user, ctx); err != nil {
			msg := "The email or token are invalid"
			span.SetAttributes(
				attribute.String("Email", login.Email),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		if !user.Verified {
			msg := "The email or token are invalid"
			span.SetAttributes(
				attribute.Bool("Verified", user.Verified),
				attribute.String("UserUuid", user.Uuid),
			)
			span.RecordError(errors.New(msg))
			span.SetStatus(codes.Error, msg)
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		var magic dao.Magic
		if err := db.FetchMagicLink(user.Uuid, login.Token, &magic, ctx); err != nil {
			msg := "The email or token are invalid"
			span.SetAttributes(
				attribute.String("Token", login.Token),
				attribute.String("UserUuid", user.Uuid),
			)
			span.RecordError(errors.New(msg))
			span.SetStatus(codes.Error, msg)
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		if magic.ExpirationDate.Before(time.Now().UTC()) {
			msg := "The email or token are invalid"
			span.SetAttributes(
				attribute.String("ExpirationDate", magic.ExpirationDate.String()),
			)
			span.RecordError(errors.New(msg))
			span.SetStatus(codes.Error, msg)
			tools.UnauthorizedErrorHandler(w, &msg)
			return
		}

		token, err := tools.GenerateJWT(
			time.Now().UTC().Add(120 * time.Hour), 
			map[string]any{ "uuid": user.Uuid, "name": user.Name })

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			tools.InternalServerErrorHandler(w, nil)
			return
		}

		if err := db.DeleteMagicLink(user.Uuid, ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}

		span.SetStatus(codes.Ok, "Login success")

		resp := tools.Message {
			Message: token,
			Data: "success",
		}

		resp.WriteMessage(w)
	})
}

