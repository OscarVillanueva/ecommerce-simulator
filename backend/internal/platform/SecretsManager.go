package platform

import (
	"context"
	"errors"
	"time"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
)

var (
	redisClient *redis.Client
)

const SecretsManager = "secrets-manager"

func InitSecretsManager(ctx context.Context) error {
	tr := otel.Tracer(SecretsManager)
	initCtx, span := tr.Start(ctx, fmt.Sprintf("%s.InitSecretsManager", SecretsManager))
	defer span.End()

	if err := godotenv.Load(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	password := os.Getenv("SECRETS_PASSWORD")
	if password == "" {
		err := errors.New("SECRETS_PASSWORD environment variable is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	host := os.Getenv("SECRETS_HOST")
	if host == "" {
		err := errors.New("SECRETS_HOST environment variable is empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to load env: %s", err.Error()))
		return err
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: host,
		Password: password,
		DB: 0,
		Protocol: 2,
	})

	pingCtx, cancel := context.WithTimeout(initCtx, 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(pingCtx).Result()
	return err
}

func SaveSecret(key string, value string, ctx context.Context) error {
	tr := otel.Tracer(EmailManagerName)
	saveCtx, span := tr.Start(ctx, fmt.Sprintf("%s.ensureBucketExists", EmailManagerName))
	defer span.End()

	if redisClient == nil {
		err := errors.New("Empty Secrets")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.String("secret-key", key),
	)

	if err := redisClient.Set(saveCtx, key, value, 0).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func GetSecret(key string, ctx context.Context) (string, error)  {
	tr := otel.Tracer(EmailManagerName)
	getCtx, span := tr.Start(ctx, fmt.Sprintf("%s.ensureBucketExists", EmailManagerName))
	defer span.End()

	if redisClient == nil {
		err := errors.New("Empty Secrets")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", errors.New("Empty Secrets")
	}

	span.SetAttributes(
		attribute.String("secret-key", key),
	)

	val, err := redisClient.Get(getCtx, key).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}

	return val, err
}

