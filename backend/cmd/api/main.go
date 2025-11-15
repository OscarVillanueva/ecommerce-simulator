package main

import (
	"log"
	"github.com/OscarVillanueva/goapi/internal/app/handlers"
)

func main()  {
	sender := handlers.EmailSenderManager {
		Host: "mailhog:1025",
		From: "no-reply@sender.com",
	}

	to := []string{"account@sender.com"}
	msg := []byte("Token: 1234")

	_, err := sender.SendEmail(to, msg)

	if err != nil {
		log.Fatal(err)
	}
}


