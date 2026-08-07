package providers

import (
	"context"
	"fmt"

	"github.com/TonyQ2k3/doctrinoz/internal/config"
)

// Message is a simple chat message,
type Message struct {
	Role    string // "system"|"user"|"assistant"
	Content string
}

// CompletionRequest is provider-agnostic,
type CompletionRequest struct {
	Messages  []Message
	MaxTokens int
	// Temperature - lower = more deterministic(good for docs)
	Temperature float64
}

// Provider is the interface every AI backend must satisfy,
type Provider interface {
	// Name returns the human-readable provider name.
	Name() string
	// Complete sends a chat completion request and returns the response text.
	Complete(ctx context.Context, req CompletionRequest) (string, error)
}

// New constructs the right provider from Config.
func New(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "openrouter":
		return NewOpenRouter(*cfg), nil
	// case "openai":
	// 	return NewOpenAI(cfg), nil
	// case "anthropic":
	// 	return NewAnthropic(cfg), nil
	// case "local":
	// 	return NewLocal(cfg), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider)
	}
}
