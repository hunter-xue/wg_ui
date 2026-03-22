package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

const (
	fieldName = iota
	fieldAddress
	fieldListenPort
	fieldMTU
	fieldDNS
	fieldPostUp
	fieldPostDown
	fieldEndpoint
	fieldComments
	fieldCount
)

type FormModel struct {
	store    *db.Store
	inputs   []textinput.Model
	focused  int
	editing  *db.Server // nil = create mode
	err      string
}

func NewFormModel(store *db.Store, existing *db.Server) FormModel {
	inputs := make([]textinput.Model, fieldCount)

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 256
	}
	inputs[fieldPostUp].CharLimit = 1024
	inputs[fieldPostDown].CharLimit = 1024

	inputs[fieldName].Placeholder = "Server name"
	inputs[fieldName].Focus()
	inputs[fieldAddress].Placeholder = "100.100.0.1/32"
	inputs[fieldListenPort].Placeholder = "51820"
	inputs[fieldMTU].Placeholder = "1420"
	inputs[fieldDNS].Placeholder = "DNS (optional)"
	inputs[fieldPostUp].Placeholder = "PostUp rules, use ; to separate multiple commands"
	inputs[fieldPostDown].Placeholder = "PostDown rules, use ; to separate multiple commands"
	inputs[fieldEndpoint].Placeholder = "ip:port (e.g. 1.2.3.4:51820)"
	inputs[fieldComments].Placeholder = "Comments (optional)"

	// Default values for new server
	if existing == nil {
		inputs[fieldMTU].SetValue("1420")
		inputs[fieldPostUp].SetValue("iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE")
		inputs[fieldPostDown].SetValue("iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE")
	}

	if existing != nil {
		inputs[fieldName].SetValue(existing.Name)
		inputs[fieldAddress].SetValue(existing.Address)
		inputs[fieldListenPort].SetValue(fmt.Sprintf("%d", existing.ListenPort))
		inputs[fieldMTU].SetValue(fmt.Sprintf("%d", existing.MTU))
		inputs[fieldDNS].SetValue(existing.DNS)
		inputs[fieldPostUp].SetValue(existing.PostUp)
		inputs[fieldPostDown].SetValue(existing.PostDown)
		inputs[fieldEndpoint].SetValue(existing.Endpoint)
		inputs[fieldComments].SetValue(existing.Comments)
	}

	return FormModel{
		store:   store,
		inputs:  inputs,
		editing: existing,
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
			return m, tui.SwitchScreen(tui.ScreenServerView)
		case msg.String() == "tab" || msg.String() == "down":
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		case msg.String() == "shift+tab" || msg.String() == "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()
		case key.Matches(msg, tui.Keys.Enter):
			if m.focused == fieldCount-1 {
				return m, m.save()
			}
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		}
	case tui.SuccessMsg:
		return m, tui.SwitchScreen(tui.ScreenServerView)
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *FormModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
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
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m FormModel) save() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[fieldName].Value()
		address := m.inputs[fieldAddress].Value()
		portStr := m.inputs[fieldListenPort].Value()
		mtuStr := m.inputs[fieldMTU].Value()

		endpoint := m.inputs[fieldEndpoint].Value()

		if name == "" || address == "" || portStr == "" || endpoint == "" {
			return tui.ErrorMsg{Err: fmt.Errorf("name, address, listen port, and endpoint are required")}
		}

		// Validate endpoint format: must be host:port
		if _, _, err := net.SplitHostPort(endpoint); err != nil {
			return tui.ErrorMsg{Err: fmt.Errorf("endpoint must be in ip:port format (e.g. 1.2.3.4:51820)")}
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			return tui.ErrorMsg{Err: fmt.Errorf("invalid port: %s", portStr)}
		}

		mtu := 1420
		if mtuStr != "" {
			mtu, err = strconv.Atoi(mtuStr)
			if err != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("invalid MTU: %s", mtuStr)}
			}
		}

		srv := &db.Server{
			Name:        name,
			Address:     address,
			ListenPort:  port,
			MTU:         mtu,
			DNS:         m.inputs[fieldDNS].Value(),
			PostUp:      m.inputs[fieldPostUp].Value(),
			PostDown:    m.inputs[fieldPostDown].Value(),
			Endpoint:    m.inputs[fieldEndpoint].Value(),
			Comments:    m.inputs[fieldComments].Value(),
		}

		if m.editing != nil {
			srv.ID = m.editing.ID
			srv.PrivateKey = m.editing.PrivateKey
			srv.PublicKey = m.editing.PublicKey
			if err := m.store.UpdateServer(srv); err != nil {
				return tui.ErrorMsg{Err: err}
			}
		} else {
			privKey, pubKey, err := wg.GenerateKeypair(context.Background())
			if err != nil {
				return tui.ErrorMsg{Err: fmt.Errorf("keypair generation failed: %w", err)}
			}
			srv.PrivateKey = privKey
			srv.PublicKey = pubKey
			if err := m.store.CreateServer(srv); err != nil {
				return tui.ErrorMsg{Err: err}
			}
		}

		// Generate and write config
		clients, _ := m.store.ListClients(srv.ID, "name")
		configContent := wg.GenerateServerConfig(srv, clients)
		if err := wg.WriteServerConfig(configContent); err != nil {
			// Non-fatal on macOS dev
			return tui.SuccessMsg{Message: fmt.Sprintf("Server saved (config write: %v)", err)}
		}

		_ = wg.RestartService(context.Background())
		return tui.SuccessMsg{Message: "Server saved and config applied"}
	}
}

func (m FormModel) View() string {
	title := "Create Server"
	if m.editing != nil {
		title = "Edit Server"
	}
	s := tui.TitleStyle.Render(title) + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}

	labels := []string{"Name", "Address", "Listen Port", "MTU", "DNS", "PostUp", "PostDown", "Endpoint", "Comments"}
	for i, label := range labels {
		s += tui.LabelStyle.Render(label+":") + " " + m.inputs[i].View() + "\n"
	}

	s += "\n" + tui.HelpStyle.Render("tab/↓: next field • shift+tab/↑: prev field • enter: save (on last field) • esc: cancel")
	return s
}
