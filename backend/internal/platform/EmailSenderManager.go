package platform

import (
	"net/smtp"
	"errors"
	"sync"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type EmailSenderManager struct {
	Host string
	From string
}

var (
	emailManager *EmailSenderManager
	emailOnce sync.Once
)

func getEmailSenderManager() *EmailSenderManager {
	emailOnce.Do(func () {
		err := godotenv.Load()

		if err != nil {
			log.Warning("Couldn't load env", err)
			return
		}

		host := os.Getenv("STMP_HOST")

		emailManager = &EmailSenderManager{
			Host: host,
			From: "no-reply@sender.com",
		}
	})

	return emailManager
}

func SendEmail(to []string, msg []byte) error {
	manager := getEmailSenderManager()

	if manager == nil {
		return errors.New("Empty manager")
	}

	err := smtp.SendMail(manager.Host, nil, manager.From, to, msg)
	
	return err
}

