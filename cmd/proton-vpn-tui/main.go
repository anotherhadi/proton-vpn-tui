// Command proton-vpn-tui is a minimal, keyboard-friendly terminal UI wrapper
// for protonvpn-cli.
package main

import (
	"errors"
	"fmt"
	"os"

	fconfig "github.com/anotherhadi/foam/config"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
	"github.com/anotherhadi/proton-vpn-tui/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "proton-vpn-tui:", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load(os.Args[1:])
	if errors.Is(err, fconfig.ErrExit) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := backend.Available(); err != nil {
		return err
	}
	return ui.Run(cfg)
}
