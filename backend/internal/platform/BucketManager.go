package platform

import (
	"os"
	"errors"
	"context"
	"io"
	"fmt"
	"sync"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	log "github.com/sirupsen/logrus"
)

var (
	minioClient *minio.Client
	minioOnce sync.Once
	bucketName string
	imagesHost string
)


func ensureBucketExists() {
	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		log.Error("Failed to check bucket status:", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Error("Failed to create bucket:", err)
		}
		// Set policy to public read (simplification)
		policy := `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::` + bucketName + `/*"]}]}`
		minioClient.SetBucketPolicy(ctx, bucketName, policy)
	}
}

func getBucketClient() *minio.Client  {
	minioOnce.Do(func () {
		err := godotenv.Load()

		if err != nil {
			log.Warning("Could't load env: ", err)
			return
		}
		
		endpoint := os.Getenv("MINIO_HOST")
		accessKey := os.Getenv("MINIO_ROOT_USER")
		secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
		bucketName = os.Getenv("MINIO_BUCKET_NAME")
		imagesHost = os.Getenv("IMAGES_HOST")
		useSSL := false

		minioClient, err = minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: useSSL,
		})

		if err != nil {
			log.Error(err)
			return
		}

		ensureBucketExists()
	})

	return minioClient
}

func PutImage(imageName string, file io.Reader, size int64, contentType string, ctx context.Context) (string, error) {
	client := getBucketClient()

	if client == nil {
		return "", errors.New("We couldn't connect to the bucket")
	}

	_, putError := client.PutObject(ctx,
		bucketName,
		imageName,
		file,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})

	if putError != nil {
		return "", putError
	}

	imageURL := fmt.Sprintf("%s/%s/%s", imagesHost, bucketName, imageName)
	return imageURL, nil
}
