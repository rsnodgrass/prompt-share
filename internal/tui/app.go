package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"crumb/internal/config"
	"crumb/internal/storage"
	"crumb/internal/tui/components"
)

// App is a simple wrapper to match the expected interface from main.go
type App struct{}

func NewApp() *App {
	return &App{}
}

func (a *App) Run() error {
	fmt.Println("crumb TUI initialized")
	fmt.Println("TODO: integrate bubbletea Model")
	return nil
}

// catppuccin mocha color palette
var (
	baseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cdd6f4"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cba6f7")).
			Bold(true)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89dceb")).
			Bold(true)

	focusedLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f9e2af")).
				Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c7086")).
			Italic(true)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#45475a")).
			Padding(1, 2)
)

type Model struct {
	prompt     textarea.Model
	tags       components.TagInput
	output     textarea.Model
	toolSelect components.Dropdown

	focusIndex int  // 0=prompt, 1=output, 2=tool, 3=tags
	showHelp   bool
	showToast  bool
	toastMsg   string
	isError    bool

	config   *config.Config
	storage  *storage.MarkdownStorage
	stayOpen bool // --stay flag
	width    int
	height   int
}

func New(cfg *config.Config, tool string, stay bool) Model {
	// initialize prompt textarea (focused first - most important field)
	promptTA := textarea.New()
	promptTA.Placeholder = "Enter your prompt here..."
	promptTA.CharLimit = 10000
	promptTA.SetHeight(8)
	promptTA.ShowLineNumbers = false
	promptTA.Focus()

	// initialize storage with output directory from config
	// resolve relative path to absolute path based on cwd
	cwd, _ := os.Getwd()
	outputDir := cfg.OutputDir
	if !filepath.IsAbs(outputDir) {
		outputDir = filepath.Join(cwd, outputDir)
	}
	markdownStorage := storage.NewMarkdownStorage(outputDir)

	// build tag suggestions: config favorites + frequent tags from existing crumbs
	tagSuggestions := mergeTagSuggestions(cfg.FavoriteTags, markdownStorage.GetFrequentTags(10))

	// initialize tags input with merged suggestions
	tagsInput := components.NewTagInput(tagSuggestions)

	// initialize output textarea
	outputTA := textarea.New()
	outputTA.Placeholder = "LLM output (optional)"
	outputTA.CharLimit = 50000
	outputTA.SetHeight(6)
	outputTA.ShowLineNumbers = false

	// initialize tool dropdown
	allTools := config.GetAllTools(cfg)
	defaultIdx := 0
	for i, t := range allTools {
		if t == tool {
			defaultIdx = i
			break
		}
	}
	toolDropdown := components.NewDropdown(allTools, defaultIdx, tool)

	return Model{
		prompt:     promptTA,
		tags:       tagsInput,
		output:     outputTA,
		toolSelect: toolDropdown,
		focusIndex: 0,
		showHelp:   false,
		showToast:  false,
		toastMsg:   "",
		isError:    false,
		config:     cfg,
		storage:    markdownStorage,
		stayOpen:   stay,
		width:      80,
		height:     24,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

type saveSuccessMsg struct {
	filename string
}

type saveErrorMsg struct {
	err error
}

type quitAfterDelayMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case saveSuccessMsg:
		m.showToast = true
		m.isError = false
		m.toastMsg = "Saved!"

		if m.stayOpen {
			// clear fields and return focus to prompt
			m.clearFields()
			return m, HideToastAfter(2 * time.Second)
		}

		// exit after brief delay
		return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return quitAfterDelayMsg{}
		})

	case saveErrorMsg:
		m.showToast = true
		m.isError = true
		m.toastMsg = "Error: " + msg.err.Error()
		return m, HideToastAfter(3 * time.Second)

	case quitAfterDelayMsg:
		return m, tea.Quit

	case ToastHideMsg:
		m.showToast = false
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// dynamically size textareas based on available height
		m.updateTextareaSizes()
		return m, nil

	case tea.KeyMsg:
		// if help is showing, toggle on '?' or close on any other key
		if m.showHelp {
			if msg.String() == "?" {
				m.showHelp = false
			} else {
				m.showHelp = false
			}
			return m, nil
		}

		// global key bindings
		switch msg.String() {
		case "ctrl+c", "ctrl+d":
			return m, tea.Quit

		case "esc":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "ctrl+s":
			return m, m.saveAndExit()

		case "/", "ctrl+t":
			m.setFocus(2) // tool selector
			return m, nil

		case "tab":
			m.focusNext()
			return m, nil

		case "shift+tab":
			m.focusPrev()
			return m, nil
		}
	}

	// update focused component
	switch m.focusIndex {
	case 0: // prompt
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		cmds = append(cmds, cmd)

	case 1: // output
		var cmd tea.Cmd
		m.output, cmd = m.output.Update(msg)
		cmds = append(cmds, cmd)

	case 2: // tool selector
		var cmd tea.Cmd
		m.toolSelect, cmd = m.toolSelect.Update(msg)
		cmds = append(cmds, cmd)

	case 3: // tags
		var cmd tea.Cmd
		m.tags, cmd = m.tags.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	// header
	b.WriteString(titleStyle.Render("crumb"))
	b.WriteString("  ")
	b.WriteString(helpStyle.Render(fmt.Sprintf("→ %s/", m.config.OutputDir)))
	b.WriteString("\n\n")

	// prompt field (index 0) - first and most important
	label := labelStyle.Render("Prompt:")
	if m.focusIndex == 0 {
		label = focusedLabelStyle.Render("→ Prompt:")
	}
	b.WriteString(label + "\n")
	b.WriteString(m.prompt.View())
	b.WriteString("\n\n")

	// output field (index 1)
	label = labelStyle.Render("Paste Output:")
	if m.focusIndex == 1 {
		label = focusedLabelStyle.Render("→ Paste Output:")
	}
	b.WriteString(label + " ")
	b.WriteString(helpStyle.Render("(optional)"))
	b.WriteString("\n")
	b.WriteString(m.output.View())
	b.WriteString("\n\n")

	// tool selector (index 2)
	label = labelStyle.Render("Tool:")
	if m.focusIndex == 2 {
		label = focusedLabelStyle.Render("→ Tool:")
	}
	b.WriteString(label + " ")
	b.WriteString(m.toolSelect.View())
	b.WriteString("\n\n")

	// tags field (index 3)
	label = labelStyle.Render("Tags:")
	if m.focusIndex == 3 {
		label = focusedLabelStyle.Render("→ Tags:")
	}
	b.WriteString(label + " ")
	b.WriteString(helpStyle.Render("(enter to add, backspace to remove)"))
	b.WriteString("\n")
	b.WriteString(m.tags.View())
	b.WriteString("\n\n")

	// help text
	b.WriteString(helpStyle.Render("Tab: next • Shift+Tab: prev • Ctrl+S: save • ?: help • Esc: cancel"))

	// use full width and height
	contentStyle := lipgloss.NewStyle().
		Width(m.width - 4).
		Height(m.height - 2).
		Padding(1, 2)

	baseView := contentStyle.Render(b.String())

	// toast message rendered at bottom, centered
	if m.showToast {
		baseView += "\n" + RenderToast(m.toastMsg, m.isError, m.width)
	}

	// overlay help on top if showing
	if m.showHelp {
		return RenderHelpOverlay(m.width, m.height)
	}

	return baseView
}

func (m Model) renderHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("crumb - Help"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Navigation:"))
	b.WriteString("\n")
	b.WriteString("  Tab / Shift+Tab     Navigate between fields\n")
	b.WriteString("  / or Ctrl+T         Focus tool selector\n")
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Editing:"))
	b.WriteString("\n")
	b.WriteString("  Enter               Add tag (in tags field)\n")
	b.WriteString("  Backspace           Remove last tag (in tags field)\n")
	b.WriteString("  Ctrl+S              Save and exit\n")
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Other:"))
	b.WriteString("\n")
	b.WriteString("  ?                   Toggle this help screen\n")
	b.WriteString("  Esc                 Cancel and exit\n")
	b.WriteString("  Ctrl+C / Ctrl+D     Force quit\n")
	b.WriteString("\n")

	b.WriteString(helpStyle.Render("Press ? to close this help screen"))

	return borderStyle.Render(b.String())
}

func (m *Model) setFocus(index int) {
	// blur current
	switch m.focusIndex {
	case 0:
		m.prompt.Blur()
	case 1:
		m.output.Blur()
	case 2:
		m.toolSelect.Blur()
	case 3:
		m.tags.Blur()
	}

	// set new focus
	m.focusIndex = index

	// focus new
	switch m.focusIndex {
	case 0:
		m.prompt.Focus()
	case 1:
		m.output.Focus()
	case 2:
		m.toolSelect.Focus()
	case 3:
		m.tags.Focus()
	}
}

func (m *Model) focusNext() {
	m.setFocus((m.focusIndex + 1) % 4)
}

func (m *Model) focusPrev() {
	m.setFocus((m.focusIndex - 1 + 4) % 4)
}

// updateTextareaSizes dynamically adjusts textarea heights based on terminal size
func (m *Model) updateTextareaSizes() {
	// fixed elements take approximately:
	// header: 2, labels/spacing: 12, title/tool/tags: 6, footer: 2, padding: 4
	fixedHeight := 26
	availableHeight := m.height - fixedHeight

	if availableHeight < 10 {
		availableHeight = 10 // minimum usable space
	}

	// split available space: 60% prompt, 40% output
	promptHeight := (availableHeight * 60) / 100
	outputHeight := availableHeight - promptHeight

	// set minimums
	if promptHeight < 4 {
		promptHeight = 4
	}
	if outputHeight < 3 {
		outputHeight = 3
	}

	m.prompt.SetHeight(promptHeight)
	m.prompt.SetWidth(m.width - 8)
	m.output.SetHeight(outputHeight)
	m.output.SetWidth(m.width - 8)
}

func (m *Model) saveAndExit() tea.Cmd {
	// validate required fields
	if strings.TrimSpace(m.prompt.Value()) == "" {
		m.showToast = true
		m.isError = true
		m.toastMsg = "Prompt is required"
		return HideToastAfter(2 * time.Second)
	}

	// auto-generate title from prompt
	title := storage.GenerateTitle(m.prompt.Value())

	// generate filename
	timestamp := storage.GetTimestamp()
	filename := storage.GenerateFilename(title, timestamp)

	// get git author
	author := storage.GetGitAuthor()
	if author == "" {
		author = "Unknown"
	}

	// build markdown content
	var content strings.Builder
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("title: %s\n", title))
	content.WriteString(fmt.Sprintf("date: %s\n", timestamp.Format(time.RFC3339)))
	content.WriteString(fmt.Sprintf("author: %s\n", author))
	content.WriteString(fmt.Sprintf("tool: %s\n", m.toolSelect.Selected()))

	if len(m.tags.Tags()) > 0 {
		content.WriteString("tags:\n")
		for _, tag := range m.tags.Tags() {
			content.WriteString(fmt.Sprintf("  - %s\n", tag))
		}
	}

	content.WriteString("---\n\n")
	content.WriteString("## Prompt\n\n")
	content.WriteString(m.prompt.Value())
	content.WriteString("\n\n")

	if strings.TrimSpace(m.output.Value()) != "" {
		content.WriteString("## Output\n\n")
		content.WriteString(m.output.Value())
		content.WriteString("\n")
	}

	// actually save to file system
	filepath, err := m.storage.Save(filename, content.String())
	if err != nil {
		m.showToast = true
		m.isError = true
		m.toastMsg = "Error: " + err.Error()
		return HideToastAfter(3 * time.Second)
	}

	// show success message
	if m.stayOpen {
		m.showToast = true
		m.isError = false
		m.toastMsg = fmt.Sprintf("Saved: %s", filepath)
		m.clearFields()
		return HideToastAfter(2 * time.Second)
	}

	// return success message which will trigger exit
	return func() tea.Msg { return saveSuccessMsg{filename: filepath} }
}

// clearFields resets all input fields to empty state
func (m *Model) clearFields() {
	m.prompt.Reset()
	// rebuild tag suggestions with fresh frequent tags
	tagSuggestions := mergeTagSuggestions(m.config.FavoriteTags, m.storage.GetFrequentTags(10))
	m.tags = components.NewTagInput(tagSuggestions)
	m.output.Reset()
	m.setFocus(0)
}

// mergeTagSuggestions combines config favorites with frequent tags, removing duplicates
func mergeTagSuggestions(favorites []string, frequent []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(favorites)+len(frequent))

	// add favorites first (higher priority)
	for _, tag := range favorites {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	// add frequent tags that aren't already in favorites
	for _, tag := range frequent {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}
