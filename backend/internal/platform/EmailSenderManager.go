package platform

import (
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type EmailSenderManager struct {
	Host string
	From string
}

type Actions interface {
	SendEmail(to []string, msg []byte) (bool, error)
}

func (esm EmailSenderManager) SendEmail(to []string, msg []byte) (bool, error) {	
	err := smtp.SendMail(esm.Host, nil, esm.From, to, msg)

	if err != nil {
		return false, err
	}

	return true, nil
}

func GetEmailSenderManager() (*EmailSenderManager, error) {
	err := godotenv.Load()

	if err != nil {
		log.Warning("Couldn't load env", err)
		return nil, err
	}

	host := os.Getenv("STMP_HOST")

	return &EmailSenderManager{
		Host: host,
		From: "no-reply@sender.com",
	}, nil
}
