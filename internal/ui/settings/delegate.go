package settings

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/anotherhadi/ilovetui/style"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/icons"
)

func clampToWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

type delegate struct{}

func (d delegate) Height() int                             { return 2 }
func (d delegate) Spacing() int                            { return 1 }
func (d delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(item)
	if !ok {
		return
	}
	selected := index == m.Index()

	var titleStyle, descStyle, valueStyle lipgloss.Style
	if selected {
		titleStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(style.S.Primary).
			Foreground(style.S.Primary).
			Bold(true).
			PaddingLeft(1)
		descStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(style.S.Primary).
			Foreground(style.S.Muted).
			PaddingLeft(1)
		valueStyle = lipgloss.NewStyle().Foreground(style.S.Success).Bold(true)
	} else {
		titleStyle = lipgloss.NewStyle().Foreground(style.S.Text).PaddingLeft(2)
		descStyle = lipgloss.NewStyle().Foreground(style.S.Muted).PaddingLeft(2)
		valueStyle = lipgloss.NewStyle().Foreground(style.S.Success)
	}

	value := renderCycleValue(valueStyle.Render(humanizeKebab(it.setting.Value)))
	if isOnOffSetting(it.setting) {
		value = renderToggle(it.setting.Value == "on")
	}

	width := m.Width()
	title := titleStyle.Render(humanizeKebab(it.key))
	line1 := padLine(justifyBetween(title, value, width), width)
	line2 := padLine(descStyle.Render(it.setting.Description), width)
	fmt.Fprintf(w, "%s\n%s", line1, line2)
}

func justifyBetween(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func padLine(s string, width int) string {
	s = clampToWidth(s, width)
	if pad := width - lipgloss.Width(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

func isOnOffSetting(s backend.Setting) bool {
	if len(s.Options) != 2 {
		return false
	}
	var hasOn, hasOff bool
	for _, o := range s.Options {
		hasOn = hasOn || o.ID == "on"
		hasOff = hasOff || o.ID == "off"
	}
	return hasOn && hasOff
}

func renderCycleValue(value string) string {
	arrow := lipgloss.NewStyle().Foreground(style.S.Muted)
	return arrow.Render("< ") + value + arrow.Render(" >")
}

func renderToggle(on bool) string {
	if on {
		return renderBadge("  ●", style.S.Success)
	}
	return renderBadge("●  ", style.S.Subtle)
}

func renderBadge(label string, fill color.Color) string {
	body := lipgloss.NewStyle().
		Background(fill).
		Foreground(style.S.Background).
		Bold(true).
		Render(label)

	if icons.I.PillLeft == "" && icons.I.PillRight == "" {
		return body
	}

	cap := lipgloss.NewStyle().Foreground(fill).Background(style.S.Background)
	return cap.Render(icons.I.PillLeft) + body + cap.Render(icons.I.PillRight)
}

var settingWordOverrides = map[string]string{
	"nat":       "NAT",
	"dns":       "DNS",
	"vpn":       "VPN",
	"ip":        "IP",
	"ipv6":      "IPv6",
	"netshield": "NetShield",
}

func humanizeKebab(s string) string {
	words := strings.Split(s, "-")
	for i, w := range words {
		if override, ok := settingWordOverrides[strings.ToLower(w)]; ok {
			words[i] = override
			continue
		}
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
