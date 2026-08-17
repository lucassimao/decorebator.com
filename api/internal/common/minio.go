package common

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minIOClientCache struct {
	mu     sync.Mutex
	config *MinIOConfig
	client *minio.Client
}

var sharedMinIOClient minIOClientCache

var minIOBucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)

const (
	defaultMinIOBucket = "decorebator"
	httpsScheme        = "https"
)

type MinIOConfig struct {
	Environment   string
	Endpoint      string
	AccessKey     string
	SecretKey     string
	Bucket        string
	LegacyBuckets []string
	Secure        bool
	PublicBaseURL string
}

func LoadMinIOConfig(environment string) (MinIOConfig, error) { //nolint:gocyclo // environment validation is intentionally fail closed
	if err := ValidateRuntimeEnvironment(environment); err != nil {
		return MinIOConfig{}, err
	}
	deployed := environment == ProductionEnv || environment == StagingEnv
	host := firstNonempty(os.Getenv("MINIO_HOST"), os.Getenv("MINIO_ENDPOINT"))
	port := strings.TrimSpace(os.Getenv("MINIO_PORT"))
	accessKey := firstNonempty(os.Getenv("MINIO_ROOT_USER"), os.Getenv("MINIO_ACCESS_KEY"))
	secretKey := firstNonempty(os.Getenv("MINIO_ROOT_PASSWORD"), os.Getenv("MINIO_SECRET_KEY"))
	if !deployed {
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "9000"
		}
		if accessKey == "" {
			accessKey = "local-minio-user"
		}
		if secretKey == "" {
			secretKey = "local-minio-password"
		}
	}
	if host == "" || accessKey == "" || secretKey == "" {
		return MinIOConfig{}, errors.New("deployed storage requires endpoint and credentials")
	}
	if strings.Contains(host, "://") || strings.ContainsAny(host, "/?#") {
		return MinIOConfig{}, errors.New("MINIO_HOST/MINIO_ENDPOINT must be a host without scheme or path")
	}
	endpoint := host
	if port != "" {
		if _, _, err := net.SplitHostPort(host); err != nil {
			endpoint = net.JoinHostPort(host, port)
		}
	}
	secure := deployed
	if rawSecure := strings.TrimSpace(os.Getenv("MINIO_USE_SSL")); rawSecure != "" {
		parsed, err := strconv.ParseBool(rawSecure)
		if err != nil {
			return MinIOConfig{}, errors.New("MINIO_USE_SSL must be true or false")
		}
		secure = parsed
	}
	if deployed && !secure {
		return MinIOConfig{}, errors.New("deployed storage requires TLS")
	}
	bucket := strings.TrimSpace(os.Getenv("MINIO_BUCKET"))
	if bucket == "" {
		if deployed {
			return MinIOConfig{}, errors.New("deployed storage requires MINIO_BUCKET")
		}
		bucket = defaultMinIOBucket
	}
	if !validMinIOBucketName(bucket) {
		return MinIOConfig{}, errors.New("MINIO_BUCKET is invalid")
	}
	legacyBuckets := make([]string, 0)
	for _, candidate := range strings.Split(os.Getenv("MINIO_LEGACY_BUCKETS"), ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate != "" && !validMinIOBucketName(candidate) {
			return MinIOConfig{}, errors.New("MINIO_LEGACY_BUCKETS contains an invalid bucket")
		}
		if candidate != "" && candidate != bucket && !slices.Contains(legacyBuckets, candidate) {
			legacyBuckets = append(legacyBuckets, candidate)
		}
	}
	publicBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("MINIO_PUBLIC_BASE_URL")), "/")
	if deployed && publicBaseURL == "" {
		return MinIOConfig{}, errors.New("deployed storage requires MINIO_PUBLIC_BASE_URL")
	}
	if publicBaseURL != "" {
		parsed, valid := ParsePublicObjectURL(publicBaseURL)
		if !valid || (deployed && parsed.Scheme != httpsScheme) {
			return MinIOConfig{}, errors.New("MINIO_PUBLIC_BASE_URL must be an absolute URL (HTTPS when deployed) without credentials, query, or fragment")
		}
	}
	return MinIOConfig{
		Environment: environment, Endpoint: endpoint, AccessKey: accessKey, SecretKey: secretKey,
		Bucket: bucket, LegacyBuckets: legacyBuckets, Secure: secure, PublicBaseURL: publicBaseURL,
	}, nil
}

func ParsePublicObjectURL(rawURL string) (*url.URL, bool) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != httpsScheme && parsed.Scheme != "http") ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, false
	}
	return parsed, true
}

func ValidPublicObjectURLForKey(rawURL, objectName string) bool {
	parsed, valid := ParsePublicObjectURL(rawURL)
	return valid && strings.HasSuffix(parsed.EscapedPath(), "/"+objectName)
}

func ConfigureMinIO(config MinIOConfig) error {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""), Secure: config.Secure,
	})
	if err != nil {
		return err
	}
	sharedMinIOClient.mu.Lock()
	defer sharedMinIOClient.mu.Unlock()
	configCopy := config
	sharedMinIOClient.config = &configCopy
	sharedMinIOClient.client = client
	return nil
}

func ConfigureMinIOFromEnvironment(environment string) error {
	config, err := LoadMinIOConfig(environment)
	if err != nil {
		return err
	}
	return ConfigureMinIO(config)
}

func reusableMinIOClient() (*minio.Client, MinIOConfig, error) {
	sharedMinIOClient.mu.Lock()
	defer sharedMinIOClient.mu.Unlock()
	if sharedMinIOClient.client == nil || sharedMinIOClient.config == nil {
		return nil, MinIOConfig{}, errors.New("MinIO is not configured")
	}
	return sharedMinIOClient.client, *sharedMinIOClient.config, nil
}

func MinIOBucketName() string {
	sharedMinIOClient.mu.Lock()
	defer sharedMinIOClient.mu.Unlock()
	if sharedMinIOClient.config == nil {
		return defaultMinIOBucket
	}
	return sharedMinIOClient.config.Bucket
}

func MinIOBucketAllowed(bucket string) bool {
	sharedMinIOClient.mu.Lock()
	defer sharedMinIOClient.mu.Unlock()
	if sharedMinIOClient.config == nil {
		return bucket == defaultMinIOBucket
	}
	return bucket == sharedMinIOClient.config.Bucket || slices.Contains(sharedMinIOClient.config.LegacyBuckets, bucket)
}

func MinIOCleanupBuckets() []string {
	sharedMinIOClient.mu.Lock()
	defer sharedMinIOClient.mu.Unlock()
	if sharedMinIOClient.config == nil {
		return []string{defaultMinIOBucket}
	}
	result := make([]string, 0, 1+len(sharedMinIOClient.config.LegacyBuckets))
	result = append(result, sharedMinIOClient.config.Bucket)
	return append(result, sharedMinIOClient.config.LegacyBuckets...)
}

func validMinIOBucketName(bucket string) bool {
	return minIOBucketPattern.MatchString(bucket) && !strings.Contains(bucket, "..") &&
		!strings.Contains(bucket, ".-") && !strings.Contains(bucket, "-.") && net.ParseIP(bucket) == nil
}

func MinIOPUT(ctx context.Context, data []byte, bucketName, objectName, contentType string) (string, error) {
	receipt, err := MinIOPUTWithReceipt(ctx, data, bucketName, objectName, contentType)
	return receipt.URL, err
}

func MinIOPUTWithReceipt(ctx context.Context, data []byte, bucketName, objectName, contentType string) (UploadReceipt, error) {
	minioClient, storageConfig, err := reusableMinIOClient()
	if err != nil {
		return UploadReceipt{}, err
	}

	dataReader := bytes.NewReader(data)

	// Upload the file to MinIO
	uploadInfo, err := minioClient.PutObject(ctx, bucketName, objectName, dataReader, int64(len(data)), minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
	})

	if err != nil {
		return UploadReceipt{}, err
	}

	publicURL := storageConfig.publicObjectURL(bucketName, objectName)

	return UploadReceipt{URL: publicURL, Bucket: bucketName, ObjectName: objectName, VersionID: uploadInfo.VersionID}, nil
}

func MinIOObjectURL(bucketName, objectName string) (string, error) {
	_, storageConfig, err := reusableMinIOClient()
	if err != nil {
		return "", err
	}
	if bucketName != storageConfig.Bucket && !slices.Contains(storageConfig.LegacyBuckets, bucketName) {
		return "", errors.New("object bucket is not configured")
	}
	return storageConfig.publicObjectURL(bucketName, objectName), nil
}

// MinIODeleteObject removes one object, including the object created by an
// upload whose following database update failed.
func MinIODeleteObject(ctx context.Context, bucketName, objectName string) error {
	return MinIODeleteObjectVersion(ctx, bucketName, objectName, "")
}

func MinIODeleteObjectVersion(ctx context.Context, bucketName, objectName, versionID string) error {
	client, _, err := reusableMinIOClient()
	if err != nil {
		return err
	}
	return client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{VersionID: versionID})
}

// MinIODeletePrefix removes every current object and historical version under
// an exact prefix. Repeating the operation is safe.
func MinIODeletePrefix(ctx context.Context, bucketName, prefix string) error {
	client, _, err := reusableMinIOClient()
	if err != nil {
		return err
	}

	listContext, cancelList := context.WithCancel(ctx)
	objects := client.ListObjects(listContext, bucketName, minio.ListObjectsOptions{
		Prefix:       prefix,
		Recursive:    true,
		WithVersions: true,
	})
	return drainMinIOListing(cancelList, objects, func(object minio.ObjectInfo) error {
		if err := client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{
			VersionID: object.VersionID,
		}); err != nil {
			return fmt.Errorf("remove %q: %w", object.Key, err)
		}
		return nil
	}, func(_ minio.ObjectInfo) bool { return true }, "prefix", prefix)
}

// MinIODeleteObjectAllVersions removes only the exact object key and each of
// its historical versions. It never treats objectName as a caller-controlled
// prefix, so a sibling key cannot be removed by reconciliation.
func MinIODeleteObjectAllVersions(ctx context.Context, bucketName, objectName string) error {
	client, _, err := reusableMinIOClient()
	if err != nil {
		return err
	}
	listContext, cancelList := context.WithCancel(ctx)
	objects := client.ListObjects(listContext, bucketName, minio.ListObjectsOptions{
		Prefix: objectName, Recursive: true, WithVersions: true,
	})
	return drainMinIOListing(cancelList, objects, func(object minio.ObjectInfo) error {
		if err := client.RemoveObject(ctx, bucketName, object.Key, minio.RemoveObjectOptions{VersionID: object.VersionID}); err != nil {
			return fmt.Errorf("remove %q version %q: %w", object.Key, object.VersionID, err)
		}
		return nil
	}, func(object minio.ObjectInfo) bool { return object.Key == objectName }, "object", objectName)
}

func drainMinIOListing(
	cancel context.CancelFunc,
	objects <-chan minio.ObjectInfo,
	remove func(minio.ObjectInfo) error,
	shouldRemove func(minio.ObjectInfo) bool,
	targetKind, target string,
) error {
	defer cancel()
	var firstErr error
	for object := range objects {
		if firstErr != nil {
			continue
		}
		if object.Err != nil {
			firstErr = fmt.Errorf("list %s %q: %w", targetKind, target, object.Err)
			cancel()
			continue
		}
		if !shouldRemove(object) {
			continue
		}
		if err := remove(object); err != nil {
			firstErr = err
			cancel()
		}
	}
	return firstErr
}

func (config MinIOConfig) publicObjectURL(bucketName, objectName string) string {
	if config.PublicBaseURL != "" && bucketName == config.Bucket {
		return config.PublicBaseURL + "/" + objectName
	}
	if config.Environment == ProductionEnv || config.Environment == StagingEnv {
		host := config.Endpoint
		if parsedHost, _, err := net.SplitHostPort(config.Endpoint); err == nil {
			host = parsedHost
		}
		return fmt.Sprintf("https://%s.%s/%s", bucketName, host, objectName)
	}
	scheme := "http"
	if config.Secure {
		scheme = httpsScheme
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, config.Endpoint, bucketName, objectName)
}

func firstNonempty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
