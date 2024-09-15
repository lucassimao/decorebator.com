package common

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinIOPUT(data []byte, bucketName, objectName, contentType string) (string, error) {

	var endpoint string
	var useSecure bool

	switch Env.Env {
	case Development:
		endpoint = fmt.Sprintf("%s:%s", Env.MinioHost, Env.MinioPort)
		useSecure = false
	case Production:
		endpoint = Env.MinioHost
		useSecure = true
	default:
		return "", fmt.Errorf("unsupported environment: %v", Env.Env)
	}

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(Env.MinioRootUser, Env.MinioRootPassword, ""),
		Secure: useSecure,
	})
	if err != nil {
		return "", err
	}

	dataReader := bytes.NewReader(data)

	// Upload the file to MinIO
	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, dataReader, int64(len(data)), minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
	})

	if err != nil {
		return "", err
	}

	var publicURL string

	switch Env.Env {
	case Development:
		publicURL = fmt.Sprintf("http://%s:%s/%s/%s", Env.MinioHost, Env.MinioPort, bucketName, objectName)
	case Production:
		publicURL = fmt.Sprintf("https://%s.%s/%s", bucketName, Env.MinioHost, objectName)
	default:
		return "", fmt.Errorf("unsupported environment: %v", Env.Env)
	}

	return publicURL, nil
}
