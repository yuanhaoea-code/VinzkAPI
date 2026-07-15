package service

import (
	"sort"
	"strconv"
	"strings"
)

const (
	ImageBillingSize1K = "1K"
	ImageBillingSize2K = "2K"
	ImageBillingSize4K = "4K"

	ImageGenerationPrice1K = 0.06
	ImageGenerationPrice2K = 0.16
	ImageGenerationPrice4K = 0.20

	ImageSizeSourceOutput          = "output"
	ImageSizeSourceInput           = "input"
	ImageSizeSourceDefault         = "default"
	ImageSizeSourceLegacy          = "legacy"
	ImageSizeSourceRequested       = "requested"
	ImageSizeSourceOutputDowngrade = "output_downgrade"
)

var imageBillingTierOrder = []string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K}

func NormalizeImageBillingTier(tier string) string {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case ImageBillingSize1K:
		return ImageBillingSize1K
	case ImageBillingSize2K:
		return ImageBillingSize2K
	case ImageBillingSize4K:
		return ImageBillingSize4K
	default:
		return ""
	}
}

func NormalizeImageAllowedTiers(tiers []string, allowImageGeneration bool) []string {
	seen := make(map[string]struct{}, len(tiers))
	for _, raw := range tiers {
		if tier := NormalizeImageBillingTier(raw); tier != "" {
			seen[tier] = struct{}{}
		}
	}
	if allowImageGeneration && len(seen) == 0 {
		seen[ImageBillingSize1K] = struct{}{}
	}
	// Standard and HD keys are separate products. An HD group never also
	// grants 1K access, which keeps each key in exactly one frontend pool.
	if _, has2K := seen[ImageBillingSize2K]; has2K {
		delete(seen, ImageBillingSize1K)
	}
	if _, has4K := seen[ImageBillingSize4K]; has4K {
		delete(seen, ImageBillingSize1K)
	}
	out := make([]string, 0, len(seen))
	for _, tier := range imageBillingTierOrder {
		if _, ok := seen[tier]; ok {
			out = append(out, tier)
		}
	}
	return out
}

func ImageGenerationUnitPrice(tier string) (float64, bool) {
	switch NormalizeImageBillingTier(tier) {
	case ImageBillingSize1K:
		return ImageGenerationPrice1K, true
	case ImageBillingSize2K:
		return ImageGenerationPrice2K, true
	case ImageBillingSize4K:
		return ImageGenerationPrice4K, true
	default:
		return 0, false
	}
}

func ImageGenerationFixedPrices() map[string]float64 {
	return map[string]float64{
		ImageBillingSize1K: ImageGenerationPrice1K,
		ImageBillingSize2K: ImageGenerationPrice2K,
		ImageBillingSize4K: ImageGenerationPrice4K,
	}
}

func NormalizeImageAspectRatio(ratio string) string {
	switch strings.TrimSpace(ratio) {
	case "1:1":
		return "1:1"
	case "2:3":
		return "2:3"
	case "3:2":
		return "3:2"
	default:
		return ""
	}
}

func InferImageAspectRatio(size string) string {
	width, height, ok := parseImageBillingDimensions(size)
	if !ok || width == height {
		return "1:1"
	}
	if width < height {
		return "2:3"
	}
	return "3:2"
}

func ResolveImageGenerationSize(tier, aspectRatio string) (string, bool) {
	tier = NormalizeImageBillingTier(tier)
	aspectRatio = NormalizeImageAspectRatio(aspectRatio)
	if tier == "" || aspectRatio == "" {
		return "", false
	}
	sizes := map[string]map[string]string{
		ImageBillingSize1K: {"1:1": "1024x1024", "2:3": "1024x1536", "3:2": "1536x1024"},
		ImageBillingSize2K: {"1:1": "2048x2048", "2:3": "1365x2048", "3:2": "2048x1365"},
		ImageBillingSize4K: {"1:1": "2880x2880", "2:3": "1920x2880", "3:2": "2880x1920"},
	}
	size, ok := sizes[tier][aspectRatio]
	return size, ok
}

type ImageBillingSizeResolution struct {
	BillingSize string
	InputSize   string
	OutputSize  string
	Source      string
	Breakdown   map[string]int
}

func ClassifyImageBillingTier(size string) (string, bool) {
	trimmed := strings.TrimSpace(size)
	normalized := strings.ToLower(trimmed)
	switch normalized {
	case "", "auto":
		return "", false
	case "1k":
		return ImageBillingSize1K, true
	case "2k":
		return ImageBillingSize2K, true
	case "4k":
		return ImageBillingSize4K, true
	case "2048x2048", "2048x1152":
		return ImageBillingSize2K, true
	case "3840x2160", "2160x3840":
		return ImageBillingSize4K, true
	}

	width, height, ok := parseImageBillingDimensions(trimmed)
	if !ok {
		return "", false
	}
	maxEdge := width
	if height > maxEdge {
		maxEdge = height
	}
	switch {
	case maxEdge <= 1024:
		return ImageBillingSize1K, true
	case maxEdge <= 2048:
		return ImageBillingSize2K, true
	default:
		return ImageBillingSize4K, true
	}
}

func NormalizeImageBillingTierOrDefault(size string) string {
	if tier, ok := ClassifyImageBillingTier(size); ok {
		return tier
	}
	return ImageBillingSize2K
}

func ResolveImageBillingSize(inputSize string, outputSizes []string) ImageBillingSizeResolution {
	inputSize = strings.TrimSpace(inputSize)
	outputSizes = compactTrimmedStrings(outputSizes)

	breakdown := map[string]int{}
	outputSize := firstDisplayImageOutputSize(outputSizes)
	outputTier := ""
	for _, output := range outputSizes {
		tier, ok := ClassifyImageBillingTier(output)
		if !ok {
			continue
		}
		breakdown[tier]++
		if imageTierRank(tier) > imageTierRank(outputTier) {
			outputTier = tier
		}
	}
	if outputTier != "" {
		return ImageBillingSizeResolution{
			BillingSize: outputTier,
			InputSize:   inputSize,
			OutputSize:  outputSize,
			Source:      ImageSizeSourceOutput,
			Breakdown:   normalizeImageSizeBreakdown(breakdown),
		}
	}

	if tier, ok := ClassifyImageBillingTier(inputSize); ok {
		return ImageBillingSizeResolution{
			BillingSize: tier,
			InputSize:   inputSize,
			OutputSize:  outputSize,
			Source:      ImageSizeSourceInput,
		}
	}

	return ImageBillingSizeResolution{
		BillingSize: ImageBillingSize2K,
		InputSize:   inputSize,
		OutputSize:  outputSize,
		Source:      ImageSizeSourceDefault,
	}
}

func ApplyOpenAIImageBillingResolution(result *OpenAIForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	inputSize := strings.TrimSpace(result.ImageInputSize)
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	resolved := ResolveImageBillingSize(inputSize, outputSizes)
	resolved = resolveRequestedImageBilling(result.ImageSize, result.ImageCount, resolved)
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func ApplyForwardImageBillingResolution(result *ForwardResult) {
	if result == nil || result.ImageCount <= 0 {
		return
	}
	inputSize := strings.TrimSpace(result.ImageInputSize)
	if inputSize == "" && strings.TrimSpace(result.ImageSize) != ImageBillingSize2K {
		inputSize = strings.TrimSpace(result.ImageSize)
	}
	outputSizes := result.ImageOutputSizes
	if len(outputSizes) == 0 && strings.TrimSpace(result.ImageOutputSize) != "" {
		outputSizes = []string{result.ImageOutputSize}
	}
	resolved := ResolveImageBillingSize(inputSize, outputSizes)
	resolved = resolveRequestedImageBilling(result.ImageSize, result.ImageCount, resolved)
	applyImageBillingResolution(
		&result.ImageSize,
		&result.ImageInputSize,
		&result.ImageOutputSize,
		&result.ImageSizeSource,
		&result.ImageSizeBreakdown,
		resolved,
	)
}

func resolveRequestedImageBilling(requestedTier string, imageCount int, resolved ImageBillingSizeResolution) ImageBillingSizeResolution {
	requestedTier = NormalizeImageBillingTier(requestedTier)
	if requestedTier == "" {
		return resolved
	}
	resolved.Source = ImageSizeSourceRequested
	resolved.BillingSize = requestedTier
	if len(resolved.Breakdown) == 0 {
		if imageCount > 0 {
			resolved.Breakdown = map[string]int{requestedTier: imageCount}
		}
		return resolved
	}

	capped := make(map[string]int, len(resolved.Breakdown))
	highest := ""
	for outputTier, count := range resolved.Breakdown {
		if count <= 0 {
			continue
		}
		billedTier := NormalizeImageBillingTier(outputTier)
		if billedTier == "" {
			continue
		}
		if imageTierRank(billedTier) > imageTierRank(requestedTier) {
			billedTier = requestedTier
		}
		capped[billedTier] += count
		if imageTierRank(billedTier) > imageTierRank(highest) {
			highest = billedTier
		}
	}
	if len(capped) == 0 {
		return resolved
	}
	resolved.Breakdown = normalizeImageSizeBreakdown(capped)
	if imageTierRank(highest) < imageTierRank(requestedTier) {
		resolved.BillingSize = highest
		resolved.Source = ImageSizeSourceOutputDowngrade
	}
	return resolved
}

func applyImageBillingResolution(
	billingSize *string,
	inputSize *string,
	outputSize *string,
	source *string,
	breakdown *map[string]int,
	resolved ImageBillingSizeResolution,
) {
	*billingSize = resolved.BillingSize
	*inputSize = resolved.InputSize
	*outputSize = resolved.OutputSize
	*source = resolved.Source
	*breakdown = resolved.Breakdown
}

func parseImageBillingDimensions(size string) (int, int, bool) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(size)), "x")
	if len(parts) != 2 {
		return 0, 0, false
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, false
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, false
	}
	if width <= 0 || height <= 0 {
		return 0, 0, false
	}
	return width, height, true
}

func compactTrimmedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func firstDisplayImageOutputSize(outputSizes []string) string {
	for _, output := range outputSizes {
		if trimmed := strings.TrimSpace(output); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func imageTierRank(tier string) int {
	switch strings.ToUpper(strings.TrimSpace(tier)) {
	case ImageBillingSize1K:
		return 1
	case ImageBillingSize2K:
		return 2
	case ImageBillingSize4K:
		return 3
	default:
		return 0
	}
}

func normalizeImageSizeBreakdown(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for _, tier := range []string{ImageBillingSize1K, ImageBillingSize2K, ImageBillingSize4K} {
		if count := in[tier]; count > 0 {
			out[tier] = count
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func SortedImageBillingBreakdownKeys(breakdown map[string]int) []string {
	keys := make([]string, 0, len(breakdown))
	for key := range breakdown {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, right := imageTierRank(keys[i]), imageTierRank(keys[j])
		if left == right {
			return keys[i] < keys[j]
		}
		return left < right
	})
	return keys
}
