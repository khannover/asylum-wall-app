package gitrepo

import (
	"os"
	"path/filepath"

	"github.com/khannover/asylum-wall-app/internal/entries"
)

func (r *Repo) CommitEntryEdit(msg string, result entries.UpdateResult) error {
	jsonPath := filepath.Join("entries", entries.JSONFileName(result.Entry.ID))
	paths := []string{jsonPath}

	if result.ProofReplaced && result.Entry.ProofFileName != "" {
		paths = append(paths, filepath.Join("entries", result.Entry.ProofFileName))
	}

	if result.ProofRemoved && result.OldProofFile != "" {
		oldRel := filepath.Join("entries", result.OldProofFile)
		_ = r.run("git", "rm", "-f", oldRel)
	}

	if result.ProofReplaced && result.OldProofFile != "" && result.OldProofFile != result.Entry.ProofFileName {
		oldRel := filepath.Join("entries", result.OldProofFile)
		_ = r.run("git", "rm", "-f", oldRel)
		_ = os.Remove(filepath.Join(r.cfg.RepoPath, oldRel))
	}

	args := append([]string{"git", "add"}, paths...)
	if err := r.run(args...); err != nil {
		return err
	}
	return r.commitAndPush(msg)
}