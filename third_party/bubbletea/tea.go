package bubbletea

import "time"

type Msg interface{}

type Cmd func() Msg

type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() string
}

type KeyMsg struct {
	val string
}

func (k KeyMsg) String() string { return k.val }

func NewKeyMsg(value string) KeyMsg { return KeyMsg{val: value} }

type WindowSizeMsg struct {
	Width  int
	Height int
}

type Program struct {
	model Model
}

type ProgramOption func(*Program)

func NewProgram(m Model, _ ...ProgramOption) *Program {
	return &Program{model: m}
}

func (p *Program) Run() (Model, error) {
	return p.model, nil
}

func WithAltScreen() ProgramOption {
	return func(*Program) {}
}

var EnterAltScreen Cmd = func() Msg { return nil }

var Quit Cmd = func() Msg { return nil }

func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		for _, cmd := range cmds {
			if cmd == nil {
				continue
			}
			cmd()
		}
		return nil
	}
}

func Tick(_ time.Duration, fn func(time.Time) Msg) Cmd {
	return func() Msg {
		return fn(time.Now())
	}
}
