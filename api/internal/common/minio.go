package common

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func MinIOPUT(data []byte, bucketName, objectName, contentType string) (string, error) {
	endpoint := fmt.Sprintf("%s:%s", Env.MinioHost, Env.MinioPort)

	// Initialize minio client object
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(Env.MinioRootUser, Env.MinioRootPassword, ""),
		Secure: false,
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

	// Generate a public URL for the uploaded object
	// Assuming the bucket is configured to allow public access
	publicURL := fmt.Sprintf("http://%s:%s/%s/%s", Env.MinioHost, Env.MinioPort, bucketName, objectName)

	return publicURL, nil
}
