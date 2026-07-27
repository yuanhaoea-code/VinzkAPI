package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const defaultImageGenerationRoot = "./data/image-generations"

var ErrImageGenerationNoAPIKey = ErrAPIKeyNotFound

type imageGenerationAPIKeySource interface {
	GetByID(ctx context.Context, id int64) (*APIKey, error)
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	List(ctx context.Context, userID int64, params pagination.PaginationParams, filters APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error)
}

type imageGenerationSelectAPIKey struct {
	ID      int64
	GroupID *int64
	Key     string
	Group   *Group
}

type imageGenerationQueue struct {
	queue chan func()
	wg    sync.WaitGroup
	mu    sync.RWMutex
	stop  bool
}

func newImageGenerationQueue(workerCount, queueSize int) *imageGenerationQueue {
	if workerCount <= 0 {
		workerCount = 1
	}
	if queueSize <= 0 {
		queueSize = 1
	}
	q := &imageGenerationQueue{queue: make(chan func(), queueSize)}
	q.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer q.wg.Done()
			for fn := range q.queue {
				func() {
					defer func() { _ = recover() }()
					if fn != nil {
						fn()
					}
				}()
			}
		}()
	}
	return q
}

func (q *imageGenerationQueue) TryEnqueue(task func()) error {
	if q == nil {
		return errors.New("image generation queue is nil")
	}
	if task == nil {
		return errors.New("image generation task is nil")
	}
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.stop {
		return errors.New("image generation queue stopped")
	}
	select {
	case q.queue <- task:
		return nil
	default:
		return errors.New("image generation queue full")
	}
}

func (q *imageGenerationQueue) Stop() {
	if q == nil {
		return
	}
	q.mu.Lock()
	if !q.stop {
		q.stop = true
		close(q.queue)
	}
	q.mu.Unlock()
	q.wg.Wait()
}

type ImageGenerationService struct {
	repo        ImageGenerationRepository
	apiKeyRepo  imageGenerationAPIKeySource
	cfg         *config.Config
	httpClient  *http.Client
	storageRoot string
	queue       *imageGenerationQueue
}

func NewImageGenerationService(repo ImageGenerationRepository, apiKeyService *APIKeyService, cfg *config.Config) *ImageGenerationService {
	var apiKeyRepo imageGenerationAPIKeySource
	if apiKeyService != nil {
		apiKeyRepo = apiKeyService
	}
	root := defaultImageGenerationRoot
	if cfg != nil && strings.TrimSpace(cfg.Pricing.DataDir) != "" {
		root = filepath.Join(cfg.Pricing.DataDir, "image-generations")
	}
	return &ImageGenerationService{
		repo:        repo,
		apiKeyRepo:  apiKeyRepo,
		cfg:         cfg,
		httpClient:  &http.Client{Timeout: 10 * time.Minute},
		storageRoot: root,
		queue:       newImageGenerationQueue(1, 128),
	}
}

func (s *ImageGenerationService) List(ctx context.Context, userID int64, params pagination.PaginationParams) ([]ImageGeneration, *pagination.PaginationResult, error) {
	if s.repo == nil {
		return nil, nil, ErrImageGenerationNotFound
	}
	return s.repo.ListByUserID(ctx, userID, params)
}

func (s *ImageGenerationService) GetOwned(ctx context.Context, userID, id int64) (*ImageGeneration, error) {
	if s.repo == nil {
		return nil, ErrImageGenerationNotFound
	}
	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if record == nil || record.UserID != userID {
		return nil, ErrImageGenerationNotFound
	}
	return record, nil
}

func (s *ImageGenerationService) Delete(ctx context.Context, userID, id int64) error {
	if s.repo == nil {
		return ErrImageGenerationNotFound
	}
	record, err := s.GetOwned(ctx, userID, id)
	if err != nil {
		return err
	}
	if record != nil {
		_ = os.RemoveAll(filepath.Join(s.storageRoot, strconv.FormatInt(userID, 10), strconv.FormatInt(id, 10)))
	}
	return s.repo.DeleteByUserID(ctx, userID, id)
}

func (s *ImageGenerationService) GetPromptVersions(ctx context.Context, userID, id int64) ([]ImageGenerationPromptVersion, error) {
	if s.repo == nil {
		return nil, ErrImageGenerationNotFound
	}
	return s.repo.ListPromptVersionsByUserID(ctx, userID, id)
}

func (s *ImageGenerationService) OpenImage(ctx context.Context, userID, id int64, index int) (io.ReadCloser, *ImageGenerationImage, error) {
	record, err := s.GetOwned(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if index < 0 || index >= len(record.Images) {
		return nil, nil, ErrImageGenerationNotFound
	}
	img := record.Images[index]
	if strings.TrimSpace(img.Path) == "" {
		return nil, nil, ErrImageGenerationNotFound
	}
	file, err := os.Open(img.Path)
	if err != nil {
		return nil, nil, err
	}
	return file, &img, nil
}

func (s *ImageGenerationService) EnqueueGenerate(ctx context.Context, input CreateImageGenerationInput) (*ImageGeneration, error) {
	resultCh := make(chan struct {
		record *ImageGeneration
		err    error
	}, 1)
	if s.queue == nil {
		return s.Generate(ctx, input)
	}
	if err := s.queue.TryEnqueue(func() {
		record, err := s.Generate(context.WithoutCancel(ctx), input)
		resultCh <- struct {
			record *ImageGeneration
			err    error
		}{record: record, err: err}
	}); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultCh:
		return res.record, res.err
	}
}

func (s *ImageGenerationService) Generate(ctx context.Context, input CreateImageGenerationInput) (*ImageGeneration, error) {
	normalized, err := normalizeImageGenerationInput(input)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	apiKey, err := s.selectAPIKey(ctx, normalized.UserID, normalized.APIKeyID, normalized.ResolutionTier)
	if err == ErrImageGenerationNoAPIKey && strings.TrimSpace(normalized.APIKey) != "" {
		apiKey, err = s.selectAPIKeyFromValue(ctx, normalized.UserID, normalized.APIKey, normalized.ResolutionTier)
	}
	if err != nil {
		return nil, err
	}

	version := ImageGenerationPromptVersion{
		Version:   1,
		Prompt:    normalized.Prompt,
		Model:     normalized.Model,
		Size:      normalized.Size,
		Quality:   normalized.Quality,
		OutputFmt: normalized.OutputFormat,
		CreatedAt: time.Now(),
	}

	requestID := newImageGenerationRequestID()
	endpoint := openAIImagesGenerationsEndpoint
	body := map[string]any{
		"model":           normalized.Model,
		"prompt":          normalized.Prompt,
		"size":            normalized.Size,
		"resolution_tier": normalized.ResolutionTier,
		"aspect_ratio":    normalized.AspectRatio,
		"n":               normalized.N,
		"quality":         normalized.Quality,
		"output_format":   normalized.OutputFormat,
	}
	if strings.TrimSpace(normalized.SourceImage) != "" {
		endpoint = openAIImagesEditsEndpoint
		body["images"] = []map[string]string{
			{"image_url": normalized.SourceImage},
		}
	}
	requestJSON, _ := json.Marshal(body)

	record, err := s.repo.Create(ctx, ImageGenerationCreateParams{
		UserID:         normalized.UserID,
		APIKeyID:       &apiKey.ID,
		GroupID:        apiKey.GroupID,
		RequestID:      requestID,
		Model:          normalized.Model,
		Prompt:         normalized.Prompt,
		SourceImage:    normalized.SourceImage,
		Size:           normalized.Size,
		ResolutionTier: normalized.ResolutionTier,
		AspectRatio:    normalized.AspectRatio,
		Quality:        normalized.Quality,
		OutputFormat:   normalized.OutputFormat,
		N:              normalized.N,
		RequestJSON:    requestJSON,
	})
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePromptVersion(ctx, record.ID, version); err != nil {
		return s.repo.UpdateResult(ctx, record.ID, ImageGenerationUpdateParams{
			Status:        ImageGenerationStatusFailed,
			StorageStatus: ImageStorageStatusFailed,
			ResponseJSON:  json.RawMessage(`{}`),
			ErrorMessage:  err.Error(),
			CompletedAt:   &now,
		})
	}
	respBody, statusCode, err := s.callLocalImagesGateway(ctx, apiKey.Key, requestID, endpoint, requestJSON)
	if err != nil {
		return s.repo.UpdateResult(ctx, record.ID, ImageGenerationUpdateParams{
			Status:        ImageGenerationStatusFailed,
			StorageStatus: ImageStorageStatusFailed,
			ResponseJSON:  json.RawMessage(`{}`),
			ErrorMessage:  friendlyImageGatewayTransportError(err),
			CompletedAt:   &now,
		})
	}
	if statusCode < 200 || statusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return s.repo.UpdateResult(ctx, record.ID, ImageGenerationUpdateParams{
			Status:        ImageGenerationStatusFailed,
			StorageStatus: ImageStorageStatusFailed,
			ResponseJSON:  json.RawMessage(respBody),
			ErrorMessage:  friendlyImageGatewayError(statusCode, msg),
			CompletedAt:   &now,
		})
	}

	images, fileSize, upstreamRequestID, err := s.persistImages(ctx, record.ID, normalized.UserID, respBody, normalized.OutputFormat)
	if err != nil {
		return s.repo.UpdateResult(ctx, record.ID, ImageGenerationUpdateParams{
			Status:            ImageGenerationStatusFailed,
			StorageStatus:     ImageStorageStatusFailed,
			ResponseJSON:      sanitizeImageGenerationResponse(respBody),
			FileSizeBytes:     fileSize,
			ImageCount:        len(images),
			Images:            images,
			ErrorMessage:      err.Error(),
			UpstreamRequestID: upstreamRequestID,
			CompletedAt:       &now,
		})
	}

	return s.repo.UpdateResult(ctx, record.ID, ImageGenerationUpdateParams{
		Status:            ImageGenerationStatusSuccess,
		StorageStatus:     ImageStorageStatusSaved,
		ResponseJSON:      sanitizeImageGenerationResponse(respBody),
		FileSizeBytes:     fileSize,
		ImageCount:        len(images),
		Images:            images,
		UpstreamRequestID: upstreamRequestID,
		CompletedAt:       &now,
	})
}

func (s *ImageGenerationService) GenerateQueued(ctx context.Context, input CreateImageGenerationInput) (*ImageGeneration, error) {
	if s.queue == nil {
		return s.Generate(ctx, input)
	}
	return s.EnqueueGenerate(ctx, input)
}

func (s *ImageGenerationService) selectAPIKey(ctx context.Context, userID int64, apiKeyID *int64, tier string) (*imageGenerationSelectAPIKey, error) {
	if s.apiKeyRepo == nil {
		return nil, ErrImageGenerationNoAPIKey
	}
	if apiKeyID != nil && *apiKeyID > 0 {
		apiKey, err := s.apiKeyRepo.GetByID(ctx, *apiKeyID)
		if err != nil {
			return nil, err
		}
		if apiKey == nil || apiKey.UserID != userID || !apiKey.IsActive() {
			return nil, ErrImageGenerationNoAPIKey
		}
		if !GroupAllowsImageTier(apiKey.Group, tier) {
			return nil, ErrImageGenerationTierNotAllowed
		}
		return &imageGenerationSelectAPIKey{ID: apiKey.ID, GroupID: apiKey.GroupID, Key: apiKey.Key, Group: apiKey.Group}, nil
	}

	keys, _, err := s.apiKeyRepo.List(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 100}, APIKeyListFilters{Status: StatusAPIKeyActive})
	if err != nil {
		return nil, err
	}
	for i := range keys {
		key := keys[i]
		if !key.IsActive() {
			continue
		}
		if !GroupAllowsImageTier(key.Group, tier) {
			continue
		}
		if strings.TrimSpace(key.Key) != "" {
			return &imageGenerationSelectAPIKey{ID: key.ID, GroupID: key.GroupID, Key: key.Key, Group: key.Group}, nil
		}
	}
	return nil, ErrImageGenerationNoAPIKey
}

func (s *ImageGenerationService) selectAPIKeyFromValue(ctx context.Context, userID int64, raw string, tier string) (*imageGenerationSelectAPIKey, error) {
	if s.apiKeyRepo == nil {
		return nil, ErrImageGenerationNoAPIKey
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, ErrImageGenerationNoAPIKey
	}
	apiKey, err := s.apiKeyRepo.GetByKey(ctx, raw)
	if err != nil {
		return nil, err
	}
	if apiKey == nil || apiKey.UserID != userID || !apiKey.IsActive() {
		return nil, ErrImageGenerationNoAPIKey
	}
	if !GroupAllowsImageTier(apiKey.Group, tier) {
		return nil, ErrImageGenerationTierNotAllowed
	}
	return &imageGenerationSelectAPIKey{ID: apiKey.ID, GroupID: apiKey.GroupID, Key: apiKey.Key, Group: apiKey.Group}, nil
}

func (s *ImageGenerationService) callLocalImagesGateway(ctx context.Context, key, requestID, endpoint string, body []byte) ([]byte, int, error) {
	port := 8080
	if s.cfg != nil && s.cfg.Server.Port > 0 {
		port = s.cfg.Server.Port
	}
	backendURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, openAIImagesGenerationsEndpoint)
	if strings.TrimSpace(endpoint) == openAIImagesEditsEndpoint {
		backendURL = fmt.Sprintf("http://127.0.0.1:%d%s", port, openAIImagesEditsEndpoint)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backendURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(key))
	req.Header.Set("X-Request-ID", strings.TrimSpace(requestID))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func (s *ImageGenerationService) persistImages(ctx context.Context, recordID, userID int64, respBody []byte, defaultOutputFormat string) ([]ImageGenerationImage, int64, string, error) {
	if len(respBody) == 0 {
		return nil, 0, "", fmt.Errorf("empty upstream image response")
	}
	mimeType, outFmt, images, upstreamRequestID, err := parseImageGenerationResponse(respBody, defaultOutputFormat)
	if err != nil {
		return nil, 0, upstreamRequestID, err
	}
	if len(images) == 0 {
		return nil, 0, upstreamRequestID, fmt.Errorf("no images returned")
	}

	stored := make([]ImageGenerationImage, 0, len(images))
	now := time.Now()
	for idx, img := range images {
		data, resolvedMimeType, err := s.resolveImagePayload(ctx, img, mimeType)
		if err != nil {
			return stored, totalBytes(stored), upstreamRequestID, err
		}
		root := filepath.Join(s.storageRoot, strconv.FormatInt(userID, 10), strconv.FormatInt(recordID, 10))
		if err := os.MkdirAll(root, 0755); err != nil {
			return stored, totalBytes(stored), upstreamRequestID, err
		}
		filename := fmt.Sprintf("%d.%s", idx, outputExt(outFmt))
		filePath := filepath.Join(root, filename)
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return stored, totalBytes(stored), upstreamRequestID, err
		}
		stored = append(stored, ImageGenerationImage{
			Index:     idx,
			Path:      filePath,
			URL:       buildStoredImageURL(recordID, idx),
			MimeType:  resolvedMimeType,
			SizeBytes: int64(len(data)),
		})
		_ = now
	}
	return stored, totalBytes(stored), upstreamRequestID, nil
}

func (s *ImageGenerationService) resolveImagePayload(ctx context.Context, payload, fallbackMimeType string) ([]byte, string, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, "", fmt.Errorf("empty image payload")
	}

	if strings.HasPrefix(payload, "data:") {
		mt, b64, ok := splitDataURL(payload)
		if !ok {
			return nil, "", fmt.Errorf("invalid data url")
		}
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, "", err
		}
		return data, mt, nil
	}

	if isHTTPOrHTTPSURL(payload) {
		data, contentType, err := s.downloadImageBytes(ctx, payload)
		if err != nil {
			return nil, "", err
		}
		if strings.TrimSpace(contentType) == "" {
			contentType = fallbackMimeType
		}
		return data, contentType, nil
	}

	data, err := decodeImagePayload(payload, fallbackMimeType)
	if err != nil {
		return nil, "", err
	}
	if strings.TrimSpace(fallbackMimeType) == "" {
		fallbackMimeType = "image/png"
	}
	return data, fallbackMimeType, nil
}

func (s *ImageGenerationService) downloadImageBytes(ctx context.Context, rawURL string) ([]byte, string, error) {
	client := s.httpClient
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "image/*,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = resp.Status
		}
		return nil, "", fmt.Errorf("download image failed: %s", msg)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, openAIImageMaxDownloadBytes))
	if err != nil {
		return nil, "", err
	}
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if semi := strings.IndexByte(contentType, ';'); semi >= 0 {
		contentType = strings.TrimSpace(contentType[:semi])
	}
	if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		contentType = ""
	}
	return data, contentType, nil
}

func parseImageGenerationResponse(body []byte, defaultOutputFormat string) (string, string, []string, string, error) {
	if !json.Valid(body) {
		return "", "", nil, "", fmt.Errorf("invalid image response")
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", "", nil, "", err
	}
	upstreamRequestID, _ := raw["request_id"].(string)
	if upstreamRequestID == "" {
		upstreamRequestID, _ = raw["id"].(string)
	}

	var items []any
	switch v := raw["data"].(type) {
	case []any:
		items = v
	default:
		if out, ok := raw["output"].([]any); ok {
			items = out
		}
	}
	if len(items) == 0 {
		if out, ok := raw["output"].([]interface{}); ok {
			items = make([]any, 0, len(out))
			for _, item := range out {
				items = append(items, item)
			}
		}
	}

	mimeType := "image/png"
	outFmt := strings.ToLower(strings.TrimSpace(defaultOutputFormat))
	if outFmt == "" {
		outFmt = "png"
	}

	images := make([]string, 0, len(items))
	for _, item := range items {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		if s := strings.TrimSpace(firstNonEmptyString(m["revised_prompt"], m["prompt"])); s != "" {
			_ = s
		}
		if format := strings.TrimSpace(firstNonEmptyString(m["output_format"], m["format"])); format != "" {
			outFmt = strings.ToLower(format)
			mimeType = outputMimeType(outFmt)
		}
		if result := strings.TrimSpace(firstNonEmptyString(m["b64_json"], m["result"], m["download_url"], m["image_url"], m["url"])); result != "" {
			if strings.HasPrefix(result, "data:") {
				if mt, payload, ok := splitDataURL(result); ok {
					mimeType = mt
					images = append(images, payload)
					continue
				}
			}
			images = append(images, result)
		}
	}

	if len(images) == 0 {
		return mimeType, outFmt, nil, upstreamRequestID, nil
	}
	return mimeType, outFmt, images, upstreamRequestID, nil
}

func decodeImagePayload(payload, mimeType string) ([]byte, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, fmt.Errorf("empty image payload")
	}
	if strings.HasPrefix(payload, "data:") {
		_, b64, ok := splitDataURL(payload)
		if !ok {
			return nil, fmt.Errorf("invalid data url")
		}
		return base64.StdEncoding.DecodeString(b64)
	}
	return base64.StdEncoding.DecodeString(payload)
}

func isHTTPOrHTTPSURL(raw string) bool {
	raw = strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://")
}

func splitDataURL(raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "data:") {
		return "", "", false
	}
	comma := strings.IndexByte(raw, ',')
	if comma <= 0 {
		return "", "", false
	}
	meta := raw[len("data:"):comma]
	payload := raw[comma+1:]
	if payload == "" {
		return "", "", false
	}
	mimeType := strings.TrimSpace(meta)
	if semi := strings.IndexByte(mimeType, ';'); semi >= 0 {
		mimeType = mimeType[:semi]
	}
	if mimeType == "" {
		mimeType = "image/png"
	}
	return mimeType, payload, true
}

func outputMimeType(outputFormat string) string {
	switch strings.ToLower(strings.TrimSpace(outputFormat)) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func outputExt(outputFormat string) string {
	switch strings.ToLower(strings.TrimSpace(outputFormat)) {
	case "jpg", "jpeg":
		return "jpg"
	case "webp":
		return "webp"
	default:
		return "png"
	}
}

func sanitizeImageGenerationResponse(body []byte) json.RawMessage {
	if len(body) == 0 {
		return json.RawMessage(`{}`)
	}
	if json.Valid(body) {
		return append(json.RawMessage(nil), body...)
	}
	return json.RawMessage(`{}`)
}

func friendlyImageGatewayError(statusCode int, rawMessage string) string {
	msg := strings.TrimSpace(rawMessage)
	lower := strings.ToLower(msg)
	upper := strings.ToUpper(msg)

	switch {
	case strings.Contains(lower, "resolved ip") && strings.Contains(lower, "not allowed"):
		return "生图上游域名被本机代理解析到了保留地址，安全校验已拦截请求。请把该上游域名加入代理的 fake-ip-filter，或改用能返回真实公网 IP 的 DNS 后重试。"
	case strings.Contains(lower, "no available compatible accounts") || strings.Contains(lower, "no available accounts"):
		return "当前生图分组没有可用上游账号，请检查渠道管理里的生图账号是否已加入该分组，并确认账号状态为启用。"
	case strings.Contains(upper, "INSUFFICIENT_BALANCE") || strings.Contains(lower, "insufficient account balance") || strings.Contains(lower, "insufficient balance"):
		return "账户余额不足，请充值后再生成图片。"
	case statusCode == http.StatusUnauthorized || strings.Contains(upper, "INVALID_API_KEY") || strings.Contains(lower, "unauthorized"):
		return "API Key 无效或已失效，请检查你选择的站内 Key 是否正确、是否启用。"
	case statusCode == http.StatusTooManyRequests || strings.Contains(lower, "rate limit"):
		return "生图请求触发上游限流，请稍后再试，或更换生图号池账号。"
	case statusCode == http.StatusGatewayTimeout || strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline exceeded"):
		return "图片生成超时了，请稍后重试；如果提示词复杂，生成时间可能会更久。"
	case isImageGatewayEOFMessage(lower):
		return "生图上游连接在返回结果前中断（EOF）。请先检查上游记录；确认没有创建生成任务后再重试。"
	case statusCode >= 500:
		return "生图上游暂时不可用或没有返回可用结果，请稍后重试；如果持续出现，请检查生图号池账号和上游服务状态。"
	case msg != "":
		return fmt.Sprintf("生图失败：%s", msg)
	default:
		return fmt.Sprintf("生图失败，上游返回 HTTP %d。", statusCode)
	}
}

func friendlyImageGatewayTransportError(err error) string {
	if err == nil {
		return "生图网关请求失败，请稍后重试。"
	}
	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	if strings.Contains(lower, "resolved ip") && strings.Contains(lower, "not allowed") {
		return "生图上游域名被本机代理解析到了保留地址，安全校验已拦截请求。请把该上游域名加入代理的 fake-ip-filter，或改用能返回真实公网 IP 的 DNS 后重试。"
	}
	if isImageGatewayEOFMessage(lower) {
		return "生图上游连接在返回结果前中断（EOF）。请先检查上游记录；确认没有创建生成任务后再重试。"
	}
	return message
}

func isImageGatewayEOFMessage(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	return lower == "eof" ||
		strings.Contains(lower, "unexpected eof") ||
		strings.HasSuffix(lower, ": eof") ||
		strings.Contains(lower, ": eof\"")
}

func newImageGenerationRequestID() string {
	return fmt.Sprintf("img_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1_000_000)
}

func buildStoredImageURL(recordID int64, index int) string {
	return fmt.Sprintf("/api/v1/image-generations/%d/images/%d/preview", recordID, index)
}

func totalBytes(images []ImageGenerationImage) int64 {
	var total int64
	for _, img := range images {
		total += img.SizeBytes
	}
	return total
}

func normalizeImageGenerationInput(input CreateImageGenerationInput) (CreateImageGenerationInput, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.Model = strings.TrimSpace(input.Model)
	input.Size = strings.TrimSpace(input.Size)
	input.ResolutionTier = strings.TrimSpace(input.ResolutionTier)
	input.AspectRatio = strings.TrimSpace(input.AspectRatio)
	input.Quality = strings.TrimSpace(input.Quality)
	input.OutputFormat = strings.TrimSpace(input.OutputFormat)
	input.APIKey = strings.TrimSpace(input.APIKey)
	input.SourceImage = strings.TrimSpace(input.SourceImage)
	if input.Prompt == "" {
		return input, fmt.Errorf("prompt is required")
	}
	if input.Model == "" {
		input.Model = "gpt-image-2"
	}
	hasExplicitProductSpec := input.ResolutionTier != "" || input.AspectRatio != ""
	if input.ResolutionTier == "" {
		if tier, ok := ClassifyImageBillingTier(input.Size); ok {
			input.ResolutionTier = tier
		} else {
			input.ResolutionTier = ImageBillingSize1K
		}
	} else {
		input.ResolutionTier = NormalizeImageBillingTier(input.ResolutionTier)
		if input.ResolutionTier == "" {
			return input, fmt.Errorf("resolution_tier must be one of 1K, 2K, 4K")
		}
	}
	if input.AspectRatio == "" {
		input.AspectRatio = InferImageAspectRatio(input.Size)
	} else {
		input.AspectRatio = NormalizeImageAspectRatio(input.AspectRatio)
		if input.AspectRatio == "" {
			return input, fmt.Errorf("aspect_ratio must be one of 1:1, 2:3, 3:2")
		}
	}
	if hasExplicitProductSpec || input.Size == "" {
		resolvedSize, ok := ResolveImageGenerationSize(input.ResolutionTier, input.AspectRatio)
		if !ok {
			return input, fmt.Errorf("unsupported image resolution and aspect ratio")
		}
		input.Size = resolvedSize
	}
	if input.Quality == "" {
		input.Quality = "auto"
	}
	if input.OutputFormat == "" {
		input.OutputFormat = "png"
	}
	if input.N <= 0 {
		input.N = 1
	}
	if input.N > 4 {
		return input, fmt.Errorf("n must not exceed 4")
	}
	return input, nil
}
