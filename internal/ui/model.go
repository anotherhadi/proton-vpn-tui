package app

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/anotherhadi/ilovetui/app"
	"github.com/anotherhadi/ilovetui/bubbles"
	"github.com/anotherhadi/ilovetui/drawer"
	"github.com/anotherhadi/ilovetui/helpbar"
	"github.com/anotherhadi/ilovetui/modal"
	"github.com/anotherhadi/ilovetui/notification"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
	"github.com/anotherhadi/proton-vpn-tui/internal/icons"
	"github.com/anotherhadi/proton-vpn-tui/internal/ui/settings"
)

type filters struct {
	Tor        bool
	SecureCore bool
	P2P        bool
	Free       bool
}

type Model struct {
	list list.Model
	help helpbar.Model

	width, height int

	servers       []backend.LogicalServer
	tree          []*countryGroup
	serversLoaded bool

	expandedCountries map[string]bool
	expandedCities    map[string]bool

	filters filters
	prefer  string

	status         backend.ConnectionStatus
	statusErr      error
	statusLoaded   bool
	quickConnected bool
	connecting     bool
	disconnecting  bool

	searching   bool
	searchInput textinput.Model

	Overlay bool
}

func New() Model {
	l := bubbles.NewList(nil, 0, 0)
	l.SetDelegate(rowDelegate{})
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()

	gk := config.Global.Keybindings.Global
	l.KeyMap.CursorUp = gk.Up
	l.KeyMap.CursorDown = gk.Down
	l.KeyMap.GoToStart = gk.GotoTop
	l.KeyMap.GoToEnd = gk.GotoBottom

	h := helpbar.New(helpbar.WithToggle(gk.Help))

	si := bubbles.NewTextInput()
	si.Prompt = icons.I.Search + " "
	si.Placeholder = "country, city, or server id…"

	focusSearch := config.Global.App.FocusSearchOnLaunch
	if focusSearch {
		si.Focus()
	}

	def := config.Global.App.Default
	return Model{
		list:              l,
		help:              h,
		searchInput:       si,
		searching:         focusSearch,
		expandedCountries: map[string]bool{},
		expandedCities:    map[string]bool{},
		filters:           filters{Tor: def.Tor, SecureCore: def.SecureCore, P2P: def.P2P, Free: def.Free},
		prefer:            def.Prefer,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchServersCache, refreshServers,
		fetchQuickStatus, fetchStatus,
		tickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	m, cmd := m.update(msg)
	m.resizeList()
	return m, cmd
}

func (m Model) update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.SetWidth(msg.Width)
		m.searchInput.SetWidth(max(msg.Width-14, 10))
		return m, nil

	case tickMsg:
		return m, tea.Batch(fetchStatus, tickCmd())

	case serversMsg:
		m.servers = msg
		m.serversLoaded = true
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, cmd

	case serversRefreshedMsg:
		m.servers = msg
		m.serversLoaded = true
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, tea.Batch(cmd, notification.Show("ProtonVPN", "Server list refreshed from cache", notification.Success))

	case serversErrMsg:
		m.serversLoaded = true
		if backend.IsAuthRequired(msg.err) {
			return m, modal.Show("Not signed in", notSignedInModal{})
		}
		return m, errorToast(msg.err)

	case serversCacheErrMsg:
		return m, nil

	case quickStatusMsg:
		if !m.statusLoaded {
			m.quickConnected = bool(msg)
		}
		return m, nil

	case quickStatusErrMsg:
		return m, nil

	case statusMsg:
		m.status = backend.ConnectionStatus(msg)
		m.statusErr = nil
		m.statusLoaded = true
		return m, nil

	case statusErrMsg:
		m.statusErr = msg.err
		m.status = backend.ConnectionStatus{}
		m.statusLoaded = true
		return m, nil

	case connectResultMsg:
		m.connecting = false
		m.disconnecting = false
		if msg.err != nil {
			return m, tea.Batch(errorToast(msg.err), fetchStatus)
		}
		title, text := "ProtonVPN", fmt.Sprintf("Connected to %s", msg.target)
		switch {
		case msg.disconnect:
			text = "Disconnected"
		case msg.freeFallback:
			title = "Free plan"
			text = "Free accounts can't pick a server or location. Connected you to an available free server instead."
		}
		return m, tea.Batch(fetchStatus, notification.Show(title, text, notification.Success))

	case tea.KeyPressMsg:
		if m.Overlay {
			return m, nil
		}
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateList(msg)

	case tea.MouseWheelMsg:
		if m.Overlay {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseWheelUp:
			m.list.CursorUp()
		case tea.MouseWheelDown:
			m.list.CursorDown()
		}
		return m, nil

	case tea.MouseClickMsg:
		if m.Overlay {
			return m, nil
		}
		return m.updateMouseClick(msg)
	}

	if m.Overlay {
		return m, nil
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateList(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	gk := config.Global.Keybindings.Global
	hk := config.Global.Keybindings.Home

	switch {
	case key.Matches(msg, gk.Quit):
		return m, app.Quit()

	case key.Matches(msg, hk.OpenSettings):
		title := "Settings"
		if icons.I.Settings != "" {
			title = icons.I.Settings + " " + title
		}
		return m, drawer.Show(title, settings.New(m.width, m.drawerContentHeight()), drawer.WithSide(drawer.Right))

	case key.Matches(msg, hk.Search):
		m.searching = true
		return m, m.searchInput.Focus()

	case key.Matches(msg, gk.Escape):
		if strings.TrimSpace(m.searchInput.Value()) == "" {
			return m, nil
		}
		m.searchInput.Reset()
		cmd := m.refreshList()
		return m, cmd

	case key.Matches(msg, hk.Refresh):
		return m, refreshServersManual

	case key.Matches(msg, hk.Disconnect):
		m.disconnecting = true
		return m, disconnectCmd()

	case key.Matches(msg, hk.ToggleTor):
		m.filters.Tor = !m.filters.Tor
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, cmd

	case key.Matches(msg, hk.ToggleSecureCore):
		m.filters.SecureCore = !m.filters.SecureCore
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, cmd

	case key.Matches(msg, hk.ToggleP2P):
		m.filters.P2P = !m.filters.P2P
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, cmd

	case key.Matches(msg, hk.ToggleFree):
		m.filters.Free = !m.filters.Free
		m.tree = buildTree(m.servers, m.filters)
		cmd := m.refreshList()
		return m, cmd

	case key.Matches(msg, hk.TogglePrefer):
		if m.prefer == "random" {
			m.prefer = "fastest"
		} else {
			m.prefer = "random"
		}
		return m, nil

	case key.Matches(msg, hk.GotoFavorite):
		m.gotoFavorite()
		return m, nil

	case key.Matches(msg, gk.Right):
		return m.expandSelected()

	case key.Matches(msg, gk.Left):
		return m.collapseSelected()

	case key.Matches(msg, hk.Connect):
		return m.connectSelected()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) selectedRow() (row, bool) {
	if item := m.list.SelectedItem(); item != nil {
		return item.(row), true
	}
	return row{}, false
}

func (m Model) expandSelected() (Model, tea.Cmd) {
	r, ok := m.selectedRow()
	if !ok {
		return m, nil
	}
	switch r.kind {
	case rowCountry:
		m.expandedCountries[r.country] = true
	case rowCity:
		m.expandedCities[cityKey(r.country, r.city)] = true
	default:
		return m, nil
	}
	cmd := m.refreshList()
	return m, cmd
}

func (m Model) collapseSelected() (Model, tea.Cmd) {
	r, ok := m.selectedRow()
	if !ok {
		return m, nil
	}
	switch r.kind {
	case rowCountry:
		m.expandedCountries[r.country] = false
	case rowCity, rowServer:
		m.expandedCities[cityKey(r.country, r.city)] = false
	}
	cmd := m.refreshList()
	return m, cmd
}

func (m *Model) gotoFavorite() {
	for i, item := range m.list.Items() {
		if r, ok := item.(row); ok && r.kind == rowCountry && r.favorite {
			m.list.Select(i)
			return
		}
	}
}

func (m Model) connectSelected() (Model, tea.Cmd) {
	r, ok := m.selectedRow()
	if !ok {
		return m, nil
	}
	m.connecting = true
	if r.kind == rowServer {
		return m, connectServer(r.server)
	}
	return m, connectGroup(r.country, r.city, m.prefer, m.filters)
}

func (m Model) favoritesSet() (countries, servers map[string]bool) {
	countries = map[string]bool{}
	servers = map[string]bool{}
	for _, raw := range config.Global.App.Favorites {
		f := strings.TrimSpace(raw)
		if f == "" {
			continue
		}
		if id := strings.ToUpper(f); backend.IsServerID(id) {
			servers[id] = true
			continue
		}
		if code, ok := backend.ResolveCountryCode(f); ok {
			countries[code] = true
		}
	}
	return
}

func (m *Model) refreshList() tea.Cmd {
	q := strings.TrimSpace(m.searchInput.Value())
	favCountries, favServers := m.favoritesSet()
	var rows []row
	if q != "" {
		rows = searchRows(m.tree, q, m.expandedCountries, m.expandedCities, favCountries, favServers)
	} else {
		rows = flattenRows(m.tree, m.expandedCountries, m.expandedCities, favCountries, favServers)
	}
	return m.setListItems(rows)
}

func (m *Model) setListItems(rows []row) tea.Cmd {
	items := make([]list.Item, len(rows))
	for i, r := range rows {
		items[i] = r
	}
	cmd := m.list.SetItems(items)
	m.resizeList()
	if idx := m.list.Index(); len(items) > 0 && (idx < 0 || idx >= len(items)) {
		m.list.Select(len(items) - 1)
	}
	return cmd
}

func (m *Model) resizeList() {
	m.list.SetDelegate(rowDelegate{showInfo: m.showServersColumn()})
	m.list.SetSize(m.width, m.listHeight())
}

func (m Model) showServersColumn() bool {
	return m.width > serversColumnMinWidth
}

func (m Model) listHeight() int {
	headerH := lipgloss.Height(m.renderHeader())
	helpH := m.help.Height(m.helpBindings()...)
	return m.height - headerH - helpH
}

func (m Model) View() string {
	header := m.renderHeader()
	listView := strings.TrimRight(m.list.View(), "\n")
	if !m.serversLoaded && len(m.list.Items()) == 0 {
		listView = m.loadingView()
	}
	return strings.Join([]string{header, listView}, "\n")
}

func (m Model) loadingView() string {
	text := mutedStyle.Render(icons.I.Loading + " loading…")
	return lipgloss.Place(m.width, max(m.listHeight(), 1), lipgloss.Center, lipgloss.Center, text)
}

func (m Model) WindowTitle() string {
	const base = "ProtonVPN TUI"
	switch {
	case m.statusLoaded && m.status.Connected:
		return base + " - Connected"
	case m.statusLoaded:
		return base + " - Disconnected"
	case m.quickConnected:
		return base + " - Connected"
	default:
		return base
	}
}

func (m Model) HelpView() string {
	return strings.TrimRight(m.help.View(m.helpBindings()...), "\n")
}

func (m *Model) ToggleHelp() {
	m.help.ShowAll = !m.help.ShowAll
	m.resizeList()
}

func (m Model) settingsHelpHeight() int {
	pages := config.Help(config.Global.Keybindings.Global, config.Global.Keybindings.Settings)
	return m.help.Height(m.bindingsFor(pages)...)
}

func (m Model) drawerContentHeight() int {
	return m.height - m.settingsHelpHeight()
}

func (m Model) DrawerWindowSizeMsg() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: m.width, Height: m.drawerContentHeight()}
}

func (m Model) bindingsFor(pages config.HelpKeyMap) []key.Binding {
	if !m.help.ShowAll {
		return pages.ShortHelp()
	}
	var all []key.Binding
	for _, col := range pages.FullHelp() {
		all = append(all, col...)
	}
	return all
}

func (m Model) helpBindings() []key.Binding {
	pages := config.Help(config.Global.Keybindings.Global, config.Global.Keybindings.Home)
	if m.Overlay {
		pages = config.Help(config.Global.Keybindings.Global, config.Global.Keybindings.Settings)
	}
	return m.bindingsFor(pages)
}
