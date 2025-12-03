package config

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DefaultTool  string   `yaml:"default_tool"`
	CustomTools  []string `yaml:"custom_tools"`
	FavoriteTags []string `yaml:"favorite_tags"`
	OutputDir    string   `yaml:"output_dir"` // defaults to "crumbs"
}

// builtInTools is the hardcoded list of built-in tools
var builtInTools = []string{
	"Claude Code",
	"Cursor",
	"Kiro",
	"ChatGPT",
	"Copilot",
	"Warp AI",
	"Windsurf",
	"Aider",
	"Gemini",
	"Perplexity",
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultTool:  "Claude Code",
		CustomTools:  []string{},
		FavoriteTags: []string{},
		OutputDir:    "crumbs",
	}
}

// Load loads configuration from XDG config path, returns defaults if not found
func Load() (*Config, error) {
	configPath, err := xdg.ConfigFile("crumb/config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	// if config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// apply defaults for empty fields
	if cfg.DefaultTool == "" {
		cfg.DefaultTool = "Claude Code"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "crumbs"
	}
	if cfg.CustomTools == nil {
		cfg.CustomTools = []string{}
	}
	if cfg.FavoriteTags == nil {
		cfg.FavoriteTags = []string{}
	}

	return &cfg, nil
}

// GetAllTools returns the combined list of built-in and custom tools
func GetAllTools(cfg *Config) []string {
	if cfg == nil {
		return builtInTools
	}

	allTools := make([]string, 0, len(builtInTools)+len(cfg.CustomTools))
	allTools = append(allTools, builtInTools...)
	allTools = append(allTools, cfg.CustomTools...)

	return allTools
}
