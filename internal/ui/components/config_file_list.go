package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// ConfigFile represents an editable configuration file entry.
type ConfigFile struct {
	Name string
	File string
	Desc string
}

// ConfigFileList renders selectable config files.
type ConfigFileList struct {
	files    []ConfigFile
	selected int
	home     string
	styles   theme.Styles
}

// NewConfigFileList constructs the list component.
func NewConfigFileList(files []ConfigFile, selected int, home string, styles theme.Styles) ConfigFileList {
	return ConfigFileList{
		files:    files,
		selected: selected,
		home:     home,
		styles:   styles,
	}
}

// Update satisfies the Bubble Tea component interface.
func (c ConfigFileList) Update(_ tea.Msg) (ConfigFileList, tea.Cmd) {
	return c, nil
}

// View renders the config file list with descriptions for the selected item.
func (c ConfigFileList) View() string {
	if len(c.files) == 0 {
		normalStyle := c.styles.Normal
		return normalStyle.PaddingLeft(2).Render("No configuration files detected.")
	}

	var b strings.Builder
	selected := c.selected
	if selected >= len(c.files) {
		selected = 0
	}

	for idx, cfg := range c.files {
		fileLabel := fmt.Sprintf(" (%s)", cfg.File)
		if idx == selected {
			b.WriteString(c.styles.Selected.Render("▶  " + cfg.Name))
			b.WriteString(c.styles.Help.Render(fileLabel))
		} else {
			b.WriteString(c.styles.Normal.Render("  " + cfg.Name))
			b.WriteString(c.styles.Help.Render(fileLabel))
		}
		b.WriteString("\n")

		if idx == selected {
			normalStyle := c.styles.Normal
			descLine := normalStyle.PaddingLeft(4).Render(cfg.Desc)
			helpStyle := c.styles.Help
			pathLine := helpStyle.PaddingLeft(4).Render(fmt.Sprintf("Full path: %s/%s", strings.TrimSpace(c.home), cfg.File))
			b.WriteString(descLine)
			b.WriteString("\n")
			b.WriteString(pathLine)
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
