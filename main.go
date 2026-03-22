package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"wg_ui/internal/db"
	"wg_ui/internal/tui"
	"wg_ui/internal/tui/client"
	"wg_ui/internal/tui/menu"
	"wg_ui/internal/tui/server"
	"wg_ui/internal/tui/status"
)

type app struct {
	root    tui.RootModel
	store   *db.Store
	menu    menu.Model
	statusM status.Model
	serverV server.ViewModel
	serverF server.FormModel
	clientL client.ListModel
	clientF client.FormModel
	clientD client.DetailModel
}

func newApp(store *db.Store) *app {
	a := &app{
		root:  tui.NewRootModel(store),
		store: store,
		menu:  menu.New(),
	}

	// Wire menu
	a.root.MenuView = func() string { return a.menu.View() }
	a.root.MenuUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.menu, cmd = a.menu.Update(msg)

		// Handle menu selection
		if sel, ok := msg.(tea.KeyMsg); ok {
			_ = sel
		}
		return cmd
	}

	// Wire server view
	a.root.ServerView = func() string { return a.serverV.View() }
	a.root.ServerUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.serverV, cmd = a.serverV.Update(msg)
		return cmd
	}
	a.root.ServerInit = func() tea.Cmd {
		a.serverV = server.NewViewModel(store)
		return a.serverV.Init()
	}

	// Wire server form
	a.root.SFormView = func() string { return a.serverF.View() }
	a.root.SFormUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.serverF, cmd = a.serverF.Update(msg)
		return cmd
	}
	a.root.SFormInit = func() tea.Cmd { return a.serverF.Init() }

	// Wire client list
	a.root.ClientLView = func() string { return a.clientL.View() }
	a.root.ClientLUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.clientL, cmd = a.clientL.Update(msg)
		return cmd
	}
	a.root.ClientLInit = func() tea.Cmd { return a.clientL.Init() }

	// Wire client form
	a.root.CFFormView = func() string { return a.clientF.View() }
	a.root.CFFormUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.clientF, cmd = a.clientF.Update(msg)
		return cmd
	}
	a.root.CFFormInit = func() tea.Cmd { return a.clientF.Init() }

	// Wire client detail
	a.root.CDetailView = func() string { return a.clientD.View() }
	a.root.CDetailUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.clientD, cmd = a.clientD.Update(msg)
		return cmd
	}
	a.root.CDetailInit = func() tea.Cmd { return a.clientD.Init() }

	// Wire status
	a.root.StatusView = func() string { return a.statusM.View() }
	a.root.StatusUpdate = func(msg tea.Msg) tea.Cmd {
		var cmd tea.Cmd
		a.statusM, cmd = a.statusM.Update(msg)
		return cmd
	}
	a.root.StatusInit = func() tea.Cmd {
		a.statusM = status.New(store)
		return a.statusM.Init()
	}

	// Screen switch handler - initialize sub-models with context
	a.root.OnSwitchScreen = func(screen tui.Screen) {
		switch screen {
		case tui.ScreenServerForm:
			// Check if editing existing server
			srv, _ := store.GetServer()
			a.serverF = server.NewFormModel(store, srv)
		case tui.ScreenClientList:
			srv, _ := store.GetServer()
			if srv != nil {
				a.clientL = client.NewListModel(store, srv.ID)
			}
		case tui.ScreenClientCreate:
			srv, _ := store.GetServer()
			if srv != nil {
				a.clientF = client.NewFormModel(store, srv.ID, nil) // nil = create mode
			}
		case tui.ScreenClientForm:
			srv, _ := store.GetServer()
			if srv != nil {
				existing := a.clientL.SelectedClient()
				a.clientF = client.NewFormModel(store, srv.ID, existing)
			}
		case tui.ScreenClientDetail:
			sel := a.clientL.SelectedClient()
			a.clientD = client.NewDetailModel(sel)
		}
	}

	return a
}

func (a *app) Init() tea.Cmd {
	return a.root.Init()
}

func (a *app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Intercept menu selections
	if sel, ok := msg.(menu.SelectMsg); ok {
		switch sel.Choice {
		case menu.InstallWG:
			return a, a.root.StartInstall()
		case menu.ServerMgmt:
			return a, tui.SwitchScreen(tui.ScreenServerView)
		case menu.ClientMgmt:
			srv, _ := a.store.GetServer()
			if srv == nil {
				// Can't manage clients without a server
				return a, nil
			}
			return a, tui.SwitchScreen(tui.ScreenClientList)
		case menu.Status:
			return a, tui.SwitchScreen(tui.ScreenStatus)
		case menu.Quit:
			return a, tea.Quit
		}
	}

	model, cmd := a.root.Update(msg)
	a.root = model.(tui.RootModel)
	return a, cmd
}

func (a *app) View() string {
	return a.root.View()
}

func main() {
	store, err := db.Open("wg.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	a := newApp(store)
	p := tea.NewProgram(a, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
