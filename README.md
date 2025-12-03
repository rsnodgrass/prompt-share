# crumb

[![Release](https://img.shields.io/github/v/release/rsnodgrass/crumb)](https://github.com/rsnodgrass/crumb/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/rsnodgrass/crumb/build.yml)](https://github.com/rsnodgrass/crumb/actions)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**leave crumbs for your teammates**

A beautiful TUI for capturing and sharing AI prompts - create a collaborative learning trail in your repo.

![crumb TUI](assets/demo.gif)

## Design Philosophy

crumb is built for **minimal friction**:
- Launch, type, save in under 10 seconds
- Auto-generated titles and timestamps
- Smart defaults from config
- Keyboard-first navigation

## Features

- **Capture-as-you-go** - Minimal friction prompt capture workflow
- **Beautiful TUI** - Catppuccin-inspired terminal interface
- **10+ AI tools** - Claude Code, Cursor, Kiro, ChatGPT, Copilot, Warp AI, Windsurf, Aider, Gemini, Perplexity
- **Auto-generated metadata** - Timestamp, author (from git), title
- **Smart tag suggestions** - Quick-select favorites with number keys (1-5)
- **README generation** - Auto-generate prompt index for discovery

## Installation

### macOS & Linux (Homebrew)

```bash
brew install rsnodgrass/tap/crumb
```

### Windows (Scoop)

```bash
scoop bucket add rsnodgrass https://github.com/rsnodgrass/scoop-bucket
scoop install crumb
```

### Go

```bash
go install github.com/rsnodgrass/crumb/cmd/crumb@latest
```

### Binary Downloads

Download pre-built binaries from [Releases](https://github.com/rsnodgrass/crumb/releases).

### From Source

```bash
git clone https://github.com/rsnodgrass/crumb.git
cd crumb
make install
```

## Quick Start

```bash
# Initialize crumbs directory in your repo
crumb init

# Capture a new prompt
crumb

# Generate README index of all prompts
crumb readme
```

## Usage

```bash
crumb              # Launch TUI to capture a prompt
crumb init         # Create crumbs/ directory
crumb readme       # Generate/update prompt index
crumb config       # Open config in $EDITOR
crumb -t Cursor    # Override default tool
crumb --stay       # Capture multiple prompts
crumb -v           # Show version
```

## Configuration

Config file: `~/.config/crumb/config.yaml`

```yaml
default_tool: Claude Code
custom_tools:
  - Internal GPT
  - Custom Model
favorite_tags:
  - debugging
  - design
  - refactoring
output_dir: crumbs
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Ctrl+S` | Save and exit |
| `Esc` | Cancel and exit |
| `/` | Open tool selector |
| `?` | Show help |

## License

[MIT](LICENSE)
