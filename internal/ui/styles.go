package app

import (
	"charm.land/lipgloss/v2"

	"github.com/anotherhadi/ilovetui/style"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(style.S.Primary).
			PaddingLeft(1)

	mutedStyle     = lipgloss.NewStyle().Foreground(style.S.Muted)
	warnStyle      = lipgloss.NewStyle().Foreground(style.S.Warning)
	connectedStyle = lipgloss.NewStyle().Foreground(style.S.Success)
	primaryStyle   = lipgloss.NewStyle().Foreground(style.S.Primary)

	infoLabelStyle = lipgloss.NewStyle().
			Foreground(style.S.Muted).
			PaddingLeft(1).
			Width(10)
)
