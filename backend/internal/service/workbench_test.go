package service

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkbenchReasoningEffort(t *testing.T) {
	tests := []struct {
		provider string
		preset   string
		want     string
	}{
		{PlatformOpenAI, WorkbenchReasoningFast, "medium"},
		{PlatformOpenAI, WorkbenchReasoningStandard, "high"},
		{PlatformOpenAI, WorkbenchReasoningDeep, "xhigh"},
		{PlatformAnthropic, WorkbenchReasoningFast, "low"},
		{PlatformAnthropic, WorkbenchReasoningStandard, "medium"},
		{PlatformAnthropic, WorkbenchReasoningDeep, "high"},
	}
	for _, test := range tests {
		require.Equal(t, test.want, workbenchReasoningEffort(test.provider, test.preset))
	}
}

func TestWorkbenchReasoningSummary(t *testing.T) {
	require.Equal(t, "concise", workbenchReasoningSummary(WorkbenchReasoningFast))
	require.Equal(t, "detailed", workbenchReasoningSummary(WorkbenchReasoningStandard))
	require.Equal(t, "detailed", workbenchReasoningSummary(WorkbenchReasoningDeep))
}

func TestWorkbenchPlatformDefaultModels(t *testing.T) {
	require.Equal(t, "gpt-5.4-mini", workbenchPlatformDefaultModel(PlatformOpenAI))
	require.Equal(t, "claude-3-5-haiku-20241022", workbenchPlatformDefaultModel(PlatformAnthropic))
	require.Equal(t, "gemini-2.0-flash", workbenchPlatformDefaultModel(PlatformGemini))
}

func TestWorkbenchPickerCatalogsAreCappedAndNewestFirst(t *testing.T) {
	require.LessOrEqual(t, len(workbenchPickerModelsByProvider[PlatformOpenAI]), 8)
	require.Equal(t, []string{
		"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna", "gpt-5.5",
		"gpt-5.4", "gpt-5.4-mini", "gpt-5.3",
	}, workbenchPickerModelsByProvider[PlatformOpenAI])
	require.NotContains(t, workbenchPickerModelsByProvider[PlatformOpenAI], "codex-auto-review")
	require.NotContains(t, workbenchPickerModelsByProvider[PlatformOpenAI], "gpt-5.3-codex-spark")
	require.LessOrEqual(t, len(workbenchPickerModelsByProvider[PlatformAnthropic]), 8)
	require.Contains(t, workbenchPickerModelsByProvider[PlatformAnthropic], "claude-3-5-haiku-20241022")
	require.LessOrEqual(t, len(workbenchPickerModelsByProvider[PlatformGemini]), 8)
	require.Contains(t, workbenchPickerModelsByProvider[PlatformGemini], "gemini-2.0-flash")
}

func TestWorkbenchConversationCatalogRejectsIncompatibleModels(t *testing.T) {
	require.True(t, isWorkbenchConversationModel("gpt-5.3-codex-spark"))
	require.True(t, isWorkbenchConversationModel("gpt-4o-vision-preview"))
	require.False(t, isWorkbenchConversationModel("codex-auto-review"))
	require.False(t, isWorkbenchConversationModel("gpt-4o-audio-preview"))
	require.False(t, isWorkbenchConversationModel("gpt-4o-realtime-preview"))
	require.False(t, isWorkbenchConversationModel("gpt-image-2"))
}

func TestFilterWorkbenchModels(t *testing.T) {
	available := []string{"gpt-5.6-sol", "gpt-5.5", "claude-*"}
	selected := []string{"gpt-5.6-sol", "claude-opus-4-8", "gemini-3", "gpt-5.6-sol"}

	require.Equal(t,
		[]string{"gpt-5.6-sol", "claude-opus-4-8"},
		filterWorkbenchModels(available, nil, selected),
	)
}

func TestWorkbenchStreamParser(t *testing.T) {
	parser := workbenchStreamParser{}
	parser.AddLine([]byte("data: {\"type\":\"response.reasoning_summary_text.delta\",\"delta\":\"先分析\"}\n"))
	parser.AddLine([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"完成\"}\n"))
	parser.AddLine([]byte("data: [DONE]\n"))

	require.Equal(t, "先分析", parser.Reasoning.String())
	require.Equal(t, "完成", parser.Text.String())
	require.False(t, parser.Failed)
	require.True(t, parser.Completed)
}

func TestWorkbenchStreamParserReturnsStandardDeltas(t *testing.T) {
	parser := workbenchStreamParser{}
	reasoningEvents := parser.AddLine([]byte(`data: {"type":"response.reasoning_summary_text.delta","delta":"plan"}`))
	answerEvents := parser.AddLine([]byte(`data: {"type":"response.output_text.delta","delta":"answer"}`))

	require.Equal(t, []workbenchParsedStreamEvent{{Type: WorkbenchStreamThinkingDelta, Delta: "plan"}}, reasoningEvents)
	require.Equal(t, []workbenchParsedStreamEvent{{Type: WorkbenchStreamAnswerDelta, Delta: "answer"}}, answerEvents)
}

func TestWorkbenchStreamParserConvertsCompletedFallbackText(t *testing.T) {
	parser := workbenchStreamParser{}
	events := parser.AddLine([]byte(`data: {"type":"response.completed","response":{"output":[{"content":[{"text":"fallback answer"}]}]}}`))

	require.True(t, parser.Completed)
	require.Equal(t, "fallback answer", parser.Text.String())
	require.Equal(t, []workbenchParsedStreamEvent{{Type: WorkbenchStreamAnswerDelta, Delta: "fallback answer"}}, events)
}

func TestWorkbenchStreamParserRequiresTerminalEvent(t *testing.T) {
	parser := workbenchStreamParser{}
	parser.AddLine([]byte("data: {\"type\":\"response.output_text.delta\",\"delta\":\"部分回答\"}\n"))

	require.Equal(t, "部分回答", parser.Text.String())
	require.False(t, parser.Completed)
}

func TestBuildWorkbenchResponsesBody(t *testing.T) {
	conversation := &WorkbenchConversation{ID: "conversation", ReasoningPreset: WorkbenchReasoningDeep}
	binding := &WorkbenchModelBinding{ModelID: "gpt-5.6-sol", Provider: PlatformOpenAI}
	messages := []WorkbenchMessage{
		{Role: "user", Content: "开始任务", Status: WorkbenchMessageCompleted},
		{Role: "assistant", Content: "上一轮结果", Status: WorkbenchMessageCompleted},
		{Role: "assistant", Content: "pending-placeholder", Status: WorkbenchMessagePending},
	}

	body, err := buildWorkbenchResponsesBody(conversation, binding, messages)
	require.NoError(t, err)

	var payload struct {
		Input []struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"input"`
		Reasoning map[string]string `json:"reasoning"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload.Input, 2)
	require.Equal(t, "开始任务", payload.Input[0].Content[0].Text)
	require.Equal(t, "上一轮结果", payload.Input[1].Content[0].Text)
	require.Equal(t, "xhigh", payload.Reasoning["effort"])
	require.Equal(t, "detailed", payload.Reasoning["summary"])
}

func TestBuildWorkbenchResponsesBodyAlwaysKeepsLatestUserMessage(t *testing.T) {
	conversation := &WorkbenchConversation{ID: "conversation", ReasoningPreset: WorkbenchReasoningStandard}
	binding := &WorkbenchModelBinding{ModelID: "gpt-5.6-sol", Provider: PlatformOpenAI}
	latest := strings.Repeat("新", workbenchMaxHistoryRunes+1)
	messages := []WorkbenchMessage{
		{Role: "user", Content: "较早的问题", Status: WorkbenchMessageCompleted},
		{Role: "assistant", Content: "较早的回答", Status: WorkbenchMessageCompleted},
		{Role: "user", Content: latest, Status: WorkbenchMessageCompleted},
		{Role: "assistant", Status: WorkbenchMessagePending},
	}

	body, err := buildWorkbenchResponsesBody(conversation, binding, messages)
	require.NoError(t, err)

	var payload struct {
		Input []struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"input"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload.Input, 1)
	require.Equal(t, "user", payload.Input[0].Role)
	require.Equal(t, latest, payload.Input[0].Content[0].Text)
}

func TestBuildWorkbenchResponsesBodyIncludesImageAttachments(t *testing.T) {
	conversation := &WorkbenchConversation{ID: "conversation", ReasoningPreset: WorkbenchReasoningStandard}
	binding := &WorkbenchModelBinding{ModelID: "gpt-5.5", Provider: PlatformOpenAI}
	messages := []WorkbenchMessage{
		{
			Role: "user", Content: "describe this", Status: WorkbenchMessageCompleted,
			Attachments: []WorkbenchAttachment{{
				ID: "f595fc23-4147-45cf-b1d9-f6d6325cc1f2", Name: "image.png",
				MIMEType: "image/png", SizeBytes: 5, DataURL: "data:image/png;base64,aGVsbG8=",
			}},
		},
		{Role: "assistant", Status: WorkbenchMessagePending},
	}

	body, err := buildWorkbenchResponsesBody(conversation, binding, messages)
	require.NoError(t, err)

	var payload struct {
		Input []struct {
			Content []struct {
				Type     string `json:"type"`
				Text     string `json:"text"`
				ImageURL string `json:"image_url"`
			} `json:"content"`
		} `json:"input"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload.Input, 1)
	require.Equal(t, "input_text", payload.Input[0].Content[0].Type)
	require.Equal(t, "describe this", payload.Input[0].Content[0].Text)
	require.Equal(t, "input_image", payload.Input[0].Content[1].Type)
	require.Equal(t, "data:image/png;base64,aGVsbG8=", payload.Input[0].Content[1].ImageURL)
}

func TestBuildWorkbenchResponsesBodyIncludesDocumentAttachments(t *testing.T) {
	conversation := &WorkbenchConversation{ID: "conversation", ReasoningPreset: WorkbenchReasoningStandard}
	binding := &WorkbenchModelBinding{ModelID: "gpt-5.5", Provider: PlatformOpenAI}
	messages := []WorkbenchMessage{
		{
			Role: "user", Content: "summarize this", Status: WorkbenchMessageCompleted,
			Attachments: []WorkbenchAttachment{{
				ID: "f595fc23-4147-45cf-b1d9-f6d6325cc1f2", Name: "brief.pdf",
				MIMEType: "application/pdf", SizeBytes: 5, DataURL: "data:application/pdf;base64,aGVsbG8=",
			}},
		},
		{Role: "assistant", Status: WorkbenchMessagePending},
	}

	body, err := buildWorkbenchResponsesBody(conversation, binding, messages)
	require.NoError(t, err)

	var payload struct {
		Input []struct {
			Content []struct {
				Type     string `json:"type"`
				Filename string `json:"filename"`
				FileData string `json:"file_data"`
			} `json:"content"`
		} `json:"input"`
	}
	require.NoError(t, json.Unmarshal(body, &payload))
	require.Len(t, payload.Input, 1)
	require.Equal(t, "input_file", payload.Input[0].Content[1].Type)
	require.Equal(t, "brief.pdf", payload.Input[0].Content[1].Filename)
	require.Equal(t, "data:application/pdf;base64,aGVsbG8=", payload.Input[0].Content[1].FileData)
}

func TestValidateWorkbenchAttachmentsRejectsUnsupportedContent(t *testing.T) {
	_, err := validateWorkbenchAttachments([]WorkbenchAttachment{{
		Name: "archive.exe", MIMEType: "application/octet-stream", DataURL: "data:application/octet-stream;base64,aGVsbG8=",
	}})
	require.ErrorIs(t, err, ErrWorkbenchInvalidInput)
}

func TestValidateWorkbenchAttachmentsAcceptsOfficeAndTextFiles(t *testing.T) {
	attachments, err := validateWorkbenchAttachments([]WorkbenchAttachment{
		{Name: "notes.txt", MIMEType: "text/plain", DataURL: "data:text/plain;base64,aGVsbG8="},
		{Name: "report.xlsx", MIMEType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", DataURL: "data:application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;base64,aGVsbG8="},
	})
	require.NoError(t, err)
	require.Len(t, attachments, 2)
}

func TestValidateWorkbenchAttachmentsRejectsOversizedBase64BeforeDecode(t *testing.T) {
	oversized := strings.Repeat("A", base64.StdEncoding.EncodedLen(workbenchMaxAttachmentBytes)+4)
	_, err := validateWorkbenchAttachments([]WorkbenchAttachment{{
		Name: "oversized.png", MIMEType: "image/png", DataURL: "data:image/png;base64," + oversized,
	}})
	require.ErrorIs(t, err, ErrWorkbenchInvalidInput)
}
