package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/foam/draw"
	"github.com/anotherhadi/foam/geom"
	"github.com/anotherhadi/foam/minsize"
	"github.com/anotherhadi/foam/theme"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
)

// minSize is the smallest terminal proton-vpn-tui draws in.
var minSize = minsize.Size{W: 30, H: 16}

const (
	infoColWidth = 14
	loadColWidth = 8
	minNameWidth = 6

	boxOverhead = 4

	serversColumnMinWidth = 45

	rowHeight = 3 // the border and one content line
	rowGap    = 1
	rowStride = rowHeight + rowGap
)

func (m *Model) showServersColumn() bool { return m.w > serversColumnMinWidth }

func (m *Model) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.WindowTitle = m.windowTitle()
	if theme.Mouse() {
		v.MouseMode = tea.MouseModeCellMotion
	}
	if m.w == 0 || m.h == 0 {
		return v
	}
	if m.fatal != "" {
		v.SetContent(lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center,
			theme.Fg().Render(m.fatal)))
		return v
	}
	if !minSize.Fits(m.w, m.h) {
		v.SetContent(minSize.View(m.w, m.h))
		return v
	}

	header := m.renderHeader()
	headerH := lipgloss.Height(header)
	bottom := 0
	if theme.HelpBar() {
		bottom = 1
	}
	listH := max(m.h-headerH-bottom, 0)
	m.listRect = geom.Rect{X: 0, Y: headerH, W: m.w, H: listH}
	m.list.SetHeight(listH / rowStride)

	c := lipgloss.NewCanvas(m.w, m.h)
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(header),
		lipgloss.NewLayer(m.renderList(listH)).Y(headerH),
	}
	if theme.HelpBar() {
		layers = append(layers, lipgloss.NewLayer(m.helpBar()).Y(m.h-1))
	}
	c.Compose(lipgloss.NewCompositor(layers...))
	if theme.HelpBar() {
		draw.FillRow(c, m.h-1, theme.Surface())
	}
	m.overlays.Draw(c)
	if t := m.toasts(); len(t) > 0 {
		c.Compose(lipgloss.NewCompositor(t...))
	}
	v.SetContent(c.Render())
	return v
}

func (m *Model) windowTitle() string {
	const base = "ProtonVPN TUI"
	switch {
	case m.statusLoaded && m.status.Connected, !m.statusLoaded && m.quickConnected:
		return base + " - Connected"
	case m.statusLoaded:
		return base + " - Disconnected"
	default:
		return base
	}
}

func (m *Model) renderHeader() string {
	title := theme.Primary().Style().Bold(true).Render("ProtonVPN")
	title = iconVPN() + title
	lines := []string{
		title,
		"",
		m.renderStatusLine(),
		m.renderModesLine(),
	}
	if l := m.renderSearchLine(); l != "" {
		lines = append(lines, l)
	}
	lines = append(lines, "", m.renderColumnHeader())
	for i, l := range lines {
		lines[i] = clampToWidth(l, m.w)
	}
	return strings.Join(lines, "\n")
}

func label(s string) string {
	return lipgloss.NewStyle().Foreground(theme.Subtle().Value()).PaddingLeft(1).Width(10).Render(s)
}

func (m *Model) renderStatusLine() string {
	l := label("Status")
	switch {
	case m.connecting:
		return l + theme.Primary().Render(iconConnecting()+" connecting…")
	case m.disconnecting:
		return l + theme.Subtle().Render(iconLoading()+" loading…")
	case m.statusErr != nil:
		return l + theme.Warning().Render(oneLine(m.statusErr.Error()))
	case !m.statusLoaded:
		if m.quickConnected {
			return l + theme.Subtle().Render(iconConnected()+" connected, "+iconLoading()+" loading…")
		}
		return l + theme.Subtle().Render(iconLoading()+" loading…")
	case !m.status.Connected:
		return l + theme.Subtle().Render(iconDisconnected()+" disconnected")
	}
	loc := m.status.Server
	if m.status.Location != "" {
		loc = fmt.Sprintf("%s (%s)", m.status.Server, m.status.Location)
	}
	text := fmt.Sprintf("%s connected: %s", iconConnected(), loc)
	if m.status.Load > 0 {
		text += fmt.Sprintf("  %d%% load", m.status.Load)
	}
	return l + theme.Success().Render(text)
}

func (m *Model) renderSearchLine() string {
	if !m.searching && strings.TrimSpace(m.search.Value()) == "" {
		return ""
	}
	return label("Search") + m.search.View(max(m.w-14, 10), m.searching)
}

// modeBadge is one toggle in the mode line, with its name for mouse hits.
type modeBadge struct {
	name string
	text string
	on   bool
}

func (m *Model) modeBadges() []modeBadge {
	return []modeBadge{
		{"fastest", iconFastest() + " fastest", m.prefer == "fastest"},
		{"random", iconRandom() + " random", m.prefer == "random"},
		{"securecore", iconSecureCore() + " secure core", m.filters.SecureCore},
		{"tor", iconTor() + " tor", m.filters.Tor},
		{"p2p", iconP2P() + " p2p", m.filters.P2P},
		{"free", iconFree() + " free", m.filters.Free},
	}
}

func (m *Model) renderModesLine() string {
	l := label("Mode")
	m.badgeRects = map[string]geom.Rect{}
	x := ansi.StringWidth(l)
	gap := theme.Subtle().Render("  ")
	var parts []string
	for _, b := range m.modeBadges() {
		rendered := modeBadgeStyle(b.text, b.on)
		m.badgeRects[b.name] = geom.Rect{X: x, Y: 3, W: ansi.StringWidth(rendered), H: 1}
		parts = append(parts, rendered)
		x += ansi.StringWidth(rendered) + ansi.StringWidth(gap)
	}
	return l + strings.Join(parts, gap)
}

func modeBadgeStyle(text string, on bool) string {
	if on {
		return theme.Primary().Style().Bold(true).Render(text)
	}
	return theme.Subtle().Render(text)
}

func (m *Model) renderColumnHeader() string {
	indent, expandIcon := "", " "
	prefixWidth := lipgloss.Width(indent) + lipgloss.Width(expandIcon) + 1
	infoW, gaps := infoColumnLayout(m.showServersColumn())
	nameWidth := max(m.w-boxOverhead-prefixWidth-infoW-loadColWidth-gaps, minNameWidth)

	hdr := lipgloss.NewStyle().Foreground(theme.Muted().Value()).Bold(true)
	line := "  " + indent + expandIcon + " " + hdr.Render(fixedWidth("Name", nameWidth))
	if m.showServersColumn() {
		line += " " + hdr.Render(fixedWidth("Servers", infoColWidth))
	}
	return line + " " + hdr.Render(fixedWidth("Load", loadColWidth))
}

func (m *Model) renderList(h int) string {
	if !m.serversLoaded && len(m.rows) == 0 {
		text := theme.Subtle().Render(iconLoading() + " loading…")
		return lipgloss.Place(m.w, max(h, 1), lipgloss.Center, lipgloss.Center, text)
	}
	start, end := m.list.Visible()
	var lines []string
	for i := start; i < end; i++ {
		lines = append(lines, strings.Split(m.renderRow(m.rows[i], i == m.list.Selected()), "\n")...)
		lines = append(lines, "")
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return strings.Join(lines[:max(h, 0)], "\n")
}

func (m *Model) renderRow(r row, selected bool) string {
	indent := strings.Repeat("  ", int(r.kind))

	expandIcon := " "
	if r.kind == rowCountry || r.kind == rowCity {
		if r.expanded {
			expandIcon = iconExpanded()
		} else {
			expandIcon = iconCollapsed()
		}
	}

	name, info, load := m.rowColumns(r)

	nameStyle := theme.Fg().Style()
	metaStyle := theme.Subtle().Style()
	if selected {
		nameStyle = theme.Primary().Style().Bold(true)
		metaStyle = theme.Primary().Style()
	}

	prefixWidth := lipgloss.Width(indent) + lipgloss.Width(expandIcon) + 1
	infoW, gaps := infoColumnLayout(m.showServersColumn())
	nameWidth := max(m.w-boxOverhead-prefixWidth-infoW-loadColWidth-gaps, minNameWidth)

	nameCol := nameStyle.Render(fixedWidth(name, nameWidth))
	loadCol := metaStyle.Render(fixedWidth(load, loadColWidth))

	border := theme.Border()
	if !selected && r.kind != rowCountry {
		border = lipgloss.HiddenBorder()
	}

	boxStyle := lipgloss.NewStyle().Border(border).BorderForeground(rowBorderColor(r, selected)).Padding(0, 1)

	line := indent + expandIcon + " " + nameCol
	if m.showServersColumn() {
		line += " " + metaStyle.Render(fixedWidth(info, infoColWidth))
	}
	line += " " + loadCol
	return boxStyle.Render(clampToWidth(line, m.w-boxOverhead))
}

func infoColumnLayout(showInfo bool) (width, gaps int) {
	if !showInfo {
		return 0, 1
	}
	return infoColWidth, 2
}

func fixedWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true).Render(ansi.Truncate(s, width, "…"))
}

func rowBorderColor(r row, selected bool) color.Color {
	if selected {
		return theme.Primary().Value()
	}
	if r.kind == rowCountry {
		return theme.Muted().Value()
	}
	return theme.Bg().Value()
}

func (m *Model) rowColumns(r row) (name, info, load string) {
	switch r.kind {
	case rowCountry:
		countryName := backend.CountryName(r.country)
		mark := ""
		switch {
		case m.cfg.App.ShowFlags:
			mark = backend.FlagEmoji(r.country) + " "
		case iconCountry() != "":
			mark = iconCountry() + " "
		}
		fav := ""
		if r.favorite {
			fav = iconFavorite() + " "
		}
		name = fav + mark + countryName
		info = fmt.Sprintf("%d servers", r.serverCount)
		load = fmt.Sprintf("%s %d%%", iconLoad(), r.avgLoad)

	case rowCity:
		cityName := r.city
		if cityName == "" {
			cityName = "Other"
		}
		cityIcon := ""
		if iconCity() != "" {
			cityIcon = iconCity() + " "
		}
		fav := ""
		if r.favorite {
			fav = iconFavorite() + " "
		}
		name = fav + cityIcon + cityName
		info = fmt.Sprintf("%d servers", r.serverCount)
		load = fmt.Sprintf("%s %d%%", iconLoad(), r.avgLoad)

	default:
		serverIcon := ""
		if iconServer() != "" {
			serverIcon = iconServer() + " "
		}
		fav := ""
		if r.favorite {
			fav = iconFavorite() + " "
		}
		name = fav + serverIcon + r.server.Name

		var badges []string
		if r.server.SecureCore {
			badges = append(badges, iconSecureCore())
		}
		if r.server.Tor {
			badges = append(badges, iconTor())
		}
		if r.server.P2P {
			badges = append(badges, iconP2P())
		}
		info = strings.Join(badges, " ")
		load = fmt.Sprintf("%s %d%%", iconLoad(), r.server.Load)
	}
	return
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

func clampToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

func (m *Model) helpBar() string {
	ks, ls := theme.Primary().Style(), theme.Subtle().Style()
	var parts []string
	for _, h := range m.reg.Hints(m.scopes()) {
		first, _, _ := strings.Cut(h.Keys, " / ")
		parts = append(parts, ks.Render(first)+" "+ls.Render(h.Label))
	}
	return draw.Pad(ansi.Truncate(" "+strings.Join(parts, ls.Render("  ·  ")), m.w, "…"), m.w)
}

// toasts draws the live notifications at the top right, newest first.
func (m *Model) toasts() []*lipgloss.Layer {
	var out []*lipgloss.Layer
	w := min(48, max(m.w/3, 24), m.w)
	y := 0
	m.toastRects = m.toastRects[:0]
	live := m.notes.Live()
	for i := len(live) - 1; i >= 0 && i >= len(live)-5; i-- {
		s := live[i].Toast(w)
		r := geom.Rect{X: max(m.w-w, 0), Y: y, W: w, H: lipgloss.Height(s)}
		if len(out) > 0 && y+r.H > m.h {
			break
		}
		m.toastRects = append(m.toastRects, toastRect{r, live[i].ID})
		out = append(out, lipgloss.NewLayer(s).X(r.X).Y(r.Y).Z(20))
		y += r.H
	}
	return out
}
