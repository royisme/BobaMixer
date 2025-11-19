// Package layouts provides layout utilities for arranging UI components in the BobaMixer TUI.
package layouts

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Row composes blocks horizontally without applying any additional styling.
func Row(blocks ...string) string {
	items := filterBlocks(blocks)
	if len(items) == 0 {
		return ""
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}

// Column stacks blocks vertically without altering their own styling.
func Column(blocks ...string) string {
	items := filterBlocks(blocks)
	if len(items) == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

// Section renders a labeled block made from a title and body content.
func Section(title string, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	switch {
	case title == "" && body == "":
		return ""
	case title == "":
		return body
	case body == "":
		return title
	default:
		return Column(title, body)
	}
}

// Gap inserts blank lines to create vertical spacing between layout blocks.
func Gap(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("\n", n)
}

// Pad applies a left padding using spaces to keep spacing logic outside components.
func Pad(padding int, content string) string {
	if padding <= 0 || content == "" {
		return content
	}

	prefix := strings.Repeat(" ", padding)
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func filterBlocks(blocks []string) []string {
	result := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if trimmed := strings.TrimSpace(block); trimmed != "" {
			result = append(result, block)
		}
	}
	return result
}

// Center places the content in the center of a box of size width x height.
func Center(width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
