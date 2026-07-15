package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateImageGenerationFixedCost(t *testing.T) {
	svc := &BillingService{}

	tests := []struct {
		tier  string
		count int
		want  float64
	}{
		{tier: ImageBillingSize1K, count: 2, want: 0.12},
		{tier: ImageBillingSize2K, count: 3, want: 0.48},
		{tier: ImageBillingSize4K, count: 4, want: 0.80},
	}

	for _, tt := range tests {
		cost := svc.CalculateImageGenerationFixedCost(tt.tier, tt.count)
		require.InDelta(t, tt.want, cost.TotalCost, 1e-12)
		require.InDelta(t, tt.want, cost.ActualCost, 1e-12)
		require.Equal(t, string(BillingModeImage), cost.BillingMode)
	}
}

func TestOpenAIImageStudioBillingIgnoresGroupPriceAndMultiplier(t *testing.T) {
	customPrice := 9.99
	svc := &OpenAIGatewayService{billingService: &BillingService{}}
	cost := svc.calculateOpenAIImageTierCost(
		context.Background(),
		"gpt-image-2",
		&APIKey{Group: &Group{
			AllowImageGeneration: true,
			ImagePrice2K:         &customPrice,
		}},
		ImageBillingSize2K,
		2,
		8,
	)

	require.InDelta(t, 0.32, cost.TotalCost, 1e-12)
	require.InDelta(t, 0.32, cost.ActualCost, 1e-12)
}

func TestApplyFixedImageGenerationPricing(t *testing.T) {
	group := &Group{
		AllowImageGeneration: true,
		RateMultiplier:       9,
		ImageRateMultiplier:  7,
	}

	ApplyFixedImageGenerationPricing(group)

	require.True(t, group.ImageRateIndependent)
	require.Equal(t, 1.0, group.ImageRateMultiplier)
	require.InDelta(t, ImageGenerationPrice1K, *group.ImagePrice1K, 1e-12)
	require.InDelta(t, ImageGenerationPrice2K, *group.ImagePrice2K, 1e-12)
	require.InDelta(t, ImageGenerationPrice4K, *group.ImagePrice4K, 1e-12)
}
