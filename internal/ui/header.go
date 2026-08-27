package app

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/anotherhadi/ilovetui/style"

	"github.com/anotherhadi/proton-vpn-tui/internal/icons"
)

func (m Model) renderHeader() string {
	title := headerStyle.Render("ProtonVPN")
	if style.S.NerdFonts {
		title = icons.I.Vpn + title
	}
	lines := []string{
		title,
		"",
		m.renderStatusLine(),
		m.renderModesLine(),
	}
	if searchLine := m.renderSearchLine(); searchLine != "" {
		lines = append(lines, searchLine)
	}
	lines = append(lines, "", m.renderColumnHeader())
	for i, l := range lines {
		lines[i] = clampToWidth(l, m.width)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderColumnHeader() string {
	indent, expandIcon := "", " "
	prefixWidth := lipgloss.Width(indent) + lipgloss.Width(expandIcon) + 1
	infoW, gaps := infoColumnLayout(m.showServersColumn())
	nameWidth := max(m.width-boxOverhead-prefixWidth-infoW-loadColWidth-gaps, minNameWidth)

	hdrStyle := lipgloss.NewStyle().Foreground(style.S.Subtle).Bold(true)

	line := "  " + indent + expandIcon + " " + hdrStyle.Render(fixedWidth("Name", nameWidth))
	if m.showServersColumn() {
		line += " " + hdrStyle.Render(fixedWidth("Servers", infoColWidth))
	}
	return line + " " + hdrStyle.Render(fixedWidth("Load", loadColWidth))
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func clampToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

func (m Model) renderStatusLine() string {
	label := infoLabelStyle.Render("Status")

	if m.connecting {
		return label + primaryStyle.Render(icons.I.Connecting+" connecting…")
	}
	if m.disconnecting {
		return label + mutedStyle.Render(icons.I.Loading+" loading…")
	}
	if m.statusErr != nil {
		return label + warnStyle.Render(oneLine(m.statusErr.Error()))
	}
	if !m.statusLoaded {
		if m.quickConnected {
			return label + mutedStyle.Render(icons.I.Connected+" connected, "+icons.I.Loading+" loading…")
		}
		return label + mutedStyle.Render(icons.I.Loading+" loading…")
	}
	if !m.status.Connected {
		return label + mutedStyle.Render(icons.I.Disconnected+" disconnected")
	}

	loc := m.status.Server
	if m.status.Location != "" {
		loc = fmt.Sprintf("%s (%s)", m.status.Server, m.status.Location)
	}
	text := fmt.Sprintf("%s connected: %s", icons.I.Connected, loc)
	if m.status.Load > 0 {
		text += fmt.Sprintf("  %d%% load", m.status.Load)
	}
	return label + connectedStyle.Render(text)
}

func (m Model) renderSearchLine() string {
	if !m.searching && strings.TrimSpace(m.searchInput.Value()) == "" {
		return ""
	}
	return infoLabelStyle.Render("Search") + m.searchInput.View()
}

func (m Model) renderModesLine() string {
	label := infoLabelStyle.Render("Mode")
	parts := []string{
		zone.Mark(zoneBadgeFastest, modeBadge(icons.I.Fastest+" fastest", m.prefer == "fastest")),
		zone.Mark(zoneBadgeRandom, modeBadge(icons.I.Random+" random", m.prefer == "random")),
		zone.Mark(zoneBadgeSecureCore, modeBadge(icons.I.SecureCore+" secure core", m.filters.SecureCore)),
		zone.Mark(zoneBadgeTor, modeBadge(icons.I.Tor+" tor", m.filters.Tor)),
		zone.Mark(zoneBadgeP2P, modeBadge(icons.I.P2p+" p2p", m.filters.P2P)),
		zone.Mark(zoneBadgeFree, modeBadge(icons.I.Free+" free", m.filters.Free)),
	}
	return label + strings.Join(parts, mutedStyle.Render("  "))
}

func modeBadge(text string, active bool) string {
	if active {
		return primaryStyle.Bold(true).Render(text)
	}
	return mutedStyle.Render(text)
}
