// Package ui is proton-vpn-tui's root model: a single list of countries,
// cities and servers, with the keys and the overlays of foam.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/proton-vpn-tui/internal/config"
)

// Run opens the TUI.
func Run(cfg *config.Config) error {
	theme.Set(cfg.Theme)
	m, err := New(cfg)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m).Run()
	return err
}
