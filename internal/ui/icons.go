package ui

import "github.com/anotherhadi/foam/theme"

// Icons are Nerd Font glyphs with an ASCII fallback; theme.Icon picks
// whichever the user's theme.nerdFonts setting asks for.
func iconVPN() string          { return theme.Icon("\uf023 ", "") }
func iconSecureCore() string   { return theme.Icon("\uf132", "") }
func iconTor() string          { return theme.Icon("\uf21b", "") }
func iconP2P() string          { return theme.Icon("\uf1e0", "") }
func iconFree() string         { return theme.Icon("\uf06b", "") }
func iconFastest() string      { return theme.Icon("\uf0e7", "") }
func iconRandom() string       { return theme.Icon("\uf074", "") }
func iconConnected() string    { return theme.Icon("\uf058", "●") }
func iconDisconnected() string { return theme.Icon("\uf057", "○") }
func iconConnecting() string   { return theme.Icon("\uf110", "…") }
func iconLoading() string      { return theme.Icon("\uf110", "…") }
func iconCountry() string      { return theme.Icon("\uf0ac", "") }
func iconCity() string         { return theme.Icon("\uf1ad", "") }
func iconServer() string       { return theme.Icon("\uf233", "") }
func iconExpanded() string     { return theme.Icon("\uf078", "v") }
func iconCollapsed() string    { return theme.Icon("\uf054", ">") }
func iconFavorite() string     { return theme.Icon("\uf005", "*") }
func iconSearch() string       { return theme.Icon("\uf002", "/") }
func iconLoad() string         { return theme.Icon("\uf0e4", "") }
