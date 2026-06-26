package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/khannover/asylum-wall-app/internal/entries"
)

type VerificationPayload struct {
	Entry      entries.Entry
	SiteURL    string
	EntryLabel string
}

func NotifyVerificationRequest(webhookURL string, p VerificationPayload) error {
	if webhookURL == "" {
		return fmt.Errorf("discord webhook not configured")
	}

	entryURL := ""
	if p.SiteURL != "" {
		entryURL = fmt.Sprintf("%s#entry-%d", stringsTrimRightSlash(p.SiteURL), p.Entry.ID)
	}

	fields := []map[string]interface{}{
		{"name": "Entry", "value": p.EntryLabel, "inline": true},
		{"name": "Platform", "value": safeField(p.Entry.Platform), "inline": true},
		{"name": "Account", "value": safeField(p.Entry.AccountName)},
		{"name": "Incident", "value": safeField(p.Entry.IncidentType)},
	}
	if p.Entry.ProofFileName != "" {
		fields = append(fields, map[string]interface{}{"name": "Proof", "value": "Attached ✓", "inline": true})
	}
	if entryURL != "" {
		fields = append(fields, map[string]interface{}{"name": "Open", "value": entryURL})
	}

	body := map[string]interface{}{
		"username": "Asylum Wall",
		"embeds": []map[string]interface{}{
			{
				"title":       "Verification requested",
				"description": "Someone asked for this case to be reviewed.",
				"color":       0xE8B84A,
				"fields":      fields,
				"timestamp":   time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %d", resp.StatusCode)
	}
	return nil
}

func safeField(s string) string {
	if s == "" {
		return "—"
	}
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

func stringsTrimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}