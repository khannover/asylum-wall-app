package entries

import (
	"testing"
)

func TestUpdateEntry(t *testing.T) {
	dir := t.TempDir()

	created, err := Create(dir, SubmitInput{
		AccountName:  "oldname",
		Platform:     "Spotify",
		IncidentType: "Account Ban",
		Story:        "original",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := Update(dir, created.ID, UpdateInput{
		AccountName:        "newname",
		PlatformProfileURL: "https://open.spotify.com/search/newname",
		Platform:           "YouTube",
		IncidentType:       "Shadowban",
		Story:              "updated story",
		SetVerified:        true,
		Verified:           true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if updated.Entry.AccountName != "newname" {
		t.Fatalf("account = %q", updated.Entry.AccountName)
	}
	if updated.Entry.EditedAt == "" {
		t.Fatal("edited_at should be set")
	}
	if !updated.Entry.Verified {
		t.Fatal("verified should be true")
	}

	loaded, err := GetByID(dir, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Platform != "YouTube" {
		t.Fatalf("platform = %q", loaded.Platform)
	}
}