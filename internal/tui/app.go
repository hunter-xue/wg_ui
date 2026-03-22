package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenServerView
	ScreenServerForm
	ScreenClientList
	ScreenClientCreate // new client (no pre-fill)
	ScreenClientForm   // edit existing client (pre-fill from selected)
	ScreenClientDetail
	ScreenInstall
	ScreenStatus
	ScreenSetupPassword
	ScreenSettings
	ScreenChangePassword
)

// Navigation messages
type SwitchScreenMsg struct {
	Screen Screen
}

type BackToMenuMsg struct{}

// ErrorMsg carries an error to display
type ErrorMsg struct {
	Err error
}

// SuccessMsg carries a success message
type SuccessMsg struct {
	Message string
}

// Store accessor for sub-models
type StoreProvider interface {
	GetStore() *db.Store
}

func SwitchScreen(s Screen) tea.Cmd {
	return func() tea.Msg { return SwitchScreenMsg{Screen: s} }
}

func BackToMenu() tea.Cmd {
	return func() tea.Msg { return BackToMenuMsg{} }
}
