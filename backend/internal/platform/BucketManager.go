package platform

import (
	"os"
	"errors"
	"context"
	"io"
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	minioClient *minio.Client
	bucketName string
	imagesHost string
)

const BucketManager = "bucket-manager"

func createMinioClient() (*minio.Client, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	
	endpoint := os.Getenv("MINIO_HOST")
	accessKey := os.Getenv("MINIO_ROOT_USER")
	secretKey := os.Getenv("MINIO_ROOT_PASSWORD")
	bucketName = os.Getenv("MINIO_BUCKET_NAME")
	imagesHost = os.Getenv("IMAGES_HOST")
	useSSL := false

	client, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	return client, err
}

func ensureBucketExists(ctx context.Context) error {
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	if !exists {
		if err := minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}

		policy := fmt.Sprintf(`{
        "Version": "2012-10-17",
        "Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"]}]
    }`, bucketName)
		minioClient.SetBucketPolicy(ctx, bucketName, policy)
	}

	return nil
}

func InitMinio(ctx context.Context) error {
	tr := otel.Tracer(BucketManager)
	getBucketCtx, span := tr.Start(ctx, fmt.Sprintf("%s.ensureBucketExists", BucketManager))
	defer span.End()

	var err error
	minioClient, err = createMinioClient();
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to create minio client: %s", err.Error()))
		return err
	}

	if err := ensureBucketExists(getBucketCtx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("Unable to create minio bucket: %s", err.Error()))
		return err
	}

	return nil
}

func PutImage(imageName string, file io.Reader, size int64, contentType string, ctx context.Context) (string, error) {
	tr := otel.Tracer(BucketManager)
	_, span := tr.Start(ctx, fmt.Sprintf("%s.ensureBucketExists", BucketManager))
	defer span.End()

	if minioClient == nil {
		err := errors.New("We couldn't connect to the bucket")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}

	_, putError := minioClient.PutObject(ctx,
		bucketName,
		imageName,
		file,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		})

	if putError != nil {
		span.RecordError(putError)
		span.SetStatus(codes.Error, putError.Error())
		return "", putError
	}

	imageURL := fmt.Sprintf("%s/%s/%s", imagesHost, bucketName, imageName)
	return imageURL, nil
}
