package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Repo        string
	Token       string
	GitHubAPI   string
	HistoryPath string
	OutputDir   string
	MaxCandles  int
}

func Load() *Config {
	return &Config{
		Repo:        requireRepo(),
		Token:       os.Getenv("GITHUB_TOKEN"),
		GitHubAPI:   getEnv("GITHUB_API_URL", "https://api.github.com"),
		HistoryPath: getEnv("HISTORY_FILE", "out/history.json"),
		OutputDir:   getEnv("OUTPUT_DIR", "out"),
		MaxCandles:  getEnvAsInt("MAX_CANDLES", 90),
	}
}

func requireRepo() string {
	if v, ok := os.LookupEnv("REPO"); ok && v != "" {
		return v
	}
	if v, ok := os.LookupEnv("GITHUB_REPOSITORY"); ok && v != "" {
		return v
	}
	log.Fatalf("Required environment variable REPO is not set")
	return ""
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultValue
	}
	return n
}
