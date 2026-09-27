package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRepository_LoadFromCustomFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	content := `
default_profile: work
log_level: debug
profiles:
  work:
    token: direct-token-123
    account_id: acc-1
    zone_id: zone-1
  personal:
    token_env: PERSONAL_CF_TOKEN
    account_id: acc-2
`
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	t.Setenv("PERSONAL_CF_TOKEN", "resolved-env-token-456")

	repo := NewConfigRepository(configPath)
	cfg, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DefaultProfile != "work" {
		t.Errorf("expected default_profile 'work', got '%s'", cfg.DefaultProfile)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log_level 'debug', got '%s'", cfg.LogLevel)
	}

	workProf, ok := cfg.Profiles["work"]
	if !ok || workProf.Token != "direct-token-123" {
		t.Errorf("expected work profile with direct-token-123, got %+v", workProf)
	}

	persProf, ok := cfg.Profiles["personal"]
	if !ok || persProf.Token != "resolved-env-token-456" {
		t.Errorf("expected personal profile with resolved token, got %+v", persProf)
	}
}

func TestConfigRepository_FallbackToEnv(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "global-env-token-789")

	repo := NewConfigRepository("/non/existent/path/config.yaml")
	cfg, err := repo.Load(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DefaultProfile != "default" {
		t.Errorf("expected default profile 'default', got '%s'", cfg.DefaultProfile)
	}

	defProf, ok := cfg.Profiles["default"]
	if !ok || defProf.Token != "global-env-token-789" {
		t.Errorf("expected default profile with global env token, got %+v", defProf)
	}
}
