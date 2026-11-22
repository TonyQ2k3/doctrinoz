package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/TonyQ2k3/doctrinoz/internal/config"
)

type OpenAIProvider struct {
	cfg    *config.Config
	client *http.Client
}

func NewOpenAI(cfg config.Config) *OpenAIProvider {
	return &OpenAIProvider{
		cfg:    &cfg,
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

func (p *OpenAIProvider) Name() string { return "OpenAI" }

type openAIRequest struct {
	Model           string          `json:"model"`
	Input           []openAIMessage `json:"input"`
	MaxOutputTokens int             `json:"max_output_tokens,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Output []struct {
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	if p.cfg.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key not set - use --api-key, OPENAI_API_KEY, or config set --provider openai --api-key <apikey>")
	}

	messages := make([]openAIMessage, len(req.Messages))
	for i, message := range req.Messages {
		messages[i] = openAIMessage{Role: message.Role, Content: message.Content}
	}

	body := openAIRequest{
		Model:           p.cfg.Model,
		Input:           messages,
		MaxOutputTokens: req.MaxTokens,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	endpoint := strings.TrimRight(p.cfg.BaseURL, "/") + "/responses"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
	}

	var result openAIResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	var output strings.Builder
	for _, item := range result.Output {
		if item.Type != "message" || item.Role != "assistant" {
			continue
		}
		for _, content := range item.Content {
			if content.Type == "output_text" {
				output.WriteString(content.Text)
			}
		}
	}
	if output.Len() == 0 {
		return "", fmt.Errorf("no text output in OpenAI response")
	}

	return output.String(), nil
}
