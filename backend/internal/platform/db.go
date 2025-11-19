package platform

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DBConnection struct {
	Host string
	Database string
	User string
}

func (db DBConnection) Connect() {
	dsn := fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.User, db.Host, db.Database)

	_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Panicf("Failed to connect database: %v", err)
	}

	log.Println("Connected to the database successfully!")
}
