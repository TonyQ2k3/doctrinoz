package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TonyQ2k3/doctrinoz/internal/config"
)

func TestOpenAICompleteUsesResponsesAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" {
			t.Errorf("request = %s %s, want POST /v1/responses", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want Bearer test-key", got)
		}

		var body struct {
			Model           string          `json:"model"`
			Input           []openAIMessage `json:"input"`
			MaxOutputTokens int             `json:"max_output_tokens"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decoding request: %v", err)
		}
		if body.Model != "gpt-test" {
			t.Errorf("model = %q, want gpt-test", body.Model)
		}
		if body.MaxOutputTokens != 42 {
			t.Errorf("max_output_tokens = %d, want 42", body.MaxOutputTokens)
		}
		if len(body.Input) != 2 || body.Input[0] != (openAIMessage{Role: "system", Content: "instructions"}) ||
			body.Input[1] != (openAIMessage{Role: "user", Content: "prompt"}) {
			t.Errorf("input = %#v, want system and user messages", body.Input)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output": [
				{"type":"reasoning","summary":[]},
				{"type":"message","role":"assistant","content":[
					{"type":"output_text","text":"first "},
					{"type":"output_text","text":"second"}
				]}
			]
		}`))
	}))
	defer server.Close()

	provider := NewOpenAI(config.Config{
		APIKey:  "test-key",
		Model:   "gpt-test",
		BaseURL: server.URL + "/v1/",
	})
	got, err := provider.Complete(context.Background(), CompletionRequest{
		Messages: []Message{
			{Role: "system", Content: "instructions"},
			{Role: "user", Content: "prompt"},
		},
		MaxTokens: 42,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != "first second" {
		t.Errorf("Complete() = %q, want %q", got, "first second")
	}
}

func TestOpenAICompleteReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"invalid key"}}`))
	}))
	defer server.Close()

	provider := NewOpenAI(config.Config{
		APIKey:  "bad-key",
		Model:   "gpt-test",
		BaseURL: server.URL,
	})
	_, err := provider.Complete(context.Background(), CompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "invalid key") {
		t.Fatalf("Complete() error = %v, want 401 invalid key error", err)
	}
}

func TestOpenAICompleteRequiresAPIKey(t *testing.T) {
	provider := NewOpenAI(config.Config{BaseURL: "https://api.openai.com/v1"})
	_, err := provider.Complete(context.Background(), CompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "OpenAI API key not set") {
		t.Fatalf("Complete() error = %v, want missing API key error", err)
	}
}

func TestOpenAICompleteRequiresTextOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"output":[{"type":"reasoning"}]}`))
	}))
	defer server.Close()

	provider := NewOpenAI(config.Config{
		APIKey:  "test-key",
		Model:   "gpt-test",
		BaseURL: server.URL,
	})
	_, err := provider.Complete(context.Background(), CompletionRequest{})
	if err == nil || !strings.Contains(err.Error(), "no text output") {
		t.Fatalf("Complete() error = %v, want no text output error", err)
	}
}
