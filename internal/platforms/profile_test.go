package platforms

import "testing"

func TestProfileURL(t *testing.T) {
	tests := []struct {
		platform string
		account  string
		want     string
	}{
		{"TikTok", "@neonvibe", "https://www.tiktok.com/@neonvibe"},
		{"Instagram", "my_band", "https://www.instagram.com/my_band/"},
		{"X / Twitter", "@artist", "https://x.com/artist"},
		{"YouTube", "coolchannel", "https://www.youtube.com/@coolchannel"},
		{"Bandcamp", "neonvibe", "https://neonvibe.bandcamp.com"},
		{"Bandcamp", "Neon Vibe", "https://bandcamp.com/search?q=Neon+Vibe"},
		{"Spotify", "Neon Vibe", "https://open.spotify.com/search/Neon%20Vibe"},
		{"Other", "someone", ""},
	}

	for _, tt := range tests {
		got := ProfileURL(tt.platform, tt.account)
		if got != tt.want {
			t.Fatalf("ProfileURL(%q, %q) = %q, want %q", tt.platform, tt.account, got, tt.want)
		}
	}
}