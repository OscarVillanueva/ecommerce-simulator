package platform

import (
	"os"
	"context"
	"io"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

const BUCKET_NAME = "ecommerce-image"
const IMAGES_HOST = "images.localhost"

func GetBucketClient() (*minio.Client, error)  {
	err := godotenv.Load()

	if err != nil {
		log.Warning("Could't load env: ", err)
		return nil, err
	}
	
	endpoint := os.Getenv("MINIO_HOST")
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Error()
	}

	return minioClient, err
}

func PutImage(imageName string, file io.Reader, size int64, contentType string, ctx context.Context) (string, error) {
	client, err := GetBucketClient()

	if err != nil {
		return "", err
	}

	if exists, _ := client.BucketExists(ctx, BUCKET_NAME); !exists {
		client.MakeBucket(ctx, BUCKET_NAME, minio.MakeBucketOptions{})
		policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::` + BUCKET_NAME + `/*"]}]}`
		client.SetBucketPolicy(ctx, BUCKET_NAME, policy)
	}

	_, putError := client.PutObject(ctx,
		BUCKET_NAME,
		imageName,
		file,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})

	if putError != nil {
		return "", err
	}

	imageURL := fmt.Sprintf("%s/%s/%s", IMAGES_HOST, BUCKET_NAME, imageName)
	return imageURL, nil
}
