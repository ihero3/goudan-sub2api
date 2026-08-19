package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestSanitizeOpenAIResponsesInputItemIDs_ReplacesLegacyIDs(t *testing.T) {
	body, changed, err := sanitizeOpenAIResponsesInputItemIDs([]byte(`{
		"model":"gpt-5.4",
		"input":[
			{"type":"message","id":"item_msg_old","role":"assistant","content":[{"type":"output_text","text":"hi"}]},
			{"type":"function_call","id":"item_fc_old","call_id":"call_1","name":"exec","arguments":"{}"},
			{"type":"function_call_output","call_id":"call_1","output":"ok"},
			{"type":"reasoning","id":"item_rs_old","summary":[{"type":"summary_text","text":"plan"}]},
			{"type":"item_reference","id":"item_fc_old"}
		]
	}`))
	require.NoError(t, err)
	require.True(t, changed)

	msgID := gjson.GetBytes(body, "input.0.id").String()
	fcID := gjson.GetBytes(body, "input.1.id").String()
	rsID := gjson.GetBytes(body, "input.3.id").String()
	refID := gjson.GetBytes(body, "input.4.id").String()

	require.Truef(t, strings.HasPrefix(msgID, "msg_"), "message id must start with msg_, got %q", msgID)
	require.Truef(t, strings.HasPrefix(fcID, "fc_"), "function_call id must start with fc_, got %q", fcID)
	require.Truef(t, strings.HasPrefix(rsID, "rs_"), "reasoning id must start with rs_, got %q", rsID)
	require.Equal(t, fcID, refID)
	require.Equal(t, "call_1", gjson.GetBytes(body, "input.1.call_id").String())
}

func TestSanitizeOpenAIResponsesInputItemIDs_KeepsValidIDs(t *testing.T) {
	original := []byte(`{
		"model":"gpt-5.4",
		"input":[
			{"type":"message","id":"msg_1","role":"assistant","content":[{"type":"output_text","text":"hi"}]},
			{"type":"function_call","id":"fc_1","call_id":"call_1","name":"exec","arguments":"{}"},
			{"type":"reasoning","id":"rs_1","summary":[{"type":"summary_text","text":"plan"}]}
		]
	}`)

	body, changed, err := sanitizeOpenAIResponsesInputItemIDs(original)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, string(original), string(body))
}

func TestSanitizeOpenAIResponsesInputItemIDs_StripsReasoningContent(t *testing.T) {
	body, changed, err := sanitizeOpenAIResponsesInputItemIDs([]byte(`{
		"model":"deepseek-v4",
		"input":[
			{"type":"message","id":"msg_1","role":"user","content":[{"type":"input_text","text":"hello"}]},
			{"type":"reasoning","id":"rs_1","content":[{"type":"reasoning_text","text":"thinking step"}],"summary":[{"type":"summary_text","text":"thinking step"}]}
		]
	}`))
	require.NoError(t, err)
	require.True(t, changed)
	// reasoning item must have no "content" field
	require.False(t, gjson.GetBytes(body, "input.1.content").Exists(), "reasoning input item should have content stripped")
	// summary must be preserved
	require.True(t, gjson.GetBytes(body, "input.1.summary").Exists(), "reasoning input item should keep summary")
	require.Equal(t, "thinking step", gjson.GetBytes(body, "input.1.summary.0.text").String())
}

func TestSanitizeOpenAIResponsesInputItemIDs_KeepsNonReasoningContent(t *testing.T) {
	body, changed, err := sanitizeOpenAIResponsesInputItemIDs([]byte(`{
		"model":"deepseek-v4",
		"input":[
			{"type":"message","id":"msg_1","role":"user","content":[{"type":"input_text","text":"hello"}]}
		]
	}`))
	require.NoError(t, err)
	require.False(t, changed)
	// message item's content must be preserved
	require.True(t, gjson.GetBytes(body, "input.0.content").Exists(), "message content should be preserved")
}
