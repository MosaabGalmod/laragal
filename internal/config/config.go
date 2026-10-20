package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectCustomSettings struct {
	Type       string `json:"type,omitempty"`
	DocRoot    string `json:"doc_root,omitempty"`
	ProxyPort  int    `json:"proxy_port,omitempty"`
	PHPVersion string `json:"php_version,omitempty"`
	IsIgnored  bool   `json:"is_ignored,omitempty"`
}

type AppConfig struct {
	ProjectsPath     string                           `json:"projects_path"`
	RecentPaths      []string                         `json:"recent_paths,omitempty"`
	DefaultPHP       string                           `json:"default_php"`
	DomainSuffix     string                           `json:"domain_suffix"`
	Services         map[string]bool                  `json:"services"`
	ProjectPHP       map[string]string                `json:"project_php,omitempty"`
	ProjectOverrides map[string]ProjectCustomSettings `json:"project_overrides,omitempty"`
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
		ProjectPHP:       make(map[string]string),
		ProjectOverrides: make(map[string]ProjectCustomSettings),
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
	if cfg.ProjectPHP == nil {
		cfg.ProjectPHP = make(map[string]string)
	}
	if cfg.ProjectOverrides == nil {
		cfg.ProjectOverrides = make(map[string]ProjectCustomSettings)
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
