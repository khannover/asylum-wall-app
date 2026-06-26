package gitrepo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/khannover/asylum-wall-app/internal/config"
)

type Repo struct {
	cfg config.Config
}

func New(cfg config.Config) *Repo {
	return &Repo{cfg: cfg}
}

func (r *Repo) Init() error {
	if r.cfg.RepoURL == "" {
		return fmt.Errorf("REPO_URL environment variable is required")
	}

	gitDir := filepath.Join(r.cfg.RepoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		if err := r.clone(); err != nil {
			return err
		}
	}

	if err := r.configure(); err != nil {
		return err
	}

	return r.sync()
}

func (r *Repo) clone() error {
	if err := os.MkdirAll(r.cfg.RepoPath, 0o755); err != nil {
		return fmt.Errorf("create repo path: %w", err)
	}

	cloneURL := r.authenticatedURL()
	cmd := exec.Command("git", "clone", "--branch", r.cfg.GitBranch, cloneURL, r.cfg.RepoPath)
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Empty GitHub repos often have no branch yet.
		if strings.Contains(string(out), "Remote branch") || strings.Contains(string(out), "not found") {
			cmd = exec.Command("git", "clone", cloneURL, r.cfg.RepoPath)
			cmd.Env = gitEnv()
			out, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("git clone failed: %s: %w", strings.TrimSpace(string(out)), err)
			}
			return nil
		}
		return fmt.Errorf("git clone failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (r *Repo) sync() error {
	hasCommits, err := r.hasCommits()
	if err != nil {
		return err
	}

	if !hasCommits {
		return r.bootstrapEmptyRepo()
	}

	_ = r.run("git", "fetch", "origin")
	return r.run("git", "pull", "--rebase", "origin", r.cfg.GitBranch)
}

func (r *Repo) bootstrapEmptyRepo() error {
	if err := r.ensureScaffold(); err != nil {
		return err
	}

	_ = r.run("git", "fetch", "origin")

	if exists, err := r.remoteBranchExists(); err == nil && exists {
		// Recover from a broken local state (e.g. staged files on an unborn branch).
		return r.run("git", "checkout", "-B", r.cfg.GitBranch, "origin/"+r.cfg.GitBranch)
	}

	if err := r.run("git", "checkout", "-B", r.cfg.GitBranch); err != nil {
		return err
	}
	if err := r.run("git", "add", "."); err != nil {
		return err
	}
	if err := r.commitIfNeeded("Initial Asylum Wall repository structure"); err != nil {
		return err
	}
	if err := r.run("git", "push", "-u", "origin", r.cfg.GitBranch); err != nil {
		return fmt.Errorf("initial push failed (ensure GITHUB_TOKEN has repo write access): %w", err)
	}
	return nil
}

func (r *Repo) ensureScaffold() error {
	entriesDir := filepath.Join(r.cfg.RepoPath, "entries")
	if err := os.MkdirAll(entriesDir, 0o755); err != nil {
		return err
	}

	readme := filepath.Join(r.cfg.RepoPath, "README.md")
	if _, err := os.Stat(readme); os.IsNotExist(err) {
		content := "# Bancamp Asylum Wall — Censorship Log\n\nPublic, append-only ledger of platform censorship incidents.\n"
		if err := os.WriteFile(readme, []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) configure() error {
	commands := [][]string{
		{"git", "config", "user.name", r.cfg.GitUserName},
		{"git", "config", "user.email", r.cfg.GitUserEmail},
		{"git", "config", "safe.directory", r.cfg.RepoPath},
	}

	for _, args := range commands {
		if err := r.run(args...); err != nil {
			return err
		}
	}

	if r.cfg.GitHubToken != "" && strings.HasPrefix(r.cfg.RepoURL, "https://") {
		return r.run("git", "remote", "set-url", "origin", r.authenticatedURL())
	}

	return nil
}

func (r *Repo) CommitAndPush(artistName, platform string, paths ...string) error {
	msg := fmt.Sprintf("New Asylum Entry: %s (%s)", artistName, platform)
	return r.CommitAndPushWithMessage(msg, paths...)
}

func (r *Repo) CommitAndPushWithMessage(msg string, paths ...string) error {
	args := append([]string{"git", "add"}, paths...)
	if err := r.run(args...); err != nil {
		return err
	}

	if err := r.run("git", "commit", "-m", msg); err != nil {
		return err
	}

	return r.run("git", "push", "origin", r.cfg.GitBranch)
}

func (r *Repo) hasCommits() (bool, error) {
	_, err := r.output("git", "rev-parse", "--verify", "HEAD")
	if err != nil {
		if isGitExitError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repo) remoteBranchExists() (bool, error) {
	out, err := r.output("git", "ls-remote", "--heads", "origin", r.cfg.GitBranch)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

func (r *Repo) commitIfNeeded(message string) error {
	out, err := r.output("git", "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "" {
		return nil
	}
	return r.run("git", "commit", "-m", message)
}

func (r *Repo) authenticatedURL() string {
	url := r.cfg.RepoURL
	if r.cfg.GitHubToken == "" || !strings.HasPrefix(url, "https://") {
		return url
	}

	trimmed := strings.TrimPrefix(url, "https://")
	return "https://x-access-token:" + r.cfg.GitHubToken + "@" + trimmed
}

func (r *Repo) run(args ...string) error {
	_, err := r.output(args...)
	return err
}

func (r *Repo) output(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = r.cfg.RepoPath
	cmd.Env = gitEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func isGitExitError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Needed a single revision") ||
		strings.Contains(msg, "bad revision") ||
		strings.Contains(msg, "unknown revision") ||
		strings.Contains(msg, "HEAD")
}

func gitEnv() []string {
	return append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=echo",
	)
}