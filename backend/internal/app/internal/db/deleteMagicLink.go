package db

import (
	"errors"
	"context"

	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	log "github.com/sirupsen/logrus"
)

func DeleteMagicLink(userUuid string, ctx context.Context) error {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		return errors.New("We couldn't connect to the database")
	}

	if err := db.Where("belongs_to = (?)", userUuid).Delete(&dao.Magic{}).Error; err != nil {
		return err
	}

	return nil
}
