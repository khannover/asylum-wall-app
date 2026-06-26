package entries

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalLegacyArtistName(t *testing.T) {
	raw := `{
		"id": 1,
		"timestamp": "2026-06-26T12:00:00Z",
		"artist_name": "legacy_handle",
		"platform": "TikTok",
		"incident_type": "Shadowban",
		"verified": false
	}`

	var entry Entry
	if err := json.Unmarshal([]byte(raw), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.AccountName != "legacy_handle" {
		t.Fatalf("account_name = %q, want legacy_handle", entry.AccountName)
	}
}