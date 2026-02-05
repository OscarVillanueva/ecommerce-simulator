package db

import (
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

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
