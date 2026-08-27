package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/anotherhadi/ilovetui/style"
	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
	"github.com/anotherhadi/proton-vpn-tui/internal/config"
	"github.com/anotherhadi/proton-vpn-tui/internal/icons"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	var (
		flagConfig           string
		flagAddDefaultConfig bool
	)

	rootCmd := &cobra.Command{
		Use:           "proton-vpn-tui",
		Short:         "A minimal, TUI and keyboard friendly wrapper for proton-vpn-cli.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagAddDefaultConfig {
				home, _ := os.UserHomeDir()
				cfgPath := filepath.Join(home, ".config", "proton-vpn-tui", "config.yaml")
				if flagConfig != "" {
					cfgPath = flagConfig
				}
				if err := config.WriteDefaultConfig(cfgPath); err != nil {
					return fmt.Errorf("add-default-config: %w", err)
				}
				fmt.Printf("default config written to %s\n", cfgPath)
				return nil
			}

			home, _ := os.UserHomeDir()
			cfgPath := filepath.Join(home, ".config", "proton-vpn-tui", "config.yaml")
			if flagConfig != "" {
				cfgPath = flagConfig
			}

			if err := config.Load(cfgPath); err != nil {
				return fmt.Errorf("config: %w", err)
			}
			config.Global.Version = version

			style.Init()
			icons.Init()
			zone.NewGlobal()

			if err := backend.Available(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			p := tea.NewProgram(newApp())
			if _, err := p.Run(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}

	rootCmd.Flags().StringVarP(&flagConfig, "config", "c", "", "path to config file")
	rootCmd.Flags().BoolVar(&flagAddDefaultConfig, "add-default-config", false, "copy the default config file to the config path and exit")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	if version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}
