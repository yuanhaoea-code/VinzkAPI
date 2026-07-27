package service

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

// ImageGenerationKeyCapability is the public, non-secret capability view of a
// user's API key. The raw key value is intentionally never included here.
type ImageGenerationKeyCapability struct {
	APIKeyID          int64    `json:"api_key_id"`
	KeyName           string   `json:"key_name"`
	GroupID           *int64   `json:"group_id,omitempty"`
	GroupName         string   `json:"group_name,omitempty"`
	Status            string   `json:"status"`
	Available         bool     `json:"available"`
	AllowedTiers      []string `json:"allowed_tiers"`
	UnavailableReason string   `json:"unavailable_reason,omitempty"`
}

// ImageGenerationCapabilities is the single source of truth used by the
// image-generation selector and by the service-side authorization check.
type ImageGenerationCapabilities struct {
	Keys  []ImageGenerationKeyCapability `json:"keys"`
	Tiers map[string][]int64             `json:"tiers"`
}

// ListImageGenerationCapabilities returns all user keys with their effective
// image-generation capability. Inactive or misconfigured keys remain visible
// with an explanation so the UI can tell the user why a key is unavailable.
func (s *ImageGenerationService) ListImageGenerationCapabilities(ctx context.Context, userID int64) (*ImageGenerationCapabilities, error) {
	result := &ImageGenerationCapabilities{
		Keys: make([]ImageGenerationKeyCapability, 0),
		Tiers: map[string][]int64{
			ImageBillingSize1K: {},
			ImageBillingSize2K: {},
			ImageBillingSize4K: {},
		},
	}
	if s.apiKeyRepo == nil {
		return result, nil
	}

	keys, _, err := s.apiKeyRepo.List(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 500}, APIKeyListFilters{})
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		capability := ImageGenerationKeyCapability{
			APIKeyID:     key.ID,
			KeyName:      key.Name,
			GroupID:      key.GroupID,
			Status:       key.Status,
			AllowedTiers: []string{},
		}
		if key.Group != nil {
			capability.GroupName = key.Group.Name
		}

		switch {
		case !key.IsActive():
			capability.UnavailableReason = "API key is not active"
		case key.Group != nil && !key.Group.IsActive():
			capability.UnavailableReason = "group is not active"
		case key.Group != nil && !key.Group.AllowImageGeneration:
			capability.UnavailableReason = "image generation is disabled for this group"
		default:
			if key.Group == nil {
				capability.AllowedTiers = []string{ImageBillingSize1K}
			} else {
				capability.AllowedTiers = NormalizeImageAllowedTiers(key.Group.ImageAllowedTiers, true)
			}
			if len(capability.AllowedTiers) == 0 {
				capability.UnavailableReason = "no image resolution is enabled for this group"
			}
		}

		capability.Available = capability.UnavailableReason == "" && len(capability.AllowedTiers) > 0
		if capability.Available {
			for _, tier := range capability.AllowedTiers {
				result.Tiers[tier] = append(result.Tiers[tier], capability.APIKeyID)
			}
		}
		result.Keys = append(result.Keys, capability)
	}
	return result, nil
}
