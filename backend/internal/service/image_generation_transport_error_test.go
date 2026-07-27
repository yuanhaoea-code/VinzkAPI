package service

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFriendlyImageGatewayErrorExplainsEOF(t *testing.T) {
	message := friendlyImageGatewayError(http.StatusBadGateway, `{"error":{"message":"upstream request failed: Post https://sub.geiliapi.com/v1/images/generations: EOF"}}`)
	require.Contains(t, message, "EOF")
	require.Contains(t, message, "检查上游记录")
}

func TestFriendlyImageGatewayTransportErrorExplainsEOF(t *testing.T) {
	message := friendlyImageGatewayTransportError(errors.New("upstream request failed: unexpected EOF"))
	require.Contains(t, message, "EOF")
	require.Contains(t, message, "确认没有创建生成任务后再重试")
}
