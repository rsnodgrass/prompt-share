package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultTool != "Claude Code" {
		t.Errorf("expected DefaultTool to be 'Claude Code', got '%s'", cfg.DefaultTool)
	}

	if cfg.OutputDir != "crumbs" {
		t.Errorf("expected OutputDir to be 'crumbs', got '%s'", cfg.OutputDir)
	}

	if cfg.CustomTools == nil {
		t.Error("expected CustomTools to be initialized")
	}

	if cfg.FavoriteTags == nil {
		t.Error("expected FavoriteTags to be initialized")
	}
}

func TestGetAllTools(t *testing.T) {
	cfg := &Config{
		CustomTools: []string{"Custom Tool 1", "Custom Tool 2"},
	}

	tools := GetAllTools(cfg)

	expectedBuiltIn := 10
	expectedCustom := 2
	expectedTotal := expectedBuiltIn + expectedCustom

	if len(tools) != expectedTotal {
		t.Errorf("expected %d tools, got %d", expectedTotal, len(tools))
	}

	// check first few built-in tools
	if tools[0] != "Claude Code" {
		t.Errorf("expected first tool to be 'Claude Code', got '%s'", tools[0])
	}

	// check custom tools are at the end
	if tools[len(tools)-2] != "Custom Tool 1" {
		t.Errorf("expected second-to-last tool to be 'Custom Tool 1', got '%s'", tools[len(tools)-2])
	}

	if tools[len(tools)-1] != "Custom Tool 2" {
		t.Errorf("expected last tool to be 'Custom Tool 2', got '%s'", tools[len(tools)-1])
	}
}

func TestGetAllTools_NilConfig(t *testing.T) {
	tools := GetAllTools(nil)

	expectedBuiltIn := 10

	if len(tools) != expectedBuiltIn {
		t.Errorf("expected %d tools, got %d", expectedBuiltIn, len(tools))
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// ensure config file doesn't exist by using a temporary XDG_CONFIG_HOME
	tempDir := t.TempDir()
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// reload xdg paths
	xdg.Reload()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error when config file doesn't exist, got: %v", err)
	}

	// should return default config
	if cfg.DefaultTool != "Claude Code" {
		t.Errorf("expected DefaultTool to be 'Claude Code', got '%s'", cfg.DefaultTool)
	}

	if cfg.OutputDir != "crumbs" {
		t.Errorf("expected OutputDir to be 'crumbs', got '%s'", cfg.OutputDir)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// create a temporary config file
	tempDir := t.TempDir()
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// reload xdg paths
	xdg.Reload()

	configDir := filepath.Join(tempDir, "crumb")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	configContent := `default_tool: "Cursor"
custom_tools:
  - "My Custom Tool"
  - "Another Tool"
favorite_tags:
  - "golang"
  - "testing"
output_dir: "my-prompts"
`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.DefaultTool != "Cursor" {
		t.Errorf("expected DefaultTool to be 'Cursor', got '%s'", cfg.DefaultTool)
	}

	if cfg.OutputDir != "my-prompts" {
		t.Errorf("expected OutputDir to be 'my-prompts', got '%s'", cfg.OutputDir)
	}

	if len(cfg.CustomTools) != 2 {
		t.Errorf("expected 2 custom tools, got %d", len(cfg.CustomTools))
	}

	if len(cfg.FavoriteTags) != 2 {
		t.Errorf("expected 2 favorite tags, got %d", len(cfg.FavoriteTags))
	}
}

func TestLoad_PartialConfig(t *testing.T) {
	// create a temporary config file with only some fields
	tempDir := t.TempDir()
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// reload xdg paths
	xdg.Reload()

	configDir := filepath.Join(tempDir, "crumb")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// only specify custom_tools, leave others empty
	configContent := `custom_tools:
  - "My Tool"
`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// should apply defaults for empty fields
	if cfg.DefaultTool != "Claude Code" {
		t.Errorf("expected DefaultTool to be 'Claude Code', got '%s'", cfg.DefaultTool)
	}

	if cfg.OutputDir != "crumbs" {
		t.Errorf("expected OutputDir to be 'crumbs', got '%s'", cfg.OutputDir)
	}

	if len(cfg.CustomTools) != 1 {
		t.Errorf("expected 1 custom tool, got %d", len(cfg.CustomTools))
	}
}
