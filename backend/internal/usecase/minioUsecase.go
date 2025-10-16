package usecase

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioUsecase struct {
	client     *minio.Client
	bucketName string
}

func NewMinioUsecase(endpoint, accessKey, secretKey, bucketName string, useSSL bool) ObjectStorageInterface {
	// Initialize MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize MinIO client: %v", err))
	}

	// Ensure the bucket exists
	mu := &MinioUsecase{
		client:     minioClient,
		bucketName: bucketName,
	}
	if err := mu.ensureBucketExists(context.Background()); err != nil {
		panic(fmt.Sprintf("Failed to ensure bucket exists: %v", err))
	}

	return mu
}

// SaveSBOM saves an SBOM (Software Bill of Materials) to object storage
func (s *MinioUsecase) SaveSBOM(ctx context.Context, appID string, appName string, sbomData []byte, format string) (string, error) {
	timestamp := time.Now().Format("2006-01-02")
	fileExtension := "json"
	if format == "xml" {
		fileExtension = "xml"
	}

	objectKey := fmt.Sprintf("sbom/%s/%s/%s_sbom.%s",
		appName,
		timestamp,
		appID,
		fileExtension)

	reader := bytes.NewReader(sbomData)
	contentType := "application/json"
	if format == "xml" {
		contentType = "application/xml"
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, int64(len(sbomData)), minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"app-id":        appID,
			"app-name":      appName,
			"document-type": "sbom",
			"format":        format,
			"generated-at":  time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload SBOM: %w", err)
	}

	slog.Info("SBOM saved to object storage",
		"object_key", objectKey,
		"app_id", appID,
		"app_name", appName,
		"size_bytes", len(sbomData))

	return objectKey, nil
}

// SaveVulnerabilityReport saves a vulnerability report to object storage
func (s *MinioUsecase) SaveVulnerabilityReport(ctx context.Context, appID string, appName string, reportData []byte, format string) (string, error) {
	timestamp := time.Now().Format("2006-01-02")
	fileExtension := "json"
	if format == "pdf" {
		fileExtension = "pdf"
	} else if format == "html" {
		fileExtension = "html"
	}

	objectKey := fmt.Sprintf("vulnerability-reports/%s/%s/%s_vuln_report.%s",
		appName,
		timestamp,
		appID,
		fileExtension)

	reader := bytes.NewReader(reportData)
	contentType := "application/json"
	if format == "pdf" {
		contentType = "application/pdf"
	} else if format == "html" {
		contentType = "text/html"
	}

	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, int64(len(reportData)), minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"app-id":        appID,
			"app-name":      appName,
			"document-type": "vulnerability-report",
			"format":        format,
			"generated-at":  time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload vulnerability report: %w", err)
	}

	slog.Info("Vulnerability report saved to object storage",
		"object_key", objectKey,
		"app_id", appID,
		"app_name", appName,
		"size_bytes", len(reportData))

	return objectKey, nil
}

// GetSBOM retrieves an SBOM from object storage
func (s *MinioUsecase) GetSBOM(ctx context.Context, objectKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get SBOM: %w", err)
	}
	defer object.Close()

	// Read the object content
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(object); err != nil {
		return nil, fmt.Errorf("failed to read SBOM: %w", err)
	}

	return buf.Bytes(), nil
}

// GetVulnerabilityReport retrieves a vulnerability report from object storage
func (s *MinioUsecase) GetVulnerabilityReport(ctx context.Context, objectKey string) ([]byte, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get vulnerability report: %w", err)
	}
	defer object.Close()

	// Read the object content
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(object); err != nil {
		return nil, fmt.Errorf("failed to read vulnerability report: %w", err)
	}

	return buf.Bytes(), nil
}

// ListSBOMs lists all SBOMs for an application
func (s *MinioUsecase) ListSBOMs(ctx context.Context, appName string) ([]string, error) {
	prefix := fmt.Sprintf("sbom/%s/", appName)

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objectKeys []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing SBOMs: %w", object.Err)
		}
		objectKeys = append(objectKeys, object.Key)
	}

	return objectKeys, nil
}

// ListVulnerabilityReports lists all vulnerability reports for an application
func (s *MinioUsecase) ListVulnerabilityReports(ctx context.Context, appName string) ([]string, error) {
	prefix := fmt.Sprintf("vulnerability-reports/%s/", appName)

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objectKeys []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing vulnerability reports: %w", object.Err)
		}
		objectKeys = append(objectKeys, object.Key)
	}
	return objectKeys, nil
}

// ensureBucketExists creates the bucket if it doesn't exist
func (s *MinioUsecase) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		slog.Info("Created bucket", "bucket", s.bucketName)
	}

	return nil
}
