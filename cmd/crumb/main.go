package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"crumb/internal/config"
	"crumb/internal/readme"
	"crumb/internal/tui"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// define flags
	var (
		toolFlag    string
		titleFlag   string
		stayFlag    bool
		versionFlag bool
		helpFlag    bool
	)

	flag.StringVar(&toolFlag, "tool", "", "override default tool for this session")
	flag.StringVar(&toolFlag, "t", "", "override default tool for this session (shorthand)")
	flag.StringVar(&titleFlag, "title", "", "set title for the prompt")
	flag.BoolVar(&stayFlag, "stay", false, "don't exit after save (capture multiple prompts)")
	flag.BoolVar(&versionFlag, "version", false, "show version")
	flag.BoolVar(&versionFlag, "v", false, "show version (shorthand)")
	flag.BoolVar(&helpFlag, "help", false, "show help")
	flag.BoolVar(&helpFlag, "h", false, "show help (shorthand)")

	flag.Usage = printUsage
	flag.Parse()

	// handle version flag
	if versionFlag {
		fmt.Printf("crumb version %s\n", version)
		return nil
	}

	// handle help flag
	if helpFlag {
		printUsage()
		return nil
	}

	// get command (first non-flag argument)
	args := flag.Args()
	command := ""
	if len(args) > 0 {
		command = args[0]
	}

	// load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// handle commands
	switch command {
	case "":
		// default: launch TUI
		return runTUI(cfg, toolFlag, titleFlag, stayFlag)
	case "readme":
		return runReadme(cfg)
	case "config":
		return runConfig()
	case "init":
		return runInit(cfg)
	default:
		// check if it's a markdown file
		if strings.HasSuffix(command, ".md") {
			return renderMarkdown(command)
		}
		return fmt.Errorf("unknown command: %s", command)
	}
}

// runTUI launches the TUI to capture a new prompt
func runTUI(cfg *config.Config, toolOverride, titleOverride string, stay bool) error {
	// determine which tool to use
	selectedTool := cfg.DefaultTool
	if toolOverride != "" {
		selectedTool = toolOverride

		// validate that tool is in known tools list
		allTools := config.GetAllTools(cfg)
		found := false
		for _, t := range allTools {
			if t == toolOverride {
				found = true
				break
			}
		}

		if !found {
			// warn but still allow unknown tool
			fmt.Fprintf(os.Stderr, "warning: tool '%s' is not in known tools list (built-in + custom)\n", toolOverride)
			fmt.Fprintf(os.Stderr, "known tools: %v\n", allTools)
			fmt.Fprintf(os.Stderr, "continuing anyway...\n\n")
		}
	}

	// create and run TUI model with tool pre-selected
	model := tui.New(cfg, selectedTool, titleOverride, stay)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}
	return nil
}

// runReadme generates/updates the README.md in the prompts directory
func runReadme(cfg *config.Config) error {
	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// prompts directory is relative to cwd
	promptsDir := filepath.Join(cwd, cfg.OutputDir)

	// check if directory exists
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		return fmt.Errorf("prompts directory does not exist: %s (run 'crumb init' first)", promptsDir)
	}

	// generate README content
	content, err := readme.Generate(promptsDir)
	if err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	// write README.md
	readmePath := filepath.Join(promptsDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	fmt.Printf("generated: %s\n", readmePath)
	return nil
}

// runConfig opens the config file in $EDITOR (or vim if not set)
func runConfig() error {
	// get config file path
	configPath, err := xdg.ConfigFile("crumb/config.yaml")
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}

	// ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := writeDefaultConfig(configPath); err != nil {
			return fmt.Errorf("failed to write default config: %w", err)
		}
	}

	// get editor from environment, default to vim
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// open config in editor
	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	return nil
}

// runInit creates the prompts directory and initial README
func runInit(cfg *config.Config) error {
	// get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// create prompts directory
	promptsDir := filepath.Join(cwd, cfg.OutputDir)
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// create initial README.md
	readmePath := filepath.Join(promptsDir, "README.md")
	initialContent := `# Prompts

A shared collection of AI prompts captured by the team. Learn from each other's techniques, discover effective patterns, and build institutional knowledge around AI-assisted development.

**What is this?** This directory contains prompts saved using [crumb](https://github.com/rsnodgrass/crumb), a tool for capturing and sharing AI prompts across a team.

## Index

| Date | Author | Tool | Tags | Title |
|------|--------|------|------|-------|

---
*Run ` + "`crumb readme`" + ` to regenerate this index.*
`

	if err := os.WriteFile(readmePath, []byte(initialContent), 0644); err != nil {
		return fmt.Errorf("failed to write README: %w", err)
	}

	fmt.Printf("initialized: %s\n", promptsDir)
	fmt.Printf("created: %s\n", readmePath)
	return nil
}

// renderMarkdown renders a markdown file using glamour
func renderMarkdown(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	out, err := renderer.Render(string(content))
	if err != nil {
		return fmt.Errorf("failed to render markdown: %w", err)
	}

	fmt.Print(out)
	return nil
}

// writeDefaultConfig writes a default config file
func writeDefaultConfig(path string) error {
	defaultContent := `# crumb configuration

# default tool to pre-select in the dropdown
default_tool: Claude Code

# custom tools to add to the dropdown (in addition to built-in tools)
custom_tools: []

# favorite tags to suggest when tagging prompts
favorite_tags: []

# output directory for prompts (relative to current working directory)
output_dir: crumbs
`

	return os.WriteFile(path, []byte(defaultContent), 0644)
}

// printUsage prints usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, `crumb - leave crumbs for your teammates

USAGE:
  crumb [command] [flags]

COMMANDS:
  (default)      launch TUI to capture a new prompt
  <file.md>      render markdown file with syntax highlighting
  readme         generate/update crumbs/README.md
  config         open config file in $EDITOR
  init           create crumbs/ directory with starter README

FLAGS:
  -t, --tool <name>    override default tool for this session
  --stay               don't exit after save (capture multiple prompts)
  -v, --version        show version
  -h, --help           show help

EXAMPLES:
  crumb                    # launch TUI
  crumb -t "ChatGPT"       # launch TUI with tool override
  crumb file.md            # render markdown file
  crumb readme             # regenerate README
  crumb config             # edit config
  crumb init               # initialize prompts directory

CONFIG:
  Config file: ~/.config/crumb/config.yaml
  Run 'crumb config' to edit

`)
}
