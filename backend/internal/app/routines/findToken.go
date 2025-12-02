package routines

import (
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	log "github.com/sirupsen/logrus"
)

func FindToken(token string, magic *dao.Magic, channel chan bool) {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		channel <- false
		return
	}

	db.Where("token = ?", token).First(magic)

	if magic.Token == "" {
		log.Error("The token not exits")
		channel <- false
		return
	}

	channel <- true
}
