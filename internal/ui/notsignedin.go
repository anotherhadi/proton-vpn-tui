package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/ilovetui/app"
)

type notSignedInModal struct{}

func (notSignedInModal) Init() tea.Cmd { return nil }

func (notSignedInModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyPressMsg); ok {
		return notSignedInModal{}, app.Quit()
	}
	return notSignedInModal{}, nil
}

func (notSignedInModal) View() tea.View {
	return tea.NewView("You're not signed in.\n\nRun `protonvpn signin` to sign in,\nthen restart proton-vpn-tui.\n\nPress any key to quit.")
}
