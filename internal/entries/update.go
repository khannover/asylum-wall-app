package entries

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type UpdateInput struct {
	AccountName        string
	PlatformProfileURL string
	Platform           string
	BancampProfile     string
	IncidentType       string
	ReasonCategory     string
	Story              string
	Verified           bool
	SetVerified        bool
	ProofExt           string
	ProofData          []byte
	RemoveProof        bool
}

type UpdateResult struct {
	Entry          Entry
	OldProofFile   string
	ProofReplaced  bool
	ProofRemoved   bool
}

func GetByID(repoPath string, id int) (Entry, error) {
	if id < 1 {
		return Entry{}, fmt.Errorf("invalid entry id")
	}

	path := filepath.Join(EntriesDir(repoPath), JSONFileName(id))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, fmt.Errorf("entry not found")
		}
		return Entry{}, err
	}

	var entry Entry
	if err := json.Unmarshal(data, &entry); err != nil {
		return Entry{}, fmt.Errorf("parse entry: %w", err)
	}
	if entry.ID == 0 {
		entry.ID = id
	}
	return entry, nil
}

func Update(repoPath string, id int, input UpdateInput) (UpdateResult, error) {
	entry, err := GetByID(repoPath, id)
	if err != nil {
		return UpdateResult{}, err
	}

	oldProof := entry.ProofFileName

	if len(input.AccountName) < 2 {
		return UpdateResult{}, fmt.Errorf("account_name is required (min 2 characters)")
	}
	if input.Platform == "" || input.IncidentType == "" {
		return UpdateResult{}, fmt.Errorf("platform and incident_type are required")
	}

	entry.AccountName = input.AccountName
	entry.PlatformProfileURL = input.PlatformProfileURL
	entry.Platform = input.Platform
	entry.BancampProfile = input.BancampProfile
	entry.IncidentType = input.IncidentType
	entry.ReasonCategory = input.ReasonCategory
	entry.Story = input.Story
	entry.EditedAt = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	if input.SetVerified {
		entry.Verified = input.Verified
	}

	result := UpdateResult{Entry: entry, OldProofFile: oldProof}
	dir := EntriesDir(repoPath)

	if input.RemoveProof && oldProof != "" {
		_ = os.Remove(filepath.Join(dir, oldProof))
		entry.ProofFileName = ""
		result.ProofRemoved = true
	}

	if input.ProofExt != "" && len(input.ProofData) > 0 {
		newProof := ProofFileName(id, input.ProofExt)
		if err := os.WriteFile(filepath.Join(dir, newProof), input.ProofData, 0o644); err != nil {
			return UpdateResult{}, fmt.Errorf("write proof file: %w", err)
		}
		if oldProof != "" && oldProof != newProof {
			_ = os.Remove(filepath.Join(dir, oldProof))
		}
		entry.ProofFileName = newProof
		result.ProofReplaced = true
	}

	if err := Save(repoPath, entry); err != nil {
		return UpdateResult{}, err
	}

	result.Entry = entry
	return result, nil
}

func Save(repoPath string, entry Entry) error {
	dir := EntriesDir(repoPath)
	jsonPath := filepath.Join(dir, JSONFileName(entry.ID))
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(jsonPath, data, 0o644)
}