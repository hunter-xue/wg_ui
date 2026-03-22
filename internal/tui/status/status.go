package status

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

type dataLoadedMsg struct {
	serviceStatus string
	server        *db.Server
	total         int
	enabled       int
	disabled      int
}

type Model struct {
	store         *db.Store
	serviceStatus string
	server        *db.Server
	total         int
	enabled       int
	disabled      int
	loading       bool
	err           string
}

func New(store *db.Store) Model {
	return Model{store: store, loading: true}
}

func (m Model) Init() tea.Cmd {
	return m.load()
}

func (m Model) load() tea.Cmd {
	return func() tea.Msg {
		srv, err := m.store.GetServer()
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}

		var total, enabled, disabled int
		if srv != nil {
			clients, err := m.store.ListClients(srv.ID, "name")
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			total = len(clients)
			for _, c := range clients {
				if c.Disabled == 0 {
					enabled++
				} else {
					disabled++
				}
			}
		}

		serviceStatus, _ := wg.ServiceStatus(context.Background())

		return dataLoadedMsg{
			serviceStatus: serviceStatus,
			server:        srv,
			total:         total,
			enabled:       enabled,
			disabled:      disabled,
		}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dataLoadedMsg:
		m.loading = false
		m.serviceStatus = msg.serviceStatus
		m.server = msg.server
		m.total = msg.total
		m.enabled = msg.enabled
		m.disabled = msg.disabled
	case tui.ErrorMsg:
		m.loading = false
		m.err = msg.Err.Error()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back), key.Matches(msg, tui.Keys.Quit):
			return m, tui.BackToMenu()
		}
	}
	return m, nil
}

func (m Model) View() string {
	s := tui.TitleStyle.Render("Status") + "\n\n"

	if m.loading {
		return s + "Loading...\n"
	}
	if m.err != "" {
		return s + tui.ErrorStyle.Render("Error: "+m.err) + "\n"
	}

	// Service status
	s += tui.LabelStyle.Render("WireGuard Service:") + "\n"
	if m.serviceStatus != "" {
		s += m.serviceStatus + "\n"
	} else {
		s += tui.ValueStyle.Render("  (not running or not installed)") + "\n"
	}

	// Server info
	s += "\n"
	if m.server == nil {
		s += tui.ValueStyle.Render("No server configured.") + "\n"
	} else {
		s += tui.LabelStyle.Render("Server:") + " " + tui.ValueStyle.Render(m.server.Name) + "\n"
		s += tui.LabelStyle.Render("Address:") + " " + tui.ValueStyle.Render(m.server.Address) + "\n"
		s += tui.LabelStyle.Render("Listen Port:") + " " + tui.ValueStyle.Render(fmt.Sprintf("%d", m.server.ListenPort)) + "\n"
	}

	// Client stats
	s += "\n"
	s += tui.LabelStyle.Render("Clients:") + " " + tui.ValueStyle.Render(fmt.Sprintf("%d total", m.total)) + "\n"
	s += tui.LabelStyle.Render("  Enabled:") + " " + tui.SuccessStyle.Render(fmt.Sprintf("%d", m.enabled)) + "\n"
	s += tui.LabelStyle.Render("  Disabled:") + " " + tui.ValueStyle.Render(fmt.Sprintf("%d", m.disabled)) + "\n"

	s += "\n" + tui.HelpStyle.Render("esc/q: back to menu")
	return s
}
