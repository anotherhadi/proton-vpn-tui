package config

import "charm.land/bubbles/v2/key"

type GlobalKeys struct {
	Quit   key.Binding `mapstructure:"quit"   desc:"quit"  help:"short"`
	Escape key.Binding `mapstructure:"escape" desc:"close" help:"none"`
	Help   key.Binding `mapstructure:"help" desc:"help" help:"none"`

	Up    key.Binding `mapstructure:"up"    desc:"up"    help:"full"`
	Down  key.Binding `mapstructure:"down"  desc:"down"  help:"full"`
	Left  key.Binding `mapstructure:"left"  desc:"left"  help:"full"`
	Right key.Binding `mapstructure:"right" desc:"right" help:"full"`

	GotoTop    key.Binding `mapstructure:"goto_top"    desc:"go to top"    help:"full"`
	GotoBottom key.Binding `mapstructure:"goto_bottom" desc:"go to bottom" help:"full"`
}

type HomeKeys struct {
	OpenSettings key.Binding `mapstructure:"open_settings" desc:"settings"         help:"full"`
	Connect      key.Binding `mapstructure:"connect"       desc:"connect"          help:"short"`
	Disconnect   key.Binding `mapstructure:"disconnect"    desc:"disconnect"       help:"short"`
	GotoFavorite key.Binding `mapstructure:"goto_favorite" desc:"go to favorites"  help:"full"`
	Search       key.Binding `mapstructure:"search"        desc:"search"           help:"short"`
	Refresh      key.Binding `mapstructure:"refresh"       desc:"refresh"          help:"full"`

	ToggleTor        key.Binding `mapstructure:"toggle_tor"        desc:"toggle tor"         help:"full"`
	ToggleSecureCore key.Binding `mapstructure:"toggle_securecore" desc:"toggle secure core" help:"full"`
	ToggleP2P        key.Binding `mapstructure:"toggle_p2p"        desc:"toggle p2p"         help:"full"`
	ToggleFree       key.Binding `mapstructure:"toggle_free"       desc:"toggle free"        help:"full"`
	TogglePrefer     key.Binding `mapstructure:"toggle_prefer"     desc:"fastest/random"     help:"full"`
}

type SettingsKeys struct {
	PreviousOption key.Binding `mapstructure:"previous_option" desc:"decrease" help:"full"`
	NextOption     key.Binding `mapstructure:"next_option"     desc:"increase" help:"full"`
	ToggleSetting  key.Binding `mapstructure:"toggle_setting" desc:"toggle" help:"short"`
}

type Keybindings struct {
	Global   GlobalKeys   `mapstructure:"global"`
	Home     HomeKeys     `mapstructure:"home"`
	Settings SettingsKeys `mapstructure:"settings"`
}
