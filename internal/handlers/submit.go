package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/khannover/asylum-wall-app/internal/config"
	"github.com/khannover/asylum-wall-app/internal/entries"
	"github.com/khannover/asylum-wall-app/internal/gitrepo"
	"github.com/khannover/asylum-wall-app/internal/platforms"
	"github.com/khannover/asylum-wall-app/internal/ratelimit"
	"github.com/khannover/asylum-wall-app/internal/templates"
	"github.com/khannover/asylum-wall-app/internal/upload"
)

type Submitter struct {
	cfg     config.Config
	repo    *gitrepo.Repo
	limiter *ratelimit.Limiter
}

func NewSubmitter(cfg config.Config, repo *gitrepo.Repo, limiter *ratelimit.Limiter) *Submitter {
	return &Submitter{cfg: cfg, repo: repo, limiter: limiter}
}

func (s *Submitter) HandleSignal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(upload.MaxProofSize + 1024*1024); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	if honeypot := strings.TrimSpace(r.FormValue("website")); honeypot != "" {
		writeError(w, http.StatusBadRequest, "submission rejected")
		return
	}

	accountName := readAccountName(r)
	platform := sanitizeField(r.FormValue("platform"), 200)
	templateID := sanitizeField(r.FormValue("template_id"), 64)
	bancampProfile := sanitizeField(r.FormValue("bancamp_profile"), 500)
	story := sanitizeField(r.FormValue("story"), 10000)

	if len(accountName) < 2 {
		writeError(w, http.StatusBadRequest, "account_name is required (min 2 characters)")
		return
	}
	if platform == "" {
		writeError(w, http.StatusBadRequest, "platform is required")
		return
	}

	tmpl, ok := templates.ByID(templateID)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid template_id")
		return
	}

	ip := clientIP(r)
	if !s.checkSignalLimits(w, ip, templateID, platform, accountName) {
		return
	}

	profileURL := platforms.ProfileURL(platform, accountName)

	proofData, proofExt, err := readOptionalProof(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	entry, err := entries.Create(s.cfg.RepoPath, entries.SubmitInput{
		SubmissionType:     "signal",
		TemplateID:         templateID,
		AccountName:        accountName,
		PlatformProfileURL: profileURL,
		BancampProfile:     bancampProfile,
		Platform:           platform,
		IncidentType:       tmpl.IncidentType,
		ReasonCategory:     tmpl.ReasonCategory,
		Story:              story,
		ProofExt:           proofExt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gitPaths := []string{filepath.Join("entries", entries.JSONFileName(entry.ID))}

	if proofExt != "" {
		proofPath := filepath.Join(entries.EntriesDir(s.cfg.RepoPath), entry.ProofFileName)
		if err := os.WriteFile(proofPath, proofData, 0o644); err != nil {
			_ = os.Remove(filepath.Join(entries.EntriesDir(s.cfg.RepoPath), entries.JSONFileName(entry.ID)))
			writeError(w, http.StatusInternalServerError, "failed to save proof file")
			return
		}
		gitPaths = append(gitPaths, filepath.Join("entries", entry.ProofFileName))
	}

	commitMsg := fmt.Sprintf("Signal: %s — %s (%s)", tmpl.Title, accountName, platform)
	if err := s.repo.CommitAndPushWithMessage(commitMsg, gitPaths...); err != nil {
		s.rollbackEntry(entry, proofExt != "")
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("git push failed: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"entry_id": entry.ID,
		"message":  fmt.Sprintf("Signal recorded as entry_%s", entries.FormatID(entry.ID)),
	})
}

func (s *Submitter) HandleReport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(upload.MaxProofSize + 1024*1024); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	if honeypot := strings.TrimSpace(r.FormValue("website")); honeypot != "" {
		writeError(w, http.StatusBadRequest, "submission rejected")
		return
	}

	accountName := readAccountName(r)
	bancampProfile := sanitizeField(r.FormValue("bancamp_profile"), 500)
	platform := sanitizeField(r.FormValue("platform"), 200)
	incidentType := sanitizeField(r.FormValue("incident_type"), 200)
	reasonCategory := sanitizeField(r.FormValue("reason_category"), 200)
	story := sanitizeField(r.FormValue("story"), 10000)

	if len(accountName) < 2 || platform == "" || incidentType == "" {
		writeError(w, http.StatusBadRequest, "account_name, platform, and incident_type are required")
		return
	}

	profileURL := platforms.ProfileURL(platform, accountName)

	ip := clientIP(r)
	if !s.limiter.Allow("report:hour:"+ip, 3, time.Hour) {
		writeError(w, http.StatusTooManyRequests, "too many full reports — try again later")
		return
	}

	proofData, proofExt, err := readOptionalProof(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	entry, err := entries.Create(s.cfg.RepoPath, entries.SubmitInput{
		SubmissionType:     "report",
		AccountName:        accountName,
		PlatformProfileURL: profileURL,
		BancampProfile:     bancampProfile,
		Platform:           platform,
		IncidentType:       incidentType,
		ReasonCategory:     reasonCategory,
		Story:              story,
		ProofExt:           proofExt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gitPaths := []string{filepath.Join("entries", entries.JSONFileName(entry.ID))}
	if proofExt != "" {
		proofPath := filepath.Join(entries.EntriesDir(s.cfg.RepoPath), entry.ProofFileName)
		if err := os.WriteFile(proofPath, proofData, 0o644); err != nil {
			_ = os.Remove(filepath.Join(entries.EntriesDir(s.cfg.RepoPath), entries.JSONFileName(entry.ID)))
			writeError(w, http.StatusInternalServerError, "failed to save proof file")
			return
		}
		gitPaths = append(gitPaths, filepath.Join("entries", entry.ProofFileName))
	}

	if err := s.repo.CommitAndPush(accountName, platform, gitPaths...); err != nil {
		s.rollbackEntry(entry, proofExt != "")
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("git push failed: %v", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"entry_id": entry.ID,
		"message":  fmt.Sprintf("Case logged as entry_%s", entries.FormatID(entry.ID)),
	})
}

func readAccountName(r *http.Request) string {
	name := sanitizeField(r.FormValue("account_name"), 200)
	if name == "" {
		name = sanitizeField(r.FormValue("artist_name"), 200)
	}
	return name
}

func (s *Submitter) checkSignalLimits(w http.ResponseWriter, ip, templateID, platform, accountName string) bool {
	if !s.limiter.Allow("signal:hour:"+ip, s.cfg.SignalMaxPerHour, time.Hour) {
		writeError(w, http.StatusTooManyRequests, "rate limit — max signals per hour reached, try again later")
		return false
	}
	if !s.limiter.Allow("signal:day:"+ip, s.cfg.SignalMaxPerDay, 24*time.Hour) {
		writeError(w, http.StatusTooManyRequests, "rate limit — max signals per day reached")
		return false
	}

	dupKey := fmt.Sprintf("signal:dup:%s:%s:%s:%s",
		ip, templateID, normalizeKey(platform), normalizeKey(accountName))
	if !s.limiter.Allow(dupKey, 1, s.cfg.SignalDuplicateTTL) {
		writeError(w, http.StatusTooManyRequests, "you already signaled this issue recently")
		return false
	}
	return true
}

func (s *Submitter) rollbackEntry(entry entries.Entry, hasProof bool) {
	dir := entries.EntriesDir(s.cfg.RepoPath)
	if hasProof {
		_ = os.Remove(filepath.Join(dir, entry.ProofFileName))
	}
	_ = os.Remove(filepath.Join(dir, entries.JSONFileName(entry.ID)))
}

func readOptionalProof(r *http.Request) ([]byte, string, error) {
	if r.MultipartForm == nil || len(r.MultipartForm.File["proof_file"]) == 0 {
		return nil, "", nil
	}

	file, header, err := r.FormFile("proof_file")
	if err != nil {
		return nil, "", nil
	}
	defer file.Close()

	ext, err := upload.SanitizeExtension(header.Filename)
	if err != nil {
		return nil, "", err
	}

	data, err := upload.ReadLimited(file, upload.MaxProofSize)
	if err != nil {
		return nil, "", err
	}

	if err := upload.ValidateContent(data, ext); err != nil {
		return nil, "", err
	}

	return data, ext, nil
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if real := r.Header.Get("X-Real-IP"); real != "" {
		return strings.TrimSpace(real)
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		return host[:i]
	}
	return host
}

func normalizeKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func sanitizeField(value string, maxLen int) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\x00", "")
	if len(value) > maxLen {
		value = value[:maxLen]
	}
	return value
}

