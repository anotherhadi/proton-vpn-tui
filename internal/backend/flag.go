package backend

import "strings"

func countryCodeFromServerId(id string) string {
	if i := strings.IndexAny(id, "-#"); i != -1 {
		return id[:i]
	}
	return id
}

func FlagEmoji(countryCode string) string {
	code := strings.ToUpper(countryCode)
	if code == "UK" {
		code = "GB"
	}
	if len(code) != 2 {
		return "🏳️"
	}

	var flag strings.Builder
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return "🏳️"
		}
		flag.WriteRune(r + 127397)
	}

	return flag.String()
}

func FlagEmojiFromServerId(serverId string) string {
	return FlagEmoji(countryCodeFromServerId(serverId))
}
