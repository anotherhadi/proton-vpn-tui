package ui

import (
	"errors"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anotherhadi/foam/form"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
)

// settingsSource adapts backend.Config to foam/form, so the settings page is
// a plain overlay instead of a hand-rolled drawer.
type settingsSource struct {
	cfg backend.Config
}

func newSettingsSource(cfg backend.Config) *settingsSource { return &settingsSource{cfg: cfg} }

func (s *settingsSource) Fields() []form.Field {
	keys := make([]string, 0, len(s.cfg))
	for k := range s.cfg {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fields := make([]form.Field, len(keys))
	for i, k := range keys {
		st := s.cfg[k]
		choices := make([]string, len(st.Options))
		for j, o := range st.Options {
			choices[j] = o.ID
		}
		fields[i] = form.Field{ID: k, Label: humanizeKebab(k), Help: st.Description, Value: st.Value, Choices: choices}
	}
	return fields
}

var errCustomDNS = errors.New("set the DNS server via the CLI: protonvpn config set custom-dns on --dns <ip[,ip...]>")

func (s *settingsSource) Set(id, value string) error {
	if id == "custom-dns" && value == "on" {
		return errCustomDNS
	}
	return s.cfg.Set(id, value)
}

// Reset has nothing to go back to: protonvpn-cli has no notion of a default.
func (s *settingsSource) Reset(id string) error {
	return errors.New("no default: set it to a value instead")
}

type settingsCacheErrMsg struct{ err error }

// refreshSettingsCache refreshes the on-disk cache in the background: the
// settings page itself already opened from whatever was cached.
func refreshSettingsCache() tea.Msg {
	if _, err := backend.LoadConfig(); err != nil {
		return settingsCacheErrMsg{err: err}
	}
	return nil
}

var settingWordOverrides = map[string]string{
	"nat": "NAT", "dns": "DNS", "vpn": "VPN", "ip": "IP",
	"ipv6": "IPv6", "netshield": "NetShield",
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
