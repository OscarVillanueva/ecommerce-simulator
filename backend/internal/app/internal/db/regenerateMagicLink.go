
package db

import (
	"errors"
	"context"
	"time"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
)

func RegenerateMagicLink(ctx context.Context, uuid string, magic *dao.Magic) error  {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	err := db.WithContext(ctx).Transaction(func (tx *gorm.DB) error  {
		if userError := tx.Where("belongs_to = (?)", uuid).Delete(&dao.Magic{}).Error; userError != nil {
			return userError
		}

		token := tools.GenerateSecureToken(3)

		magic = &dao.Magic {
			Token: token,
			ExpirationDate: time.Now().UTC().Add(15 * time.Minute),
			BelongsTo: uuid,
		}

		if magicError := tx.Create(&magic).Error; magicError != nil {
			return magicError
		}

		return nil
	})

	return err
}
