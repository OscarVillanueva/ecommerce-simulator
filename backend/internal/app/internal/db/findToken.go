package db

import (
	"errors"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	log "github.com/sirupsen/logrus"
)

func FindToken(token string, email string, magic *dao.Magic, user *dao.User) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	userSubQuery := db.Model(user).Select("uuid").Where("email = ?", email)

	err := db.Where("token = ? AND belongs_to = (?)", token, userSubQuery).First(magic).Error

	if err != nil {
		log.Warning("query failed")
		log.Error(err)
		return errors.New("query failed")
	}

	if magic.Token == "" {
		log.Error("The token not exits", err)
		return errors.New("The token not exists")
	}

	return nil
}
