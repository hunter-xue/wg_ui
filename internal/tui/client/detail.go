package client

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
)

type DetailModel struct {
	client   *db.Client
	viewport viewport.Model
}

func NewDetailModel(c *db.Client) DetailModel {
	vp := viewport.New(80, 20)
	if c != nil {
		vp.SetContent(c.Description)
	}
	return DetailModel{client: c, viewport: vp}
}

func (m DetailModel) Init() tea.Cmd {
	return nil
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back):
			return m, tui.SwitchScreen(tui.ScreenClientList)
		case key.Matches(msg, tui.Keys.Quit):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	s := tui.TitleStyle.Render("Client Config") + "\n"

	if m.client != nil {
		s += tui.LabelStyle.Render("Client: "+m.client.Name) + "\n\n"
	}

	s += m.viewport.View() + "\n"
	s += tui.HelpStyle.Render("↑/↓: scroll • esc: back • q: quit")
	return s
}
