package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/google/uuid"
)

const (
	WorkbenchMaxAttachmentsPerMessage       = 8
	WorkbenchMaxAttachmentBytes       int64 = 10 << 20
	WorkbenchMaxAttachmentTotalBytes  int64 = 24 << 20
)

var (
	ErrWorkbenchAttachmentNotFound = infraerrors.New(http.StatusNotFound, "WORKBENCH_ATTACHMENT_NOT_FOUND", "附件不存在或已失效")
	ErrWorkbenchAttachmentTooLarge = infraerrors.New(http.StatusBadRequest, "WORKBENCH_ATTACHMENT_TOO_LARGE", "单个附件不能超过 10 MB")
	ErrWorkbenchAttachmentType     = infraerrors.New(http.StatusBadRequest, "WORKBENCH_ATTACHMENT_TYPE_UNSUPPORTED", "不支持该附件类型")
	ErrWorkbenchAttachmentInvalid  = infraerrors.New(http.StatusBadRequest, "WORKBENCH_ATTACHMENT_INVALID", "附件内容与文件类型不匹配")
	ErrWorkbenchAttachmentQuota    = infraerrors.New(http.StatusTooManyRequests, "WORKBENCH_ATTACHMENT_QUOTA_EXCEEDED", "附件上传额度已用完，请稍后再试")
	ErrWorkbenchAttachmentState    = infraerrors.New(http.StatusConflict, "WORKBENCH_ATTACHMENT_STATE_INVALID", "附件尚未上传完成或已被使用")
)

type WorkbenchAttachmentUploadPlan struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers,omitempty"`
	RequiresAuth bool              `json:"requires_auth"`
}

type WorkbenchAttachmentUploadTicket struct {
	Attachment WorkbenchAttachment           `json:"attachment"`
	Upload     WorkbenchAttachmentUploadPlan `json:"upload"`
	ExpiresAt  time.Time                     `json:"expires_at"`
}

type WorkbenchAttachmentObjectMetadata struct {
	SizeBytes   int64
	ContentType string
	ETag        string
}

type WorkbenchAttachmentObjectStore interface {
	Provider() string
	PresignPut(ctx context.Context, key, contentType string, sizeBytes int64, expiry time.Duration) (WorkbenchAttachmentUploadPlan, bool, error)
	Put(ctx context.Context, key, contentType string, sizeBytes int64, body io.Reader) (int64, error)
	Head(ctx context.Context, key string) (*WorkbenchAttachmentObjectMetadata, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

func (s *WorkbenchService) CreateAttachmentUpload(ctx context.Context, userID int64, name, mimeType string, sizeBytes int64) (*WorkbenchAttachmentUploadTicket, error) {
	name = truncateRunes(strings.TrimSpace(name), 160)
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	if name == "" || sizeBytes <= 0 {
		return nil, ErrWorkbenchAttachmentInvalid
	}
	if sizeBytes > WorkbenchMaxAttachmentBytes {
		return nil, ErrWorkbenchAttachmentTooLarge
	}
	if !workbenchAttachmentMIMEAllowed(mimeType) {
		return nil, ErrWorkbenchAttachmentType
	}

	id := uuid.NewString()
	prefix := strings.Trim(strings.TrimSpace(s.cfg.Workbench.Attachments.Prefix), "/")
	if prefix == "" {
		prefix = "workbench-attachments"
	}
	objectKey := fmt.Sprintf("%s/%d/%s", prefix, userID, id)
	pendingTTL := time.Duration(s.cfg.Workbench.Attachments.PendingTTLSeconds) * time.Second
	expiresAt := time.Now().UTC().Add(pendingTTL)
	attachment := WorkbenchAttachment{
		ID: id, Name: name, MIMEType: mimeType, SizeBytes: sizeBytes,
		Provider: s.attachmentStore.Provider(), ObjectKey: objectKey, Status: "pending", ExpiresAt: &expiresAt,
	}
	reserved, err := s.repo.ReserveAttachment(ctx, userID, attachment, WorkbenchAttachmentLimits{
		DailyUploadBytes: s.cfg.Workbench.Attachments.DailyUploadLimitBytes,
		DailyUploadCount: s.cfg.Workbench.Attachments.DailyUploadLimitCount,
		MaxStoredBytes:   s.cfg.Workbench.Attachments.MaxStoredBytes,
		MaxStoredCount:   s.cfg.Workbench.Attachments.MaxStoredCount,
	})
	if err != nil {
		return nil, err
	}

	presignTTL := time.Duration(s.cfg.Workbench.Attachments.PresignTTLSeconds) * time.Second
	plan, direct, err := s.attachmentStore.PresignPut(ctx, objectKey, mimeType, sizeBytes, presignTTL)
	if err != nil {
		_, _ = s.repo.DeleteAttachment(ctx, userID, id)
		return nil, fmt.Errorf("presign workbench attachment: %w", err)
	}
	if !direct {
		plan = WorkbenchAttachmentUploadPlan{
			URL: "/workbench/attachments/" + id + "/content", Method: http.MethodPut,
			Headers: map[string]string{"Content-Type": mimeType}, RequiresAuth: true,
		}
	}
	return &WorkbenchAttachmentUploadTicket{Attachment: *reserved, Upload: plan, ExpiresAt: expiresAt}, nil
}

func (s *WorkbenchService) UploadAttachmentContent(ctx context.Context, userID int64, attachmentID string, contentLength int64, body io.Reader) (*WorkbenchAttachment, error) {
	if s.attachmentStore.Provider() != "local" {
		return nil, ErrWorkbenchAttachmentState
	}
	attachment, err := s.repo.GetAttachment(ctx, userID, attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment.Status != "pending" || attachment.ExpiresAt == nil || time.Now().After(*attachment.ExpiresAt) {
		return nil, ErrWorkbenchAttachmentState
	}
	if contentLength >= 0 && contentLength != attachment.SizeBytes {
		return nil, ErrWorkbenchAttachmentInvalid
	}
	limited := &io.LimitedReader{R: body, N: attachment.SizeBytes + 1}
	written, err := s.attachmentStore.Put(ctx, attachment.ObjectKey, attachment.MIMEType, attachment.SizeBytes, limited)
	if err != nil {
		return nil, fmt.Errorf("upload workbench attachment: %w", err)
	}
	if written != attachment.SizeBytes || limited.N == 0 {
		_ = s.attachmentStore.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
		return nil, ErrWorkbenchAttachmentInvalid
	}
	return s.CompleteAttachmentUpload(ctx, userID, attachmentID)
}

func (s *WorkbenchService) CompleteAttachmentUpload(ctx context.Context, userID int64, attachmentID string) (*WorkbenchAttachment, error) {
	attachment, err := s.repo.GetAttachment(ctx, userID, attachmentID)
	if err != nil {
		return nil, err
	}
	if attachment.Status == "ready" || attachment.Status == "attached" {
		return attachment, nil
	}
	if attachment.Status != "pending" || attachment.ExpiresAt == nil || time.Now().After(*attachment.ExpiresAt) {
		return nil, ErrWorkbenchAttachmentState
	}
	metadata, err := s.attachmentStore.Head(ctx, attachment.ObjectKey)
	if err != nil {
		return nil, ErrWorkbenchAttachmentNotFound
	}
	if metadata.SizeBytes != attachment.SizeBytes ||
		(metadata.ContentType != "" && !strings.EqualFold(strings.TrimSpace(metadata.ContentType), attachment.MIMEType)) {
		_ = s.attachmentStore.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
		return nil, ErrWorkbenchAttachmentInvalid
	}
	body, err := s.attachmentStore.Open(ctx, attachment.ObjectKey)
	if err != nil {
		return nil, ErrWorkbenchAttachmentNotFound
	}
	header := make([]byte, 512)
	n, readErr := io.ReadFull(body, header)
	_ = body.Close()
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		return nil, ErrWorkbenchAttachmentInvalid
	}
	if !workbenchAttachmentContentMatches(header[:n], attachment.MIMEType) {
		_ = s.attachmentStore.Delete(context.WithoutCancel(ctx), attachment.ObjectKey)
		return nil, ErrWorkbenchAttachmentInvalid
	}
	readyTTL := time.Duration(s.cfg.Workbench.Attachments.ReadyTTLSeconds) * time.Second
	return s.repo.MarkAttachmentReady(ctx, userID, attachmentID, metadata.ETag, time.Now().UTC().Add(readyTTL))
}

func (s *WorkbenchService) OpenAttachmentContent(ctx context.Context, userID int64, attachmentID string) (*WorkbenchAttachment, io.ReadCloser, error) {
	attachment, err := s.repo.GetAttachment(ctx, userID, attachmentID)
	if err != nil {
		return nil, nil, err
	}
	if attachment.Status != "ready" && attachment.Status != "attached" {
		return nil, nil, ErrWorkbenchAttachmentState
	}
	body, err := s.attachmentStore.Open(ctx, attachment.ObjectKey)
	if err != nil {
		return nil, nil, ErrWorkbenchAttachmentNotFound
	}
	return attachment, body, nil
}

func (s *WorkbenchService) DeleteAttachment(ctx context.Context, userID int64, attachmentID string) error {
	attachment, err := s.repo.GetAttachment(ctx, userID, attachmentID)
	if err != nil {
		return err
	}
	if attachment.Status == "attached" {
		return ErrWorkbenchAttachmentState
	}
	if err := s.attachmentStore.Delete(ctx, attachment.ObjectKey); err != nil {
		return fmt.Errorf("delete workbench attachment object: %w", err)
	}
	_, err = s.repo.DeleteAttachment(ctx, userID, attachmentID)
	return err
}

func workbenchAttachmentMIMEAllowed(mimeType string) bool {
	_, ok := workbenchAllowedAttachmentMIMEs[mimeType]
	return ok
}

var workbenchAllowedAttachmentMIMEs = map[string]struct{}{
	"image/png": {}, "image/jpeg": {}, "image/webp": {}, "image/gif": {},
	"application/pdf": {}, "application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/vnd.ms-excel": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.ms-powerpoint":                                             {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/json": {}, "application/xml": {}, "application/rtf": {},
	"text/plain": {}, "text/markdown": {}, "text/csv": {}, "text/html": {},
	"text/xml": {}, "text/yaml": {}, "application/x-yaml": {},
}

func workbenchAttachmentContentMatches(header []byte, mimeType string) bool {
	switch mimeType {
	case "image/png":
		return bytes.HasPrefix(header, []byte("\x89PNG\r\n\x1a\n"))
	case "image/jpeg":
		return len(header) >= 3 && header[0] == 0xff && header[1] == 0xd8 && header[2] == 0xff
	case "image/webp":
		return len(header) >= 12 && string(header[:4]) == "RIFF" && string(header[8:12]) == "WEBP"
	case "image/gif":
		return bytes.HasPrefix(header, []byte("GIF87a")) || bytes.HasPrefix(header, []byte("GIF89a"))
	case "application/pdf":
		return bytes.HasPrefix(bytes.TrimSpace(header), []byte("%PDF-"))
	case "application/msword", "application/vnd.ms-excel", "application/vnd.ms-powerpoint":
		return bytes.HasPrefix(header, []byte("\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1"))
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation":
		return bytes.HasPrefix(header, []byte("PK\x03\x04")) || bytes.HasPrefix(header, []byte("PK\x05\x06"))
	case "application/rtf":
		return bytes.HasPrefix(bytes.TrimSpace(header), []byte("{\\rtf"))
	case "application/json":
		trimmed := bytes.TrimSpace(header)
		return utf8.Valid(header) && (bytes.HasPrefix(trimmed, []byte("{")) || bytes.HasPrefix(trimmed, []byte("[")))
	case "application/xml", "text/xml":
		return utf8.Valid(header) && bytes.HasPrefix(bytes.TrimSpace(header), []byte("<"))
	default:
		detected := strings.SplitN(http.DetectContentType(header), ";", 2)[0]
		return utf8.Valid(header) && !bytes.ContainsRune(header, '\x00') && strings.HasPrefix(detected, "text/")
	}
}

func encodeWorkbenchAttachmentDataURL(mimeType string, data []byte) string {
	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
}
