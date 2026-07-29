package repository

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLocalWorkbenchAttachmentStoreLifecycle(t *testing.T) {
	store := &localWorkbenchAttachmentStore{root: t.TempDir()}
	ctx := context.Background()
	key := "workbench-attachments/42/attachment-id"

	written, err := store.Put(ctx, key, "text/plain", 5, strings.NewReader("hello"))
	require.NoError(t, err)
	require.EqualValues(t, 5, written)

	metadata, err := store.Head(ctx, key)
	require.NoError(t, err)
	require.EqualValues(t, 5, metadata.SizeBytes)

	body, err := store.Open(ctx, key)
	require.NoError(t, err)
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())
	require.Equal(t, "hello", string(data))

	require.NoError(t, store.Delete(ctx, key))
	_, err = store.Head(ctx, key)
	require.Error(t, err)
}

func TestLocalWorkbenchAttachmentStoreRejectsTraversal(t *testing.T) {
	store := &localWorkbenchAttachmentStore{root: t.TempDir()}
	_, err := store.path("../../outside")
	require.Error(t, err)
}

func TestS3WorkbenchAttachmentStorePresignsFixedObject(t *testing.T) {
	store, err := newS3WorkbenchAttachmentStore(context.Background(), config.WorkbenchAttachmentConfig{
		Endpoint: "https://s3.example.com", Region: "test-region", Bucket: "private-bucket",
		AccessKeyID: "test-access", SecretAccessKey: "test-secret", ForcePathStyle: true,
	})
	require.NoError(t, err)

	plan, direct, err := store.PresignPut(
		context.Background(), "workbench-attachments/42/attachment-id", "application/pdf", 1024, 5*time.Minute,
	)
	require.NoError(t, err)
	require.True(t, direct)
	require.Equal(t, "PUT", plan.Method)
	require.Contains(t, plan.URL, "private-bucket/workbench-attachments/42/attachment-id")
	require.NotContains(t, plan.URL, "test-secret")
}
