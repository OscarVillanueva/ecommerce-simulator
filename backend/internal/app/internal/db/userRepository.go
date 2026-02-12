package db

import (
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
)

func FetchUser(email string, user *dao.User, ctx context.Context) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	if err := db.WithContext(ctx).Where("email = ?", email).First(user).Error; err != nil {
		return err
	}

	return nil
}

func VerifyUserAccount(magic dao.Magic, ctx context.Context) error  {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		userError := tx.Model(&dao.User{}).Where("uuid = ?", magic.BelongsTo).Update("verified", true).Error
		if userError != nil {
			return userError
		}

		if magicError := tx.Where("token = ?", magic.Token).Delete(&magic).Error; magicError != nil {
			return magicError
		}

		return nil
	})
	
	return err
}

