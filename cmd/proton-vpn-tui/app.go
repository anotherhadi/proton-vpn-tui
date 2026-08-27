package main

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	ilovetuiapp "github.com/anotherhadi/ilovetui/app"
	"github.com/anotherhadi/ilovetui/drawer"
	"github.com/anotherhadi/ilovetui/minsize"
	"github.com/anotherhadi/ilovetui/modal"
	"github.com/anotherhadi/ilovetui/notification"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/anotherhadi/proton-vpn-tui/internal/config"
	"github.com/anotherhadi/proton-vpn-tui/internal/ui"
)

const (
	minWidth  = 30
	minHeight = 16
)

type appModel struct {
	core    app.Model
	drawers drawer.Model
	modals  modal.Model
	notif   notification.Model
	minSize minsize.Model

	width, height int
}

func newApp() appModel {
	return appModel{
		core:    app.New(),
		drawers: drawer.New(drawer.WithMaxWidth(100000)),
		modals:  modal.New(),
		notif:   notification.New(),
		minSize: minsize.New(minWidth, minHeight),
	}
}

func (a appModel) Init() tea.Cmd {
	return tea.Batch(a.core.Init(), a.drawers.Init(), a.modals.Init(), a.notif.Init())
}

func (a appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(ilovetuiapp.QuitMsg); ok {
		return a, tea.Quit
	}

	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.width, a.height = sizeMsg.Width, sizeMsg.Height
	}

	var drawerCmd, modalCmd tea.Cmd
	a.drawers, drawerCmd = a.drawers.Update(msg)
	a.modals, modalCmd = a.modals.Update(msg)
	a.core.Overlay = a.drawers.Open() || a.modals.Open()

	helpToggled := false
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, config.Global.Keybindings.Global.Help) {
		a.core.ToggleHelp()
		helpToggled = true
	}

	var coreCmd, notifCmd tea.Cmd
	a.core, coreCmd = a.core.Update(msg)
	a.notif, notifCmd = a.notif.Update(msg)

	if _, isResize := msg.(tea.WindowSizeMsg); a.drawers.Open() && (isResize || helpToggled) {
		var cmd tea.Cmd
		a.drawers, cmd = a.drawers.Update(a.core.DrawerWindowSizeMsg())
		drawerCmd = tea.Batch(drawerCmd, cmd)
	}

	return a, tea.Batch(coreCmd, drawerCmd, modalCmd, notifCmd)
}

func (a appModel) View() tea.View {
	bg := a.core.View()
	bg = zone.Scan(bg)
	bg = a.drawers.Render(bg)
	bg = a.modals.Render(bg)
	bg = a.notif.Render(bg)
	bg = strings.Join([]string{bg, a.core.HelpView()}, "\n")
	if !a.minSize.Fits(a.width, a.height) {
		bg = a.minSize.View(a.width, a.height)
	}
	return tea.View{
		Content:     bg,
		AltScreen:   true,
		WindowTitle: a.core.WindowTitle(),
		MouseMode:   tea.MouseModeCellMotion,
	}
}
