package settings

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

type syncDoneMsg struct {
	err error
}

const (
	choiceChangePassword = iota
	choiceSyncConfig
	choiceCount
)

var choiceLabels = []string{
	"Change Password",
	"Sync Server Config",
}

type Model struct {
	store  *db.Store
	cursor int
	msg    string
	err    string
}

func NewModel(store *db.Store) Model {
	return Model{store: store}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < choiceCount-1 {
				m.cursor++
			}
		case "enter":
			m.msg = ""
			m.err = ""
			switch m.cursor {
			case choiceChangePassword:
				return m, tui.SwitchScreen(tui.ScreenChangePassword)
			case choiceSyncConfig:
				return m, m.syncConfig()
			}
		case "esc", "q":
			return m, tui.BackToMenu()
		}
	case syncDoneMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.msg = "Server config backed up and regenerated successfully."
		}
	}
	return m, nil
}

func (m Model) syncConfig() tea.Cmd {
	return func() tea.Msg {
		srv, err := m.store.GetServer()
		if err != nil {
			return syncDoneMsg{err: err}
		}
		if srv == nil {
			return syncDoneMsg{err: fmt.Errorf("no server configured")}
		}
		clients, err := m.store.ListClients(srv.ID, "name")
		if err != nil {
			return syncDoneMsg{err: err}
		}
		content := wg.GenerateServerConfig(srv, clients)
		if err := wg.BackupAndWriteServerConfig(content); err != nil {
			return syncDoneMsg{err: err}
		}
		_ = wg.SyncConfig(context.Background())
		return syncDoneMsg{}
	}
}

func (m Model) View() string {
	s := tui.TitleStyle.Render("Settings") + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}
	if m.msg != "" {
		s += tui.SuccessStyle.Render(m.msg) + "\n\n"
	}

	for i, label := range choiceLabels {
		if i == m.cursor {
			s += tui.SelectedItemStyle.Render("> "+label) + "\n"
		} else {
			s += tui.MenuItemStyle.Render("  "+label) + "\n"
		}
	}

	s += "\n" + tui.HelpStyle.Render("↑/↓: move • enter: select • esc: back")
	return s
}
