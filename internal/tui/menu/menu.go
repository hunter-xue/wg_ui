package menu

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/tui"
)

type Choice int

const (
	InstallWG Choice = iota
	ServerMgmt
	ClientMgmt
	Status
	Settings
	Quit
)

var choices = []string{
	"Install WireGuard",
	"Server Management",
	"Client Management",
	"Status",
	"Settings",
	"Quit",
}

type SelectMsg struct {
	Choice Choice
}

type Model struct {
	cursor int
}

func New() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, tui.Keys.Down):
			if m.cursor < len(choices)-1 {
				m.cursor++
			}
		case key.Matches(msg, tui.Keys.Enter):
			return m, func() tea.Msg { return SelectMsg{Choice: Choice(m.cursor)} }
		case key.Matches(msg, tui.Keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := tui.TitleStyle.Render("KSYun WireGuard Manager") + "\n\n"

	for i, choice := range choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "▸ "
			s += tui.SelectedItemStyle.Render(cursor+choice) + "\n"
		} else {
			s += tui.MenuItemStyle.Render(cursor+choice) + "\n"
		}
	}

	s += "\n" + tui.HelpStyle.Render(fmt.Sprintf("↑/↓: navigate • enter: select • q: quit"))

	return s
}
