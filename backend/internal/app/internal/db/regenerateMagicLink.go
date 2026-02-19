
package db

import (
	"errors"
	"context"
	"time"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

func RegenerateMagicLink(ctx context.Context, uuid string, magic *dao.Magic) error  {
	tr := otel.Tracer(PurchaseRepositoryName)
	trContext, span := tr.Start(ctx, "RegenerateMagicLink")
	defer span.End()

	db := platform.GetInstance()

	if db == nil {
		dbErr := errors.New("We couldn't connect to the database")
		span.RecordError(dbErr)
		span.SetStatus(codes.Error, dbErr.Error())
		return dbErr
	}

	err := db.WithContext(trContext).Transaction(func (tx *gorm.DB) error  {
		if userError := tx.Where("belongs_to = (?)", uuid).Delete(&dao.Magic{}).Error; userError != nil {
			span.SetAttributes(
				attribute.String("MagicUuid", uuid),
			)
			span.RecordError(userError)
			span.SetStatus(codes.Error, userError.Error())
			return userError
		}

		token := tools.GenerateSecureToken(3)

		*magic = dao.Magic {
			Token: token,
			ExpirationDate: time.Now().UTC().Add(15 * time.Minute),
			BelongsTo: uuid,
		}

		if magicError := tx.Create(&magic).Error; magicError != nil {
			span.RecordError(magicError)
			span.SetStatus(codes.Error, magicError.Error())
			return magicError
		}

		return nil
	})

	return err
}
