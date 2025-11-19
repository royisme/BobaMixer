package components

import (
	"github.com/royisme/bobamixer/internal/ui/theme"
)

// Card wraps content in a styled box.
type Card struct {
	styles theme.Styles
	width  int
	height int
}

// NewCard creates a new Card component.
func NewCard(styles theme.Styles) Card {
	return Card{
		styles: styles,
	}
}

// WithWidth sets the width of the card.
func (c Card) WithWidth(width int) Card {
	c.width = width
	return c
}

// WithHeight sets the height of the card.
func (c Card) WithHeight(height int) Card {
	c.height = height
	return c
}

// Render renders the content within the card style.
func (c Card) Render(content string) string {
	style := c.styles.Card
	if c.width > 0 {
		style = style.Width(c.width)
	}
	if c.height > 0 {
		style = style.Height(c.height)
	}
	return style.Render(content)
}
