package service

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageGenerationParseResponsePrefersUrlField(t *testing.T) {
	mimeType, outFmt, images, upstreamRequestID, err := parseImageGenerationResponse([]byte(`{"id":"req_123","data":[{"url":"https://example.com/generated.png","output_format":"png"}]}`), "png")
	require.NoError(t, err)
	require.Equal(t, "image/png", mimeType)
	require.Equal(t, "png", outFmt)
	require.Equal(t, "req_123", upstreamRequestID)
	require.Len(t, images, 1)
	require.Equal(t, "https://example.com/generated.png", images[0])
}

func TestImageGenerationResolveImagePayloadDownloadsRemoteURL(t *testing.T) {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-bytes"))
	}))
	defer upstream.Close()

	svc := &ImageGenerationService{httpClient: upstream.Client()}
	data, mimeType, err := svc.resolveImagePayload(context.Background(), upstream.URL, "image/png")
	require.NoError(t, err)
	require.Equal(t, []byte("png-bytes"), data)
	require.Equal(t, "image/png", mimeType)
}

func TestImageGenerationResolveImagePayloadSupportsDataURL(t *testing.T) {
	svc := &ImageGenerationService{}
	data, mimeType, err := svc.resolveImagePayload(context.Background(), "data:image/png;base64,UE5H", "image/png")
	require.NoError(t, err)
	require.Equal(t, []byte("PNG"), data)
	require.Equal(t, "image/png", mimeType)
}

func TestImageGenerationPersistImagesSupportsRemoteURL(t *testing.T) {
	t.Helper()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("remote-image"))
	}))
	defer upstream.Close()

	tmp := t.TempDir()
	svc := &ImageGenerationService{
		httpClient:  upstream.Client(),
		storageRoot: tmp,
	}

	resp := []byte(`{"request_id":"req_abc","data":[{"url":"` + upstream.URL + `","output_format":"png"}]}`)
	images, size, upstreamRequestID, err := svc.persistImages(context.Background(), 10, 20, resp, "png")
	require.NoError(t, err)
	require.Equal(t, "req_abc", upstreamRequestID)
	require.Len(t, images, 1)
	require.Equal(t, int64(len("remote-image")), size)
	if assertFileExists := func(path string) {
		_, err := os.Stat(path)
		require.NoError(t, err)
	}; assertFileExists != nil {
		assertFileExists(filepath.Join(tmp, "20", "10", "0.png"))
	}
}

func TestImageGenerationResolveImagePayloadRejectsInvalidBase64(t *testing.T) {
	svc := &ImageGenerationService{}
	_, _, err := svc.resolveImagePayload(context.Background(), "not-base64", "image/png")
	require.Error(t, err)
}

func TestImageGenerationResolveImagePayloadDataURLRoundTrip(t *testing.T) {
	data := []byte("hello")
	encoded := base64.StdEncoding.EncodeToString(data)
	svc := &ImageGenerationService{}
	got, mimeType, err := svc.resolveImagePayload(context.Background(), "data:image/png;base64,"+encoded, "image/png")
	require.NoError(t, err)
	require.Equal(t, data, got)
	require.Equal(t, "image/png", mimeType)
}

func TestFriendlyImageGatewayErrorExplainsProxyFakeIP(t *testing.T) {
	message := friendlyImageGatewayError(http.StatusBadGateway, `{"error":{"message":"upstream request failed: resolved ip 198.18.0.37 is not allowed"}}`)
	require.Contains(t, message, "fake-ip-filter")
	require.Contains(t, message, "安全校验已拦截")
}

func TestFriendlyImageGatewayTransportErrorExplainsProxyFakeIP(t *testing.T) {
	message := friendlyImageGatewayTransportError(errors.New("upstream request failed: resolved ip 198.18.0.37 is not allowed"))
	require.Contains(t, message, "fake-ip-filter")
	require.Contains(t, message, "安全校验已拦截")
}
