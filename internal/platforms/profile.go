package platforms

import (
	"net/url"
	"strings"
)

func ProfileURL(platform, accountName string) string {
	accountName = strings.TrimSpace(accountName)
	if accountName == "" {
		return ""
	}

	handle := normalizeHandle(accountName)
	platform = strings.TrimSpace(platform)

	switch platform {
	case "Bandcamp":
		if handle != "" {
			return "https://" + strings.ToLower(handle) + ".bandcamp.com"
		}
		return "https://bandcamp.com/search?q=" + url.QueryEscape(accountName)
	case "Spotify":
		return "https://open.spotify.com/search/" + url.PathEscape(accountName)
	case "YouTube":
		if looksLikeHandle(handle) {
			return "https://www.youtube.com/@" + url.PathEscape(handle)
		}
		return "https://www.youtube.com/results?search_query=" + url.QueryEscape(accountName)
	case "TikTok":
		if handle != "" {
			return "https://www.tiktok.com/@" + url.PathEscape(handle)
		}
		return "https://www.tiktok.com/search?q=" + url.QueryEscape(accountName)
	case "Instagram":
		if handle != "" {
			return "https://www.instagram.com/" + url.PathEscape(handle) + "/"
		}
		return "https://www.instagram.com/explore/search/keyword/?q=" + url.QueryEscape(accountName)
	case "Apple Music":
		return "https://music.apple.com/search?term=" + url.QueryEscape(accountName)
	case "SoundCloud":
		if handle != "" {
			return "https://soundcloud.com/" + url.PathEscape(strings.ToLower(handle))
		}
		return "https://soundcloud.com/search?q=" + url.QueryEscape(accountName)
	case "DistroKid":
		return "https://www.google.com/search?q=" + url.QueryEscape("site:distrokid.com "+accountName)
	case "Facebook":
		return "https://www.facebook.com/search/top?q=" + url.QueryEscape(accountName)
	case "X / Twitter":
		if handle != "" {
			return "https://x.com/" + url.PathEscape(handle)
		}
		return "https://x.com/search?q=" + url.QueryEscape(accountName) + "&src=typed_query"
	default:
		return ""
	}
}

func normalizeHandle(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimPrefix(name, "@")
	name = strings.TrimSpace(name)
	if strings.Contains(name, " ") {
		return ""
	}
	return name
}

func looksLikeHandle(handle string) bool {
	if handle == "" {
		return false
	}
	if len(handle) > 64 {
		return false
	}
	for _, r := range handle {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
			continue
		}
		return false
	}
	return true
}