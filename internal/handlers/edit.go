package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/khannover/asylum-wall-app/internal/config"
	"github.com/khannover/asylum-wall-app/internal/entries"
	"github.com/khannover/asylum-wall-app/internal/gitrepo"
	"github.com/khannover/asylum-wall-app/internal/platforms"
	"github.com/khannover/asylum-wall-app/internal/ratelimit"
	"github.com/khannover/asylum-wall-app/internal/upload"
)

type Editor struct {
	cfg     config.Config
	repo    *gitrepo.Repo
	limiter *ratelimit.Limiter
}

func NewEditor(cfg config.Config, repo *gitrepo.Repo, limiter *ratelimit.Limiter) *Editor {
	return &Editor{cfg: cfg, repo: repo, limiter: limiter}
}

func (e *Editor) HandleEdit(w http.ResponseWriter, r *http.Request) {
	if e.cfg.EditToken == "" {
		writeError(w, http.StatusForbidden, "editing is not enabled on this server")
		return
	}

	token := strings.TrimSpace(r.Header.Get("X-Edit-Token"))
	if token == "" || token != e.cfg.EditToken {
		writeError(w, http.StatusUnauthorized, "invalid edit token")
		return
	}

	id, err := parseEntryID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	ip := clientIP(r)
	if !e.limiter.Allow("edit:hour:"+ip, 30, time.Hour) {
		writeError(w, http.StatusTooManyRequests, "too many edits — try again later")
		return
	}

	if err := r.ParseMultipartForm(upload.MaxProofSize + 1024*1024); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	accountName := readAccountName(r)
	platform := sanitizeField(r.FormValue("platform"), 200)
	incidentType := sanitizeField(r.FormValue("incident_type"), 200)
	if len(accountName) < 2 || platform == "" || incidentType == "" {
		writeError(w, http.StatusBadRequest, "account_name, platform, and incident_type are required")
		return
	}

	input := entries.UpdateInput{
		AccountName:        accountName,
		PlatformProfileURL: platforms.ProfileURL(platform, accountName),
		Platform:           platform,
		BancampProfile:     sanitizeField(r.FormValue("bancamp_profile"), 500),
		IncidentType:       incidentType,
		ReasonCategory:     sanitizeField(r.FormValue("reason_category"), 200),
		Story:              sanitizeField(r.FormValue("story"), 10000),
		RemoveProof:        r.FormValue("remove_proof") == "true",
	}

	input.SetVerified = true
	input.Verified = r.FormValue("verified") == "true"

	proofData, proofExt, err := readOptionalProof(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	input.ProofExt = proofExt
	input.ProofData = proofData

	result, err := entries.Update(e.cfg.RepoPath, id, input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	msg := fmt.Sprintf("Edit entry_%s: %s (%s)", entries.FormatID(id), accountName, platform)
	if err := e.repo.CommitEntryEdit(msg, result); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("git push failed: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"entry_id": id,
		"entry":    result.Entry,
		"message":  fmt.Sprintf("Entry_%s updated and committed to Git", entries.FormatID(id)),
	})
}

func parseEntryID(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(strings.ToLower(raw), "entry_")
	return strconv.Atoi(raw)
}