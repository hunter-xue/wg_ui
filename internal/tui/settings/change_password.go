package settings

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/crypto/bcrypt"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
)

const (
	cpFieldCurrent = iota
	cpFieldNew
	cpFieldConfirm
	cpFieldCount
)

type ChangePasswordModel struct {
	store   *db.Store
	inputs  []textinput.Model
	focused int
	err     string
}

func NewChangePasswordModel(store *db.Store) ChangePasswordModel {
	inputs := make([]textinput.Model, cpFieldCount)

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].EchoMode = textinput.EchoPassword
		inputs[i].EchoCharacter = '•'
	}
	inputs[cpFieldCurrent].Placeholder = "Current password"
	inputs[cpFieldNew].Placeholder = "New password"
	inputs[cpFieldConfirm].Placeholder = "Confirm new password"
	inputs[cpFieldCurrent].Focus()

	return ChangePasswordModel{
		store:  store,
		inputs: inputs,
	}
}

func (m ChangePasswordModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ChangePasswordModel) Update(msg tea.Msg) (ChangePasswordModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, tui.SwitchScreen(tui.ScreenSettings)
		case "tab", "down":
			m.focused = (m.focused + 1) % cpFieldCount
			return m, m.updateFocus()
		case "shift+tab", "up":
			m.focused = (m.focused - 1 + cpFieldCount) % cpFieldCount
			return m, m.updateFocus()
		case "enter":
			if m.focused == cpFieldCount-1 {
				return m, m.save()
			}
			m.focused = (m.focused + 1) % cpFieldCount
			return m, m.updateFocus()
		}
	case tui.SuccessMsg:
		return m, tui.SwitchScreen(tui.ScreenSettings)
	case tui.ErrorMsg:
		m.err = msg.Err.Error()
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *ChangePasswordModel) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, cpFieldCount)
	for i := range m.inputs {
		if i == m.focused {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m *ChangePasswordModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, cpFieldCount)
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m ChangePasswordModel) save() tea.Cmd {
	return func() tea.Msg {
		current := m.inputs[cpFieldCurrent].Value()
		newPw := m.inputs[cpFieldNew].Value()
		confirm := m.inputs[cpFieldConfirm].Value()

		if current == "" || newPw == "" {
			return tui.ErrorMsg{Err: fmt.Errorf("all fields are required")}
		}
		if newPw != confirm {
			return tui.ErrorMsg{Err: fmt.Errorf("new passwords do not match")}
		}

		user, err := m.store.GetAdminUser()
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		if user == nil {
			return tui.ErrorMsg{Err: fmt.Errorf("admin user not found")}
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Passwd), []byte(current)); err != nil {
			return tui.ErrorMsg{Err: fmt.Errorf("current password is incorrect")}
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(newPw), bcrypt.DefaultCost)
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		if err := m.store.UpdateAdminPassword(string(hash)); err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.SuccessMsg{Message: "Password changed successfully"}
	}
}

func (m ChangePasswordModel) View() string {
	s := tui.TitleStyle.Render("Change Password") + "\n\n"

	if m.err != "" {
		s += tui.ErrorStyle.Render("Error: "+m.err) + "\n\n"
	}

	labels := []string{"Current Password ", "New Password     ", "Confirm Password "}
	for i, input := range m.inputs {
		s += tui.LabelStyle.Render(labels[i]) + " " + input.View() + "\n"
	}

	s += "\n" + tui.HelpStyle.Render("tab: next field • enter: confirm • esc: back")
	return s
}
