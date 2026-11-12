package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/form"
	"github.com/anotherhadi/foam/help"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"
	"github.com/anotherhadi/foam/palette"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
)

func (m *Model) global() []commands.Command {
	return []commands.Command{
		{ID: "app.quit", Title: "Quit", Help: "quit", Run: func() tea.Cmd { return tea.Quit }},
		{ID: "app.help", Title: "Show the shortcuts", Help: "help", Run: func() tea.Cmd {
			return overlay.Open(help.New(m.reg.Groups(m.scopes())))
		}},
		{ID: "app.palette", Title: "Command palette", Run: func() tea.Cmd {
			return overlay.Open(palette.New(m.reg.Active(m.scopes()), "app.palette"))
		}},
	}
}

func (m *Model) serverCommands() []commands.Command {
	connected := func() bool { return m.statusLoaded && m.status.Connected }
	return append(m.list.Commands(),
		commands.Command{ID: "server.connect", Title: "Connect to the selected server or location", Help: "connect", Run: m.connectSelected},
		commands.Command{ID: "server.disconnect", Title: "Disconnect", Help: "disconnect", When: connected, Run: m.disconnectSelected},
		commands.Command{ID: "server.refresh", Title: "Refresh the server list", Help: "refresh", Run: func() tea.Cmd { return refreshServersManual }},
		commands.Command{ID: "server.goto-favorite", Title: "Go to the first favorite", Run: m.gotoFavorite},
		commands.Command{ID: "server.expand", Run: m.expandSelected},
		commands.Command{ID: "server.collapse", Run: m.collapseSelected},
		commands.Command{ID: "settings.open", Title: "Settings", Help: "settings", Run: m.openSettings},
		commands.Command{ID: "filter.tor", Title: "Toggle routing through Tor", Run: func() tea.Cmd {
			m.filters.Tor = !m.filters.Tor
			m.rebuildTree()
			return nil
		}},
		commands.Command{ID: "filter.securecore", Title: "Toggle Secure Core", Run: func() tea.Cmd {
			m.filters.SecureCore = !m.filters.SecureCore
			m.rebuildTree()
			return nil
		}},
		commands.Command{ID: "filter.p2p", Title: "Toggle P2P-optimized servers", Run: func() tea.Cmd {
			m.filters.P2P = !m.filters.P2P
			m.rebuildTree()
			return nil
		}},
		commands.Command{ID: "filter.free", Title: "Toggle free servers", Run: func() tea.Cmd {
			m.filters.Free = !m.filters.Free
			m.rebuildTree()
			return nil
		}},
		commands.Command{ID: "filter.prefer", Title: "Switch between fastest and random", Run: func() tea.Cmd {
			if m.prefer == "random" {
				m.prefer = "fastest"
			} else {
				m.prefer = "random"
			}
			return nil
		}},
	)
}

func (m *Model) searchCommands() []commands.Command {
	typing := func() bool { return m.searching }
	active := func() bool { return m.searching || m.search.Value() != "" }
	return []commands.Command{
		{ID: "search.open", Title: "Search servers", Help: "search", When: func() bool { return !m.searching }, Run: m.openSearch},
		{ID: "search.submit", Help: "ok", When: typing, Run: m.submitSearch},
		{ID: "search.cancel", Help: "cancel", When: active, Run: m.cancelSearch},
	}
}

// openSettings shows the settings cached on disk at once, and refreshes the
// cache in the background for next time: the CLI is too slow to block on.
func (m *Model) openSettings() tea.Cmd {
	cfg, err := backend.LoadConfigCache()
	if err != nil {
		if cfg, err = backend.LoadConfig(); err != nil {
			return notify.Err(err)
		}
	}
	return tea.Batch(overlay.Open(form.New("Settings", newSettingsSource(cfg))), refreshSettingsCache)
}

// scopes is the chain of command scopes, outermost first; search is deepest
// so its enter/esc shadow the server list's while it is active.
func (m *Model) scopes() []commands.Scope {
	return []commands.Scope{
		{Name: "proton-vpn-tui", Commands: m.global()},
		{Name: "Servers", Commands: m.serverCommands()},
		{Name: "Search", Commands: m.searchCommands()},
	}
}

// allScopes lists every scope, including foam's own overlays, for checking
// the keys at startup.
func (m *Model) allScopes() []commands.Scope {
	return append(m.scopes(),
		help.New(nil).Scope(),
		palette.New(nil, "").Scope(),
		form.New("", newSettingsSource(nil)).Scope(),
	)
}
