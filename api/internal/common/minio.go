package common

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func newMinIOClient() (*minio.Client, error) {
	var endpoint string
	var useSecure bool

	env := os.Getenv("ENV")
	minioHost := os.Getenv("MINIO_HOST")
	minioPort := os.Getenv("MINIO_PORT")
	switch env {
	case DevelopmentEnv, TestEnv:
		endpoint = fmt.Sprintf("%s:%s", minioHost, minioPort)
	case ProductionEnv:
		endpoint = minioHost
		useSecure = true
	default:
		return nil, fmt.Errorf("unsupported environment: %v", env)
	}

	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ROOT_PASSWORD"), ""),
		Secure: useSecure,
	})
}

func MinIOPUT(ctx context.Context, data []byte, bucketName, objectName, contentType string) (string, error) {
	env := os.Getenv("ENV")
	minioHost := os.Getenv("MINIO_HOST")
	minioPort := os.Getenv("MINIO_PORT")

	minioClient, err := newMinIOClient()
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
	case DevelopmentEnv, TestEnv:
		publicURL = fmt.Sprintf("http://%s:%s/%s/%s", minioHost, minioPort, bucketName, objectName)
	case ProductionEnv:
		publicURL = fmt.Sprintf("https://%s.%s/%s", bucketName, minioHost, objectName)
	default:
		return "", fmt.Errorf("unsupported environment: %v", env)
	}

	return publicURL, nil
}

// MinIODeleteObject removes one object, including the object created by an
// upload whose following database update failed.
func MinIODeleteObject(ctx context.Context, bucketName, objectName string) error {
	client, err := newMinIOClient()
	if err != nil {
		return err
	}
	return client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

// MinIODeletePrefix removes every current object and historical version under
// an exact prefix. Repeating the operation is safe.
func MinIODeletePrefix(ctx context.Context, bucketName, prefix string) error {
	client, err := newMinIOClient()
	if err != nil {
		return err
	}

	for object := range client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:       prefix,
		Recursive:    true,
		WithVersions: true,
	}) {
		if object.Err != nil {
			return fmt.Errorf("list prefix %q: %w", prefix, object.Err)
		}
		if err := client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{
			VersionID: object.VersionID,
		}); err != nil {
			return fmt.Errorf("remove %q: %w", object.Key, err)
		}
	}
	return nil
}
