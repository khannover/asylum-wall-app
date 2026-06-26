package entries

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRequestVerification(t *testing.T) {
	dir := t.TempDir()
	entry, err := Create(dir, SubmitInput{
		AccountName:  "testartist",
		Platform:     "Spotify",
		IncidentType: "Account Ban",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := RequestVerification(dir, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.VerificationRequestedAt == "" {
		t.Fatal("expected verification_requested_at to be set")
	}

	_, err = RequestVerification(dir, entry.ID)
	if err == nil {
		t.Fatal("expected cooldown error on second request")
	}

	path := filepath.Join(EntriesDir(dir), JSONFileName(entry.ID))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var stored Entry
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if stored.VerificationRequestedAt == "" {
		t.Fatal("expected persisted verification_requested_at")
	}

	stored.VerificationRequestedAt = time.Now().UTC().Add(-8 * 24 * time.Hour).Format(time.RFC3339)
	if err := Save(dir, stored); err != nil {
		t.Fatal(err)
	}

	_, err = RequestVerification(dir, entry.ID)
	if err != nil {
		t.Fatalf("expected request after cooldown, got %v", err)
	}
}