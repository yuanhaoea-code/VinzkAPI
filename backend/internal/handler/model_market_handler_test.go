package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestModelMarketProviderClassifiesFableAsAnthropic(t *testing.T) {
	require.Equal(t, service.PlatformAnthropic, modelMarketProvider("claude-fable-5", service.PlatformOpenAI))
	require.Equal(t, "Anthropic", modelMarketProviderLabel(service.PlatformAnthropic))
}

func TestMultiplyModelMarketPriceUsesEffectiveGroupRate(t *testing.T) {
	price := multiplyModelMarketPrice(&modelMarketTokenPrice{Input: 5, Output: 30, CacheRead: 0.5}, 0.2)
	require.Equal(t, &modelMarketTokenPrice{Input: 1, Output: 6, CacheRead: 0.1}, price)
}

func TestModelMarketImagePricesUsesAllowedFixedTiers(t *testing.T) {
	customPrice := 9.99
	prices := modelMarketImagePrices(&service.Group{
		ImageAllowedTiers: []string{"2K", "4K"},
		ImagePrice1K:      &customPrice,
		ImagePrice2K:      &customPrice,
		ImagePrice4K:      &customPrice,
	})
	require.Equal(t, map[string]float64{"2K": 0.16, "4K": 0.20}, prices)
}

func TestModelMarketDisplayNameUsesCanonicalCatalog(t *testing.T) {
	require.Equal(t, "Claude Fable 5", modelMarketDisplayName("claude-fable-5"))
	require.Equal(t, "GPT-5.6 Sol", modelMarketDisplayName("gpt-5.6-sol"))
}
