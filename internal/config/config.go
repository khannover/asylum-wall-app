package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               int
	RepoPath           string
	RepoURL            string
	GitHubToken        string
	GitBranch          string
	GitUserName        string
	GitUserEmail       string
	MaxProofSize       int64
	SignalMaxPerHour   int
	SignalMaxPerDay    int
	SignalDuplicateTTL time.Duration
	EditToken          string
}

func Load() Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	maxSize, _ := strconv.ParseInt(getEnv("MAX_PROOF_SIZE_MB", "5"), 10, 64)

	signalHour, _ := strconv.Atoi(getEnv("SIGNAL_MAX_PER_HOUR", "5"))
	signalDay, _ := strconv.Atoi(getEnv("SIGNAL_MAX_PER_DAY", "20"))
	dupHours, _ := strconv.Atoi(getEnv("SIGNAL_DUPLICATE_HOURS", "24"))

	return Config{
		Port:               port,
		RepoPath:           getEnv("REPO_PATH", "/repo"),
		RepoURL:            os.Getenv("REPO_URL"),
		GitHubToken:        os.Getenv("GITHUB_TOKEN"),
		GitBranch:          getEnv("GIT_BRANCH", "main"),
		GitUserName:        getEnv("GIT_USER_NAME", "Asylum Wall Bot"),
		GitUserEmail:       getEnv("GIT_USER_EMAIL", "asylum-wall@bancamp.de"),
		MaxProofSize:       maxSize * 1024 * 1024,
		SignalMaxPerHour:   signalHour,
		SignalMaxPerDay:    signalDay,
		SignalDuplicateTTL: time.Duration(dupHours) * time.Hour,
		EditToken:          os.Getenv("EDIT_TOKEN"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}