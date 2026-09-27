package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/sekai-labs/kumo/internal/core/ports"
)

type ConfigRepository struct {
	customPath string
}

var _ ports.ConfigRepository = (*ConfigRepository)(nil)

func NewConfigRepository(customPath string) *ConfigRepository {
	return &ConfigRepository{
		customPath: customPath,
	}
}

func CandidateConfigPaths() []string {
	var paths []string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "kumo", "config.yaml"))
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		paths = append(paths, filepath.Join(home, ".config", "kumo", "config.yaml"))
	}

	paths = append(paths, ".kumo.yaml")

	return paths
}

func (r *ConfigRepository) FindConfigFile() string {
	if r.customPath != "" {
		if _, err := os.Stat(r.customPath); err == nil {
			return r.customPath
		}
		return ""
	}

	for _, path := range CandidateConfigPaths() {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func (r *ConfigRepository) Load(ctx context.Context) (ports.Config, error) {
	if err := ctx.Err(); err != nil {
		return ports.Config{}, err
	}

	var cfg ports.Config
	filePath := r.FindConfigFile()

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return ports.Config{}, fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}

		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return ports.Config{}, fmt.Errorf("failed to parse config file %s: %w", filePath, err)
		}
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]ports.Profile)
	}

	envToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	for name, prof := range cfg.Profiles {
		prof.Name = name
		if prof.Token == "" && prof.TokenEnv != "" {
			prof.Token = os.Getenv(prof.TokenEnv)
		}
		if prof.Token == "" && envToken != "" {
			prof.Token = envToken
		}
		cfg.Profiles[name] = prof
	}

	if len(cfg.Profiles) == 0 && envToken != "" {
		cfg.DefaultProfile = "default"
		cfg.Profiles["default"] = ports.Profile{
			Name:  "default",
			Token: envToken,
		}
	}

	if cfg.DefaultProfile == "" && len(cfg.Profiles) == 1 {
		for name := range cfg.Profiles {
			cfg.DefaultProfile = name
		}
	}

	if strings.TrimSpace(cfg.LogLevel) == "" {
		cfg.LogLevel = "info"
	}

	return cfg, nil
}
