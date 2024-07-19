package common

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinIOPUT(data []byte, bucketName, objectName, contentType string) (string, error) {
	if Env.Env != Development {
		return "", errors.New("MinIO should be only used in development mode")
	}

	endpoint := fmt.Sprintf("%s:%s", Env.MinioHost, Env.MinioPort)

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(Env.MinioRootUser, Env.MinioRootPassword, ""),
		Secure: false,
	})
	if err != nil {
		return "", err
	}

	// Create a temporary file to store the decoded image
	tempFile, err := os.CreateTemp(os.TempDir(), "upload-*.png")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write decoded data to the temporary file
	if _, err := tempFile.Write(data); err != nil {
		return "", err
	}

	// Upload the file to MinIO
	_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, tempFile.Name(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	// Generate a public URL for the uploaded object
	// Assuming the bucket is configured to allow public access
	publicURL := fmt.Sprintf("http://%s:%s/%s/%s", Env.MinioHost, Env.MinioPort, bucketName, objectName)

	return publicURL, nil
}
