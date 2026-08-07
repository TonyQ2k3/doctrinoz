package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/TonyQ2k3/doctrinoz/internal/config"
)

type OpenRouterProvider struct {
	cfg    *config.Config
	client *http.Client
}

func NewOpenRouter(cfg config.Config) *OpenRouterProvider {
	return &OpenRouterProvider{
		cfg:    &cfg,
		client: &http.Client{},
	}
}

func (p *OpenRouterProvider) Name() string { return "OpenRouter" }

type openRouterRequest struct {
	Model       string          `json:"model"`
	Messages    []openRouterMsg `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`

	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (p *OpenRouterProvider) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	if p.cfg.APIKey == "" {
		return "", fmt.Errorf("OpenRouter API key not set - use --api-key, REPOWIKIAPIKEY, or OPENROUTER")
	}

	msgs := make([]openRouterMsg, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openRouterMsg{Role: m.Role, Content: m.Content}
	}

	body := openRouterRequest{
		Model:       p.cfg.Model,
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/TonyQ2k3/doctrinoz")
	httpReq.Header.Set("X-Title", "doctrinoz")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenRouter API error %d: %s", resp.StatusCode, rawBody)
	}

	var result openRouterResponse
	if err := json.Unmarshal(rawBody, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("OpenRouter error %d: %s", result.Error.Code, result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return result.Choices[0].Message.Content, nil
}
