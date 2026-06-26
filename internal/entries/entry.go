package entries

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var entryJSONPattern = regexp.MustCompile(`^entry_(\d+)\.json$`)

type Entry struct {
	ID                 int    `json:"id"`
	Timestamp          string `json:"timestamp"`
	SubmissionType     string `json:"submission_type,omitempty"`
	TemplateID         string `json:"template_id,omitempty"`
	AccountName        string `json:"account_name"`
	PlatformProfileURL string `json:"platform_profile_url,omitempty"`
	BancampProfile     string `json:"bancamp_profile"`
	Platform           string `json:"platform"`
	IncidentType       string `json:"incident_type"`
	ReasonCategory     string `json:"reason_category"`
	Story              string `json:"story,omitempty"`
	ProofFileName      string `json:"proof_file_name,omitempty"`
	Verified           bool   `json:"verified"`
}

type SubmitInput struct {
	SubmissionType     string
	TemplateID         string
	AccountName        string
	PlatformProfileURL string
	BancampProfile     string
	Platform           string
	IncidentType       string
	ReasonCategory     string
	Story              string
	ProofExt           string
}

func (e *Entry) UnmarshalJSON(data []byte) error {
	type Alias Entry
	aux := &struct {
		ArtistName string `json:"artist_name"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if e.AccountName == "" && aux.ArtistName != "" {
		e.AccountName = aux.ArtistName
	}
	return nil
}

func EntriesDir(repoPath string) string {
	return filepath.Join(repoPath, "entries")
}

func NextID(repoPath string) (int, error) {
	dir := EntriesDir(repoPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("create entries dir: %w", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("read entries dir: %w", err)
	}

	maxID := 0
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		matches := entryJSONPattern.FindStringSubmatch(f.Name())
		if len(matches) != 2 {
			continue
		}
		id, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		if id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

func FormatID(id int) string {
	return fmt.Sprintf("%04d", id)
}

func ProofFileName(id int, ext string) string {
	return fmt.Sprintf("entry_%s_proof%s", FormatID(id), ext)
}

func JSONFileName(id int) string {
	return fmt.Sprintf("entry_%s.json", FormatID(id))
}

func Create(repoPath string, input SubmitInput) (Entry, error) {
	id, err := NextID(repoPath)
	if err != nil {
		return Entry{}, err
	}

	submissionType := input.SubmissionType
	if submissionType == "" {
		submissionType = "report"
	}

	entry := Entry{
		ID:                 id,
		Timestamp:          time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		SubmissionType:     submissionType,
		TemplateID:         input.TemplateID,
		AccountName:        input.AccountName,
		PlatformProfileURL: input.PlatformProfileURL,
		BancampProfile:     input.BancampProfile,
		Platform:           input.Platform,
		IncidentType:       input.IncidentType,
		ReasonCategory:     input.ReasonCategory,
		Story:              input.Story,
		Verified:           false,
	}
	if input.ProofExt != "" {
		entry.ProofFileName = ProofFileName(id, input.ProofExt)
	}

	dir := EntriesDir(repoPath)
	jsonPath := filepath.Join(dir, JSONFileName(id))
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return Entry{}, fmt.Errorf("marshal entry: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
		return Entry{}, fmt.Errorf("write entry json: %w", err)
	}

	return entry, nil
}

func ListAll(repoPath string) ([]Entry, error) {
	dir := EntriesDir(repoPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create entries dir: %w", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read entries dir: %w", err)
	}

	var entries []Entry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		if !entryJSONPattern.MatchString(f.Name()) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name(), err)
		}

		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil {
			return nil, fmt.Errorf("parse %s: %w", f.Name(), err)
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp > entries[j].Timestamp
	})

	return entries, nil
}