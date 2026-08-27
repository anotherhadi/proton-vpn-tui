package config

import (
	_ "embed"
)

//go:embed default_config.yaml
var defaultConfig []byte

type Config struct {
	Version string `mapstructure:"-"`

	App struct {
		ShowFlags           bool `mapstructure:"show_flags"`
		FocusSearchOnLaunch bool `mapstructure:"focus_search_on_launch"`

		Default struct {
			Tor        bool   `mapstructure:"tor"`
			P2P        bool   `mapstructure:"p2p"`
			SecureCore bool   `mapstructure:"securecore"`
			Free       bool   `mapstructure:"free"`
			Prefer     string `mapstructure:"prefer"`
		} `mapstructure:"default"`

		Favorites []string `mapstructure:"favorites"`
	} `mapstructure:"app"`

	Keybindings Keybindings `mapstructure:"keybindings"`
}

var Global *Config
