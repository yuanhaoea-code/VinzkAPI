package handler

import (
	"context"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type ModelMarketHandler struct {
	apiKeyService  *service.APIKeyService
	gatewayService *service.GatewayService
	billingService *service.BillingService
}

func NewModelMarketHandler(
	apiKeyService *service.APIKeyService,
	gatewayService *service.GatewayService,
	billingService *service.BillingService,
) *ModelMarketHandler {
	return &ModelMarketHandler{
		apiKeyService:  apiKeyService,
		gatewayService: gatewayService,
		billingService: billingService,
	}
}

type modelMarketTokenPrice struct {
	Input     float64 `json:"input"`
	Output    float64 `json:"output"`
	CacheRead float64 `json:"cache_read"`
}

type modelMarketGroupPrice struct {
	GroupID     int64                  `json:"group_id"`
	GroupName   string                 `json:"group_name"`
	Platform    string                 `json:"platform"`
	Rate        float64                `json:"rate"`
	TokenPrice  *modelMarketTokenPrice `json:"token_price,omitempty"`
	ImagePrices map[string]float64     `json:"image_prices,omitempty"`
}

type modelMarketModel struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Provider      string                  `json:"provider"`
	ProviderLabel string                  `json:"provider_label"`
	Billing       string                  `json:"billing"`
	OfficialPrice *modelMarketTokenPrice  `json:"official_price,omitempty"`
	Groups        []modelMarketGroupPrice `json:"groups"`
	EndpointTypes []string                `json:"endpoint_types"`
	Tags          []string                `json:"tags"`
}

type modelMarketCatalogResponse struct {
	Models []modelMarketModel `json:"models"`
}

type modelMarketAccumulator struct {
	model         modelMarketModel
	endpointTypes map[string]struct{}
}

// Catalog returns only models reachable through groups available to the current user.
// Token prices use the same dynamic pricing catalog as billing, multiplied by the
// user's effective group rate. Image prices are the fixed per-resolution group prices.
func (h *ModelMarketHandler) Catalog(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	groups, err := h.apiKeyService.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	rates, err := h.apiKeyService.GetUserGroupRates(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	models := make(map[string]*modelMarketAccumulator)
	pricingCache := make(map[string]*modelMarketTokenPrice)
	for i := range groups {
		group := &groups[i]
		effectiveRate := group.RateMultiplier
		if customRate, exists := rates[group.ID]; exists {
			effectiveRate = customRate
		}

		modelIDs := h.availableModelsForGroup(c.Request.Context(), group)
		if group.AllowImageGeneration {
			modelIDs = append(modelIDs, "gpt-image-2")
		}
		for _, modelID := range uniqueModelIDs(modelIDs) {
			isImage := isImageGenerationModel(modelID)
			if isImage && !group.AllowImageGeneration {
				continue
			}

			acc := models[modelID]
			if acc == nil {
				provider := modelMarketProvider(modelID, group.Platform)
				billing := "token"
				if isImage {
					billing = "request"
				}
				acc = &modelMarketAccumulator{
					model: modelMarketModel{
						ID:            modelID,
						Name:          modelMarketDisplayName(modelID),
						Provider:      provider,
						ProviderLabel: modelMarketProviderLabel(provider),
						Billing:       billing,
						Groups:        make([]modelMarketGroupPrice, 0),
						Tags:          modelMarketTags(modelID, billing),
					},
					endpointTypes: make(map[string]struct{}),
				}
				models[modelID] = acc
			}

			groupPrice := modelMarketGroupPrice{
				GroupID:   group.ID,
				GroupName: group.Name,
				Platform:  group.Platform,
				Rate:      effectiveRate,
			}
			if isImage {
				groupPrice.ImagePrices = modelMarketImagePrices(group)
				if len(groupPrice.ImagePrices) == 0 {
					continue
				}
			} else {
				officialPrice, exists := pricingCache[modelID]
				if !exists {
					officialPrice = h.modelTokenPrice(modelID)
					pricingCache[modelID] = officialPrice
				}
				if officialPrice == nil {
					continue
				}
				acc.model.OfficialPrice = officialPrice
				groupPrice.TokenPrice = multiplyModelMarketPrice(officialPrice, effectiveRate)
			}

			acc.endpointTypes[modelMarketEndpointType(group.Platform)] = struct{}{}
			acc.model.Groups = append(acc.model.Groups, groupPrice)
		}
	}

	result := make([]modelMarketModel, 0, len(models))
	for _, acc := range models {
		if len(acc.model.Groups) == 0 {
			continue
		}
		sort.Slice(acc.model.Groups, func(i, j int) bool {
			if acc.model.Groups[i].Rate == acc.model.Groups[j].Rate {
				return acc.model.Groups[i].GroupName < acc.model.Groups[j].GroupName
			}
			return acc.model.Groups[i].Rate < acc.model.Groups[j].Rate
		})
		acc.model.EndpointTypes = make([]string, 0, len(acc.endpointTypes))
		for endpointType := range acc.endpointTypes {
			acc.model.EndpointTypes = append(acc.model.EndpointTypes, endpointType)
		}
		sort.Strings(acc.model.EndpointTypes)
		result = append(result, acc.model)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Provider == result[j].Provider {
			return result[i].ID < result[j].ID
		}
		return result[i].Provider < result[j].Provider
	})

	response.Success(c, modelMarketCatalogResponse{Models: result})
}

func (h *ModelMarketHandler) availableModelsForGroup(ctx context.Context, group *service.Group) []string {
	available := h.gatewayService.GetAvailableModels(ctx, &group.ID, group.Platform)
	if group.CustomModelsListEnabled() {
		return filterModelsByCustomList(available, defaultModelIDsForPlatform(group.Platform), group.ModelsListConfig.Models)
	}
	if len(available) > 0 {
		return available
	}
	return defaultModelIDsForPlatform(group.Platform)
}

func (h *ModelMarketHandler) modelTokenPrice(modelID string) *modelMarketTokenPrice {
	pricing, err := h.billingService.GetModelPricing(modelID)
	if err != nil || pricing == nil {
		return nil
	}
	return &modelMarketTokenPrice{
		Input:     pricing.InputPricePerToken * 1_000_000,
		Output:    pricing.OutputPricePerToken * 1_000_000,
		CacheRead: pricing.CacheReadPricePerToken * 1_000_000,
	}
}

func multiplyModelMarketPrice(price *modelMarketTokenPrice, rate float64) *modelMarketTokenPrice {
	if price == nil {
		return nil
	}
	return &modelMarketTokenPrice{
		Input:     price.Input * rate,
		Output:    price.Output * rate,
		CacheRead: price.CacheRead * rate,
	}
}

func modelMarketImagePrices(group *service.Group) map[string]float64 {
	allowedTiers := service.NormalizeImageAllowedTiers(group.ImageAllowedTiers, group.AllowImageGeneration)
	allowed := make(map[string]struct{}, len(allowedTiers))
	for _, tier := range allowedTiers {
		allowed[strings.ToUpper(strings.TrimSpace(tier))] = struct{}{}
	}
	prices := make(map[string]float64)
	add := func(tier string) {
		if _, ok := allowed[tier]; !ok {
			return
		}
		price, ok := service.ImageGenerationUnitPrice(tier)
		if ok {
			prices[tier] = price
		}
	}
	add("1K")
	add("2K")
	add("4K")
	return prices
}

func uniqueModelIDs(modelIDs []string) []string {
	seen := make(map[string]struct{}, len(modelIDs))
	result := make([]string, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			continue
		}
		if _, exists := seen[modelID]; exists {
			continue
		}
		seen[modelID] = struct{}{}
		result = append(result, modelID)
	}
	return result
}

func isImageGenerationModel(modelID string) bool {
	lower := strings.ToLower(modelID)
	return strings.Contains(lower, "image") && !strings.Contains(lower, "vision")
}

func modelMarketProvider(modelID, groupPlatform string) string {
	lower := strings.ToLower(modelID)
	switch {
	case strings.HasPrefix(lower, "claude-") || strings.Contains(lower, "fable"):
		return service.PlatformAnthropic
	case strings.HasPrefix(lower, "gemini-"):
		return service.PlatformGemini
	case strings.HasPrefix(lower, "grok-"):
		return service.PlatformGrok
	case strings.HasPrefix(lower, "gpt-") || strings.HasPrefix(lower, "o1") || strings.HasPrefix(lower, "o3"):
		return service.PlatformOpenAI
	default:
		return groupPlatform
	}
}

func modelMarketProviderLabel(provider string) string {
	switch provider {
	case service.PlatformAnthropic, service.PlatformAntigravity:
		return "Anthropic"
	case service.PlatformGemini:
		return "Google"
	case service.PlatformGrok:
		return "xAI"
	default:
		return "OpenAI"
	}
}

func modelMarketEndpointType(platform string) string {
	switch platform {
	case service.PlatformAnthropic, service.PlatformAntigravity:
		return "/v1/messages"
	case service.PlatformGemini:
		return "/v1beta/models"
	default:
		return "/v1/responses"
	}
}

func modelMarketTags(modelID, billing string) []string {
	if billing == "request" {
		return []string{"image", "vision"}
	}
	tags := []string{"reasoning", "tools"}
	lower := strings.ToLower(modelID)
	if strings.Contains(lower, "claude") || strings.Contains(lower, "gpt") || strings.Contains(lower, "gemini") {
		tags = append(tags, "files")
	}
	return tags
}

func modelMarketDisplayName(modelID string) string {
	for _, model := range openai.DefaultModels {
		if model.ID == modelID {
			return model.DisplayName
		}
	}
	for _, model := range claude.DefaultModels {
		if model.ID == modelID {
			return model.DisplayName
		}
	}
	for _, model := range geminicli.DefaultModels {
		if model.ID == modelID {
			return model.DisplayName
		}
	}
	for _, model := range antigravity.DefaultModels() {
		if model.ID == modelID {
			return model.DisplayName
		}
	}
	for _, model := range xai.DefaultModels() {
		if model.ID == modelID {
			return model.DisplayName
		}
	}
	return modelID
}
