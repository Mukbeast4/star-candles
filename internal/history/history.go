package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type Sample struct {
	Time  time.Time `json:"time"`
	Stars int       `json:"stars"`
}

type History struct {
	Samples []Sample `json:"samples"`
}

func Load(path string) (*History, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &History{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read history: %w", err)
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return nil, fmt.Errorf("decode history: %w", err)
	}
	return &h, nil
}

func (h *History) Append(s Sample) {
	if n := len(h.Samples); n > 0 && !s.Time.After(h.Samples[n-1].Time) {
		return
	}
	h.Samples = append(h.Samples, s)
}

func (h *History) Save(path string) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("encode history: %w", err)
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create history dir: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}

func FromEvents(events []time.Time) []Sample {
	var samples []Sample
	for i, e := range events {
		day := e.UTC().Truncate(24 * time.Hour)
		if len(samples) > 0 && samples[len(samples)-1].Time.Truncate(24*time.Hour).Equal(day) {
			samples[len(samples)-1] = Sample{Time: e.UTC(), Stars: i + 1}
			continue
		}
		samples = append(samples, Sample{Time: e.UTC(), Stars: i + 1})
	}
	return samples
}
