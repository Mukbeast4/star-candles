package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"star-candles/internal/candles"
	"star-candles/internal/config"
	"star-candles/internal/github"
	"star-candles/internal/history"
	"star-candles/internal/svg"
)

func main() {
	cfg := config.Load()
	client := github.NewClient(cfg.GitHubAPI, cfg.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	hist, err := history.Load(cfg.HistoryPath)
	if err != nil {
		log.Fatalf("Failed to load history: %v", err)
	}

	if len(hist.Samples) == 0 {
		events, err := client.StarEvents(ctx, cfg.Repo)
		if err != nil {
			log.Printf("WARN: backfill unavailable, starting from current count: %v", err)
		} else {
			hist.Samples = history.FromEvents(events)
			log.Printf("Backfilled %d daily samples from %d star events", len(hist.Samples), len(events))
		}
	}

	count, err := client.StarCount(ctx, cfg.Repo)
	if err != nil {
		log.Fatalf("Failed to fetch star count: %v", err)
	}
	hist.Append(history.Sample{Time: time.Now().UTC(), Stars: count})

	if err := hist.Save(cfg.HistoryPath); err != nil {
		log.Fatalf("Failed to save history: %v", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output dir: %v", err)
	}
	themes := map[string]svg.Theme{
		"light": svg.Light(),
		"dark":  svg.Dark(),
	}
	for _, period := range []candles.Period{candles.Daily, candles.Monthly, candles.Yearly} {
		cs := candles.FromSamples(hist.Samples, period)
		if len(cs) > cfg.MaxCandles {
			cs = cs[len(cs)-cfg.MaxCandles:]
		}
		for name, theme := range themes {
			doc := svg.Render(cfg.Repo, cs, period, theme)
			path := filepath.Join(cfg.OutputDir, fmt.Sprintf("chart-%s-%s.svg", period, name))
			if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
				log.Fatalf("Failed to write %s: %v", path, err)
			}
		}
		log.Printf("Rendered %d %s candles for %s", len(cs), period, cfg.Repo)
	}

	log.Printf("Done for %s (stars: %d)", cfg.Repo, count)
}
