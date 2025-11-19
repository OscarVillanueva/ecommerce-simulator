package main

import (
	"log"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/OscarVillanueva/goapi/internal/platform"
)

func main()  {
	err := godotenv.Load()

	if err != nil {
		log.Println("Couldn't load env file")
	}

	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_DATABASE")
	dbUser := fmt.Sprintf("%s:%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))

	db := platform.DBConnection {
		Host: dbHost,
		Database: dbName,
		User: dbUser,
	}

	db.Connect()
}


