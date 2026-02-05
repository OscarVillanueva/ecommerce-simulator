package db

import (
	"time"
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/requests"
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/app/tools"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"gorm.io/gorm"
	"github.com/google/uuid"
)

func CreateAccount(account requests.CreateAccount, ctx context.Context) (string, error)  {
	db := platform.GetInstance()

	if db == nil {
		return "", errors.New("We couldn't connect to the database")
	}

	var token string
	err := db.WithContext(ctx).Transaction(func (tx *gorm.DB) error {
		user := dao.User{
			Uuid: uuid.New().String(),
			Name: account.Name,
			Email: account.Email,
			Verified: false,
			CreatedAt: time.Now().UTC(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		token = tools.GenerateSecureToken(3)
		magic := dao.Magic {
			Token: token,
			ExpirationDate: time.Now().UTC().Add(15 * time.Minute),
			BelongsTo: user.Uuid,
		}

		if err := tx.Create(&magic).Error; err != nil {
			return err
		}

		return nil
	})

  return token, err
}
