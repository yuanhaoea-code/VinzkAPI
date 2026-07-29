package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type workbenchAttachmentCleanupRepoStub struct {
	items   []WorkbenchAttachment
	deleted []string
}

func (r *workbenchAttachmentCleanupRepoStub) ListExpiredAttachments(_ context.Context, provider string, limit int) ([]WorkbenchAttachment, error) {
	items := make([]WorkbenchAttachment, 0, len(r.items))
	for _, item := range r.items {
		if item.Provider == provider && len(items) < limit {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *workbenchAttachmentCleanupRepoStub) DeleteExpiredAttachment(_ context.Context, attachmentID string) error {
	r.deleted = append(r.deleted, attachmentID)
	return nil
}

type workbenchAttachmentCleanupStoreStub struct {
	deleteErrorFor string
	deleted        []string
}

func (s *workbenchAttachmentCleanupStoreStub) Provider() string { return "s3" }
func (s *workbenchAttachmentCleanupStoreStub) PresignPut(context.Context, string, string, int64, time.Duration) (WorkbenchAttachmentUploadPlan, bool, error) {
	return WorkbenchAttachmentUploadPlan{}, false, nil
}
func (s *workbenchAttachmentCleanupStoreStub) Put(context.Context, string, string, int64, io.Reader) (int64, error) {
	return 0, nil
}
func (s *workbenchAttachmentCleanupStoreStub) Head(context.Context, string) (*WorkbenchAttachmentObjectMetadata, error) {
	return nil, nil
}
func (s *workbenchAttachmentCleanupStoreStub) Open(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}
func (s *workbenchAttachmentCleanupStoreStub) Delete(_ context.Context, key string) error {
	s.deleted = append(s.deleted, key)
	if key == s.deleteErrorFor {
		return errors.New("delete failed")
	}
	return nil
}

func TestWorkbenchAttachmentCleanupDeletesRowOnlyAfterObject(t *testing.T) {
	repo := &workbenchAttachmentCleanupRepoStub{items: []WorkbenchAttachment{
		{ID: "delete", Provider: "s3", ObjectKey: "objects/delete"},
		{ID: "retry", Provider: "s3", ObjectKey: "objects/retry"},
		{ID: "other-provider", Provider: "local", ObjectKey: "objects/local"},
	}}
	store := &workbenchAttachmentCleanupStoreStub{deleteErrorFor: "objects/retry"}
	service := NewWorkbenchAttachmentCleanupService(repo, store, nil)

	service.cleanupOnce()

	require.Equal(t, []string{"objects/delete", "objects/retry"}, store.deleted)
	require.Equal(t, []string{"delete"}, repo.deleted)
}
