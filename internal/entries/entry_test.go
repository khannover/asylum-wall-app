package entries

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNextIDAndCreate(t *testing.T) {
	dir := t.TempDir()

	id1, err := NextID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id1 != 1 {
		t.Fatalf("first id = %d, want 1", id1)
	}

	entry, err := Create(dir, SubmitInput{
		AccountName:        "test_artist",
		PlatformProfileURL: "https://open.spotify.com/search/test_artist",
		BancampProfile:     "https://bancamp.de/artist/test",
		Platform:           "Spotify",
		IncidentType:       "Account Ban",
		ReasonCategory:     "Manual Abuse",
		Story:              "Test story",
		ProofExt:           ".png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.ID != 1 {
		t.Fatalf("entry id = %d, want 1", entry.ID)
	}
	if entry.Verified {
		t.Fatal("new entries should not be verified")
	}

	id2, err := NextID(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id2 != 2 {
		t.Fatalf("second id = %d, want 2", id2)
	}

	jsonPath := filepath.Join(EntriesDir(dir), JSONFileName(1))
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("json file missing: %v", err)
	}
}

func TestListAllSortsByTimestamp(t *testing.T) {
	dir := t.TempDir()

	for _, artist := range []string{"A", "B"} {
		if _, err := Create(dir, SubmitInput{
			AccountName:  artist,
			Platform:     "YouTube",
			IncidentType: "Shadowban",
			Story:        "story",
			ProofExt:     ".pdf",
		}); err != nil {
			t.Fatal(err)
		}
	}

	list, err := ListAll(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d entries, want 2", len(list))
	}
	if list[0].Timestamp < list[1].Timestamp {
		t.Fatal("expected descending timestamp order")
	}
}