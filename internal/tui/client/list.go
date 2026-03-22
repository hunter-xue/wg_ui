package client

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

type clientsLoadedMsg struct {
	clients []db.Client
}

type ListModel struct {
	store    *db.Store
	serverID int64
	clients  []db.Client
	cursor   int
	err      string
	msg      string
}

func NewListModel(store *db.Store, serverID int64) ListModel {
	return ListModel{store: store, serverID: serverID}
}

func (m ListModel) Init() tea.Cmd {
	return m.loadClients()
}

func (m ListModel) loadClients() tea.Cmd {
	return func() tea.Msg {
		clients, err := m.store.ListClients(m.serverID, "name")
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return clientsLoadedMsg{clients: clients}
	}
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case clientsLoadedMsg:
		m.clients = msg.clients
		m.err = ""
		if m.cursor >= len(m.clients) && m.cursor > 0 {
			m.cursor = len(m.clients) - 1
		}
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
	case tui.SuccessMsg:
		m.msg = msg.Message
		return m, m.loadClients()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back):
			return m, tui.BackToMenu()
		case key.Matches(msg, tui.Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, tui.Keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, tui.Keys.Down):
			if m.cursor < len(m.clients)-1 {
				m.cursor++
			}
		case key.Matches(msg, tui.Keys.Create):
			return m, tui.SwitchScreen(tui.ScreenClientCreate)
		case key.Matches(msg, tui.Keys.Edit):
			if len(m.clients) > 0 {
				return m, tui.SwitchScreen(tui.ScreenClientForm)
			}
		case key.Matches(msg, tui.Keys.Enter):
			if len(m.clients) > 0 {
				return m, tui.SwitchScreen(tui.ScreenClientDetail)
			}
		case key.Matches(msg, tui.Keys.Toggle):
			if len(m.clients) > 0 {
				return m, m.toggleClient()
			}
		case key.Matches(msg, tui.Keys.Regen):
			if len(m.clients) > 0 {
				return m, m.regenKeypair()
			}
		}
	}
	return m, nil
}

func (m ListModel) toggleClient() tea.Cmd {
	c := m.clients[m.cursor]
	return func() tea.Msg {
		newDisabled := c.Disabled == 0
		if err := m.store.SetClientDisabled(c.ID, newDisabled); err != nil {
			return tui.ErrorMsg{Err: err}
		}

		// Regenerate server config to add/remove the [Peer] section
		srv, err := m.store.GetServer()
		if err == nil && srv != nil {
			clients, _ := m.store.ListClients(srv.ID, "name")
			configContent := wg.GenerateServerConfig(srv, clients)
			if err := wg.WriteServerConfig(configContent); err == nil {
				_ = wg.SyncConfig(context.Background())
			}
		}

		action := "enabled"
		if newDisabled {
			action = "disabled"
		}
		return tui.SuccessMsg{Message: fmt.Sprintf("Client %s %s", c.Name, action)}
	}
}

func (m ListModel) regenKeypair() tea.Cmd {
	c := m.clients[m.cursor]
	return func() tea.Msg {
		// Generate new keypair
		privKey, pubKey, err := wg.GenerateKeypair(context.Background())
		if err != nil {
			return tui.ErrorMsg{Err: fmt.Errorf("keypair generation failed: %w", err)}
		}
		c.PrivateKey = privKey
		c.PublicKey = pubKey

		// Regenerate client config with new keypair
		srv, err := m.store.GetServer()
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		if srv != nil {
			endpoint := srv.Description
			if endpoint == "" {
				endpoint = "YOUR_SERVER_IP:PORT"
			}
			c.Description = wg.GenerateClientConfig(srv, &c, endpoint)
		}

		if err := m.store.UpdateClient(&c); err != nil {
			return tui.ErrorMsg{Err: err}
		}

		// Sync server config (public key in [Peer] has changed)
		if srv != nil {
			clients, _ := m.store.ListClients(srv.ID, "name")
			configContent := wg.GenerateServerConfig(srv, clients)
			if err := wg.WriteServerConfig(configContent); err == nil {
				_ = wg.SyncConfig(context.Background())
			}
		}

		return tui.SuccessMsg{Message: fmt.Sprintf("Keypair regenerated for %s", c.Name)}
	}
}

func (m ListModel) SelectedClient() *db.Client {
	if len(m.clients) == 0 || m.cursor >= len(m.clients) {
		return nil
	}
	c := m.clients[m.cursor]
	return &c
}

func (m ListModel) View() string {
	s := tui.TitleStyle.Render("Client Management") + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}
	if m.msg != "" {
		s += tui.SuccessStyle.Render(m.msg) + "\n\n"
	}

	if len(m.clients) == 0 {
		s += "No clients configured.\n\n"
		s += tui.HelpStyle.Render("c: create • esc: back • q: quit")
		return s
	}

	// Header
	s += fmt.Sprintf("  %-20s %-18s %-10s\n", "Name", "Address", "Status")
	s += fmt.Sprintf("  %-20s %-18s %-10s\n", "────────────────────", "──────────────────", "──────────")

	for i, c := range m.clients {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		status := "enabled"
		if c.Disabled != 0 {
			status = "disabled"
		}

		line := fmt.Sprintf("%-20s %-18s %-10s", c.Name, c.Address, status)
		if i == m.cursor {
			s += tui.SelectedItemStyle.Render(cursor+line) + "\n"
		} else if c.Disabled != 0 {
			s += tui.DisabledStyle.Render(cursor+line) + "\n"
		} else {
			s += tui.MenuItemStyle.Render(cursor+line) + "\n"
		}
	}

	s += "\n" + tui.HelpStyle.Render("c: create • e: edit • enter: view config • space: toggle • r: regen keypair • esc: back • q: quit")
	return s
}
