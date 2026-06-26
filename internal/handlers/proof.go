package handlers

import (
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/khannover/asylum-wall-app/internal/entries"
)

var proofFilePattern = regexp.MustCompile(`^entry_\d{4}_proof\.(png|jpe?g|gif|webp|pdf|eml)$`)

func ServeProof(repoPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/proof/")
		name = filepath.Base(name)

		if !proofFilePattern.MatchString(name) {
			http.NotFound(w, r)
			return
		}

		proofPath := filepath.Join(entries.EntriesDir(repoPath), name)
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".gif":
			w.Header().Set("Content-Type", "image/gif")
		case ".webp":
			w.Header().Set("Content-Type", "image/webp")
		case ".pdf":
			w.Header().Set("Content-Type", "application/pdf")
		case ".eml":
			w.Header().Set("Content-Type", "message/rfc822")
		default:
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, proofPath)
	}
}