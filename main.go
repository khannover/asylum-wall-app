package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/khannover/asylum-wall-app/internal/config"
	"github.com/khannover/asylum-wall-app/internal/entries"
	"github.com/khannover/asylum-wall-app/internal/gitrepo"
	"github.com/khannover/asylum-wall-app/internal/handlers"
	"github.com/khannover/asylum-wall-app/internal/ratelimit"
	"github.com/khannover/asylum-wall-app/internal/templates"
)

type server struct {
	cfg       config.Config
	submitter *handlers.Submitter
}

func main() {
	cfg := config.Load()
	repo := gitrepo.New(cfg)

	log.Println("initializing git repository...")
	if err := repo.Init(); err != nil {
		log.Fatalf("git init failed: %v", err)
	}
	log.Println("git repository ready")

	limiter := ratelimit.New()
	s := &server{
		cfg:       cfg,
		submitter: handlers.NewSubmitter(cfg, repo, limiter),
	}

	mux := http.NewServeMux()

	mux.Handle("GET /{$}", handlers.Static())
	mux.Handle("GET /assets/", handlers.Static())
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /api/templates", s.handleTemplates)
	mux.HandleFunc("GET /api/cases", s.handleListCases)
	mux.HandleFunc("GET /api/proof/{file}", handlers.ServeProof(cfg.RepoPath))
	mux.HandleFunc("POST /api/signal", s.submitter.HandleSignal)
	mux.HandleFunc("POST /api/submit-case", s.submitter.HandleReport)

	handler := withCORS(mux)
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("asylum wall server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleListCases(w http.ResponseWriter, r *http.Request) {
	cases, err := entries.ListAll(s.cfg.RepoPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if cases == nil {
		cases = []entries.Entry{}
	}
	writeJSON(w, http.StatusOK, cases)
}

func (s *server) handleTemplates(w http.ResponseWriter, r *http.Request) {
	cases, err := entries.ListAll(s.cfg.RepoPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load templates")
		return
	}
	writeJSON(w, http.StatusOK, templates.WithCounts(entries.TemplateCounts(cases)))
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}