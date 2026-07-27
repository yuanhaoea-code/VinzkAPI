package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/google/uuid"
)

const (
	WorkbenchReasoningFast     = "fast"
	WorkbenchReasoningStandard = "standard"
	WorkbenchReasoningDeep     = "deep"

	WorkbenchMessagePending    = "pending"
	WorkbenchMessageInProgress = "in_progress"
	WorkbenchMessageCompleted  = "completed"
	WorkbenchMessageFailed     = "failed"
	WorkbenchMessageCanceled   = "canceled"

	workbenchMaxMessageRunes          = 100000
	workbenchMaxHistoryRunes          = 60000
	workbenchMaxHistoryItems          = 32
	workbenchMaxAttachmentsPerMessage = 8
	workbenchMaxAttachmentBytes       = 10 << 20
	workbenchMaxAttachmentTotalBytes  = 24 << 20
)

var (
	ErrWorkbenchConversationNotFound = infraerrors.New(http.StatusNotFound, "WORKBENCH_CONVERSATION_NOT_FOUND", "Conversation not found")
	ErrWorkbenchMessageNotFound      = infraerrors.New(http.StatusNotFound, "WORKBENCH_MESSAGE_NOT_FOUND", "Message not found")
	ErrWorkbenchModelNotFound        = infraerrors.New(http.StatusNotFound, "WORKBENCH_MODEL_NOT_FOUND", "Model is not available for this API key")
	ErrWorkbenchBindingNotFound      = infraerrors.New(http.StatusNotFound, "WORKBENCH_BINDING_NOT_FOUND", "Model binding not found")
	ErrWorkbenchInvalidInput         = infraerrors.New(http.StatusBadRequest, "WORKBENCH_INVALID_INPUT", "Invalid workbench request")
	ErrWorkbenchGenerationStarted    = infraerrors.New(http.StatusConflict, "WORKBENCH_GENERATION_STARTED", "Generation has already started")
	ErrWorkbenchGenerationCanceled   = errors.New("workbench generation canceled")
)

type WorkbenchModelBinding struct {
	ID          string    `json:"id"`
	UserID      int64     `json:"-"`
	APIKeyID    int64     `json:"api_key_id"`
	ModelID     string    `json:"model_id"`
	DisplayName string    `json:"display_name"`
	Provider    string    `json:"provider"`
	Source      string    `json:"source"`
	Hidden      bool      `json:"hidden"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkbenchEligibleKey struct {
	APIKeyID  int64   `json:"api_key_id"`
	KeyName   string  `json:"key_name"`
	GroupID   int64   `json:"group_id"`
	GroupName string  `json:"group_name"`
	Platform  string  `json:"platform"`
	Rate      float64 `json:"rate"`
}

type WorkbenchKeyOption struct {
	APIKeyID         int64                  `json:"api_key_id"`
	KeyName          string                 `json:"key_name"`
	GroupID          int64                  `json:"group_id"`
	GroupName        string                 `json:"group_name"`
	Platform         string                 `json:"platform"`
	ProviderLabel    string                 `json:"provider_label"`
	Rate             float64                `json:"rate"`
	DefaultModelID   string                 `json:"default_model_id"`
	DefaultModelName string                 `json:"default_model_name"`
	Models           []WorkbenchModelOption `json:"models"`
	AvailableModels  []WorkbenchModelOption `json:"available_models"`
}

type WorkbenchModelOption struct {
	ID                string                 `json:"id"`
	DisplayName       string                 `json:"display_name"`
	Provider          string                 `json:"provider"`
	ProviderLabel     string                 `json:"provider_label"`
	DefaultVisible    bool                   `json:"default_visible"`
	Added             bool                   `json:"added"`
	Available         bool                   `json:"available"`
	UnavailableReason string                 `json:"unavailable_reason,omitempty"`
	BindingID         string                 `json:"binding_id,omitempty"`
	APIKeyID          int64                  `json:"api_key_id,omitempty"`
	KeyName           string                 `json:"key_name,omitempty"`
	GroupName         string                 `json:"group_name,omitempty"`
	SortOrder         int                    `json:"sort_order"`
	ReasoningPresets  []string               `json:"reasoning_presets"`
	EligibleKeys      []WorkbenchEligibleKey `json:"eligible_keys,omitempty"`
}

type WorkbenchModels struct {
	Keys      []WorkbenchKeyOption   `json:"keys"`
	Selected  []WorkbenchModelOption `json:"selected"`
	Available []WorkbenchModelOption `json:"available"`
}

type WorkbenchConversation struct {
	ID              string             `json:"id"`
	UserID          int64              `json:"-"`
	ModelBindingID  *string            `json:"model_binding_id,omitempty"`
	Title           string             `json:"title"`
	ReasoningPreset string             `json:"reasoning_preset"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	Messages        []WorkbenchMessage `json:"messages,omitempty"`
}

type WorkbenchMessage struct {
	ID               string                `json:"id"`
	ConversationID   string                `json:"conversation_id"`
	Sequence         int64                 `json:"sequence"`
	Role             string                `json:"role"`
	Content          string                `json:"content"`
	Status           string                `json:"status"`
	ModelBindingID   *string               `json:"model_binding_id,omitempty"`
	ReasoningPreset  string                `json:"reasoning_preset,omitempty"`
	ReasoningSummary string                `json:"reasoning_summary,omitempty"`
	ErrorMessage     string                `json:"error_message,omitempty"`
	RequestID        string                `json:"request_id,omitempty"`
	Attachments      []WorkbenchAttachment `json:"attachments,omitempty"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
}

type WorkbenchAttachment struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	MIMEType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	DataURL   string `json:"data_url"`
}

type WorkbenchTurn struct {
	Conversation     *WorkbenchConversation `json:"conversation"`
	UserMessage      *WorkbenchMessage      `json:"user_message"`
	AssistantMessage *WorkbenchMessage      `json:"assistant_message"`
}

type WorkbenchConversationPatch struct {
	Title           *string
	ModelBindingID  *string
	ReasoningPreset *string
}

type WorkbenchRepository interface {
	ListModelBindings(ctx context.Context, userID int64) ([]WorkbenchModelBinding, error)
	GetModelBinding(ctx context.Context, userID int64, bindingID string) (*WorkbenchModelBinding, error)
	EnsureModelBinding(ctx context.Context, binding WorkbenchModelBinding) (*WorkbenchModelBinding, error)
	UpsertModelBinding(ctx context.Context, binding WorkbenchModelBinding) (*WorkbenchModelBinding, error)
	HideModelBinding(ctx context.Context, userID int64, bindingID string) error

	CreateConversation(ctx context.Context, conversation WorkbenchConversation) (*WorkbenchConversation, error)
	ListConversations(ctx context.Context, userID int64, limit int) ([]WorkbenchConversation, error)
	GetConversation(ctx context.Context, userID int64, conversationID string) (*WorkbenchConversation, error)
	UpdateConversation(ctx context.Context, userID int64, conversationID string, patch WorkbenchConversationPatch) (*WorkbenchConversation, error)
	DeleteConversation(ctx context.Context, userID int64, conversationID string) error
	ListMessages(ctx context.Context, userID int64, conversationID string) ([]WorkbenchMessage, error)
	GetMessage(ctx context.Context, userID int64, messageID string) (*WorkbenchMessage, error)
	CreateTurn(ctx context.Context, userID int64, conversationID, content, title string, attachments []WorkbenchAttachment, modelBindingID *string, reasoningPreset string) (*WorkbenchTurn, error)
	ClaimAssistantMessage(ctx context.Context, userID int64, messageID, requestID string) error
	FinishAssistantMessage(ctx context.Context, userID int64, messageID, status, content, reasoningSummary, errorMessage string) error
	CancelAssistantMessage(ctx context.Context, userID int64, messageID string) error
}

type workbenchDefaultModel struct {
	ID          string
	DisplayName string
	SortOrder   int
}

var workbenchDefaultModels = []workbenchDefaultModel{
	{ID: "gpt-5.6-sol", DisplayName: "GPT-5.6 Sol", SortOrder: 10},
	{ID: "gpt-5.6-terra", DisplayName: "GPT-5.6 Terra", SortOrder: 20},
	{ID: "gpt-5.6-luna", DisplayName: "GPT-5.6 Luna", SortOrder: 30},
	{ID: "gpt-5.5", DisplayName: "GPT-5.5", SortOrder: 40},
	{ID: "gpt-5.4", DisplayName: "GPT-5.4", SortOrder: 50},
	{ID: "gpt-5.4-mini", DisplayName: "GPT-5.4 Mini", SortOrder: 60},
	{ID: "gpt-5.3", DisplayName: "GPT-5.3", SortOrder: 70},
	{ID: "claude-sonnet-5", DisplayName: "Claude Sonnet 5", SortOrder: 10},
	{ID: "claude-fable-5", DisplayName: "Claude Fable 5", SortOrder: 20},
	{ID: "claude-opus-4-8", DisplayName: "Claude Opus 4.8", SortOrder: 30},
	{ID: "claude-opus-4-7", DisplayName: "Claude Opus 4.7", SortOrder: 40},
	{ID: "claude-sonnet-4-6", DisplayName: "Claude Sonnet 4.6", SortOrder: 50},
	{ID: "claude-opus-4-6", DisplayName: "Claude Opus 4.6", SortOrder: 60},
	{ID: "claude-haiku-4-5-20251001", DisplayName: "Claude Haiku 4.5", SortOrder: 70},
	{ID: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku", SortOrder: 80},
	{ID: "gemini-3.5-flash", DisplayName: "Gemini 3.5 Flash", SortOrder: 10},
	{ID: "gemini-3.1-pro-preview", DisplayName: "Gemini 3.1 Pro Preview", SortOrder: 20},
	{ID: "gemini-3-pro-preview", DisplayName: "Gemini 3 Pro Preview", SortOrder: 30},
	{ID: "gemini-3-flash-preview", DisplayName: "Gemini 3 Flash Preview", SortOrder: 40},
	{ID: "gemini-2.5-pro", DisplayName: "Gemini 2.5 Pro", SortOrder: 50},
	{ID: "gemini-2.5-flash", DisplayName: "Gemini 2.5 Flash", SortOrder: 60},
	{ID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash", SortOrder: 70},
}

var workbenchPickerModelsByProvider = map[string][]string{
	PlatformOpenAI: {
		"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "gpt-5.5",
		"gpt-5.4", "gpt-5.4-mini", "gpt-5.3",
	},
	PlatformAnthropic: {
		"claude-sonnet-5", "claude-fable-5", "claude-opus-4-8", "claude-opus-4-7",
		"claude-sonnet-4-6", "claude-opus-4-6", "claude-haiku-4-5-20251001", "claude-3-5-haiku-20241022",
	},
	PlatformGemini: {
		"gemini-3.5-flash", "gemini-3.1-pro-preview", "gemini-3-pro-preview", "gemini-3-flash-preview",
		"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash",
	},
}

type WorkbenchService struct {
	repo           WorkbenchRepository
	apiKeyService  *APIKeyService
	gatewayService *GatewayService
	cfg            *config.Config
	httpClient     *http.Client
}

func NewWorkbenchService(repo WorkbenchRepository, apiKeyService *APIKeyService, gatewayService *GatewayService, cfg *config.Config) *WorkbenchService {
	return &WorkbenchService{
		repo:           repo,
		apiKeyService:  apiKeyService,
		gatewayService: gatewayService,
		cfg:            cfg,
		httpClient:     &http.Client{Timeout: 30 * time.Minute},
	}
}

func (s *WorkbenchService) Models(ctx context.Context, userID int64) (*WorkbenchModels, error) {
	keys, err := s.listUserKeys(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	bindings, err := s.repo.ListModelBindings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list workbench model bindings: %w", err)
	}
	byKeyModel := make(map[string]WorkbenchModelBinding, len(bindings))
	for _, binding := range bindings {
		byKeyModel[workbenchBindingKey(binding.APIKeyID, binding.ModelID)] = binding
	}

	keyOptions := make([]WorkbenchKeyOption, 0, len(keys))
	selected := make([]WorkbenchModelOption, 0)
	available := make([]WorkbenchModelOption, 0)
	for i := range keys {
		key := &keys[i]
		if validateWorkbenchAPIKey(key, userID) != nil || key.Group == nil {
			continue
		}
		catalogIDs := s.workbenchCatalogModelIDs(ctx, key.Group)
		if len(catalogIDs) == 0 {
			continue
		}
		defaultIDs := s.workbenchPickerModelIDs(ctx, key.Group)
		defaultSet := make(map[string]struct{}, len(defaultIDs))
		for _, modelID := range defaultIDs {
			defaultSet[modelID] = struct{}{}
		}

		for index, modelID := range defaultIDs {
			definition := workbenchModelDefinition(modelID, key.Group.Platform)
			binding, exists := byKeyModel[workbenchBindingKey(key.ID, modelID)]
			if !exists {
				ensured, ensureErr := s.repo.UpsertModelBinding(ctx, WorkbenchModelBinding{
					ID:          uuid.NewString(),
					UserID:      userID,
					APIKeyID:    key.ID,
					ModelID:     modelID,
					DisplayName: definition.DisplayName,
					Provider:    definition.Provider,
					Source:      "default",
					SortOrder:   index,
				})
				if ensureErr != nil {
					return nil, fmt.Errorf("ensure workbench model binding: %w", ensureErr)
				}
				binding = *ensured
				byKeyModel[workbenchBindingKey(key.ID, modelID)] = binding
			}
		}

		visibleIDs := append([]string(nil), defaultIDs...)
		for _, modelID := range catalogIDs {
			if _, isDefault := defaultSet[modelID]; isDefault {
				continue
			}
			binding, exists := byKeyModel[workbenchBindingKey(key.ID, modelID)]
			if exists && !binding.Hidden && binding.Source == "manual" {
				visibleIDs = append(visibleIDs, modelID)
			}
		}

		models := make([]WorkbenchModelOption, 0, len(visibleIDs))
		visibleSet := make(map[string]struct{}, len(visibleIDs))
		for index, modelID := range visibleIDs {
			binding, exists := byKeyModel[workbenchBindingKey(key.ID, modelID)]
			if !exists || binding.Hidden {
				continue
			}
			definition := workbenchModelDefinition(modelID, key.Group.Platform)
			option := workbenchModelOptionFromBinding(binding, *key, definition, index)
			models = append(models, option)
			selected = append(selected, option)
			visibleSet[modelID] = struct{}{}
		}

		catalog := make([]WorkbenchModelOption, 0, len(catalogIDs))
		for index, modelID := range catalogIDs {
			definition := workbenchModelDefinition(modelID, key.Group.Platform)
			option := workbenchCatalogModelOption(modelID, *key, definition, index)
			if _, added := visibleSet[modelID]; added {
				binding := byKeyModel[workbenchBindingKey(key.ID, modelID)]
				option.Added = true
				option.BindingID = binding.ID
			}
			catalog = append(catalog, option)
			available = append(available, option)
		}

		defaultID := workbenchPlatformDefaultModel(key.Group.Platform)
		defaultModel := findWorkbenchModel(models, defaultID)
		if defaultModel == nil && len(models) > 0 {
			defaultModel = &models[0]
		}
		groupID := int64(0)
		if key.GroupID != nil {
			groupID = *key.GroupID
		}
		keyOptions = append(keyOptions, WorkbenchKeyOption{
			APIKeyID:         key.ID,
			KeyName:          key.Name,
			GroupID:          groupID,
			GroupName:        key.Group.Name,
			Platform:         key.Group.Platform,
			ProviderLabel:    workbenchProviderLabel(workbenchProvider("", key.Group.Platform)),
			Rate:             key.Group.RateMultiplier,
			DefaultModelID:   defaultModelID(defaultModel),
			DefaultModelName: defaultModelName(defaultModel),
			Models:           models,
			AvailableModels:  catalog,
		})
	}

	sort.SliceStable(keyOptions, func(i, j int) bool {
		if keyOptions[i].GroupName == keyOptions[j].GroupName {
			return keyOptions[i].KeyName < keyOptions[j].KeyName
		}
		return keyOptions[i].GroupName < keyOptions[j].GroupName
	})
	return &WorkbenchModels{Keys: keyOptions, Selected: selected, Available: available}, nil
}

func (s *WorkbenchService) legacyModels(ctx context.Context, userID int64) (*WorkbenchModels, error) {
	keys, err := s.listUserKeys(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	candidates := s.modelCandidates(ctx, keys)
	bindings, err := s.repo.ListModelBindings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list workbench model bindings: %w", err)
	}

	byModel := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		byModel[binding.ModelID] = struct{}{}
	}
	for _, spec := range workbenchDefaultModels {
		candidate := candidates[spec.ID]
		if candidate == nil || len(candidate.EligibleKeys) == 0 {
			continue
		}
		if _, exists := byModel[spec.ID]; exists {
			continue
		}
		eligible := candidate.EligibleKeys[0]
		binding, ensureErr := s.repo.EnsureModelBinding(ctx, WorkbenchModelBinding{
			ID:          uuid.NewString(),
			UserID:      userID,
			APIKeyID:    eligible.APIKeyID,
			ModelID:     spec.ID,
			DisplayName: spec.DisplayName,
			Provider:    candidate.Provider,
			Source:      "default",
			SortOrder:   spec.SortOrder,
		})
		if ensureErr != nil {
			return nil, fmt.Errorf("ensure default workbench model: %w", ensureErr)
		}
		bindings = append(bindings, *binding)
		byModel[spec.ID] = struct{}{}
	}

	keyByID := make(map[int64]APIKey, len(keys))
	for _, key := range keys {
		keyByID[key.ID] = key
	}
	selected := make([]WorkbenchModelOption, 0, len(bindings))
	visibleModels := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.Hidden {
			continue
		}
		visibleModels[binding.ModelID] = struct{}{}
		option := optionFromBinding(binding, keyByID[binding.APIKeyID], candidates[binding.ModelID])
		selected = append(selected, option)
	}
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].SortOrder == selected[j].SortOrder {
			return selected[i].DisplayName < selected[j].DisplayName
		}
		return selected[i].SortOrder < selected[j].SortOrder
	})

	available := make([]WorkbenchModelOption, 0, len(candidates))
	availableModels := make(map[string]struct{}, len(candidates))
	for modelID, candidate := range candidates {
		if _, exists := visibleModels[modelID]; exists {
			continue
		}
		copyOption := *candidate
		copyOption.ReasoningPresets = reasoningPresets()
		available = append(available, copyOption)
		availableModels[modelID] = struct{}{}
	}
	for _, binding := range bindings {
		if !binding.Hidden {
			continue
		}
		if _, exists := availableModels[binding.ModelID]; exists {
			continue
		}
		option := optionFromBinding(binding, keyByID[binding.APIKeyID], nil)
		option.Added = false
		option.BindingID = ""
		option.UnavailableReason = "当前模型白名单不再包含此模型"
		available = append(available, option)
		availableModels[binding.ModelID] = struct{}{}
	}
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].DefaultVisible != available[j].DefaultVisible {
			return available[i].DefaultVisible
		}
		if available[i].Provider == available[j].Provider {
			return available[i].DisplayName < available[j].DisplayName
		}
		return available[i].Provider < available[j].Provider
	})

	return &WorkbenchModels{Selected: selected, Available: available}, nil
}

func (s *WorkbenchService) AddModel(ctx context.Context, userID int64, modelID string, apiKeyID int64) (*WorkbenchModelBinding, error) {
	bindings, err := s.AddModels(ctx, userID, []string{modelID}, apiKeyID)
	if err != nil {
		return nil, err
	}
	return &bindings[0], nil
}

func (s *WorkbenchService) AddModels(ctx context.Context, userID int64, modelIDs []string, apiKeyID int64) ([]WorkbenchModelBinding, error) {
	modelIDs = uniqueWorkbenchModels(modelIDs)
	if len(modelIDs) == 0 || len(modelIDs) > 100 || apiKeyID <= 0 {
		return nil, ErrWorkbenchInvalidInput
	}
	keys, err := s.listUserKeys(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	var selectedKey *APIKey
	for i := range keys {
		if keys[i].ID == apiKeyID && validateWorkbenchAPIKey(&keys[i], userID) == nil {
			selectedKey = &keys[i]
			break
		}
	}
	if selectedKey == nil || selectedKey.Group == nil {
		return nil, ErrWorkbenchModelNotFound
	}
	supported := make(map[string]struct{})
	for _, modelID := range s.workbenchCatalogModelIDs(ctx, selectedKey.Group) {
		supported[modelID] = struct{}{}
	}
	for _, modelID := range modelIDs {
		if _, ok := supported[modelID]; !ok {
			return nil, ErrWorkbenchModelNotFound
		}
	}

	bindings := make([]WorkbenchModelBinding, 0, len(modelIDs))
	for index, modelID := range modelIDs {
		definition := workbenchModelDefinition(modelID, selectedKey.Group.Platform)
		binding, upsertErr := s.repo.UpsertModelBinding(ctx, WorkbenchModelBinding{
			ID:          uuid.NewString(),
			UserID:      userID,
			APIKeyID:    apiKeyID,
			ModelID:     modelID,
			DisplayName: definition.DisplayName,
			Provider:    definition.Provider,
			Source:      "manual",
			SortOrder:   1000 + index,
		})
		if upsertErr != nil {
			return nil, fmt.Errorf("add workbench model: %w", upsertErr)
		}
		bindings = append(bindings, *binding)
	}
	return bindings, nil
}

func (s *WorkbenchService) HideModel(ctx context.Context, userID int64, bindingID string) error {
	if strings.TrimSpace(bindingID) == "" {
		return ErrWorkbenchInvalidInput
	}
	return s.repo.HideModelBinding(ctx, userID, bindingID)
}

func (s *WorkbenchService) CreateConversation(ctx context.Context, userID int64, bindingID, preset string) (*WorkbenchConversation, error) {
	preset = normalizeReasoningPreset(preset)
	if bindingID == "" {
		models, err := s.Models(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, model := range models.Selected {
			if model.Available {
				bindingID = model.BindingID
				break
			}
		}
	}
	if bindingID != "" {
		if _, err := s.validBinding(ctx, userID, bindingID); err != nil {
			return nil, err
		}
	}
	var binding *string
	if bindingID != "" {
		binding = &bindingID
	}
	return s.repo.CreateConversation(ctx, WorkbenchConversation{
		ID:              uuid.NewString(),
		UserID:          userID,
		ModelBindingID:  binding,
		Title:           "新对话",
		ReasoningPreset: preset,
	})
}

func (s *WorkbenchService) ListConversations(ctx context.Context, userID int64) ([]WorkbenchConversation, error) {
	return s.repo.ListConversations(ctx, userID, 100)
}

func (s *WorkbenchService) GetConversation(ctx context.Context, userID int64, conversationID string) (*WorkbenchConversation, error) {
	conversation, err := s.repo.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repo.ListMessages(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	conversation.Messages = messages
	return conversation, nil
}

func (s *WorkbenchService) UpdateConversation(ctx context.Context, userID int64, conversationID string, patch WorkbenchConversationPatch) (*WorkbenchConversation, error) {
	if patch.ModelBindingID != nil && strings.TrimSpace(*patch.ModelBindingID) != "" {
		if _, err := s.validBinding(ctx, userID, strings.TrimSpace(*patch.ModelBindingID)); err != nil {
			return nil, err
		}
	}
	if patch.ReasoningPreset != nil {
		normalized := normalizeReasoningPreset(*patch.ReasoningPreset)
		patch.ReasoningPreset = &normalized
	}
	if patch.Title != nil {
		title := truncateRunes(strings.TrimSpace(*patch.Title), 80)
		if title == "" {
			title = "新对话"
		}
		patch.Title = &title
	}
	return s.repo.UpdateConversation(ctx, userID, conversationID, patch)
}

func (s *WorkbenchService) DeleteConversation(ctx context.Context, userID int64, conversationID string) error {
	return s.repo.DeleteConversation(ctx, userID, conversationID)
}

func (s *WorkbenchService) CreateTurn(ctx context.Context, userID int64, conversationID, content, bindingID, preset string, attachments []WorkbenchAttachment) (*WorkbenchTurn, error) {
	content = strings.TrimSpace(content)
	validatedAttachments, err := validateWorkbenchAttachments(attachments)
	if err != nil || (content == "" && len(validatedAttachments) == 0) || utf8.RuneCountInString(content) > workbenchMaxMessageRunes {
		return nil, ErrWorkbenchInvalidInput
	}
	conversation, err := s.repo.GetConversation(ctx, userID, conversationID)
	if err != nil {
		return nil, err
	}
	selectedBindingID := ""
	if conversation.ModelBindingID != nil {
		selectedBindingID = strings.TrimSpace(*conversation.ModelBindingID)
	}
	if requestedBindingID := strings.TrimSpace(bindingID); requestedBindingID != "" {
		selectedBindingID = requestedBindingID
	}
	if selectedBindingID == "" {
		return nil, ErrWorkbenchBindingNotFound
	}
	if _, err := s.validBinding(ctx, userID, selectedBindingID); err != nil {
		return nil, err
	}
	selectedPreset := conversation.ReasoningPreset
	if strings.TrimSpace(preset) != "" {
		selectedPreset = preset
	}
	selectedPreset = normalizeReasoningPreset(selectedPreset)
	title := conversation.Title
	if title == "" || title == "新对话" {
		if content != "" {
			title = truncateRunes(strings.Join(strings.Fields(content), " "), 28)
		} else {
			title = "图片对话"
		}
	}
	return s.repo.CreateTurn(ctx, userID, conversationID, content, title, validatedAttachments, &selectedBindingID, selectedPreset)
}

func validateWorkbenchAttachments(attachments []WorkbenchAttachment) ([]WorkbenchAttachment, error) {
	if len(attachments) > workbenchMaxAttachmentsPerMessage {
		return nil, ErrWorkbenchInvalidInput
	}
	allowed := map[string]struct{}{
		"image/png": {}, "image/jpeg": {}, "image/webp": {}, "image/gif": {},
		"application/pdf":    {},
		"application/msword": {},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
		"application/vnd.ms-excel": {},
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
		"application/vnd.ms-powerpoint":                                             {},
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
		"application/json": {}, "application/xml": {}, "application/rtf": {},
		"text/plain": {}, "text/markdown": {}, "text/csv": {}, "text/html": {},
		"text/xml": {}, "text/yaml": {}, "application/x-yaml": {},
	}
	result := make([]WorkbenchAttachment, 0, len(attachments))
	var total int64
	for _, attachment := range attachments {
		mimeType := strings.ToLower(strings.TrimSpace(attachment.MIMEType))
		if _, ok := allowed[mimeType]; !ok {
			return nil, ErrWorkbenchInvalidInput
		}
		prefix := "data:" + mimeType + ";base64,"
		if !strings.HasPrefix(attachment.DataURL, prefix) {
			return nil, ErrWorkbenchInvalidInput
		}
		decodedSize, err := workbenchBase64DecodedSize(strings.TrimPrefix(attachment.DataURL, prefix))
		if err != nil || decodedSize <= 0 || decodedSize > workbenchMaxAttachmentBytes {
			return nil, ErrWorkbenchInvalidInput
		}
		total += decodedSize
		if total > workbenchMaxAttachmentTotalBytes {
			return nil, ErrWorkbenchInvalidInput
		}
		id := strings.TrimSpace(attachment.ID)
		if _, err := uuid.Parse(id); err != nil {
			id = uuid.NewString()
		}
		name := truncateRunes(strings.TrimSpace(attachment.Name), 160)
		if name == "" {
			name = "attachment"
		}
		result = append(result, WorkbenchAttachment{
			ID: id, Name: name, MIMEType: mimeType, SizeBytes: decodedSize, DataURL: attachment.DataURL,
		})
	}
	return result, nil
}

func workbenchBase64DecodedSize(value string) (int64, error) {
	if len(value) > base64.StdEncoding.EncodedLen(workbenchMaxAttachmentBytes) {
		return 0, ErrWorkbenchInvalidInput
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return 0, err
	}
	return int64(len(decoded)), nil
}

func (s *WorkbenchService) CancelGeneration(ctx context.Context, userID int64, messageID string) error {
	return s.repo.CancelAssistantMessage(ctx, userID, messageID)
}

type WorkbenchGatewayError struct {
	StatusCode int
	Message    string
}

func (e *WorkbenchGatewayError) Error() string {
	if e == nil {
		return "workbench gateway error"
	}
	return e.Message
}

const (
	WorkbenchStreamThinkingStarted = "workbench.thinking.started"
	WorkbenchStreamThinkingDelta   = "workbench.thinking.delta"
	WorkbenchStreamAnswerStarted   = "workbench.answer.started"
	WorkbenchStreamAnswerDelta     = "workbench.answer.delta"
)

type WorkbenchStreamEvent struct {
	Type      string `json:"type"`
	MessageID string `json:"message_id"`
	Delta     string `json:"delta,omitempty"`
}

func (s *WorkbenchService) StreamGeneration(ctx context.Context, userID int64, messageID string, emit func(WorkbenchStreamEvent) error) error {
	requestID := "wb_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := s.repo.ClaimAssistantMessage(ctx, userID, messageID, requestID); err != nil {
		return err
	}
	fail := func(status, message, content, summary string) error {
		_ = s.repo.FinishAssistantMessage(context.WithoutCancel(ctx), userID, messageID, status, content, summary, truncateRunes(message, 800))
		if status == WorkbenchMessageCanceled {
			return context.Canceled
		}
		return errors.New(message)
	}

	message, conversation, binding, apiKey, history, err := s.generationContext(ctx, userID, messageID)
	if err != nil {
		return fail(WorkbenchMessageFailed, err.Error(), "", "")
	}
	body, err := buildWorkbenchResponsesBody(conversation, binding, history)
	if err != nil {
		return fail(WorkbenchMessageFailed, err.Error(), "", "")
	}

	endpoint := s.localGatewayURL("/v1/responses")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fail(WorkbenchMessageFailed, err.Error(), "", "")
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey.Key))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Request-ID", requestID)
	req.Header.Set("User-Agent", "VinzkAPI-Workbench/2")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			return fail(WorkbenchMessageCanceled, "Generation canceled", "", "")
		}
		return fail(WorkbenchMessageFailed, err.Error(), "", "")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		gatewayMessage := workbenchGatewayErrorMessage(raw, resp.Status)
		_ = s.repo.FinishAssistantMessage(context.WithoutCancel(ctx), userID, message.ID, WorkbenchMessageFailed, "", "", gatewayMessage)
		return &WorkbenchGatewayError{StatusCode: resp.StatusCode, Message: gatewayMessage}
	}
	if emit != nil {
		if err := emit(WorkbenchStreamEvent{Type: WorkbenchStreamThinkingStarted, MessageID: messageID}); err != nil {
			return fail(WorkbenchMessageCanceled, "Client disconnected", "", "")
		}
	}

	parser := workbenchStreamParser{}
	answerStarted := false
	reader := bufio.NewReaderSize(resp.Body, 32*1024)
	for {
		chunk, readErr := reader.ReadBytes('\n')
		if len(chunk) > 0 {
			for _, parsedEvent := range parser.AddLine(chunk) {
				if emit == nil {
					continue
				}
				if parsedEvent.Type == WorkbenchStreamAnswerDelta && !answerStarted {
					if err := emit(WorkbenchStreamEvent{Type: WorkbenchStreamAnswerStarted, MessageID: messageID}); err != nil {
						return fail(WorkbenchMessageCanceled, "Client disconnected", parser.Text.String(), parser.Reasoning.String())
					}
					answerStarted = true
				}
				if err := emit(WorkbenchStreamEvent{Type: parsedEvent.Type, MessageID: messageID, Delta: parsedEvent.Delta}); err != nil {
					return fail(WorkbenchMessageCanceled, "Client disconnected", parser.Text.String(), parser.Reasoning.String())
				}
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				if errors.Is(readErr, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
					return fail(WorkbenchMessageCanceled, "Generation canceled", parser.Text.String(), parser.Reasoning.String())
				}
				return fail(WorkbenchMessageFailed, readErr.Error(), parser.Text.String(), parser.Reasoning.String())
			}
			break
		}
	}

	if parser.Failed {
		return fail(WorkbenchMessageFailed, workbenchFirstNonEmpty(parser.ErrorMessage, "Generation failed"), parser.Text.String(), parser.Reasoning.String())
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return fail(WorkbenchMessageCanceled, "Generation canceled", parser.Text.String(), parser.Reasoning.String())
	}
	if !parser.Completed {
		return fail(WorkbenchMessageFailed, "Response stream ended before completion", parser.Text.String(), parser.Reasoning.String())
	}
	if err := s.repo.FinishAssistantMessage(context.WithoutCancel(ctx), userID, message.ID, WorkbenchMessageCompleted, parser.Text.String(), parser.Reasoning.String(), ""); err != nil {
		if errors.Is(err, ErrWorkbenchGenerationCanceled) {
			return context.Canceled
		}
		return err
	}
	return nil
}

func (s *WorkbenchService) generationContext(ctx context.Context, userID int64, messageID string) (*WorkbenchMessage, *WorkbenchConversation, *WorkbenchModelBinding, *APIKey, []WorkbenchMessage, error) {
	message, err := s.repo.GetMessage(ctx, userID, messageID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if message == nil || message.Role != "assistant" {
		return nil, nil, nil, nil, nil, ErrWorkbenchMessageNotFound
	}
	conversation, err := s.repo.GetConversation(ctx, userID, message.ConversationID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	history, err := s.repo.ListMessages(ctx, userID, conversation.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	bindingID := conversation.ModelBindingID
	if message.ModelBindingID != nil && strings.TrimSpace(*message.ModelBindingID) != "" {
		bindingID = message.ModelBindingID
	}
	if bindingID == nil {
		return nil, nil, nil, nil, nil, ErrWorkbenchBindingNotFound
	}
	binding, err := s.validBinding(ctx, userID, *bindingID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	apiKey, err := s.apiKeyService.GetByID(ctx, binding.APIKeyID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	if err := validateWorkbenchAPIKey(apiKey, userID); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	generationConversation := *conversation
	generationConversation.ModelBindingID = bindingID
	if strings.TrimSpace(message.ReasoningPreset) != "" {
		generationConversation.ReasoningPreset = normalizeReasoningPreset(message.ReasoningPreset)
	}
	return message, &generationConversation, binding, apiKey, history, nil
}

func (s *WorkbenchService) validBinding(ctx context.Context, userID int64, bindingID string) (*WorkbenchModelBinding, error) {
	binding, err := s.repo.GetModelBinding(ctx, userID, bindingID)
	if err != nil {
		return nil, err
	}
	if binding.Hidden {
		return nil, ErrWorkbenchBindingNotFound
	}
	apiKey, err := s.apiKeyService.GetByID(ctx, binding.APIKeyID)
	if err != nil || validateWorkbenchAPIKey(apiKey, userID) != nil {
		return nil, ErrWorkbenchBindingNotFound
	}
	models := s.workbenchCatalogModelIDs(ctx, apiKey.Group)
	for _, model := range models {
		if model == binding.ModelID {
			return binding, nil
		}
	}
	return nil, ErrWorkbenchModelNotFound
}

func validateWorkbenchAPIKey(apiKey *APIKey, userID int64) error {
	if apiKey == nil || apiKey.UserID != userID || apiKey.Group == nil || apiKey.GroupID == nil {
		return ErrWorkbenchBindingNotFound
	}
	if !apiKey.IsActive() || apiKey.IsExpired() || apiKey.IsQuotaExhausted() || apiKey.Group.Status != StatusActive {
		return ErrWorkbenchBindingNotFound
	}
	return nil
}

func (s *WorkbenchService) listUserKeys(ctx context.Context, userID int64, status string) ([]APIKey, error) {
	keys, _, err := s.apiKeyService.List(ctx, userID, pagination.PaginationParams{
		Page: 1, PageSize: 1000, SortBy: "created_at", SortOrder: pagination.SortOrderAsc,
	}, APIKeyListFilters{Status: status})
	return keys, err
}

func workbenchBindingKey(apiKeyID int64, modelID string) string {
	return fmt.Sprintf("%d\x00%s", apiKeyID, modelID)
}

func (s *WorkbenchService) workbenchPickerModelIDs(ctx context.Context, group *Group) []string {
	if group == nil {
		return nil
	}
	available := s.workbenchCatalogModelIDs(ctx, group)
	availableSet := make(map[string]struct{}, len(available)+1)
	for _, modelID := range available {
		availableSet[modelID] = struct{}{}
	}
	provider := group.Platform
	if _, exists := workbenchPickerModelsByProvider[provider]; !exists {
		for _, modelID := range available {
			candidateProvider := workbenchProvider(modelID, group.Platform)
			if _, supported := workbenchPickerModelsByProvider[candidateProvider]; supported {
				provider = candidateProvider
				break
			}
		}
	}
	defaultID := workbenchPlatformDefaultModel(provider)
	if defaultID != "" && !group.CustomModelsListEnabled() {
		availableSet[defaultID] = struct{}{}
	}

	result := make([]string, 0, len(workbenchPickerModelsByProvider[provider]))
	for _, modelID := range workbenchPickerModelsByProvider[provider] {
		if _, available := availableSet[modelID]; !available {
			continue
		}
		result = append(result, modelID)
	}
	return result
}

func (s *WorkbenchService) workbenchCatalogModelIDs(ctx context.Context, group *Group) []string {
	if group == nil {
		return nil
	}
	models := s.modelsForGroup(ctx, group)
	result := make([]string, 0, len(models)+1)
	for _, modelID := range models {
		if isWorkbenchConversationModel(modelID) {
			result = append(result, modelID)
		}
	}
	defaultID := workbenchPlatformDefaultModel(group.Platform)
	if defaultID != "" && !group.CustomModelsListEnabled() && isWorkbenchConversationModel(defaultID) {
		result = append(result, defaultID)
	}
	return uniqueWorkbenchModels(result)
}

func workbenchPlatformDefaultModel(platform string) string {
	switch platform {
	case PlatformOpenAI:
		return "gpt-5.4-mini"
	case PlatformGemini:
		return "gemini-2.0-flash"
	case PlatformAnthropic:
		return "claude-3-5-haiku-20241022"
	default:
		return ""
	}
}

func workbenchModelOptionFromBinding(binding WorkbenchModelBinding, key APIKey, definition workbenchModelInfo, sortOrder int) WorkbenchModelOption {
	groupName := ""
	if key.Group != nil {
		groupName = key.Group.Name
	}
	return WorkbenchModelOption{
		ID:               binding.ModelID,
		DisplayName:      definition.DisplayName,
		Provider:         definition.Provider,
		ProviderLabel:    workbenchProviderLabel(definition.Provider),
		DefaultVisible:   key.Group != nil && binding.ModelID == workbenchPlatformDefaultModel(key.Group.Platform),
		Added:            true,
		Available:        true,
		BindingID:        binding.ID,
		APIKeyID:         key.ID,
		KeyName:          key.Name,
		GroupName:        groupName,
		SortOrder:        sortOrder,
		ReasoningPresets: reasoningPresets(),
	}
}

func workbenchCatalogModelOption(modelID string, key APIKey, definition workbenchModelInfo, sortOrder int) WorkbenchModelOption {
	groupName := ""
	if key.Group != nil {
		groupName = key.Group.Name
	}
	return WorkbenchModelOption{
		ID:               modelID,
		DisplayName:      definition.DisplayName,
		Provider:         definition.Provider,
		ProviderLabel:    workbenchProviderLabel(definition.Provider),
		DefaultVisible:   workbenchDefaultSpec(modelID) != nil,
		Available:        true,
		APIKeyID:         key.ID,
		KeyName:          key.Name,
		GroupName:        groupName,
		SortOrder:        sortOrder,
		ReasoningPresets: reasoningPresets(),
	}
}

func findWorkbenchModel(models []WorkbenchModelOption, modelID string) *WorkbenchModelOption {
	for i := range models {
		if models[i].ID == modelID {
			return &models[i]
		}
	}
	return nil
}

func defaultModelID(model *WorkbenchModelOption) string {
	if model == nil {
		return ""
	}
	return model.ID
}

func defaultModelName(model *WorkbenchModelOption) string {
	if model == nil {
		return ""
	}
	return model.DisplayName
}

func (s *WorkbenchService) modelCandidates(ctx context.Context, keys []APIKey) map[string]*WorkbenchModelOption {
	result := make(map[string]*WorkbenchModelOption)
	for i := range keys {
		key := &keys[i]
		if validateWorkbenchAPIKey(key, key.UserID) != nil {
			continue
		}
		for _, modelID := range s.modelsForGroup(ctx, key.Group) {
			if isWorkbenchImageModel(modelID) {
				continue
			}
			option := result[modelID]
			if option == nil {
				definition := workbenchModelDefinition(modelID, key.Group.Platform)
				option = &WorkbenchModelOption{
					ID:               modelID,
					DisplayName:      definition.DisplayName,
					Provider:         definition.Provider,
					ProviderLabel:    workbenchProviderLabel(definition.Provider),
					DefaultVisible:   workbenchDefaultSpec(modelID) != nil,
					Available:        true,
					SortOrder:        definition.SortOrder,
					ReasoningPresets: reasoningPresets(),
					EligibleKeys:     make([]WorkbenchEligibleKey, 0, 1),
				}
				result[modelID] = option
			}
			groupID := int64(0)
			if key.GroupID != nil {
				groupID = *key.GroupID
			}
			option.EligibleKeys = append(option.EligibleKeys, WorkbenchEligibleKey{
				APIKeyID:  key.ID,
				KeyName:   key.Name,
				GroupID:   groupID,
				GroupName: key.Group.Name,
				Platform:  key.Group.Platform,
				Rate:      key.Group.RateMultiplier,
			})
		}
	}
	return result
}

func (s *WorkbenchService) modelsForGroup(ctx context.Context, group *Group) []string {
	if group == nil {
		return nil
	}
	available := s.gatewayService.GetAvailableModels(ctx, &group.ID, group.Platform)
	if len(available) == 0 {
		return nil
	}
	fallback := workbenchDefaultModelIDsForPlatform(group.Platform)
	if group.CustomModelsListEnabled() {
		return filterWorkbenchModels(available, fallback, group.ModelsListConfig.Models)
	}
	if len(available) > 0 {
		return uniqueWorkbenchModels(available)
	}
	return fallback
}

func (s *WorkbenchService) localGatewayURL(path string) string {
	port := 8080
	if s.cfg != nil && s.cfg.Server.Port > 0 {
		port = s.cfg.Server.Port
	}
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
}

type workbenchModelInfo struct {
	DisplayName string
	Provider    string
	SortOrder   int
}

func workbenchModelDefinition(modelID, groupPlatform string) workbenchModelInfo {
	if spec := workbenchDefaultSpec(modelID); spec != nil {
		return workbenchModelInfo{DisplayName: spec.DisplayName, Provider: workbenchProvider(modelID, groupPlatform), SortOrder: spec.SortOrder}
	}
	for _, model := range openai.DefaultModels {
		if model.ID == modelID {
			return workbenchModelInfo{DisplayName: model.DisplayName, Provider: PlatformOpenAI, SortOrder: 1000}
		}
	}
	for _, model := range claude.DefaultModels {
		if model.ID == modelID {
			return workbenchModelInfo{DisplayName: model.DisplayName, Provider: PlatformAnthropic, SortOrder: 1000}
		}
	}
	for _, model := range geminicli.DefaultModels {
		if model.ID == modelID {
			return workbenchModelInfo{DisplayName: model.DisplayName, Provider: PlatformGemini, SortOrder: 1000}
		}
	}
	for _, model := range antigravity.DefaultModels() {
		if model.ID == modelID {
			return workbenchModelInfo{DisplayName: model.DisplayName, Provider: workbenchProvider(modelID, groupPlatform), SortOrder: 1000}
		}
	}
	return workbenchModelInfo{DisplayName: modelID, Provider: workbenchProvider(modelID, groupPlatform), SortOrder: 1000}
}

func workbenchDefaultSpec(modelID string) *workbenchDefaultModel {
	for i := range workbenchDefaultModels {
		if workbenchDefaultModels[i].ID == modelID {
			return &workbenchDefaultModels[i]
		}
	}
	return nil
}

func optionFromBinding(binding WorkbenchModelBinding, key APIKey, candidate *WorkbenchModelOption) WorkbenchModelOption {
	option := WorkbenchModelOption{
		ID: binding.ModelID, DisplayName: binding.DisplayName, Provider: binding.Provider,
		ProviderLabel: workbenchProviderLabel(binding.Provider), Added: true, BindingID: binding.ID,
		APIKeyID: binding.APIKeyID, SortOrder: binding.SortOrder, ReasoningPresets: reasoningPresets(),
		DefaultVisible: workbenchDefaultSpec(binding.ModelID) != nil,
	}
	if option.DisplayName == "" {
		option.DisplayName = binding.ModelID
	}
	if key.ID == 0 {
		option.UnavailableReason = "API Key 已删除"
		return option
	}
	option.KeyName = key.Name
	if key.Group != nil {
		option.GroupName = key.Group.Name
	}
	if candidate == nil || validateWorkbenchAPIKey(&key, binding.UserID) != nil {
		option.UnavailableReason = "API Key 或分组当前不可用"
		return option
	}
	for _, eligible := range candidate.EligibleKeys {
		if eligible.APIKeyID == binding.APIKeyID {
			option.Available = true
			return option
		}
	}
	option.UnavailableReason = "当前分组不支持该模型"
	return option
}

func workbenchProvider(modelID, groupPlatform string) string {
	lower := strings.ToLower(modelID)
	switch {
	case strings.HasPrefix(lower, "gpt-"), strings.HasPrefix(lower, "o1"), strings.HasPrefix(lower, "o3"):
		return PlatformOpenAI
	case strings.HasPrefix(lower, "claude-"), strings.Contains(lower, "fable"):
		return PlatformAnthropic
	case strings.HasPrefix(lower, "gemini-"):
		return PlatformGemini
	case strings.HasPrefix(lower, "grok-"):
		return PlatformGrok
	default:
		return groupPlatform
	}
}

func workbenchProviderLabel(provider string) string {
	switch provider {
	case PlatformAnthropic, PlatformAntigravity:
		return "Anthropic"
	case PlatformGemini:
		return "Google"
	case PlatformGrok:
		return "xAI"
	default:
		return "OpenAI"
	}
}

func reasoningPresets() []string {
	return []string{WorkbenchReasoningFast, WorkbenchReasoningStandard, WorkbenchReasoningDeep}
}

func normalizeReasoningPreset(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case WorkbenchReasoningFast:
		return WorkbenchReasoningFast
	case WorkbenchReasoningDeep:
		return WorkbenchReasoningDeep
	default:
		return WorkbenchReasoningStandard
	}
}

func workbenchReasoningEffort(provider, preset string) string {
	preset = normalizeReasoningPreset(preset)
	if provider == PlatformAnthropic || provider == PlatformAntigravity {
		switch preset {
		case WorkbenchReasoningFast:
			return "low"
		case WorkbenchReasoningDeep:
			return "high"
		default:
			return "medium"
		}
	}
	switch preset {
	case WorkbenchReasoningFast:
		return "medium"
	case WorkbenchReasoningDeep:
		return "xhigh"
	default:
		return "high"
	}
}

func workbenchReasoningSummary(preset string) string {
	if normalizeReasoningPreset(preset) == WorkbenchReasoningFast {
		return "concise"
	}
	return "detailed"
}

func buildWorkbenchResponsesBody(conversation *WorkbenchConversation, binding *WorkbenchModelBinding, messages []WorkbenchMessage) ([]byte, error) {
	selected := make([]WorkbenchMessage, 0, min(len(messages), workbenchMaxHistoryItems))
	runeBudget := 0
	for i := len(messages) - 1; i >= 0 && len(selected) < workbenchMaxHistoryItems; i-- {
		message := messages[i]
		if message.Role != "user" && message.Role != "assistant" {
			continue
		}
		if message.Role == "assistant" && message.Status != WorkbenchMessageCompleted {
			continue
		}
		messageRunes := utf8.RuneCountInString(message.Content)
		if len(selected) > 0 && runeBudget+messageRunes > workbenchMaxHistoryRunes {
			break
		}
		selected = append(selected, message)
		runeBudget += messageRunes
	}
	for left, right := 0, len(selected)-1; left < right; left, right = left+1, right-1 {
		selected[left], selected[right] = selected[right], selected[left]
	}
	input := make([]map[string]any, 0, len(selected))
	for _, message := range selected {
		contentType := "input_text"
		if message.Role == "assistant" {
			contentType = "output_text"
		}
		content := make([]map[string]any, 0, 1+len(message.Attachments))
		if message.Content != "" {
			content = append(content, map[string]any{"type": contentType, "text": message.Content})
		}
		if message.Role == "user" {
			for _, attachment := range message.Attachments {
				if strings.HasPrefix(attachment.MIMEType, "image/") {
					content = append(content, map[string]any{"type": "input_image", "image_url": attachment.DataURL})
					continue
				}
				content = append(content, map[string]any{
					"type": "input_file", "filename": attachment.Name, "file_data": attachment.DataURL,
				})
			}
		}
		if len(content) == 0 {
			continue
		}
		input = append(input, map[string]any{
			"type":    "message",
			"role":    message.Role,
			"content": content,
		})
	}
	payload := map[string]any{
		"model":            binding.ModelID,
		"instructions":     "You are 维枢AI. Complete the user's task directly and clearly. Provide a useful user-visible reasoning summary that describes the approach, checks, and progress without exposing hidden chain-of-thought.",
		"input":            input,
		"stream":           true,
		"store":            false,
		"prompt_cache_key": conversation.ID,
		"reasoning": map[string]string{
			"effort":  workbenchReasoningEffort(binding.Provider, conversation.ReasoningPreset),
			"summary": workbenchReasoningSummary(conversation.ReasoningPreset),
		},
	}
	return json.Marshal(payload)
}

type workbenchStreamParser struct {
	Text         strings.Builder
	Reasoning    strings.Builder
	Failed       bool
	Completed    bool
	ErrorMessage string
}

type workbenchParsedStreamEvent struct {
	Type  string
	Delta string
}

func (p *workbenchStreamParser) AddLine(line []byte) []workbenchParsedStreamEvent {
	trimmed := bytes.TrimSpace(line)
	if !bytes.HasPrefix(trimmed, []byte("data:")) {
		return nil
	}
	data := bytes.TrimSpace(bytes.TrimPrefix(trimmed, []byte("data:")))
	if len(data) == 0 {
		return nil
	}
	if bytes.Equal(data, []byte("[DONE]")) {
		p.Completed = true
		return nil
	}
	if !json.Valid(data) {
		return nil
	}
	var event map[string]any
	if err := json.Unmarshal(data, &event); err != nil {
		return nil
	}
	parsed := make([]workbenchParsedStreamEvent, 0, 1)
	typeName, _ := event["type"].(string)
	switch typeName {
	case "response.output_text.delta":
		if delta, _ := event["delta"].(string); delta != "" {
			p.Text.WriteString(delta)
			parsed = append(parsed, workbenchParsedStreamEvent{Type: WorkbenchStreamAnswerDelta, Delta: delta})
		}
	case "response.reasoning_summary_text.delta":
		if delta, _ := event["delta"].(string); delta != "" {
			p.Reasoning.WriteString(delta)
			parsed = append(parsed, workbenchParsedStreamEvent{Type: WorkbenchStreamThinkingDelta, Delta: delta})
		}
	case "response.completed", "response.done":
		p.Completed = true
		if p.Text.Len() == 0 {
			if text := extractWorkbenchCompletedText(event); text != "" {
				p.Text.WriteString(text)
				parsed = append(parsed, workbenchParsedStreamEvent{Type: WorkbenchStreamAnswerDelta, Delta: text})
			}
		}
	case "response.failed", "response.incomplete", "response.cancelled", "response.canceled", "error":
		p.Failed = true
		p.ErrorMessage = extractWorkbenchEventError(event)
	}
	return parsed
}

func extractWorkbenchCompletedText(event map[string]any) string {
	responseValue, _ := event["response"].(map[string]any)
	output, _ := responseValue["output"].([]any)
	var texts []string
	for _, rawItem := range output {
		item, _ := rawItem.(map[string]any)
		content, _ := item["content"].([]any)
		for _, rawPart := range content {
			part, _ := rawPart.(map[string]any)
			if text, _ := part["text"].(string); text != "" {
				texts = append(texts, text)
			}
		}
	}
	return strings.Join(texts, "\n\n")
}

func extractWorkbenchEventError(event map[string]any) string {
	if value, ok := event["error"].(map[string]any); ok {
		if message, _ := value["message"].(string); message != "" {
			return message
		}
	}
	if responseValue, ok := event["response"].(map[string]any); ok {
		if value, ok := responseValue["error"].(map[string]any); ok {
			if message, _ := value["message"].(string); message != "" {
				return message
			}
		}
	}
	if message, _ := event["message"].(string); message != "" {
		return message
	}
	return "Generation failed"
}

func workbenchGatewayErrorMessage(raw []byte, fallback string) string {
	if json.Valid(raw) {
		var body map[string]any
		if json.Unmarshal(raw, &body) == nil {
			if value, ok := body["error"].(map[string]any); ok {
				if message, _ := value["message"].(string); message != "" {
					return message
				}
			}
			if message, _ := body["message"].(string); message != "" {
				return message
			}
		}
	}
	if text := strings.TrimSpace(string(raw)); text != "" {
		return truncateRunes(text, 800)
	}
	return fallback
}

func workbenchDefaultModelIDsForPlatform(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return openai.DefaultModelIDs()
	case PlatformGemini:
		ids := make([]string, 0, len(geminicli.DefaultModels))
		for _, model := range geminicli.DefaultModels {
			ids = append(ids, model.ID)
		}
		return ids
	case PlatformAntigravity:
		models := antigravity.DefaultModels()
		ids := make([]string, 0, len(models))
		for _, model := range models {
			ids = append(ids, model.ID)
		}
		return ids
	case PlatformGrok:
		return xai.DefaultModelIDs()
	default:
		return claude.DefaultModelIDs()
	}
}

func filterWorkbenchModels(available, fallback, selected []string) []string {
	if len(selected) == 0 {
		return uniqueWorkbenchModels(available)
	}
	source := available
	if len(source) == 0 {
		source = fallback
	}
	allowed := uniqueWorkbenchModels(source)
	result := make([]string, 0, len(selected))
	seen := make(map[string]struct{}, len(selected))
	for _, model := range selected {
		model = strings.TrimSpace(model)
		if model == "" || !workbenchPatternListAllows(allowed, model) {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		result = append(result, model)
	}
	return result
}

func workbenchPatternListAllows(patterns []string, model string) bool {
	for _, pattern := range patterns {
		if pattern == model || (strings.HasSuffix(pattern, "*") && strings.HasPrefix(model, strings.TrimSuffix(pattern, "*"))) {
			return true
		}
	}
	return false
}

func uniqueWorkbenchModels(models []string) []string {
	result := make([]string, 0, len(models))
	seen := make(map[string]struct{}, len(models))
	for _, model := range models {
		model = strings.TrimSpace(model)
		if model == "" {
			continue
		}
		if _, exists := seen[model]; exists {
			continue
		}
		seen[model] = struct{}{}
		result = append(result, model)
	}
	return result
}

func isWorkbenchImageModel(modelID string) bool {
	lower := strings.ToLower(modelID)
	return strings.Contains(lower, "image") && !strings.Contains(lower, "vision")
}

func isWorkbenchConversationModel(modelID string) bool {
	lower := strings.ToLower(strings.TrimSpace(modelID))
	if lower == "" || isWorkbenchImageModel(lower) || lower == "codex-auto-review" {
		return false
	}
	for _, marker := range []string{"audio", "realtime", "transcribe", "text-to-speech", "tts", "embedding", "moderation"} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	return true
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 || utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit])) + "…"
}

func workbenchFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
