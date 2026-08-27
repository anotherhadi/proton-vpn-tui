package app

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

const (
	zoneBadgeFastest    = "badge-fastest"
	zoneBadgeRandom     = "badge-random"
	zoneBadgeSecureCore = "badge-securecore"
	zoneBadgeTor        = "badge-tor"
	zoneBadgeP2P        = "badge-p2p"
	zoneBadgeFree       = "badge-free"
)

func rowZoneID(index int) string {
	return fmt.Sprintf("row-%d", index)
}

func (m Model) updateMouseClick(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	if msg.Button != tea.MouseLeft {
		return m, nil
	}

	switch {
	case zone.Get(zoneBadgeFastest).InBounds(msg):
		m.prefer = "fastest"
		return m, nil

	case zone.Get(zoneBadgeRandom).InBounds(msg):
		m.prefer = "random"
		return m, nil

	case zone.Get(zoneBadgeSecureCore).InBounds(msg):
		m.filters.SecureCore = !m.filters.SecureCore
		m.tree = buildTree(m.servers, m.filters)
		return m, m.refreshList()

	case zone.Get(zoneBadgeTor).InBounds(msg):
		m.filters.Tor = !m.filters.Tor
		m.tree = buildTree(m.servers, m.filters)
		return m, m.refreshList()

	case zone.Get(zoneBadgeP2P).InBounds(msg):
		m.filters.P2P = !m.filters.P2P
		m.tree = buildTree(m.servers, m.filters)
		return m, m.refreshList()

	case zone.Get(zoneBadgeFree).InBounds(msg):
		m.filters.Free = !m.filters.Free
		m.tree = buildTree(m.servers, m.filters)
		return m, m.refreshList()
	}

	return m.clickRow(msg)
}

func (m Model) clickRow(msg tea.MouseClickMsg) (Model, tea.Cmd) {
	items := m.list.Items()
	for i := range items {
		if !zone.Get(rowZoneID(i)).InBounds(msg) {
			continue
		}
		if i != m.list.Index() {
			m.list.Select(i)
			return m, nil
		}
		return m.activateSelected()
	}
	return m, nil
}

func (m Model) activateSelected() (Model, tea.Cmd) {
	r, ok := m.selectedRow()
	if !ok {
		return m, nil
	}
	if r.kind == rowServer {
		return m.connectSelected()
	}
	if r.expanded {
		return m.collapseSelected()
	}
	return m.expandSelected()
}
