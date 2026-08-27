package app

import (
	"sort"
	"strings"

	"github.com/anotherhadi/proton-vpn-tui/internal/backend"
)

type rowKind int

const (
	rowCountry rowKind = iota
	rowCity
	rowServer
)

type row struct {
	kind rowKind

	country string
	city    string
	server  backend.LogicalServer

	expanded bool
	favorite bool

	serverCount int
	avgLoad     int
}

func (r row) FilterValue() string {
	switch r.kind {
	case rowServer:
		return r.server.Name + " " + r.city + " " + backend.CountryName(r.country)
	case rowCity:
		return r.city + " " + backend.CountryName(r.country)
	default:
		return r.country + " " + backend.CountryName(r.country)
	}
}

func cityKey(country, city string) string { return country + "\x00" + city }

type countryGroup struct {
	code      string
	cities    map[string][]backend.LogicalServer
	cityOrder []string
	servers   []backend.LogicalServer
}

func buildTree(servers []backend.LogicalServer, f filters) []*countryGroup {
	groups := make(map[string]*countryGroup)
	var order []string

	for _, s := range servers {
		if s.Tor != f.Tor || s.SecureCore != f.SecureCore {
			continue
		}
		if f.P2P && !s.P2P {
			continue
		}
		if f.Free && !s.IsFree() {
			continue
		}

		code := s.EntryCountry
		if s.SecureCore {
			code = s.ExitCountry
		}

		g, ok := groups[code]
		if !ok {
			g = &countryGroup{code: code, cities: make(map[string][]backend.LogicalServer)}
			groups[code] = g
			order = append(order, code)
		}

		city := ""
		if s.City != nil {
			city = *s.City
		}
		if _, ok := g.cities[city]; !ok {
			g.cityOrder = append(g.cityOrder, city)
		}
		g.cities[city] = append(g.cities[city], s)
		g.servers = append(g.servers, s)
	}

	result := make([]*countryGroup, 0, len(order))
	for _, code := range order {
		result = append(result, groups[code])
	}
	sort.Slice(result, func(i, j int) bool {
		return backend.CountryName(result[i].code) < backend.CountryName(result[j].code)
	})
	for _, g := range result {
		sort.Slice(g.cityOrder, func(i, j int) bool {
			a, b := g.cityOrder[i], g.cityOrder[j]
			if a == "" {
				return false
			}
			if b == "" {
				return true
			}
			return a < b
		})
		for _, servers := range g.cities {
			sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
		}
	}
	return result
}

func avgLoad(servers []backend.LogicalServer) int {
	if len(servers) == 0 {
		return 0
	}
	total := 0
	for _, s := range servers {
		total += s.Load
	}
	return total / len(servers)
}

func serverIsFavorite(s backend.LogicalServer, favServers map[string]bool) bool {
	return favServers[strings.ToUpper(s.Name)]
}

func serversHaveFavorite(servers []backend.LogicalServer, favServers map[string]bool) bool {
	for _, s := range servers {
		if serverIsFavorite(s, favServers) {
			return true
		}
	}
	return false
}

func countryIsFavorite(g *countryGroup, favCountries, favServers map[string]bool) bool {
	return favCountries[g.code] || serversHaveFavorite(g.servers, favServers)
}

func orderWithFavoritesFirst(tree []*countryGroup, favCountries, favServers map[string]bool) []*countryGroup {
	ordered := append([]*countryGroup(nil), tree...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return countryIsFavorite(ordered[i], favCountries, favServers) &&
			!countryIsFavorite(ordered[j], favCountries, favServers)
	})
	return ordered
}

func orderCitiesWithFavoritesFirst(g *countryGroup, favServers map[string]bool) []string {
	cities := append([]string(nil), g.cityOrder...)
	sort.SliceStable(cities, func(i, j int) bool {
		return serversHaveFavorite(g.cities[cities[i]], favServers) &&
			!serversHaveFavorite(g.cities[cities[j]], favServers)
	})
	return cities
}

func orderServersWithFavoritesFirst(servers []backend.LogicalServer, favServers map[string]bool) []backend.LogicalServer {
	ordered := append([]backend.LogicalServer(nil), servers...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return serverIsFavorite(ordered[i], favServers) && !serverIsFavorite(ordered[j], favServers)
	})
	return ordered
}

func flattenRows(tree []*countryGroup, expandedCountries, expandedCities, favCountries, favServers map[string]bool) []row {
	ordered := orderWithFavoritesFirst(tree, favCountries, favServers)

	var rows []row
	for _, g := range ordered {
		rows = append(rows, row{
			kind:        rowCountry,
			country:     g.code,
			expanded:    expandedCountries[g.code],
			favorite:    countryIsFavorite(g, favCountries, favServers),
			serverCount: len(g.servers),
			avgLoad:     avgLoad(g.servers),
		})
		if !expandedCountries[g.code] {
			continue
		}
		for _, city := range orderCitiesWithFavoritesFirst(g, favServers) {
			servers := orderServersWithFavoritesFirst(g.cities[city], favServers)
			key := cityKey(g.code, city)
			rows = append(rows, row{
				kind:        rowCity,
				country:     g.code,
				city:        city,
				expanded:    expandedCities[key],
				favorite:    serversHaveFavorite(servers, favServers),
				serverCount: len(servers),
				avgLoad:     avgLoad(servers),
			})
			if !expandedCities[key] {
				continue
			}
			for _, s := range servers {
				rows = append(rows, row{
					kind:     rowServer,
					country:  g.code,
					city:     city,
					server:   s,
					favorite: serverIsFavorite(s, favServers),
				})
			}
		}
	}
	return rows
}

type cityMatch struct {
	city    string
	servers []backend.LogicalServer
}

func searchRows(tree []*countryGroup, query string, expandedCountries, expandedCities, favCountries, favServers map[string]bool) []row {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	ordered := orderWithFavoritesFirst(tree, favCountries, favServers)
	matchCities := len([]rune(q)) > 3

	var rows []row
	for _, g := range ordered {
		countryMatch := strings.Contains(strings.ToLower(backend.CountryName(g.code)), q) ||
			strings.Contains(strings.ToLower(g.code), q)

		var matches []cityMatch
		if matchCities {
			for _, city := range g.cityOrder {
				if servers := matchingCityServers(city, g.cities[city], q); len(servers) > 0 {
					matches = append(matches, cityMatch{city: city, servers: servers})
				}
			}
		}

		if !countryMatch && len(matches) == 0 {
			continue
		}

		expanded := len(matches) > 0
		if v, ok := expandedCountries[g.code]; ok {
			expanded = v
		}

		rows = append(rows, row{
			kind:        rowCountry,
			country:     g.code,
			expanded:    expanded,
			favorite:    countryIsFavorite(g, favCountries, favServers),
			serverCount: len(g.servers),
			avgLoad:     avgLoad(g.servers),
		})
		if !expanded {
			continue
		}

		if len(matches) == 0 {
			for _, city := range g.cityOrder {
				matches = append(matches, cityMatch{city: city, servers: g.cities[city]})
			}
		}
		sort.SliceStable(matches, func(i, j int) bool {
			return serversHaveFavorite(matches[i].servers, favServers) &&
				!serversHaveFavorite(matches[j].servers, favServers)
		})
		for _, cm := range matches {
			cityExpanded := true
			if v, ok := expandedCities[cityKey(g.code, cm.city)]; ok {
				cityExpanded = v
			}
			servers := orderServersWithFavoritesFirst(cm.servers, favServers)
			rows = append(rows, row{
				kind:        rowCity,
				country:     g.code,
				city:        cm.city,
				expanded:    cityExpanded,
				favorite:    serversHaveFavorite(servers, favServers),
				serverCount: len(servers),
				avgLoad:     avgLoad(servers),
			})
			if !cityExpanded {
				continue
			}
			for _, s := range servers {
				rows = append(rows, row{
					kind:     rowServer,
					country:  g.code,
					city:     cm.city,
					server:   s,
					favorite: serverIsFavorite(s, favServers),
				})
			}
		}
	}
	return rows
}

func matchingCityServers(city string, servers []backend.LogicalServer, q string) []backend.LogicalServer {
	label := city
	if label == "" {
		label = "other"
	}
	if strings.Contains(strings.ToLower(label), q) {
		return servers
	}
	var matched []backend.LogicalServer
	for _, s := range servers {
		if strings.Contains(strings.ToLower(s.Name), q) {
			matched = append(matched, s)
		}
	}
	return matched
}
