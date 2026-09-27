package ports

import "context"

type Profile struct {
	Name      string `yaml:"name" json:"name"`
	Token     string `yaml:"token,omitempty" json:"token,omitempty"`
	TokenEnv  string `yaml:"token_env,omitempty" json:"token_env,omitempty"`
	AccountID string `yaml:"account_id,omitempty" json:"account_id,omitempty"`
	ZoneID    string `yaml:"zone_id,omitempty" json:"zone_id,omitempty"`
}

type Config struct {
	DefaultProfile string             `yaml:"default_profile" json:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles" json:"profiles"`
	LogLevel       string             `yaml:"log_level" json:"log_level"`
	LogPath        string             `yaml:"log_path" json:"log_path"`
}

type ConfigRepository interface {
	Load(ctx context.Context) (Config, error)
}
