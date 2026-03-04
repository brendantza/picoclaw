package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicoption "github.com/anthropics/anthropic-sdk-go/option"

	anthropicprovider "github.com/sipeed/picoclaw/pkg/providers/anthropic"
)

func TestClaudeProvider_ChatRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var reqBody map[string]any
		json.NewDecoder(r.Body).Decode(&reqBody)

		// Return SSE streaming format
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		// Send message_start event
		event1 := map[string]any{
			"type": "message_start",
			"message": map[string]any{
				"id":           "msg_test",
				"type":         "message",
				"role":         "assistant",
				"model":        reqBody["model"],
				"content":      []map[string]any{},
				"stop_reason":  nil,
				"usage":        map[string]any{"input_tokens": 15, "output_tokens": 0},
			},
		}
		data1, _ := json.Marshal(event1)
		fmt.Fprintf(w, "event: message_start\ndata: %s\n\n", data1)
		w.(http.Flusher).Flush()

		// Send content_block_start event
		event2 := map[string]any{
			"type":          "content_block_start",
			"index":         0,
			"content_block": map[string]any{"type": "text", "text": ""},
		}
		data2, _ := json.Marshal(event2)
		fmt.Fprintf(w, "event: content_block_start\ndata: %s\n\n", data2)
		w.(http.Flusher).Flush()

		// Send content_block_delta events
		text := "Hello! How can I help you?"
		for i := 0; i < len(text); i += 5 {
			end := i + 5
			if end > len(text) {
				end = len(text)
			}
			event := map[string]any{
				"type":  "content_block_delta",
				"index": 0,
				"delta": map[string]any{"type": "text_delta", "text": text[i:end]},
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: content_block_delta\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()
		}

		// Send content_block_stop event
		event4 := map[string]any{
			"type":  "content_block_stop",
			"index": 0,
		}
		data4, _ := json.Marshal(event4)
		fmt.Fprintf(w, "event: content_block_stop\ndata: %s\n\n", data4)
		w.(http.Flusher).Flush()

		// Send message_delta event
		event5 := map[string]any{
			"type": "message_delta",
			"delta": map[string]any{
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
			},
			"usage": map[string]any{"output_tokens": 8},
		}
		data5, _ := json.Marshal(event5)
		fmt.Fprintf(w, "event: message_delta\ndata: %s\n\n", data5)
		w.(http.Flusher).Flush()

		// Send message_stop event
		event6 := map[string]any{"type": "message_stop"}
		data6, _ := json.Marshal(event6)
		fmt.Fprintf(w, "event: message_stop\ndata: %s\n\n", data6)
	}))
	defer server.Close()

	delegate := anthropicprovider.NewProviderWithClient(createAnthropicTestClient(server.URL, "test-token"))
	provider := newClaudeProviderWithDelegate(delegate)

	messages := []Message{{Role: "user", Content: "Hello"}}
	resp, err := provider.Chat(t.Context(), messages, nil, "claude-sonnet-4.6", map[string]any{"max_tokens": 1024})
	if err != nil {
		t.Fatalf("Chat() error: %v", err)
	}
	if resp.Content != "Hello! How can I help you?" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello! How can I help you?")
	}
	if resp.FinishReason != "stop" {
		t.Errorf("FinishReason = %q, want %q", resp.FinishReason, "stop")
	}
	if resp.Usage.PromptTokens != 15 {
		t.Errorf("PromptTokens = %d, want 15", resp.Usage.PromptTokens)
	}
}

func TestClaudeProvider_GetDefaultModel(t *testing.T) {
	p := NewClaudeProvider("test-token")
	if got := p.GetDefaultModel(); got != "claude-sonnet-4.6" {
		t.Errorf("GetDefaultModel() = %q, want %q", got, "claude-sonnet-4.6")
	}
}

func createAnthropicTestClient(baseURL, token string) *anthropic.Client {
	c := anthropic.NewClient(
		anthropicoption.WithAuthToken(token),
		anthropicoption.WithBaseURL(baseURL),
	)
	return &c
}
