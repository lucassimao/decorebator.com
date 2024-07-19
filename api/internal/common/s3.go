package common

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func S3Upload(data []byte, bucketName, objectName string, contentType *string) (string, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(Env.Aws.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(Env.Aws.AccessKeyId, Env.Aws.SecretAccessKey, "")),
	)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Create an S3 client
	svc := s3.NewFromConfig(cfg)

	// Create the PutObject input
	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectName),
		ContentType: contentType,
		Body:        bytes.NewReader(data),
		ACL:         types.ObjectCannedACLPublicRead, // Optional: Set ACL to public read
	}

	// Upload the object to S3
	_, err = svc.PutObject(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %v", err)
	}

	// Generate the public URL
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, Env.Aws.Region, objectName), nil
}
