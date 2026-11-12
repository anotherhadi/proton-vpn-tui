package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/commands"
	"github.com/anotherhadi/foam/geom"
	"github.com/anotherhadi/foam/input"
	"github.com/anotherhadi/foam/list"
	"github.com/anotherhadi/foam/notify"
	"github.com/anotherhadi/foam/overlay"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
)

type filters struct {
	Tor        bool
	SecureCore bool
	P2P        bool
	Free       bool
}

type toastRect struct {
	geom.Rect
	id int
}

type Model struct {
	cfg *config.Config
	reg *commands.Registry

	rows []row
	list list.Model

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

	searching bool
	search    input.Model

	overlays overlay.Stack
	notes    notify.Center

	// fatal replaces the whole UI with a message; any key quits.
	fatal string

	w, h int

	listRect   geom.Rect
	badgeRects map[string]geom.Rect
	toastRects []toastRect
}

func New(cfg *config.Config) (*Model, error) {
	reg, err := commands.New(cfg.Keys)
	if err != nil {
		return nil, fmt.Errorf("invalid config:\n%w", err)
	}

	def := cfg.App.Default
	m := &Model{
		cfg:               cfg,
		reg:               reg,
		expandedCountries: map[string]bool{},
		expandedCities:    map[string]bool{},
		filters:           filters{Tor: def.Tor, SecureCore: def.SecureCore, P2P: def.P2P, Free: def.Free},
		prefer:            def.Prefer,
	}
	m.search = input.New(iconSearch() + " ")
	m.search.Placeholder = "country, city, or server id…"
	m.searching = cfg.App.FocusSearchOnLaunch
	m.notes.Max = 100

	if err := reg.Check(m.allScopes()); err != nil {
		return nil, fmt.Errorf("invalid config:\n%w", err)
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(fetchServersCache, refreshServers, fetchQuickStatus, fetchStatus, tickCmd())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, m.update(msg)
}

func (m *Model) update(msg tea.Msg) tea.Cmd {
	if m.fatal != "" {
		if _, ok := msg.(tea.KeyPressMsg); ok {
			return tea.Quit
		}
		return nil
	}
	if m.overlays.Handle(msg) {
		return nil
	}
	if cmd, ok := m.notes.Update(msg); ok {
		return cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height

	case tea.KeyPressMsg:
		return m.key(msg)

	case tea.PasteMsg:
		return m.paste(msg)

	case tea.MouseMsg:
		return m.mouse(msg)

	case tickMsg:
		return tea.Batch(fetchStatus, tickCmd())

	case serversMsg:
		m.servers = msg
		m.serversLoaded = true
		m.rebuildTree()

	case serversRefreshedMsg:
		m.servers = msg
		m.serversLoaded = true
		m.rebuildTree()
		return notify.Success("Server list refreshed from cache")

	case serversErrMsg:
		m.serversLoaded = true
		if backend.IsAuthRequired(msg.err) {
			m.fatal = "You're not signed in.\n\n" +
				"Run `protonvpn signin` to sign in,\nthen restart proton-vpn-tui.\n\n" +
				"Press any key to quit."
			return nil
		}
		return notify.Err(msg.err)

	case serversCacheErrMsg:
		return nil

	case settingsCacheErrMsg:
		return notify.Err(msg.err)

	case quickStatusMsg:
		if !m.statusLoaded {
			m.quickConnected = bool(msg)
		}

	case quickStatusErrMsg:
		return nil

	case statusMsg:
		m.status = backend.ConnectionStatus(msg)
		m.statusErr = nil
		m.statusLoaded = true

	case statusErrMsg:
		m.statusErr = msg.err
		m.status = backend.ConnectionStatus{}
		m.statusLoaded = true

	case connectResultMsg:
		m.connecting = false
		m.disconnecting = false
		if msg.err != nil {
			return tea.Batch(notify.Err(msg.err), fetchStatus)
		}
		text := fmt.Sprintf("Connected to %s", msg.target)
		switch {
		case msg.disconnect:
			text = "Disconnected"
		case msg.freeFallback:
			text = "Free accounts can't pick a server or location. Connected you to an available free server instead."
		}
		return tea.Batch(fetchStatus, notify.Success(text))
	}
	return nil
}

func (m *Model) key(msg tea.KeyPressMsg) tea.Cmd {
	if !minSize.Fits(m.w, m.h) {
		if b, ok := m.reg.Match(m.scopes()[:1], msg); ok && b.ID == "app.quit" {
			return b.Run()
		}
		return nil
	}
	if m.overlays.Top() != nil {
		return m.overlays.Key(m.reg, msg)
	}
	if m.searching {
		before := m.search.Value()
		if m.search.Update(msg) {
			if m.search.Value() != before {
				m.clearExpanded()
				m.refreshRows()
			}
			return nil
		}
	}
	if b, ok := m.reg.Match(m.scopes(), msg); ok {
		return b.Run()
	}
	return nil
}

func (m *Model) paste(msg tea.PasteMsg) tea.Cmd {
	if o := m.overlays.Top(); o != nil {
		return o.Update(msg)
	}
	if !m.searching {
		return nil
	}
	before := m.search.Value()
	if m.search.Update(msg) && m.search.Value() != before {
		m.clearExpanded()
		m.refreshRows()
	}
	return nil
}

func (m *Model) clearExpanded() {
	m.expandedCountries = map[string]bool{}
	m.expandedCities = map[string]bool{}
}

func (m *Model) rebuildTree() {
	m.tree = buildTree(m.servers, m.filters)
	m.refreshRows()
}

func (m *Model) favoritesSet() (countries, servers map[string]bool) {
	countries = map[string]bool{}
	servers = map[string]bool{}
	for _, raw := range m.cfg.App.Favorites {
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

func (m *Model) refreshRows() {
	q := strings.TrimSpace(m.search.Value())
	favCountries, favServers := m.favoritesSet()
	if q != "" {
		m.rows = searchRows(m.tree, q, m.expandedCountries, m.expandedCities, favCountries, favServers)
	} else {
		m.rows = flattenRows(m.tree, m.expandedCountries, m.expandedCities, favCountries, favServers)
	}
	m.list.SetLen(len(m.rows))
}

func (m *Model) selectedRow() (row, bool) {
	if len(m.rows) == 0 {
		return row{}, false
	}
	return m.rows[m.list.Selected()], true
}

func (m *Model) gotoFavorite() tea.Cmd {
	for i, r := range m.rows {
		if r.kind == rowCountry && r.favorite {
			m.list.Select(i)
			return nil
		}
	}
	return nil
}

func (m *Model) expandSelected() tea.Cmd {
	r, ok := m.selectedRow()
	if !ok {
		return nil
	}
	switch r.kind {
	case rowCountry:
		m.expandedCountries[r.country] = true
	case rowCity:
		m.expandedCities[cityKey(r.country, r.city)] = true
	default:
		return nil
	}
	m.refreshRows()
	return nil
}

func (m *Model) collapseSelected() tea.Cmd {
	r, ok := m.selectedRow()
	if !ok {
		return nil
	}
	switch r.kind {
	case rowCountry:
		m.expandedCountries[r.country] = false
	case rowCity, rowServer:
		m.expandedCities[cityKey(r.country, r.city)] = false
	}
	m.refreshRows()
	return nil
}

func (m *Model) activateSelected() tea.Cmd {
	r, ok := m.selectedRow()
	if !ok {
		return nil
	}
	if r.kind == rowServer {
		return m.connectSelected()
	}
	if r.expanded {
		return m.collapseSelected()
	}
	return m.expandSelected()
}

func (m *Model) connectSelected() tea.Cmd {
	r, ok := m.selectedRow()
	if !ok {
		return nil
	}
	m.connecting = true
	if r.kind == rowServer {
		return connectServer(r.server)
	}
	return connectGroup(r.country, r.city, m.prefer, m.filters)
}

func (m *Model) disconnectSelected() tea.Cmd {
	m.disconnecting = true
	return disconnectCmd()
}

func (m *Model) openSearch() tea.Cmd {
	m.searching = true
	return nil
}

func (m *Model) cancelSearch() tea.Cmd {
	m.searching = false
	m.search.Reset()
	m.refreshRows()
	return nil
}

func (m *Model) submitSearch() tea.Cmd {
	q := strings.TrimSpace(m.search.Value())
	if q == "" {
		return m.cancelSearch()
	}
	m.searching = false
	if id := strings.ToUpper(q); backend.IsServerID(id) {
		m.selectServerByID(id)
	}
	return nil
}

func (m *Model) selectServerByID(id string) {
	for i, r := range m.rows {
		if r.kind == rowServer && strings.EqualFold(r.server.Name, id) {
			m.list.Select(i)
			return
		}
	}
}
