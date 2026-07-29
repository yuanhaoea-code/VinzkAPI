package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewWorkbenchAttachmentObjectStore(cfg *config.Config) (service.WorkbenchAttachmentObjectStore, error) {
	attachmentCfg := cfg.Workbench.Attachments
	if strings.EqualFold(strings.TrimSpace(attachmentCfg.Provider), "s3") {
		return newS3WorkbenchAttachmentStore(context.Background(), attachmentCfg)
	}
	root, err := filepath.Abs(attachmentCfg.LocalDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workbench attachment directory: %w", err)
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create workbench attachment directory: %w", err)
	}
	return &localWorkbenchAttachmentStore{root: root}, nil
}

type localWorkbenchAttachmentStore struct {
	root string
}

func (s *localWorkbenchAttachmentStore) Provider() string { return "local" }

func (s *localWorkbenchAttachmentStore) PresignPut(_ context.Context, _, _ string, _ int64, _ time.Duration) (service.WorkbenchAttachmentUploadPlan, bool, error) {
	return service.WorkbenchAttachmentUploadPlan{}, false, nil
}

func (s *localWorkbenchAttachmentStore) Put(_ context.Context, key, _ string, _ int64, body io.Reader) (int64, error) {
	target, err := s.path(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return 0, err
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".workbench-upload-*")
	if err != nil {
		return 0, err
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	written, copyErr := io.Copy(temporary, body)
	closeErr := temporary.Close()
	if copyErr != nil {
		return written, copyErr
	}
	if closeErr != nil {
		return written, closeErr
	}
	if err := os.Rename(temporaryName, target); err != nil {
		return written, err
	}
	return written, nil
}

func (s *localWorkbenchAttachmentStore) Head(_ context.Context, key string) (*service.WorkbenchAttachmentObjectMetadata, error) {
	target, err := s.path(key)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("workbench attachment is not a regular file")
	}
	return &service.WorkbenchAttachmentObjectMetadata{SizeBytes: info.Size()}, nil
}

func (s *localWorkbenchAttachmentStore) Open(_ context.Context, key string) (io.ReadCloser, error) {
	target, err := s.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(target)
}

func (s *localWorkbenchAttachmentStore) Delete(_ context.Context, key string) error {
	target, err := s.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(target)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *localWorkbenchAttachmentStore) path(key string) (string, error) {
	normalized := filepath.FromSlash(strings.TrimLeft(key, "/"))
	target := filepath.Join(s.root, normalized)
	relative, err := filepath.Rel(s.root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("invalid workbench attachment key")
	}
	return target, nil
}

type s3WorkbenchAttachmentStore struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func newS3WorkbenchAttachmentStore(ctx context.Context, cfg config.WorkbenchAttachmentConfig) (*s3WorkbenchAttachmentStore, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load workbench attachment s3 config: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
		options.UsePathStyle = cfg.ForcePathStyle
		options.APIOptions = append(options.APIOptions, v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
		options.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	return &s3WorkbenchAttachmentStore{client: client, presign: s3.NewPresignClient(client), bucket: cfg.Bucket}, nil
}

func (s *s3WorkbenchAttachmentStore) Provider() string { return "s3" }

func (s *s3WorkbenchAttachmentStore) PresignPut(ctx context.Context, key, contentType string, sizeBytes int64, expiry time.Duration) (service.WorkbenchAttachmentUploadPlan, bool, error) {
	result, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType), ContentLength: aws.Int64(sizeBytes),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return service.WorkbenchAttachmentUploadPlan{}, true, err
	}
	headers := make(map[string]string, len(result.SignedHeader))
	for name, values := range result.SignedHeader {
		if len(values) > 0 && !strings.EqualFold(name, "host") && !strings.EqualFold(name, "content-length") {
			headers[name] = values[0]
		}
	}
	return service.WorkbenchAttachmentUploadPlan{URL: result.URL, Method: result.Method, Headers: headers}, true, nil
}

func (s *s3WorkbenchAttachmentStore) Put(ctx context.Context, key, contentType string, sizeBytes int64, body io.Reader) (int64, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), Body: body, ContentType: aws.String(contentType), ContentLength: aws.Int64(sizeBytes),
	})
	if err != nil {
		return 0, err
	}
	return sizeBytes, nil
}

func (s *s3WorkbenchAttachmentStore) Head(ctx context.Context, key string) (*service.WorkbenchAttachmentObjectMetadata, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	return &service.WorkbenchAttachmentObjectMetadata{
		SizeBytes: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType), ETag: strings.Trim(aws.ToString(result.ETag), "\""),
	}, nil
}

func (s *s3WorkbenchAttachmentStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (s *s3WorkbenchAttachmentStore) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	return err
}

var _ service.WorkbenchAttachmentObjectStore = (*localWorkbenchAttachmentStore)(nil)
var _ service.WorkbenchAttachmentObjectStore = (*s3WorkbenchAttachmentStore)(nil)
