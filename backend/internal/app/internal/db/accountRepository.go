package db

import (
	"fmt"
	"time"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var AccountRepositoryName = "account-repository"

func CreateAccount(account requests.CreateAccount, ctx context.Context) (string, error)  {
	tr := otel.Tracer(AccountRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.CreateAccount", AccountRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return "", dbErr
	}

	var token string
	err := db.WithContext(trContext).Transaction(func (tx *gorm.DB) error {
		user := dao.User{
			Uuid: uuid.New().String(),
			Name: account.Name,
			Email: account.Email,
			Verified: false,
			CreatedAt: time.Now().UTC(),
		}

		if err := tx.Create(&user).Error; err != nil {
			span.SetAttributes(
				attribute.String("Uuid", user.Uuid),
				attribute.String("Name", user.Name),
				attribute.String("Email", user.Email),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("%s: %v", "CreateUser", err.Error()))
			return err
		}

		token = tools.GenerateSecureToken(3)
		magic := dao.Magic {
			Token: token,
			ExpirationDate: time.Now().UTC().Add(15 * time.Minute),
			BelongsTo: user.Uuid,
		}

		if err := tx.Create(&magic).Error; err != nil {
			span.SetAttributes(
				attribute.String("Token", magic.Token),
				attribute.String("ExpirationDate", magic.ExpirationDate.String()),
				attribute.String("BelongsTo", user.Uuid),
			)
			span.RecordError(err)
			span.SetStatus(codes.Error, fmt.Sprintf("%s: %v", "CreateMagicLink", err.Error()))
			return err
		}

		return nil
	})

  return token, err
}
