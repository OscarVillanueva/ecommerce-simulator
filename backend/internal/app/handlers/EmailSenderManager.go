package handlers

import (
	"net/smtp"
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


