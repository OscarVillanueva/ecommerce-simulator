package platform

import (
	"fmt"
	"sync"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var lock = &sync.Mutex{}
var SharedInstance *gorm.DB

type dbConnection struct {
	Host string
	Database string
	User string
}

func (db dbConnection) Connect() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db.User, db.Host, db.Database)

	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	return connection, err
}

func GetInstance() {
	if SharedInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if SharedInstance == nil {
			log.Info("Creating a db instance")

			err := godotenv.Load()

			dbHost := os.Getenv("DB_HOST")
			dbName := os.Getenv("DB_DATABASE")
			dbUser := fmt.Sprintf("%s:%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))

			if err != nil {
				log.Warning("Could't load env: ", err)
				return
			}

			dbConn := dbConnection {
				Host: dbHost,
				Database: dbName,
				User: dbUser,
			}

			db, err := dbConn.Connect()

			if err != nil {
				log.Warning("Couldn't connect to the database: ", err)
				return
			}

			SharedInstance = db

			log.Info("Instance created")
		} else {
			log.Info("Instance already created")
		}
	} else {
		log.Info("Instance already created")
	}
}
