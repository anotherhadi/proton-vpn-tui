```
 █▀█ █▀█ █▀█ ▀█▀ █▀█ █▄ █ ▄▄ █ █ █▀█ █▄ █ ▄▄ ▀█▀ █ █ █
 █▀▀ █▀▄ █▄█  █  █▄█ █ ▀█    ▀▄▀ █▀▀ █ ▀█     █  █▄█ █
```

# ProtonVPN TUI

> A minimal, TUI and keyboard friendly wrapper for proton-vpn-cli.

![GitHub Stars](https://www.shieldcn.dev/github/stars/anotherhadi/proton-vpn-tui.svg?variant=outline&theme=violet)
![Release](https://www.shieldcn.dev/github/release/anotherhadi/proton-vpn-tui.svg?variant=outline&theme=violet)
![CI](https://www.shieldcn.dev/github/ci/anotherhadi/proton-vpn-tui.svg?variant=outline&theme=violet)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Features](#features)
- [Installation](#installation)
- [Configuration](#configuration)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

<img alt="demo" src="./.github/assets/demo.gif" width="700" />

> **⚠️ DISCLAIMER**
> This project is **NOT affiliated with**, endorsed by, or connected to Proton AG or ProtonVPN in any way. It is an unofficial, community-driven tool.

## Features

- **Server selection**: Connect to any ProtonVPN server
- **Filters**: Filter servers by Secure Core, P2P, Tor, and more
- **Free account compatible**: Works with free ProtonVPN accounts
- **Mouse support**: Navigate and interact using your mouse
- **Settings editor**: Edit settings such as Kill Switch, NetShield, and more

## Installation

<details>
<summary>Go install</summary>

```sh
go install github.com/anotherhadi/proton-vpn-tui/cmd/proton-vpn-tui@latest
```

Requires Go 1.26+. The binary will be placed in `$GOPATH/bin` (or `~/go/bin`).

</details>

<details>
<summary>NUR (Nix/NixOS)</summary>

Available via [NUR](https://github.com/nix-community/NUR), under the `anotherhadi` repo:

```nix
# configuration.nix / home.nix
environment.systemPackages = [ nur.repos.anotherhadi.proton-vpn-tui ];
```

</details>

## Configuration

Proton-VPN-TUI is fully configured via a YAML file at `~/.config/proton-vpn-tui/config.yaml`.
Check the default configuration with all the options [here](./internal/config/default.yaml), or run
`proton-vpn-tui --print-config` to see the merged configuration.

It's built on [foam](https://github.com/anotherhadi/foam), so colors, theme and keybindings shared with
other foam-based tools live in `~/.config/foam/config.yaml` and apply across all of them at once. Press
`?` for the full list of shortcuts, or `ctrl+p` for the command palette.

---

<div align="center">
  <a href="https://github.com/anotherhadi/proton-vpn-tui">github</a> |
  <a href="https://gitlab.com/anotherhadi_mirror/proton-vpn-tui">gitlab (mirror)</a> |
  <a href="https://git.hadi.icu/anotherhadi/proton-vpn-tui">gitea (mirror)</a>
</div>
