package config

import "testing"

func TestDiscordWebhookValidation(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"", false},
		{"https://discord.api/webhooks/your_id/your_token", false},
		{"https://discord.com/api/webhooks/123/abc", true},
		{"https://discordapp.com/api/webhooks/123/abc", true},
	}

	for _, tc := range tests {
		if got := isValidDiscordWebhookURL(normalizeWebhookURL(tc.url)); got != tc.valid {
			t.Fatalf("url %q valid=%v want %v", tc.url, got, tc.valid)
		}
	}
}