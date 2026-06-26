package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/khannover/asylum-wall-app/internal/config"
	"github.com/khannover/asylum-wall-app/internal/discord"
	"github.com/khannover/asylum-wall-app/internal/entries"
	"github.com/khannover/asylum-wall-app/internal/gitrepo"
	"github.com/khannover/asylum-wall-app/internal/ratelimit"
)

type VerifyRequester struct {
	cfg     config.Config
	repo    *gitrepo.Repo
	limiter *ratelimit.Limiter
}

func NewVerifyRequester(cfg config.Config, repo *gitrepo.Repo, limiter *ratelimit.Limiter) *VerifyRequester {
	return &VerifyRequester{cfg: cfg, repo: repo, limiter: limiter}
}

func (v *VerifyRequester) HandleMetaFields() map[string]bool {
	return map[string]bool{
		"verification_requests_enabled": true,
	}
}

func (v *VerifyRequester) HandleRequest(w http.ResponseWriter, r *http.Request) {
	id, err := parseEntryID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	ip := clientIP(r)
	if !v.limiter.Allow("verify:hour:"+ip, v.cfg.VerifyMaxPerHour, time.Hour) {
		writeError(w, http.StatusTooManyRequests, "too many verification requests — try again later")
		return
	}
	if !v.limiter.Allow("verify:entry:"+fmt.Sprintf("%d:%s", id, ip), 1, 7*24*time.Hour) {
		writeError(w, http.StatusTooManyRequests, "you already requested verification for this case recently")
		return
	}

	entry, err := entries.RequestVerification(v.cfg.RepoPath, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	label := fmt.Sprintf("entry_%s", entries.FormatID(id))
	msg := fmt.Sprintf("Request verification: %s — %s (%s)", label, entry.AccountName, entry.Platform)
	if err := v.repo.CommitVerificationRequest(msg, entry.ID); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("git push failed: %v", err))
		return
	}

	if v.cfg.DiscordNotifyEnabled() {
		payload := discord.VerificationPayload{
			Entry:      entry,
			SiteURL:    v.cfg.PublicSiteURL,
			EntryLabel: strings.ToUpper(label),
		}
		if err := discord.NotifyVerificationRequest(v.cfg.DiscordWebhookURL, payload); err != nil {
			log.Printf("discord verification notify failed for %s: %v", label, err)
		} else {
			log.Printf("discord verification notify sent for %s", label)
		}
	} else if v.cfg.DiscordWebhookURL != "" {
		log.Printf("discord verification notify skipped for %s: invalid DISCORD_WEBHOOK_URL (use https://discord.com/api/webhooks/...)", label)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":                   true,
		"entry_id":                  id,
		"entry":                     entry,
		"verification_requested_at": entry.VerificationRequestedAt,
		"message":                   "Verification requested — a moderator will review this case",
	})
}