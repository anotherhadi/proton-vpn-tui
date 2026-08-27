package backend

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	featureSecureCore = 1 << iota
	featureTor
	featureP2P
	featureStreaming
	featureIPv6
)

type LogicalServer struct {
	Name         string
	EntryCountry string
	ExitCountry  string
	City         *string
	Tier         int
	SecureCore   bool
	Tor          bool
	P2P          bool
	Streaming    bool
	IPv6         bool
	Score        float64
	ID           string
	Status       int
	Load         int
}

func (s LogicalServer) IsFree() bool {
	return s.Tier == 0
}

func (s *LogicalServer) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name         string
		EntryCountry string
		ExitCountry  string
		City         *string
		Tier         int
		Features     int
		Score        float64
		ID           string
		Status       int
		Load         int
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*s = LogicalServer{
		Name:         raw.Name,
		EntryCountry: raw.EntryCountry,
		ExitCountry:  raw.ExitCountry,
		City:         raw.City,
		Tier:         raw.Tier,
		SecureCore:   raw.Features&featureSecureCore != 0,
		Tor:          raw.Features&featureTor != 0,
		P2P:          raw.Features&featureP2P != 0,
		Streaming:    raw.Features&featureStreaming != 0,
		IPv6:         raw.Features&featureIPv6 != 0,
		Score:        raw.Score,
		ID:           raw.ID,
		Status:       raw.Status,
		Load:         raw.Load,
	}
	return nil
}

func RefreshServerList() error {
	_, err := Run(context.Background(), "cities", "list", "France")
	return err
}

func ParseCache() ([]LogicalServer, error) {
	cacheHome, err := xdgCacheHome()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(cacheHome, "Proton", "VPN", "serverlist.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache struct {
		LogicalServers []LogicalServer
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return cache.LogicalServers, nil
}

func xdgCacheHome() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return dir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".cache"), nil
}
