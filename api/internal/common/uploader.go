package common

func Upload(data []byte, bucketName, objectName, contentType string) (string, error) {
	return MinIOPUT(data, bucketName, objectName, contentType)
}
