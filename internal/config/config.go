package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type Config struct {
	Provider string
	Model    string
	APIKey   string
	Verbose  bool

	// generation settings
	MaxTokens       int
	ConcurrentPages int
	IgnorePatterns  []string

	// Provider-specific base URLs
	BaseURL string
}

// Map each provider to its default
var ProviderDefaults = map[string]struct {
	Model   string
	BaseURL string
	EnvKey  string
}{
	"openrouter": {
		Model:   "openrouter/free",
		BaseURL: "https://openrouter.ai/api/v1",
		EnvKey:  "",
	},
}

// Init is called by cobra
func Init(cfgFile string) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(".doc")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("DOCTRINOZ")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Ignore missing configs
	_ = viper.ReadInConfig()
}

// Resolves the final cofnig from Viper (file/env/flags)
func Load() (*Config, error) {
	provider := viper.GetString("provider")
	if provider == "" {
		provider = "openrouter"
	}

	defaults, ok := ProviderDefaults[provider]
	if !ok {
		return nil, fmt.Errorf("Unknown provider: %s", provider)
	}

	model := viper.GetString("model")
	if model == "" {
		model = defaults.Model
	}

	apiKey := viper.GetString("api_key")
	if apiKey == "" && defaults.EnvKey == "" {
		apiKey = os.Getenv("DOCTRINOZ_API_KEY")
	}

	baseURL := viper.GetString("base_url")
	if baseURL == "" {
		baseURL = defaults.BaseURL
	}

	maxTokens := viper.GetInt("max_tokens")
	if maxTokens == 0 {
		maxTokens = 1024
	}

	concurrency := viper.GetInt("concurrent_pages")
	if concurrency == 0 {
		concurrency = 3
	}

	ignorePatterns := viper.GetStringSlice("ignore_patterns")
	if len(ignorePatterns) == 0 {
		ignorePatterns = defaultIgnorePatterns()
	}

	return &Config{
		Provider:        provider,
		Model:           model,
		APIKey:          apiKey,
		BaseURL:         baseURL,
		Verbose:         viper.GetBool("verbose"),
		MaxTokens:       maxTokens,
		ConcurrentPages: concurrency,
		IgnorePatterns:  ignorePatterns,
	}, nil
}

func defaultIgnorePatterns() []string {
	return []string{
		"node_modules/", ".git/", "dist/", "build/", ".doc.yaml",
	}
}

// SetKey persist API kay to the config file
func SetKey(provider, key string) error {
	if _, ok := ProviderDefaults[provider]; !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	viper.Set(fmt.Sprintf("keys.%s", provider), key)
	path := ".doc.yaml"

	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("Writing config: %w", err)
	}

	color.Green("API key for %s saved to %s", provider, path)
	return nil
}

func (c *Config) Print() {
	bold := color.New(color.Bold)
	bold.Println("Resolved configuration: ")
	fmt.Printf("Provider     : %s\n", c.Provider)
	fmt.Printf("Model        : %s\n", c.Model)
	fmt.Printf("Base URL     : %s\n", c.BaseURL)
	fmt.Printf("Max Tokens   : %d\n", c.MaxTokens)
	fmt.Printf("Concurrenncy : %d\n", c.ConcurrentPages)
	key := c.APIKey

	if len(key) > 8 {
		key = key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
	} else if len(key) > 0 {
		key = "********"
	} else {
		key = "(not set)"
	}
	fmt.Printf("API Key      : %s\n", key)
}
