package common

func Upload(data []byte, bucketName, objectName, contentType string) (string, error) {
	return MinIOPUT(data, bucketName, objectName, contentType)
	// switch Env.Env {
	// case Development:
	// 	return MinIOPUT(data, bucketName, objectName, contentType)
	// case Production:
	// 	return S3Upload(data, bucketName, objectName, &contentType)
	// default:
	// 	return "", fmt.Errorf("unsupported environment: %v", Env.Env)
	// }
}
