package common

import "context"

func Upload(ctx context.Context, data []byte, bucketName, objectName, contentType string) (string, error) {
	return MinIOPUT(ctx, data, bucketName, objectName, contentType)
}
