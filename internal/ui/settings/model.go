package settings

import (
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/ilovetui/bubbles"
	"github.com/anotherhadi/ilovetui/drawer"
	"github.com/anotherhadi/ilovetui/notification"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
)

type item struct {
	key     string
	setting backend.Setting
}

func (i item) FilterValue() string { return i.key }

type (
	loadedMsg backend.Config
	errMsg    struct{ err error }
)

type Model struct {
	list   list.Model
	cfg    backend.Config
	width  int
	height int
}

func New(width, height int) tea.Model {
	l := bubbles.NewList(nil, 0, 0)
	l.SetDelegate(delegate{})
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	l.SetFilteringEnabled(false)

	gk := config.Global.Keybindings.Global
	l.KeyMap.CursorUp = gk.Up
	l.KeyMap.CursorDown = gk.Down
	l.KeyMap.GoToStart = gk.GotoTop
	l.KeyMap.GoToEnd = gk.GotoBottom

	m := &Model{list: l}
	m.applySize(width, height)
	return m
}

const maxDrawerWidth = 80

func (m *Model) applySize(width, height int) {
	m.width, m.height = width, height

	boxWidth := min(int(float64(width)*0.8), maxDrawerWidth)
	inner := boxWidth - 4
	if floor := min(24, max(width-4, 1)); inner < floor {
		inner = floor
	}

	listH := height - 3
	if listH < 3 {
		listH = 3
	}
	m.list.SetSize(inner, listH)
}

func (m *Model) rebuildItems() tea.Cmd {
	keys := make([]string, 0, len(m.cfg))
	for k := range m.cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	items := make([]list.Item, len(keys))
	for i, k := range keys {
		items[i] = item{key: k, setting: m.cfg[k]}
	}
	return m.list.SetItems(items)
}

func manualSetupNotice(it item) (tea.Cmd, bool) {
	if it.key != "custom-dns" || it.setting.Value != "off" {
		return nil, false
	}
	return notification.Show(
		"Settings",
		"Custom DNS needs a server address, set it via the CLI: "+
			"protonvpn config set custom-dns on --dns <ip[,ip...]>",
		notification.Info,
	), true
}

func (m Model) selectedSetting() (item, bool) {
	if it := m.list.SelectedItem(); it != nil {
		return it.(item), true
	}
	return item{}, false
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(loadCache, loadFresh)
}

func loadCache() tea.Msg {
	cfg, err := backend.LoadConfigCache()
	if err != nil {
		return errMsg{err: err}
	}
	return loadedMsg(cfg)
}

func loadFresh() tea.Msg {
	cfg, err := backend.LoadConfig()
	if err != nil {
		return errMsg{err: err}
	}
	return loadedMsg(cfg)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applySize(msg.Width, msg.Height)
		return m, nil

	case loadedMsg:
		m.cfg = backend.Config(msg)
		cmd := m.rebuildItems()
		return m, cmd

	case errMsg:
		return m, notification.Show("Settings", msg.err.Error(), notification.Error)

	case tea.KeyPressMsg:
		gk := config.Global.Keybindings.Global
		sk := config.Global.Keybindings.Settings
		switch {
		case key.Matches(msg, gk.Escape), key.Matches(msg, gk.Quit):
			return m, drawer.Close()

		case key.Matches(msg, sk.NextOption), key.Matches(msg, sk.ToggleSetting):
			if it, ok := m.selectedSetting(); ok {
				if cmd, handled := manualSetupNotice(it); handled {
					return m, cmd
				}
				if err := m.cfg.CycleSettingUp(it.key); err != nil {
					return m, notification.Show("Settings", err.Error(), notification.Error)
				}
				cmd := m.rebuildItems()
				return m, cmd
			}
			return m, nil

		case key.Matches(msg, sk.PreviousOption):
			if it, ok := m.selectedSetting(); ok {
				if cmd, handled := manualSetupNotice(it); handled {
					return m, cmd
				}
				if err := m.cfg.CycleSettingDown(it.key); err != nil {
					return m, notification.Show("Settings", err.Error(), notification.Error)
				}
				cmd := m.rebuildItems()
				return m, cmd
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() tea.View {
	listView := strings.TrimRight(m.list.View(), "\n")
	return tea.NewView(strings.Join([]string{"", listView}, "\n"))
}
