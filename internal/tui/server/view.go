package server

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/wg"
)

type serverLoadedMsg struct {
	server *db.Server
}

type statusMsg struct {
	output string
}

type ViewModel struct {
	store  *db.Store
	server *db.Server
	status string
	err    string
	msg    string
}

func NewViewModel(store *db.Store) ViewModel {
	return ViewModel{store: store}
}

func (m ViewModel) Init() tea.Cmd {
	return m.loadServer()
}

func (m ViewModel) loadServer() tea.Cmd {
	return func() tea.Msg {
		srv, err := m.store.GetServer()
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return serverLoadedMsg{server: srv}
	}
}

func (m ViewModel) Update(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case serverLoadedMsg:
		m.server = msg.server
		m.err = ""
	case statusMsg:
		m.status = msg.output
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
	case tui.SuccessMsg:
		m.msg = msg.Message
		return m, m.loadServer()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, tui.Keys.Back):
			return m, tui.BackToMenu()
		case key.Matches(msg, tui.Keys.Create):
			if m.server == nil {
				return m, tui.SwitchScreen(tui.ScreenServerForm)
			}
		case key.Matches(msg, tui.Keys.Edit):
			if m.server != nil {
				return m, tui.SwitchScreen(tui.ScreenServerForm)
			}
		case key.Matches(msg, tui.Keys.Delete):
			if m.server != nil {
				return m, m.deleteServer()
			}
		case key.Matches(msg, tui.Keys.Status):
			return m, m.checkStatus()
		case key.Matches(msg, tui.Keys.Quit):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ViewModel) deleteServer() tea.Cmd {
	return func() tea.Msg {
		if m.server == nil {
			return nil
		}
		if err := m.store.DeleteServer(m.server.ID); err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SuccessMsg{Message: "Server deleted"}
	}
}

func (m ViewModel) checkStatus() tea.Cmd {
	return func() tea.Msg {
		output, _ := wg.ServiceStatus(context.Background())
		return statusMsg{output: output}
	}
}

func (m ViewModel) View() string {
	s := tui.TitleStyle.Render("Server Management") + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}
	if m.msg != "" {
		s += tui.SuccessStyle.Render(m.msg) + "\n\n"
	}

	if m.server == nil {
		s += "No server configured.\n\n"
		s += tui.HelpStyle.Render("c: create • esc: back • q: quit")
		return s
	}

	srv := m.server
	s += renderField("Name", srv.Name)
	s += renderField("Address", srv.Address)
	s += renderField("Listen Port", fmt.Sprintf("%d", srv.ListenPort))
	s += renderField("Public Key", srv.PublicKey)
	s += renderField("MTU", fmt.Sprintf("%d", srv.MTU))
	if srv.DNS != "" {
		s += renderField("DNS", srv.DNS)
	}
	if srv.PostUp != "" {
		s += renderField("PostUp", srv.PostUp)
	}
	if srv.PostDown != "" {
		s += renderField("PostDown", srv.PostDown)
	}
	if srv.Description != "" {
		s += renderField("Endpoint", srv.Description)
	}

	if m.status != "" {
		s += "\n" + tui.LabelStyle.Render("Service Status:") + "\n" + m.status + "\n"
	}

	s += "\n" + tui.HelpStyle.Render("e: edit • d: delete • s: status • esc: back • q: quit")
	return s
}

func renderField(label, value string) string {
	return tui.LabelStyle.Render(label+":") + " " + tui.ValueStyle.Render(value) + "\n"
}
