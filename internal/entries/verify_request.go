package entries

import (
	"fmt"
	"time"
)

const verificationRequestCooldown = 7 * 24 * time.Hour

func RequestVerification(repoPath string, id int) (Entry, error) {
	entry, err := GetByID(repoPath, id)
	if err != nil {
		return Entry{}, err
	}

	if entry.Verified {
		return Entry{}, fmt.Errorf("entry is already verified")
	}

	if entry.VerificationRequestedAt != "" {
		t, err := time.Parse(time.RFC3339, entry.VerificationRequestedAt)
		if err == nil && time.Since(t) < verificationRequestCooldown {
			return Entry{}, fmt.Errorf("verification already requested — please wait before asking again")
		}
	}

	entry.VerificationRequestedAt = time.Now().UTC().Format(time.RFC3339)
	if err := Save(repoPath, entry); err != nil {
		return Entry{}, err
	}

	return entry, nil
}