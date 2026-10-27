package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// Resolves the final config from Viper (file/env/flags)
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
	if apiKey == "" {
		apiKey = os.Getenv("DOCTRINOZ_API_KEY")
	}
	if apiKey == "" {
		var err error
		apiKey, err = readSecretKey(provider)
		if err != nil {
			return nil, err
		}
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
		"node_modules/", ".git/", "dist/", "build/", ".doc.yaml", ".venv", ".env", ".cache",
	}
}

func secretConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving user home directory: %w", err)
	}
	return filepath.Join(home, ".doctrinoz", "secret.yaml"), nil
}

func newSecretConfig(path string) *viper.Viper {
	secretConfig := viper.New()
	secretConfig.SetConfigFile(path)
	secretConfig.SetConfigType("yaml")
	secretConfig.SetConfigPermissions(0o600)
	return secretConfig
}

// EnsureSecretConfig creates the user's secret config file if it does not exist.
func EnsureSecretConfig() (string, error) {
	path, err := secretConfigPath()
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("creating secret config directory %s: %w", dir, err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0o700); err != nil {
			return "", fmt.Errorf("setting permissions on secret config directory %s: %w", dir, err)
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return "", fmt.Errorf("creating secret config %s: %w", path, err)
	}
	if err == nil {
		if err := file.Close(); err != nil {
			return "", fmt.Errorf("closing secret config %s: %w", path, err)
		}
	}

	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("checking secret config %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("secret config path %s is not a regular file", path)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o600); err != nil {
			return "", fmt.Errorf("setting permissions on secret config %s: %w", path, err)
		}
	}
	return path, nil
}

func readSecretKey(provider string) (string, error) {
	path, err := secretConfigPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("checking secret config %s: %w", path, err)
	}

	secretConfig := newSecretConfig(path)
	if err := secretConfig.ReadInConfig(); err != nil {
		return "", fmt.Errorf("reading secret config %s: %w", path, err)
	}
	return secretConfig.GetString(fmt.Sprintf("keys.%s", provider)), nil
}

// SetKey persists an API key in the user's secret config file.
func SetKey(provider, key string) error {
	if _, ok := ProviderDefaults[provider]; !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	path, err := secretConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating secret config directory %s: %w", dir, err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("setting permissions on secret config directory %s: %w", dir, err)
	}

	secretConfig := newSecretConfig(path)
	if _, err := os.Stat(path); err == nil {
		if err := secretConfig.ReadInConfig(); err != nil {
			return fmt.Errorf("reading secret config %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking secret config %s: %w", path, err)
	}

	secretConfig.Set(fmt.Sprintf("keys.%s", provider), key)
	if err := secretConfig.WriteConfigAs(path); err != nil {
		return fmt.Errorf("writing secret config %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("setting permissions on secret config %s: %w", path, err)
	}

	color.Green("API key for %s saved to %s", provider, path)
	return nil
}

func SetProvider(providerName string) error {
	if _, ok := ProviderDefaults[providerName]; !ok {
		return fmt.Errorf(`unknown provider: %s
accepted providers are:
  - openrouter
  - openai
  - anthropic`, providerName)
	}

	path := ".doc.yaml"
	viper.Set("provider", providerName)

	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("Writing config: %w", err)
	}

	color.Green("Provider saved to %s", path)
	return nil
}

func SetModel(modelName string) error {
	path := ".doc.yaml"
	viper.Set("model", modelName)

	if err := viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("Writing config: %w", err)
	}

	color.Green("Model saved to %s", path)
	return nil
}

func (c *Config) Print() {
	bold := color.New(color.Bold)
	bold.Println("Resolved configuration: ")
	fmt.Printf("Provider     : %s\n", c.Provider)
	fmt.Printf("Model        : %s\n", c.Model)
	fmt.Printf("Base URL     : %s\n", c.BaseURL)
	fmt.Printf("Max Tokens   : %d\n", c.MaxTokens)
	fmt.Printf("Concurrency : %d\n", c.ConcurrentPages)
	key := c.APIKey

	if len(key) > 8 {
		key = key[:4] + "********" + key[len(key)-4:]
	} else if len(key) > 0 {
		key = "********"
	} else {
		key = "(not set)"
	}
	fmt.Printf("API Key      : %s\n", key)
}
