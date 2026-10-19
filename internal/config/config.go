package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type AppConfig struct {
	ProjectsPath string          `json:"projects_path"`
	RecentPaths  []string        `json:"recent_paths,omitempty"`
	DefaultPHP   string          `json:"default_php"`
	DomainSuffix string          `json:"domain_suffix"`
	Services     map[string]bool `json:"services"`
}

func NewDefaultConfig() *AppConfig {
	defaultPath := "C:\\laragal\\sites"
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		defaultPath = filepath.Join(home, "laragal", "sites")
	}
	return &AppConfig{
		ProjectsPath: defaultPath,
		RecentPaths:  []string{defaultPath},
		DefaultPHP:   "8.3",
		DomainSuffix: ".test",
		Services: map[string]bool{
			"caddy":   true,
			"php":     true,
			"mysql":   true,
			"redis":   false,
			"mailpit": true,
		},
	}
}

func LoadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *AppConfig) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
