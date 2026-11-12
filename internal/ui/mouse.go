package ui

import tea "charm.land/bubbletea/v2"

func (m *Model) mouse(msg tea.MouseMsg) tea.Cmd {
	if m.fatal != "" || !minSize.Fits(m.w, m.h) {
		return nil
	}
	if m.overlays.Top() != nil {
		return m.overlays.Mouse(msg)
	}

	mm := msg.Mouse()
	if _, click := msg.(tea.MouseClickMsg); click {
		for _, t := range m.toastRects {
			if t.Contains(mm.X, mm.Y) {
				m.notes.DismissID(t.id)
				return nil
			}
		}
		for name, r := range m.badgeRects {
			if r.Contains(mm.X, mm.Y) {
				return m.clickBadge(name)
			}
		}
	}

	if !m.listRect.Contains(mm.X, mm.Y) {
		return nil
	}
	_, y := m.listRect.Local(mm.X, mm.Y)
	switch msg.(type) {
	case tea.MouseClickMsg:
		return m.clickRow(y)
	case tea.MouseWheelMsg:
		m.list.Mouse(msg)
	}
	return nil
}

// clickRow maps a click at local line y (within the list area) to a row: each
// row is rowHeight lines tall with rowGap blank lines after it.
func (m *Model) clickRow(y int) tea.Cmd {
	if y%rowStride >= rowHeight {
		return nil // the gap between rows
	}
	start, _ := m.list.Visible()
	idx := start + y/rowStride
	if idx < 0 || idx >= len(m.rows) {
		return nil
	}
	before := m.list.Selected()
	m.list.Select(idx)
	if idx == before {
		return m.activateSelected()
	}
	return nil
}

func (m *Model) clickBadge(name string) tea.Cmd {
	switch name {
	case "fastest":
		m.prefer = "fastest"
	case "random":
		m.prefer = "random"
	case "securecore":
		m.filters.SecureCore = !m.filters.SecureCore
		m.rebuildTree()
	case "tor":
		m.filters.Tor = !m.filters.Tor
		m.rebuildTree()
	case "p2p":
		m.filters.P2P = !m.filters.P2P
		m.rebuildTree()
	case "free":
		m.filters.Free = !m.filters.Free
		m.rebuildTree()
	}
	return nil
}
