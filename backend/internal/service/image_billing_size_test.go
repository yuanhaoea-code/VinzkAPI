package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyImageBillingTier(t *testing.T) {
	tests := []struct {
		name     string
		size     string
		wantTier string
		wantOK   bool
	}{
		{name: "explicit 2k square", size: "2048x2048", wantTier: "2K", wantOK: true},
		{name: "explicit 2k landscape", size: "2048x1152", wantTier: "2K", wantOK: true},
		{name: "explicit 4k landscape", size: "3840x2160", wantTier: "4K", wantOK: true},
		{name: "explicit 4k portrait", size: "2160x3840", wantTier: "4K", wantOK: true},
		{name: "long edge 1k", size: "1024X768", wantTier: "1K", wantOK: true},
		{name: "long edge 2k", size: "1280x768", wantTier: "2K", wantOK: true},
		{name: "long edge 4k", size: "2560x1600", wantTier: "4K", wantOK: true},
		{name: "tier string 1k", size: "1k", wantTier: "1K", wantOK: true},
		{name: "empty", size: "", wantOK: false},
		{name: "auto", size: "auto", wantOK: false},
		{name: "invalid", size: "not-a-size", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTier, gotOK := ClassifyImageBillingTier(tt.size)
			require.Equal(t, tt.wantOK, gotOK)
			require.Equal(t, tt.wantTier, gotTier)
		})
	}
}

func TestResolveImageBillingSize(t *testing.T) {
	tests := []struct {
		name          string
		inputSize     string
		outputSizes   []string
		wantBilling   string
		wantOutput    string
		wantSource    string
		wantBreakdown map[string]int
	}{
		{
			name:          "output wins over input",
			inputSize:     "1024x1024",
			outputSizes:   []string{"3840x2160"},
			wantBilling:   "4K",
			wantOutput:    "3840x2160",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"4K": 1},
		},
		{
			name:        "input fallback",
			inputSize:   "1024x1024",
			wantBilling: "1K",
			wantSource:  ImageSizeSourceInput,
		},
		{
			name:        "auto defaults",
			inputSize:   "auto",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:        "empty defaults",
			inputSize:   "",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:        "invalid defaults",
			inputSize:   "largest",
			wantBilling: "2K",
			wantSource:  ImageSizeSourceDefault,
		},
		{
			name:          "mixed output chooses highest tier",
			inputSize:     "1024x1024",
			outputSizes:   []string{"1024x1024", "3840x2160", "1280x720"},
			wantBilling:   "4K",
			wantOutput:    "1024x1024",
			wantSource:    ImageSizeSourceOutput,
			wantBreakdown: map[string]int{"1K": 1, "2K": 1, "4K": 1},
		},
		{
			name:        "unparseable output falls back to parseable input",
			inputSize:   "2048x1152",
			outputSizes: []string{"auto"},
			wantBilling: "2K",
			wantOutput:  "auto",
			wantSource:  ImageSizeSourceInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveImageBillingSize(tt.inputSize, tt.outputSizes)
			require.Equal(t, tt.wantBilling, got.BillingSize)
			require.Equal(t, tt.inputSize, got.InputSize)
			require.Equal(t, tt.wantOutput, got.OutputSize)
			require.Equal(t, tt.wantSource, got.Source)
			require.Equal(t, tt.wantBreakdown, got.Breakdown)
		})
	}
}

func TestResolveRequestedImageBilling(t *testing.T) {
	tests := []struct {
		name          string
		requestedTier string
		imageCount    int
		resolved      ImageBillingSizeResolution
		wantTier      string
		wantSource    string
		wantBreakdown map[string]int
	}{
		{
			name:          "requested 4k charges lower returned 1k output",
			requestedTier: ImageBillingSize4K,
			imageCount:    1,
			resolved: ImageBillingSizeResolution{
				BillingSize: ImageBillingSize1K,
				Source:      ImageSizeSourceOutput,
				Breakdown:   map[string]int{ImageBillingSize1K: 1},
			},
			wantTier:      ImageBillingSize1K,
			wantSource:    ImageSizeSourceOutputDowngrade,
			wantBreakdown: map[string]int{ImageBillingSize1K: 1},
		},
		{
			name:          "requested 1k caps unexpected returned 4k output",
			requestedTier: ImageBillingSize1K,
			imageCount:    1,
			resolved: ImageBillingSizeResolution{
				BillingSize: ImageBillingSize4K,
				Source:      ImageSizeSourceOutput,
				Breakdown:   map[string]int{ImageBillingSize4K: 1},
			},
			wantTier:      ImageBillingSize1K,
			wantSource:    ImageSizeSourceRequested,
			wantBreakdown: map[string]int{ImageBillingSize1K: 1},
		},
		{
			name:          "mixed output charges each image without exceeding request",
			requestedTier: ImageBillingSize4K,
			imageCount:    3,
			resolved: ImageBillingSizeResolution{
				BillingSize: ImageBillingSize4K,
				Source:      ImageSizeSourceOutput,
				Breakdown: map[string]int{
					ImageBillingSize1K: 1,
					ImageBillingSize2K: 1,
					ImageBillingSize4K: 1,
				},
			},
			wantTier:   ImageBillingSize4K,
			wantSource: ImageSizeSourceRequested,
			wantBreakdown: map[string]int{
				ImageBillingSize1K: 1,
				ImageBillingSize2K: 1,
				ImageBillingSize4K: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRequestedImageBilling(tt.requestedTier, tt.imageCount, tt.resolved)
			require.Equal(t, tt.wantTier, got.BillingSize)
			require.Equal(t, tt.wantSource, got.Source)
			require.Equal(t, tt.wantBreakdown, got.Breakdown)
		})
	}
}

func TestGroupAllowsImageTier(t *testing.T) {
	group := &Group{
		Status:               "active",
		AllowImageGeneration: true,
		ImageAllowedTiers:    []string{ImageBillingSize2K, ImageBillingSize4K},
	}

	require.False(t, group.AllowsImageTier(ImageBillingSize1K))
	require.True(t, group.AllowsImageTier(ImageBillingSize2K))
	require.True(t, group.AllowsImageTier(ImageBillingSize4K))
	require.False(t, group.AllowsImageTier("8K"))
}

func TestGroupAllowsImageTierLegacyDefaultsTo1K(t *testing.T) {
	group := &Group{
		Status:               "active",
		AllowImageGeneration: true,
	}

	require.True(t, group.AllowsImageTier(ImageBillingSize1K))
	require.False(t, group.AllowsImageTier(ImageBillingSize2K))
}

func TestGroupAllowsImageTierUngroupedKeyDefaultsTo1K(t *testing.T) {
	require.True(t, GroupAllowsImageTier(nil, ImageBillingSize1K))
	require.False(t, GroupAllowsImageTier(nil, ImageBillingSize2K))
	require.False(t, GroupAllowsImageTier(nil, ImageBillingSize4K))
}

func TestNormalizeImageAllowedTiersKeepsStandardAndHDPoolsExclusive(t *testing.T) {
	require.Equal(t,
		[]string{ImageBillingSize2K, ImageBillingSize4K},
		NormalizeImageAllowedTiers([]string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K}, true),
	)
	require.Equal(t,
		[]string{ImageBillingSize1K},
		NormalizeImageAllowedTiers([]string{ImageBillingSize1K}, true),
	)
}

func TestImageGenerationUnitPrice(t *testing.T) {
	tests := map[string]float64{
		ImageBillingSize1K: ImageGenerationPrice1K,
		ImageBillingSize2K: ImageGenerationPrice2K,
		ImageBillingSize4K: ImageGenerationPrice4K,
	}
	for tier, expected := range tests {
		price, ok := ImageGenerationUnitPrice(tier)
		require.True(t, ok)
		require.InDelta(t, expected, price, 1e-12)
	}
	_, ok := ImageGenerationUnitPrice("8K")
	require.False(t, ok)
}
