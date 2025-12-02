package routines

import (
	"github.com/OscarVillanueva/goapi/internal/app/models/dao"
	"github.com/OscarVillanueva/goapi/internal/platform"

	log "github.com/sirupsen/logrus"
)

func FindUser(email string, user *dao.User, channel chan bool) {
	db := platform.GetInstance()

	if db == nil {
		log.Error("We couldn't connect to the database")
		channel <- false
		return
	}

	db.Where("email = ?", email).First(&user)

	if user.Uuid == "" {
		log.Error("The user not exits")
		channel <- false
		return
	}

	channel <- true
}


