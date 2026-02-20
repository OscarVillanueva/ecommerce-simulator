
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
)

var RepositoryName = "magic-link-repository"

func FindMagicLinkForUser(token string, email string, magic *dao.Magic, ctx context.Context) error {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.FindMagicLinkForUser", RepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	userSubQuery := db.Model(&dao.User{}).Select("uuid").Where("email = ?", email)

	err := db.WithContext(trContext).Where("token = ? AND belongs_to = (?)", token, userSubQuery).First(magic).Error
	if err != nil {
		span.SetAttributes(
			attribute.String("Token", token),
			attribute.String("Email", email),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	if magic.Token == "" {
		mgErr := errors.New("The token not exits")
		span.RecordError(mgErr)
		span.SetStatus(codes.Error, mgErr.Error())
		return mgErr
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.FindMagicLinkForUser successfully", RepositoryName))

	return nil
}

func FetchMagicLink(userUuid string, token string, magic *dao.Magic, ctx context.Context) error {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.FetchMagicLink", RepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	if err := db.WithContext(trContext).Where("token = ? AND belongs_to = ?", token, userUuid).First(magic).Error; err != nil {
		span.SetAttributes(
			attribute.String("Token", token),
			attribute.String("UserUuid", userUuid),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.FetchMagicLink successfully", RepositoryName))

	return nil
}

func DeleteMagicLink(userUuid string, ctx context.Context) error {
	tr := otel.Tracer(RepositoryName)
	trContext, span := tr.Start(ctx, fmt.Sprintf("%s.DeleteMagicLink", RepositoryName))
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	if err := db.WithContext(trContext).Where("belongs_to = (?)", userUuid).Delete(&dao.Magic{}).Error; err != nil {
		span.SetAttributes(
			attribute.String("BelongsTo", userUuid),
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("%s.DeleteMagicLink successfully", RepositoryName))

	return nil
}

