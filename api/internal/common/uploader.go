package common

import (
	"fmt"
)

func Upload(data []byte, bucketName, objectName, contentType string) (string, error) {
	switch Env.Env {
	case Development:
		return MinIOPUT(data, bucketName, objectName, contentType)
	case Production:
		return S3Upload(data, bucketName, objectName, &contentType)
	default:
		return "", fmt.Errorf("unsupported environment: %v", Env.Env)
	}
}
