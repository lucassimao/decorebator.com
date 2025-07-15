package common

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinIOPUT(ctx context.Context, data []byte, bucketName, objectName, contentType string) (string, error) {
	var endpoint string
	var useSecure bool

	env := os.Getenv("ENV")
	minioHost := os.Getenv("MINIO_HOST")
	minioPort := os.Getenv("MINIO_PORT")

	switch env {
	case "development":
		endpoint = fmt.Sprintf("%s:%s", minioHost, minioPort)
		useSecure = false
	case "production":
		endpoint = minioHost
		useSecure = true
	default:
		return "", fmt.Errorf("unsupported environment: %v", env)
	}

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: useSecure,
	})
	if err != nil {
		return "", err
	}

	dataReader := bytes.NewReader(data)

	// Upload the file to MinIO
	_, err = minioClient.PutObject(ctx, bucketName, objectName, dataReader, int64(len(data)), minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
	})

	if err != nil {
		return "", err
	}

	var publicURL string

	switch env {
	case "development":
		publicURL = fmt.Sprintf("http://%s:%s/%s/%s", minioHost, minioPort, bucketName, objectName)
	case "production":
		publicURL = fmt.Sprintf("https://%s.%s/%s", bucketName, minioHost, objectName)
	default:
		return "", fmt.Errorf("unsupported environment: %v", env)
	}

	return publicURL, nil
}
