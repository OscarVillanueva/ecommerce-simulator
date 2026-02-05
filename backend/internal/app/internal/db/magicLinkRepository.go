
package db

import (
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	log "github.com/sirupsen/logrus"
)

func FindMagicLinkForUser(token string, email string, magic *dao.Magic) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	userSubQuery := db.Model(&dao.User{}).Select("uuid").Where("email = ?", email)

	err := db.Where("token = ? AND belongs_to = (?)", token, userSubQuery).First(magic).Error
	if err != nil {
		log.Error(err)
		return err
	}

	if magic.Token == "" {
		log.Error("The token not exits", err)
		return errors.New("The token not exists")
	}

	return nil
}

func FetchMagicLink(userUuid string, token string, magic *dao.Magic, ctx context.Context) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	if err := db.WithContext(ctx).Where("token = ? AND belongs_to = ?", token, userUuid).First(magic).Error; err != nil {
		return err
	}

	return nil
}

func DeleteMagicLink(userUuid string, ctx context.Context) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	if err := db.WithContext(ctx).Where("belongs_to = (?)", userUuid).Delete(&dao.Magic{}).Error; err != nil {
		return err
	}

	return nil
}

