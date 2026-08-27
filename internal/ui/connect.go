package app

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/ilovetui/notification"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
)

type (
	tickMsg             time.Time
	serversMsg          []backend.LogicalServer
	serversRefreshedMsg []backend.LogicalServer

	serversErrMsg struct{ err error }
	statusMsg     backend.ConnectionStatus
	statusErrMsg  struct{ err error }

	serversCacheErrMsg struct{ err error }

	quickStatusMsg    bool
	quickStatusErrMsg struct{ err error }

	connectResultMsg struct {
		err        error
		target     string
		disconnect bool

		freeFallback bool
	}
)

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func fetchServersCache() tea.Msg {
	servers, err := backend.ParseCache()
	if err != nil {
		return serversCacheErrMsg{err: err}
	}
	return serversMsg(servers)
}

func refreshServers() tea.Msg {
	if err := backend.RefreshServerList(); err != nil {
		return serversErrMsg{err: err}
	}
	servers, err := parseCacheAfterRefresh()
	if err != nil {
		return serversErrMsg{err: err}
	}
	return serversMsg(servers)
}

func refreshServersManual() tea.Msg {
	if err := backend.RefreshServerList(); err != nil {
		return serversErrMsg{err: err}
	}
	servers, err := parseCacheAfterRefresh()
	if err != nil {
		return serversErrMsg{err: err}
	}
	return serversRefreshedMsg(servers)
}

func parseCacheAfterRefresh() ([]backend.LogicalServer, error) {
	servers, err := backend.ParseCache()
	if err == nil {
		return servers, nil
	}
	time.Sleep(300 * time.Millisecond)
	return backend.ParseCache()
}

func fetchStatus() tea.Msg {
	status, err := backend.Status()
	if err != nil {
		return statusErrMsg{err: err}
	}
	return statusMsg(status)
}

func fetchQuickStatus() tea.Msg {
	connected, err := backend.IsConnected()
	if err != nil {
		return quickStatusErrMsg{err: err}
	}
	return quickStatusMsg(connected)
}

func connectGroup(country, city, prefer string, f filters) tea.Cmd {
	return func() tea.Msg {
		target := backend.CountryName(country)
		if city != "" {
			target = city + ", " + target
		}

		err, fallback := connectWithFreeFallback(func() error {
			if prefer == "random" {
				return backend.ConnectRandom(country, city, f.P2P, f.SecureCore, f.Tor)
			}
			return backend.ConnectFastest(country, city, f.P2P, f.SecureCore, f.Tor)
		})
		if fallback {
			target = "an available free server"
		}
		return connectResultMsg{err: err, target: target, freeFallback: fallback}
	}
}

func connectServer(s backend.LogicalServer) tea.Cmd {
	return func() tea.Msg {
		target := s.Name

		err, fallback := connectWithFreeFallback(func() error { return backend.ConnectId(s.Name) })
		if fallback {
			target = "an available free server"
		}
		return connectResultMsg{err: err, target: target, freeFallback: fallback}
	}
}

func connectWithFreeFallback(attempt func() error) (err error, fallback bool) {
	err = attempt()
	if !backend.IsFreePlanRestricted(err) {
		return err, false
	}
	return backend.ConnectAny(), true
}

func disconnectCmd() tea.Cmd {
	return func() tea.Msg {
		err := backend.Disconnect()
		return connectResultMsg{err: err, disconnect: true}
	}
}

func errorToast(err error) tea.Cmd {
	return notification.Show("Error", err.Error(), notification.Error)
}
