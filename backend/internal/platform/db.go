package platform

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/joho/godotenv"
)

var sharedInstance *gorm.DB

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

func InitDbConnection (ctx context.Context) error {
	DBManager := "db-manger"
	tr := otel.Tracer(DBManager)
	_, span := tr.Start(ctx, fmt.Sprintf("%s.InitSecretsManager", DBManager))
	defer span.End()

	if err := godotenv.Load(); err != nil {
		err := errors.New("Unable to load godot")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}
	
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		err := errors.New("DB_HOST environment variable is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	dbName := os.Getenv("DB_DATABASE")
	if dbName == "" {
		err := errors.New("DB_DATABASE environment variable is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	dbUser := fmt.Sprintf("%s:%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))	

	dbConn := dbConnection {
		Host: dbHost,
		Database: dbName,
		User: dbUser,
	}

	var err error
	sharedInstance, err = dbConn.Connect()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return err
}

func GetInstance() *gorm.DB {
	return sharedInstance
}

/*
func GetInstance() (*gorm.DB) {
	/* DLC singleton
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

	once.Do(func() {
		log.Info("Creating a db instance")

		err := godotenv.Load()

		if err != nil {
			log.Warning("Could't load env: ", err)
			return
		}
		
		dbHost := os.Getenv("DB_HOST")
		dbName := os.Getenv("DB_DATABASE")
		dbUser := fmt.Sprintf("%s:%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))	

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

		sharedInstance = db

		log.Info("Instance Created")
	})

	return sharedInstance
}*/
