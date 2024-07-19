package common

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinIOPUT(base64Image, bucketName, objectName string) (string, error) {
	if Config.Env != Development {
		return "", errors.New("MinIO should be only used in development mode")
	}

	endpoint := fmt.Sprintf("%s:%s", Config.MinioHost, Config.MinioPort)

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(Config.MinioRootUser, Config.MinioRootPassword, ""),
		Secure: false,
	})
	if err != nil {
		return "", err
	}

	// Decode base64 string
	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return "", nil
	}

	// Create a temporary file to store the decoded image
	tempFile, err := os.CreateTemp(os.TempDir(), "upload-*.png")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write decoded data to the temporary file
	if _, err := tempFile.Write(imageData); err != nil {
		return "", err
	}

	// Upload the file to MinIO
	_, err = minioClient.FPutObject(context.Background(), bucketName, objectName, tempFile.Name(), minio.PutObjectOptions{
		ContentType: "image/png",
	})
	if err != nil {
		return "", err
	}

	// Generate a public URL for the uploaded object
	// Assuming the bucket is configured to allow public access
	publicURL := fmt.Sprintf("http://%s:%s/%s/%s", Config.MinioHost, Config.MinioPort, bucketName, objectName)

	return publicURL, nil
}
