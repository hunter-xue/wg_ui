package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/wg"
)

type installCheckMsg struct {
	installed bool
	status    string
}

type installStepMsg struct {
	stepName string
	output   string
	err      error
	done     bool // true when all steps finished successfully
}

type RootModel struct {
	store          *db.Store
	screen         Screen
	width          int
	height         int
	err            string
	msg            string
	spinner        spinner.Model
	loading        bool
	installStatus    string // service status when already installed
	alreadyInstalled bool
	installLogs      []string // step-by-step log lines
	installDone      bool     // installation finished (success or fail)
	installErr       string   // non-empty if install failed
	installStep      int      // index of next step to run

	// Sub-model interfaces - will be set by the app package
	MenuView      func() string
	MenuUpdate    func(tea.Msg) tea.Cmd
	MenuInit      func() tea.Cmd
	ServerView    func() string
	ServerUpdate  func(tea.Msg) tea.Cmd
	ServerInit    func() tea.Cmd
	SFormView     func() string
	SFormUpdate   func(tea.Msg) tea.Cmd
	SFormInit     func() tea.Cmd
	ClientLView   func() string
	ClientLUpdate func(tea.Msg) tea.Cmd
	ClientLInit   func() tea.Cmd
	CFFormView    func() string
	CFFormUpdate  func(tea.Msg) tea.Cmd
	CFFormInit    func() tea.Cmd
	CDetailView    func() string
	CDetailUpdate  func(tea.Msg) tea.Cmd
	CDetailInit    func() tea.Cmd
	StatusView     func() string
	StatusUpdate   func(tea.Msg) tea.Cmd
	StatusInit     func() tea.Cmd
	SetupView      func() string
	SetupUpdate    func(tea.Msg) tea.Cmd
	SetupInit      func() tea.Cmd
	SettingsView   func() string
	SettingsUpdate func(tea.Msg) tea.Cmd
	SettingsInit   func() tea.Cmd
	ChPassView     func() string
	ChPassUpdate   func(tea.Msg) tea.Cmd
	ChPassInit     func() tea.Cmd

	// Callbacks for creating sub-models with context
	OnSwitchScreen func(screen Screen)
}

func NewRootModel(store *db.Store) RootModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return RootModel{
		store:   store,
		screen:  ScreenMenu,
		spinner: sp,
	}
}

func (m RootModel) Init() tea.Cmd {
	return nil
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case installCheckMsg:
		m.loading = false
		if msg.installed {
			m.alreadyInstalled = true
			m.installStatus = msg.status
		} else {
			// Not installed — run first step
			m.installStep = 0
			return m, m.runNextInstallStep()
		}
		return m, nil
	case installStepMsg:
		m.loading = false
		if msg.err != nil {
			// Step failed — append error log and stop
			m.installLogs = append(m.installLogs,
				"✗ "+msg.stepName,
				"  "+msg.output,
			)
			m.installDone = true
			m.installErr = msg.err.Error()
		} else if msg.done {
			// All steps succeeded
			m.installLogs = append(m.installLogs, "✓ "+msg.stepName)
			m.installDone = true
		} else {
			// Step succeeded, run next
			m.installLogs = append(m.installLogs, "✓ "+msg.stepName)
			m.installStep++
			return m, m.runNextInstallStep()
		}
		return m, nil
	case SwitchScreenMsg:
		m.screen = msg.Screen
		m.err = ""
		m.msg = ""
		if m.OnSwitchScreen != nil {
			m.OnSwitchScreen(msg.Screen)
		}
		return m, m.initScreen(msg.Screen)
	case BackToMenuMsg:
		m.screen = ScreenMenu
		m.err = ""
		m.msg = ""
		return m, nil
	}

	var cmd tea.Cmd
	switch m.screen {
	case ScreenMenu:
		if m.MenuUpdate != nil {
			cmd = m.MenuUpdate(msg)
		}
	case ScreenInstall:
		if kmsg, ok := msg.(tea.KeyMsg); ok {
			if kmsg.String() == "esc" || kmsg.String() == "q" {
				m.screen = ScreenMenu
				m.alreadyInstalled = false
				m.installStatus = ""
				return m, nil
			}
		}
	case ScreenServerView:
		if m.ServerUpdate != nil {
			cmd = m.ServerUpdate(msg)
		}
	case ScreenServerForm:
		if m.SFormUpdate != nil {
			cmd = m.SFormUpdate(msg)
		}
	case ScreenClientList:
		if m.ClientLUpdate != nil {
			cmd = m.ClientLUpdate(msg)
		}
	case ScreenClientCreate, ScreenClientForm:
		if m.CFFormUpdate != nil {
			cmd = m.CFFormUpdate(msg)
		}
	case ScreenClientDetail:
		if m.CDetailUpdate != nil {
			cmd = m.CDetailUpdate(msg)
		}
	case ScreenStatus:
		if m.StatusUpdate != nil {
			cmd = m.StatusUpdate(msg)
		}
	case ScreenSetupPassword:
		if m.SetupUpdate != nil {
			cmd = m.SetupUpdate(msg)
		}
	case ScreenSettings:
		if m.SettingsUpdate != nil {
			cmd = m.SettingsUpdate(msg)
		}
	case ScreenChangePassword:
		if m.ChPassUpdate != nil {
			cmd = m.ChPassUpdate(msg)
		}
	}

	if m.loading {
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		cmd = tea.Batch(cmd, spinCmd)
	}

	return m, cmd
}

func (m RootModel) installView() string {
	s := TitleStyle.Render("Install WireGuard") + "\n\n"

	// Already installed branch
	if m.alreadyInstalled {
		s += SuccessStyle.Render("WireGuard is already installed on this system.") + "\n\n"
		s += LabelStyle.Render("Service Status:") + "\n"
		if m.installStatus != "" {
			s += ValueStyle.Render(m.installStatus) + "\n"
		} else {
			s += ValueStyle.Render("(unable to retrieve status)") + "\n"
		}
		s += "\n" + HelpStyle.Render("esc/q: back to menu")
		return s
	}

	// Checking phase
	if m.loading && len(m.installLogs) == 0 {
		return s + fmt.Sprintf("  %s Checking installation...\n", m.spinner.View())
	}

	// Show completed step logs
	for _, line := range m.installLogs {
		s += line + "\n"
	}

	// Show current step spinner
	if !m.installDone {
		steps := wg.InstallSteps
		if m.installStep < len(steps) {
			s += fmt.Sprintf("  %s %s\n", m.spinner.View(), steps[m.installStep].Name)
		}
	}

	// Final result
	if m.installDone {
		s += "\n"
		if m.installErr != "" {
			s += ErrorStyle.Render("Installation failed: "+m.installErr) + "\n"
		} else {
			s += SuccessStyle.Render("WireGuard installed successfully!") + "\n"
		}
		s += "\n" + HelpStyle.Render("esc/q: back to menu")
	}

	return s
}

func (m *RootModel) runNextInstallStep() tea.Cmd {
	steps := wg.InstallSteps
	idx := m.installStep
	m.loading = true
	step := steps[idx]
	isLast := idx == len(steps)-1
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			output, err := step.Run(context.Background())
			if err != nil {
				return installStepMsg{stepName: step.Name, output: output, err: err}
			}
			return installStepMsg{stepName: step.Name, output: output, done: isLast}
		},
	)
}

func (m RootModel) initScreen(s Screen) tea.Cmd {
	switch s {
	case ScreenServerView:
		if m.ServerInit != nil {
			return m.ServerInit()
		}
	case ScreenServerForm:
		if m.SFormInit != nil {
			return m.SFormInit()
		}
	case ScreenClientList:
		if m.ClientLInit != nil {
			return m.ClientLInit()
		}
	case ScreenClientCreate, ScreenClientForm:
		if m.CFFormInit != nil {
			return m.CFFormInit()
		}
	case ScreenClientDetail:
		if m.CDetailInit != nil {
			return m.CDetailInit()
		}
	case ScreenStatus:
		if m.StatusInit != nil {
			return m.StatusInit()
		}
	case ScreenSetupPassword:
		if m.SetupInit != nil {
			return m.SetupInit()
		}
	case ScreenSettings:
		if m.SettingsInit != nil {
			return m.SettingsInit()
		}
	case ScreenChangePassword:
		if m.ChPassInit != nil {
			return m.ChPassInit()
		}
	}
	return nil
}

func (m RootModel) View() string {
	if m.screen == ScreenInstall {
		return m.installView()
	}

	switch m.screen {
	case ScreenMenu:
		if m.MenuView != nil {
			return m.MenuView()
		}
	case ScreenServerView:
		if m.ServerView != nil {
			return m.ServerView()
		}
	case ScreenServerForm:
		if m.SFormView != nil {
			return m.SFormView()
		}
	case ScreenClientList:
		if m.ClientLView != nil {
			return m.ClientLView()
		}
	case ScreenClientCreate, ScreenClientForm:
		if m.CFFormView != nil {
			return m.CFFormView()
		}
	case ScreenClientDetail:
		if m.CDetailView != nil {
			return m.CDetailView()
		}
	case ScreenStatus:
		if m.StatusView != nil {
			return m.StatusView()
		}
	case ScreenSetupPassword:
		if m.SetupView != nil {
			return m.SetupView()
		}
	case ScreenSettings:
		if m.SettingsView != nil {
			return m.SettingsView()
		}
	case ScreenChangePassword:
		if m.ChPassView != nil {
			return m.ChPassView()
		}
	}

	return "Loading..."
}

func (m *RootModel) SetInitialScreen(s Screen) {
	m.screen = s
}

func (m *RootModel) StartInstall() tea.Cmd {
	m.loading = true
	m.alreadyInstalled = false
	m.installStatus = ""
	m.installLogs = nil
	m.installDone = false
	m.installErr = ""
	m.installStep = 0
	m.screen = ScreenInstall
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			ctx := context.Background()
			if wg.IsInstalled(ctx) {
				status, _ := wg.ServiceStatus(ctx)
				return installCheckMsg{installed: true, status: status}
			}
			return installCheckMsg{installed: false}
		},
	)
}
