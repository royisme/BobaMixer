package lipgloss

import "strings"

type Color string

func (c Color) Render(str string) string { return str }

type Style struct{}

type Border struct{}

func RoundedBorder() Border { return Border{} }

func NewStyle() Style { return Style{} }

func (s Style) Bold(_ bool) Style              { return s }
func (s Style) Foreground(_ Color) Style       { return s }
func (s Style) Background(_ Color) Style       { return s }
func (s Style) MarginBottom(_ int) Style       { return s }
func (s Style) BorderStyle(_ Border) Style     { return s }
func (s Style) BorderForeground(_ Color) Style { return s }
func (s Style) Padding(_ ...int) Style         { return s }
func (s Style) PaddingLeft(_ int) Style        { return s }
func (s Style) Italic(_ bool) Style            { return s }
func (s Style) Render(str string) string       { return str }

func (s Style) Width(_ int) Style { return s }

func (s Style) Align(_ Alignment) Style { return s }

type Alignment int

const (
	Left Alignment = iota
	Top
)

func JoinVertical(_ Alignment, values ...string) string {
	return strings.Join(values, "\n")
}

func JoinHorizontal(_ Alignment, values ...string) string {
	return strings.Join(values, "")
}
