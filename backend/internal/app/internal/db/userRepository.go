package db

import (
	"fmt"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

var UserRepositoryName = "user-repository"

func FetchUser(email string, user *dao.User, ctx context.Context) error {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.FetchUser", UserRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	if err := db.WithContext(trContext).Where("email = ?", email).First(user).Error; err != nil {
		span.SetAttributes(
			attribute.String("Email", email),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func VerifyUserAccount(magic dao.Magic, ctx context.Context) error  {
	tr := otel.Tracer(ProductRepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.VerifyUserAccount", UserRepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	err := db.WithContext(trContext).Transaction(func(tx *gorm.DB) error {

		userError := tx.Model(&dao.User{}).Where("uuid = ?", magic.BelongsTo).Update("verified", true).Error
		if userError != nil {
			span.SetAttributes(
				attribute.String("BelongsTo", magic.BelongsTo),
			)
			span.RecordError(userError)
			span.SetStatus(codes.Error, userError.Error())
			return userError
		}

		if magicError := tx.Where("token = ?", magic.Token).Delete(&magic).Error; magicError != nil {
			span.SetAttributes(
				attribute.String("Token", magic.Token),
			)
			span.RecordError(magicError)
			span.SetStatus(codes.Error, magicError.Error())
			return magicError
		}

		return nil
	})
	
	return err
}

