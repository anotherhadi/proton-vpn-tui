// Package config is proton-vpn-tui's configuration: the shared foam theme
// and keys, and its own app settings, from the same layered files as
// steep and rg-tui.
package config

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"runtime/debug"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/anotherhadi/foam"
	"github.com/anotherhadi/foam/commands"
	fconfig "github.com/anotherhadi/foam/config"
	"github.com/anotherhadi/foam/theme"
)

//go:embed default.yaml
var Defaults []byte

type Config struct {
	Theme theme.Theme       `koanf:",squash"`
	Keys  commands.Bindings `koanf:"keys"`
	App   App               `koanf:"app"`
}

type App struct {
	ShowFlags           bool       `koanf:"showFlags"`
	FocusSearchOnLaunch bool       `koanf:"focusSearchOnLaunch"`
	Default             AppDefault `koanf:"default"`
	// Country code (e.g. "CH"), country name, or a server ID (e.g. "CH#20").
	Favorites []string `koanf:"favorites"`
}

type AppDefault struct {
	Tor        bool   `koanf:"tor"`
	P2P        bool   `koanf:"p2p"`
	SecureCore bool   `koanf:"securecore"`
	Free       bool   `koanf:"free"`
	Prefer     string `koanf:"prefer" validate:"oneof=fastest random"`
}

const usage = `proton-vpn-tui: a terminal UI for ProtonVPN's official CLI

Usage: proton-vpn-tui [options]`

// Load reads every layer and the command line. It returns fconfig.ErrExit
// when the command line asked for something else than running (--help,
// --version, --print-config).
func Load(args []string) (*Config, error) {
	var version bool
	cfg, _, err := fconfig.Load[Config](fconfig.Options{
		Name:  "proton-vpn-tui",
		Usage: usage,
		Base:  [][]byte{foam.Defaults, Defaults},
		Files: []string{
			filepath.Join(fconfig.Dir("foam"), "config.yaml"),
			filepath.Join(fconfig.Dir("proton-vpn-tui"), "config.yaml"),
		},
		Flags: func(fs *flag.FlagSet) {
			fs.BoolVarP(&version, "version", "v", false, "show the version")
		},
	}, args)
	if version {
		fmt.Println("proton-vpn-tui", getVersion())
		return nil, fconfig.ErrExit
	}
	return cfg, err
}

// Version is set at build time with -ldflags -X.
var Version string

func getVersion() string {
	if Version != "" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return "dev"
}
