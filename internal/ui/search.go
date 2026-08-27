package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
)

var searchEnterKey = key.NewBinding(key.WithKeys("enter"))

func (m Model) updateSearch(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	gk := config.Global.Keybindings.Global

	switch {
	case key.Matches(msg, gk.Escape):
		return m.exitSearch(true)
	case key.Matches(msg, searchEnterKey):
		return m.submitSearch()
	}

	prev := m.searchInput.Value()
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	if m.searchInput.Value() != prev {
		m.expandedCountries = map[string]bool{}
		m.expandedCities = map[string]bool{}
	}
	return m, tea.Batch(cmd, m.refreshList())
}

func (m Model) submitSearch() (Model, tea.Cmd) {
	q := strings.TrimSpace(m.searchInput.Value())
	if q == "" {
		return m.exitSearch(true)
	}

	m.searching = false
	m.searchInput.Blur()
	if id := strings.ToUpper(q); backend.IsServerID(id) {
		m.selectServerByID(id)
	}
	return m, nil
}

func (m *Model) selectServerByID(id string) {
	for i, it := range m.list.Items() {
		if r, ok := it.(row); ok && r.kind == rowServer && strings.EqualFold(r.server.Name, id) {
			m.list.Select(i)
			return
		}
	}
}

func (m Model) exitSearch(clear bool) (Model, tea.Cmd) {
	m.searching = false
	m.searchInput.Blur()
	if clear {
		m.searchInput.Reset()
	}
	cmd := m.refreshList()
	return m, cmd
}
