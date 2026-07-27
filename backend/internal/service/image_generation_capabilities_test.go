package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type imageCapabilityKeySource struct {
	keys []APIKey
}

func (f imageCapabilityKeySource) GetByID(context.Context, int64) (*APIKey, error)   { return nil, nil }
func (f imageCapabilityKeySource) GetByKey(context.Context, string) (*APIKey, error) { return nil, nil }
func (f imageCapabilityKeySource) List(context.Context, int64, pagination.PaginationParams, APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	return f.keys, &pagination.PaginationResult{Total: int64(len(f.keys)), Page: 1, PageSize: len(f.keys)}, nil
}

func TestListImageGenerationCapabilitiesUsesEffectiveGroupTiers(t *testing.T) {
	standardGroup := &Group{ID: 10, Name: "standard", Status: StatusActive, AllowImageGeneration: true, ImageAllowedTiers: []string{"1K"}}
	hdGroup := &Group{ID: 11, Name: "hd", Status: StatusActive, AllowImageGeneration: true, ImageAllowedTiers: []string{"1K", "2K", "4K"}}
	disabledGroup := &Group{ID: 12, Name: "disabled", Status: StatusActive, AllowImageGeneration: false}
	source := imageCapabilityKeySource{keys: []APIKey{
		{ID: 1, Name: "standard-key", Status: StatusAPIKeyActive, GroupID: &standardGroup.ID, Group: standardGroup},
		{ID: 2, Name: "hd-key", Status: StatusAPIKeyActive, GroupID: &hdGroup.ID, Group: hdGroup},
		{ID: 3, Name: "disabled-key", Status: StatusAPIKeyActive, GroupID: &disabledGroup.ID, Group: disabledGroup},
		{ID: 4, Name: "inactive-key", Status: StatusAPIKeyDisabled, GroupID: &standardGroup.ID, Group: standardGroup},
		{ID: 5, Name: "ungrouped-key", Status: StatusAPIKeyActive},
	}}
	svc := &ImageGenerationService{apiKeyRepo: source}

	capabilities, err := svc.ListImageGenerationCapabilities(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, []int64{1, 5}, capabilities.Tiers[ImageBillingSize1K])
	require.Equal(t, []int64{2}, capabilities.Tiers[ImageBillingSize2K])
	require.Equal(t, []int64{2}, capabilities.Tiers[ImageBillingSize4K])
	require.True(t, capabilities.Keys[0].Available)
	require.Equal(t, []string{ImageBillingSize2K, ImageBillingSize4K}, capabilities.Keys[1].AllowedTiers)
	require.False(t, capabilities.Keys[2].Available)
	require.NotEmpty(t, capabilities.Keys[2].UnavailableReason)
	require.False(t, capabilities.Keys[3].Available)
	require.Equal(t, "standard-key", capabilities.Keys[0].KeyName)
}
