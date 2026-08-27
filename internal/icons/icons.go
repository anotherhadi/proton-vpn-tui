package icons

import (
	"github.com/anotherhadi/ilovetui/style"
)

type Icons struct {
	Settings   string
	Vpn        string
	SecureCore string
	Tor        string
	P2p        string
	Free       string
	Fastest    string
	Random     string

	Connected    string
	Disconnected string
	Connecting   string
	Loading      string

	Country string
	City    string
	Server  string

	Expanded  string
	Collapsed string

	Favorite string

	Search string
	Load   string

	PillLeft  string
	PillRight string
}

var I *Icons

func Init() {
	if style.S.NerdFonts {
		I = &Icons{
			Settings:   "",
			Vpn:        "",
			SecureCore: "",
			Tor:        "",
			P2p:        "",
			Free:       "",
			Fastest:    "",
			Random:     "",

			Connected:    "",
			Disconnected: "",
			Connecting:   "",
			Loading:      "",

			Country: "",
			City:    "",
			Server:  "",

			Expanded:  "",
			Collapsed: "",

			Favorite: "",

			Search: "",
			Load:   "",

			PillLeft:  "",
			PillRight: "",
		}
	} else {
		I = &Icons{
			Connected:    "●",
			Disconnected: "○",
			Connecting:   "…",
			Loading:      "…",

			Expanded:  "v",
			Collapsed: ">",

			Favorite: "*",

			Search: "/",
		}
	}
}
