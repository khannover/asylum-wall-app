package gitrepo

import (
	"path/filepath"

	"github.com/khannover/asylum-wall-app/internal/entries"
)

func (r *Repo) CommitVerificationRequest(msg string, entryID int) error {
	jsonPath := filepath.Join("entries", entries.JSONFileName(entryID))
	return r.CommitAndPushWithMessage(msg, jsonPath)
}