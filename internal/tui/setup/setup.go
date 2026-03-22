package setup

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/bcrypt"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
)

const (
	fieldPassword = iota
	fieldConfirm
	fieldCount
)

type Model struct {
	store  *db.Store
	inputs []textinput.Model
	focused int
	err    string
}

func NewModel(store *db.Store) Model {
	inputs := make([]textinput.Model, fieldCount)

	inputs[fieldPassword] = textinput.New()
	inputs[fieldPassword].Placeholder = "New password"
	inputs[fieldPassword].EchoMode = textinput.EchoPassword
	inputs[fieldPassword].EchoCharacter = '•'
	inputs[fieldPassword].Focus()

	inputs[fieldConfirm] = textinput.New()
	inputs[fieldConfirm].Placeholder = "Confirm password"
	inputs[fieldConfirm].EchoMode = textinput.EchoPassword
	inputs[fieldConfirm].EchoCharacter = '•'

	return Model{
		store:  store,
		inputs: inputs,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + fieldCount) % fieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focused == fieldCount-1 {
				return m, m.save()
			}
			m.focused = (m.focused + 1) % fieldCount
			return m, m.updateFocus()
		}
	case tui.SuccessMsg:
		return m, tui.SwitchScreen(tui.ScreenMenu)
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateFocus() tea.Cmd {
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

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, fieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m Model) save() tea.Cmd {
	return func() tea.Msg {
		pw := m.inputs[fieldPassword].Value()
		confirm := m.inputs[fieldConfirm].Value()
		if pw == "" {
			return tui.ErrorMsg{Err: fmt.Errorf("password cannot be empty")}
		}
		if pw != confirm {
			return tui.ErrorMsg{Err: fmt.Errorf("passwords do not match")}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		if err := m.store.CreateAdminUser(string(hash)); err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SuccessMsg{Message: "Password set successfully"}
	}
}

func (m Model) View() string {
	s := tui.TitleStyle.Render("First Run Setup") + "\n\n"
	s += "Please set an admin password to protect this application.\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}

	labels := []string{"Password       ", "Confirm Password"}
	for i, input := range m.inputs {
		s += tui.LabelStyle.Render(labels[i]) + " " + input.View() + "\n"
	}

	s += "\n" + tui.HelpStyle.Render("tab: next field • enter: confirm")
	return s
}
