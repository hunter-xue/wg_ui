package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

const (
	cFieldName = iota
	cFieldAddress
	cFieldAllowIPs
	cFieldMTU
	cFieldDNS
	cFieldKeepalive
	cFieldComments
	cFieldCount
)

type FormModel struct {
	store    *db.Store
	serverID int64
	inputs   []textinput.Model
	focused  int
	editing  *db.Client // nil = create mode
	err      string
}

func NewFormModel(store *db.Store, serverID int64, existing *db.Client) FormModel {
	inputs := make([]textinput.Model, cFieldCount)

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 256
	}

	inputs[cFieldName].Placeholder = "Client name"
	inputs[cFieldName].Focus()
	inputs[cFieldAddress].Placeholder = "100.100.0.x/24"
	inputs[cFieldAllowIPs].Placeholder = "100.100.0.0/24, 10.100.0.0/16"
	inputs[cFieldMTU].Placeholder = "MTU"
	inputs[cFieldDNS].Placeholder = "DNS (optional)"
	inputs[cFieldKeepalive].Placeholder = "Keepalive (seconds)"
	inputs[cFieldComments].Placeholder = "Comments (optional)"

	// Default values for new client
	if existing == nil {
		inputs[cFieldMTU].SetValue("1420")
		inputs[cFieldKeepalive].SetValue("25")
	}

	if existing != nil {
		inputs[cFieldName].SetValue(existing.Name)
		inputs[cFieldAddress].SetValue(existing.Address)
		inputs[cFieldAllowIPs].SetValue(existing.AllowIPs)
		inputs[cFieldMTU].SetValue(fmt.Sprintf("%d", existing.MTU))
		inputs[cFieldDNS].SetValue(existing.DNS)
		inputs[cFieldKeepalive].SetValue(fmt.Sprintf("%d", existing.PersistentKeepalive))
		inputs[cFieldComments].SetValue(existing.Comments)
	}

	return FormModel{
		store:    store,
		serverID: serverID,
		inputs:   inputs,
		editing:  existing,
	}
}

func (m FormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormModel) Update(msg tea.Msg) (FormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back):
			return m, tui.SwitchScreen(tui.ScreenClientList)
		case msg.String() == "tab" || msg.String() == "down":
			m.focused = (m.focused + 1) % cFieldCount
			return m, m.updateFocus()
		case msg.String() == "shift+tab" || msg.String() == "up":
			m.focused = (m.focused - 1 + cFieldCount) % cFieldCount
			return m, m.updateFocus()
		case key.Matches(msg, tui.Keys.Enter):
			if m.focused == cFieldCount-1 {
				return m, m.save()
			}
			m.focused = (m.focused + 1) % cFieldCount
			return m, m.updateFocus()
		}
	case tui.SuccessMsg:
		return m, tui.SwitchScreen(tui.ScreenClientList)
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *FormModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, cFieldCount)
	for i := range m.inputs {
		if i == m.focused {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *FormModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, cFieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m FormModel) save() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[cFieldName].Value()
		address := m.inputs[cFieldAddress].Value()
		allowIPs := m.inputs[cFieldAllowIPs].Value()
		mtuStr := m.inputs[cFieldMTU].Value()
		keepaliveStr := m.inputs[cFieldKeepalive].Value()

		if name == "" || address == "" || allowIPs == "" {
			return tui.ErrorMsg{Err: fmt.Errorf("name, address, and allowed IPs are required")}
		}

		// Load server for IP validation
		srv, err := m.store.GetServer()
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		if srv != nil {
			// 1. Client IP must not duplicate server IP
			if wg.SameIP(srv.Address, address) {
				return tui.ErrorMsg{Err: fmt.Errorf("address %s conflicts with server address", ipOnly(address))}
			}
			// 2. Client IP must be in the same subnet as server
			ok, err := wg.SameSubnet(srv.Address, address)
			if err != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("invalid address: %w", err)}
			}
			if !ok {
				return tui.ErrorMsg{Err: fmt.Errorf("address %s is not in the same subnet as server (%s)", ipOnly(address), ipOnly(srv.Address))}
			}
			// 3. Client IP must not duplicate any existing client
			existing, err := m.store.ListClients(srv.ID, "name")
			if err != nil {
				return tui.ErrorMsg{Err: err}
			}
			for _, c := range existing {
				if m.editing != nil && c.ID == m.editing.ID {
					continue // skip self when editing
				}
				if wg.SameIP(c.Address, address) {
					return tui.ErrorMsg{Err: fmt.Errorf("address %s is already used by client %q", ipOnly(address), c.Name)}
				}
			}
		}

		mtu := 1420
		if mtuStr != "" {
			var e error
			mtu, e = strconv.Atoi(mtuStr)
			if e != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("invalid MTU: %s", mtuStr)}
			}
		}

		keepalive := 25
		if keepaliveStr != "" {
			var e error
			keepalive, e = strconv.Atoi(keepaliveStr)
			if e != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("invalid keepalive: %s", keepaliveStr)}
			}
		}

		c := &db.Client{
			ServerID:            m.serverID,
			Name:                name,
			Address:             address,
			AllowIPs:            allowIPs,
			MTU:                 mtu,
			DNS:                 m.inputs[cFieldDNS].Value(),
			PersistentKeepalive: keepalive,
			Comments:            m.inputs[cFieldComments].Value(),
		}

		if m.editing != nil {
			c.ID = m.editing.ID
			c.PrivateKey = m.editing.PrivateKey
			c.PublicKey = m.editing.PublicKey
			c.Disabled = m.editing.Disabled
			if err := m.store.UpdateClient(c); err != nil {
				return tui.ErrorMsg{Err: err}
			}
		} else {
			privKey, pubKey, err := wg.GenerateKeypair(context.Background())
			if err != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("keypair generation failed: %w", err)}
			}
			c.PrivateKey = privKey
			c.PublicKey = pubKey
			if err := m.store.CreateClient(c); err != nil {
				return tui.ErrorMsg{Err: err}
			}
		}

		// Generate client config and store in description (reuse srv loaded above)
		if err == nil && srv != nil {
			endpoint := srv.Endpoint
			if endpoint == "" {
				endpoint = "YOUR_SERVER_IP:PORT"
			}
			clientConfig := wg.GenerateClientConfig(srv, c, endpoint)
			c.Description = clientConfig
			_ = m.store.UpdateClient(c)

			// Regenerate server config with all clients
			clients, _ := m.store.ListClients(srv.ID, "name")
			serverConfig := wg.GenerateServerConfig(srv, clients)
			if err := wg.WriteServerConfig(serverConfig); err == nil {
				_ = wg.SyncConfig(context.Background())
			}
		}

		return tui.SuccessMsg{Message: "Client saved"}
	}
}

func (m FormModel) View() string {
	title := "Create Client"
	if m.editing != nil {
		title = "Edit Client"
	}
	s := tui.TitleStyle.Render(title) + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}

	labels := []string{"Name", "Address", "Allowed IPs", "MTU", "DNS", "Keepalive", "Comments"}
	for i, label := range labels {
		s += tui.LabelStyle.Render(label+":") + " " + m.inputs[i].View() + "\n"
	}

	s += "\n" + tui.HelpStyle.Render("tab/↓: next field • shift+tab/↑: prev field • enter: save (on last field) • esc: cancel")
	return s
}

// ipOnly strips the prefix length from a CIDR string for display purposes.
func ipOnly(cidr string) string {
	if idx := strings.Index(cidr, "/"); idx >= 0 {
		return cidr[:idx]
	}
	return cidr
}
