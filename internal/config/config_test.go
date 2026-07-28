package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("REPO", "owner/repo")

	cfg := Load()

	if cfg.Repo != "owner/repo" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "owner/repo")
	}
	if cfg.GitHubAPI != "https://api.github.com" {
		t.Errorf("GitHubAPI = %q, want default", cfg.GitHubAPI)
	}
	if cfg.HistoryPath != "out/history.json" {
		t.Errorf("HistoryPath = %q, want default", cfg.HistoryPath)
	}
	if cfg.OutputDir != "out" {
		t.Errorf("OutputDir = %q, want default", cfg.OutputDir)
	}
	if cfg.MaxCandles != 90 {
		t.Errorf("MaxCandles = %d, want 90", cfg.MaxCandles)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("REPO", "owner/repo")
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GITHUB_API_URL", "http://localhost:9999")
	t.Setenv("HISTORY_FILE", "data/h.json")
	t.Setenv("OUTPUT_DIR", "data")
	t.Setenv("MAX_CANDLES", "30")

	cfg := Load()

	if cfg.Token != "tok" {
		t.Errorf("Token = %q, want %q", cfg.Token, "tok")
	}
	if cfg.GitHubAPI != "http://localhost:9999" {
		t.Errorf("GitHubAPI = %q, want override", cfg.GitHubAPI)
	}
	if cfg.HistoryPath != "data/h.json" {
		t.Errorf("HistoryPath = %q, want override", cfg.HistoryPath)
	}
	if cfg.OutputDir != "data" {
		t.Errorf("OutputDir = %q, want override", cfg.OutputDir)
	}
	if cfg.MaxCandles != 30 {
		t.Errorf("MaxCandles = %d, want 30", cfg.MaxCandles)
	}
}

func TestRepoFallsBackToGitHubRepository(t *testing.T) {
	t.Setenv("REPO", "")
	t.Setenv("GITHUB_REPOSITORY", "actions/owner-repo")

	cfg := Load()

	if cfg.Repo != "actions/owner-repo" {
		t.Errorf("Repo = %q, want %q", cfg.Repo, "actions/owner-repo")
	}
}

func TestGetEnvAsIntRejectsInvalid(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{"non numeric", "abc", 90},
		{"zero", "0", 90},
		{"negative", "-5", 90},
		{"valid", "120", 120},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("MAX_CANDLES", tt.value)
			if got := getEnvAsInt("MAX_CANDLES", 90); got != tt.want {
				t.Errorf("getEnvAsInt(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}
