package service

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

type WorkbenchAttachmentCleanupRepository interface {
	ListExpiredAttachments(ctx context.Context, provider string, limit int) ([]WorkbenchAttachment, error)
	DeleteExpiredAttachment(ctx context.Context, attachmentID string) error
}

type WorkbenchAttachmentCleanupService struct {
	repo     WorkbenchAttachmentCleanupRepository
	store    WorkbenchAttachmentObjectStore
	interval time.Duration
	batch    int

	startOnce sync.Once
	stopOnce  sync.Once
	stopCh    chan struct{}
}

func NewWorkbenchAttachmentCleanupService(
	repo WorkbenchAttachmentCleanupRepository,
	store WorkbenchAttachmentObjectStore,
	cfg *config.Config,
) *WorkbenchAttachmentCleanupService {
	interval := 5 * time.Minute
	batch := 200
	if cfg != nil {
		if cfg.Workbench.Attachments.CleanupIntervalSeconds > 0 {
			interval = time.Duration(cfg.Workbench.Attachments.CleanupIntervalSeconds) * time.Second
		}
		if cfg.Workbench.Attachments.CleanupBatchSize > 0 {
			batch = cfg.Workbench.Attachments.CleanupBatchSize
		}
	}
	return &WorkbenchAttachmentCleanupService{
		repo: repo, store: store, interval: interval, batch: batch, stopCh: make(chan struct{}),
	}
}

func ProvideWorkbenchAttachmentCleanupService(
	repo WorkbenchAttachmentCleanupRepository,
	store WorkbenchAttachmentObjectStore,
	cfg *config.Config,
) *WorkbenchAttachmentCleanupService {
	service := NewWorkbenchAttachmentCleanupService(repo, store, cfg)
	service.Start()
	return service
}

func (s *WorkbenchAttachmentCleanupService) Start() {
	if s == nil || s.repo == nil || s.store == nil {
		return
	}
	s.startOnce.Do(func() {
		logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] started interval=%s batch=%d", s.interval, s.batch)
		go s.runLoop()
	})
}

func (s *WorkbenchAttachmentCleanupService) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		close(s.stopCh)
		logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] stopped")
	})
}

func (s *WorkbenchAttachmentCleanupService) runLoop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.cleanupOnce()
	for {
		select {
		case <-ticker.C:
			s.cleanupOnce()
		case <-s.stopCh:
			return
		}
	}
}

func (s *WorkbenchAttachmentCleanupService) cleanupOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	attachments, err := s.repo.ListExpiredAttachments(ctx, s.store.Provider(), s.batch)
	if err != nil {
		logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] list failed err=%v", err)
		return
	}
	deleted := 0
	for _, attachment := range attachments {
		if err := s.store.Delete(ctx, attachment.ObjectKey); err != nil {
			logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] object delete failed attachment_id=%s err=%v", attachment.ID, err)
			continue
		}
		if err := s.repo.DeleteExpiredAttachment(ctx, attachment.ID); err != nil {
			logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] row delete failed attachment_id=%s err=%v", attachment.ID, err)
			continue
		}
		deleted++
	}
	if deleted > 0 {
		logger.LegacyPrintf("service.workbench_attachment_cleanup", "[WorkbenchAttachmentCleanup] cleaned expired attachments count=%d", deleted)
	}
}
