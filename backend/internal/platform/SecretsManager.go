package platform

import (
	"context"
	"errors"
	"sync"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var (
	redisClient *redis.Client
	redisOnce sync.Once
)

func getSecretsClient() *redis.Client {
	redisOnce.Do(func ()  {
		err := godotenv.Load()

		if err != nil {
			log.Warning("Couldn't load env", err)
			return
		}


		password := os.Getenv("SECRETS_PASSWORD")
		host := os.Getenv("SECRETS_HOST")

		redisClient = redis.NewClient(&redis.Options{
			Addr: host,
			Password: password,
			DB: 0,
			Protocol: 2,
		})

		log.Info("Secrets connection enabled")
	})

	return redisClient
}

func SaveSecret(key string, value string, ctx context.Context) error {
	manager := getSecretsClient()

	if manager == nil {
		return errors.New("Empty Secrets")
	}

	err := manager.Set(ctx, key, value, 0).Err()

	return err
}

func GetSecret(key string, ctx context.Context) (string, error)  {
	manager := getSecretsClient()

	if manager == nil {
		return "", errors.New("Empty Secrets")
	}

	val, err := manager.Get(ctx, key).Result()

	return val, err
}

