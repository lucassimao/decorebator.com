package common

import "context"

func Upload(ctx context.Context, data []byte, bucketName, objectName, contentType string) (string, error) {
	return MinIOPUT(ctx, data, bucketName, objectName, contentType)
}

type UploadReceipt struct {
	URL        string
	Bucket     string
	ObjectName string
	VersionID  string
}

func UploadWithReceipt(ctx context.Context, data []byte, bucketName, objectName, contentType string) (UploadReceipt, error) {
	return MinIOPUTWithReceipt(ctx, data, bucketName, objectName, contentType)
}
