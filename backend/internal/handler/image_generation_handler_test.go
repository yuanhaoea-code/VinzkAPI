package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMapImageGenerationDoesNotExposeStoragePath(t *testing.T) {
	record := &service.ImageGeneration{
		ID: 1,
		Images: []service.ImageGenerationImage{
			{Index: 0, Path: "/var/lib/sub2api/private/0.png", URL: "/preview/0", MimeType: "image/png"},
		},
	}

	response := mapImageGeneration(record)

	require.Len(t, response.Images, 1)
	require.Empty(t, response.Images[0].Path)
	require.Equal(t, "/preview/0", response.Images[0].URL)
	require.Equal(t, "/var/lib/sub2api/private/0.png", record.Images[0].Path)
}

func TestImageGenerationPricingReturnsFixedTiers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	handler := &ImageGenerationHandler{}

	handler.Pricing(ctx)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Data struct {
			Currency string             `json:"currency"`
			Tiers    map[string]float64 `json:"tiers"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "USD", body.Data.Currency)
	require.InDelta(t, 0.06, body.Data.Tiers[service.ImageBillingSize1K], 1e-12)
	require.InDelta(t, 0.16, body.Data.Tiers[service.ImageBillingSize2K], 1e-12)
	require.InDelta(t, 0.20, body.Data.Tiers[service.ImageBillingSize4K], 1e-12)
}
