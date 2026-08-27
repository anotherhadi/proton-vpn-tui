package app

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/anotherhadi/ilovetui/style"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
	"github.com/anotherhadi/proton-vpn-tui/internal/icons"
)

const (
	infoColWidth = 14
	loadColWidth = 8
	minNameWidth = 6

	boxOverhead = 4

	serversColumnMinWidth = 45
)

type rowDelegate struct {
	showInfo bool
}

func (d rowDelegate) Height() int { return 3 }

func (d rowDelegate) Spacing() int                            { return 1 }
func (d rowDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d rowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	r, ok := item.(row)
	if !ok {
		return
	}
	selected := index == m.Index()

	indent := strings.Repeat("  ", int(r.kind))

	expandIcon := " "
	if r.kind == rowCountry || r.kind == rowCity {
		if r.expanded {
			expandIcon = icons.I.Expanded
		} else {
			expandIcon = icons.I.Collapsed
		}
	}

	label, info, load := d.columns(r)

	nameStyle := lipgloss.NewStyle().Foreground(style.S.Text)
	metaStyle := lipgloss.NewStyle().Foreground(style.S.Muted)
	if selected {
		nameStyle = nameStyle.Foreground(style.S.Primary).Bold(true)
		metaStyle = metaStyle.Foreground(style.S.Primary)
	}

	prefixWidth := lipgloss.Width(indent) + lipgloss.Width(expandIcon) + 1
	infoW, gaps := infoColumnLayout(d.showInfo)
	nameWidth := max(m.Width()-boxOverhead-prefixWidth-infoW-loadColWidth-gaps, minNameWidth)

	nameCol := nameStyle.Render(fixedWidth(label, nameWidth))
	loadCol := metaStyle.Render(fixedWidth(load, loadColWidth))

	borderType := style.S.BorderType
	if !selected && r.kind != rowCountry {
		borderType = lipgloss.HiddenBorder()
	}

	boxStyle := lipgloss.NewStyle().
		Border(borderType).
		BorderForeground(rowBorderColor(r, selected)).
		Padding(0, 1)

	line := indent + expandIcon + " " + nameCol
	if d.showInfo {
		line += " " + metaStyle.Render(fixedWidth(info, infoColWidth))
	}
	line += " " + loadCol
	rendered := boxStyle.Render(clampToWidth(line, m.Width()-boxOverhead))
	fmt.Fprint(w, zone.Mark(rowZoneID(index), rendered))
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
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true).
		Render(ansi.Truncate(s, width, "…"))
}

func rowBorderColor(r row, selected bool) color.Color {
	if selected {
		return style.S.Primary
	}
	if r.kind == rowCountry {
		return style.S.Subtle
	}
	return style.S.Background
}

func (d rowDelegate) columns(r row) (name, info, load string) {
	switch r.kind {
	case rowCountry:
		countryName := backend.CountryName(r.country)
		mark := ""
		switch {
		case config.Global.App.ShowFlags:
			mark = backend.FlagEmoji(r.country) + " "
		case icons.I.Country != "":
			mark = icons.I.Country + " "
		}
		fav := ""
		if r.favorite {
			fav = icons.I.Favorite + " "
		}
		name = fav + mark + countryName
		info = fmt.Sprintf("%d servers", r.serverCount)
		load = fmt.Sprintf("%s %d%%", icons.I.Load, r.avgLoad)

	case rowCity:
		cityName := r.city
		if cityName == "" {
			cityName = "Other"
		}
		cityIcon := ""
		if icons.I.City != "" {
			cityIcon = icons.I.City + " "
		}
		fav := ""
		if r.favorite {
			fav = icons.I.Favorite + " "
		}
		name = fav + cityIcon + cityName
		info = fmt.Sprintf("%d servers", r.serverCount)
		load = fmt.Sprintf("%s %d%%", icons.I.Load, r.avgLoad)

	default:
		serverIcon := ""
		if icons.I.Server != "" {
			serverIcon = icons.I.Server + " "
		}
		fav := ""
		if r.favorite {
			fav = icons.I.Favorite + " "
		}
		name = fav + serverIcon + r.server.Name

		var badges []string
		if r.server.SecureCore {
			badges = append(badges, icons.I.SecureCore)
		}
		if r.server.Tor {
			badges = append(badges, icons.I.Tor)
		}
		if r.server.P2P {
			badges = append(badges, icons.I.P2p)
		}
		info = strings.Join(badges, " ")
		load = fmt.Sprintf("%s %d%%", icons.I.Load, r.server.Load)
	}
	return
}
