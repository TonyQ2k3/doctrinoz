package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigSettersDoNotWriteAPIKey(t *testing.T) {
	tests := []struct {
		name         string
		set          func() error
		wantProvider string
		wantModel    string
	}{
		{
			name: "provider",
			set: func() error {
				return SetProvider("openrouter")
			},
			wantProvider: "openrouter",
			wantModel:    "existing-model",
		},
		{
			name: "model",
			set: func() error {
				return SetModel("new-model")
			},
			wantProvider: "existing-provider",
			wantModel:    "new-model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			testDir := t.TempDir()
			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Errorf("restoring working directory: %v", err)
				}
			})

			viper.Reset()
			t.Cleanup(viper.Reset)
			configPath := filepath.Join(testDir, ".doc.yaml")
			if err := os.WriteFile(configPath, []byte(
				"provider: existing-provider\nmodel: existing-model\napi_key: existing-key\nmax_tokens: 512\n",
			), 0o600); err != nil {
				t.Fatal(err)
			}
			viper.SetConfigFile(configPath)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatal(err)
			}
			viper.Set("api_key", "runtime-key")

			if err := tt.set(); err != nil {
				t.Fatal(err)
			}
			if got := viper.GetString("api_key"); got != "runtime-key" {
				t.Errorf("runtime api_key = %q, want %q", got, "runtime-key")
			}

			content, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(content), "api_key") || strings.Contains(string(content), "runtime-key") {
				t.Fatalf("config file contains an API key: %s", content)
			}

			writtenConfig := viper.New()
			writtenConfig.SetConfigFile(configPath)
			if err := writtenConfig.ReadInConfig(); err != nil {
				t.Fatal(err)
			}
			if got := writtenConfig.GetString("provider"); got != tt.wantProvider {
				t.Errorf("provider = %q, want %q", got, tt.wantProvider)
			}
			if got := writtenConfig.GetString("model"); got != tt.wantModel {
				t.Errorf("model = %q, want %q", got, tt.wantModel)
			}
			if got := writtenConfig.GetInt("max_tokens"); got != 512 {
				t.Errorf("max_tokens = %d, want 512", got)
			}
		})
	}
}
