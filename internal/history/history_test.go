package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	h, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(h.Samples) != 0 {
		t.Errorf("len(Samples) = %d, want 0", len(h.Samples))
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "history.json")
	h := &History{Samples: []Sample{
		{Time: time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC), Stars: 5},
		{Time: time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC), Stars: 8},
	}}

	if err := h.Save(path); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Samples) != 2 {
		t.Fatalf("len(Samples) = %d, want 2", len(loaded.Samples))
	}
	if !loaded.Samples[1].Time.Equal(h.Samples[1].Time) || loaded.Samples[1].Stars != 8 {
		t.Errorf("Samples[1] = %+v, want %+v", loaded.Samples[1], h.Samples[1])
	}
}

func TestLoadCorruptFileFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestAppendRejectsNonMonotonicTime(t *testing.T) {
	base := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	h := &History{}

	h.Append(Sample{Time: base, Stars: 1})
	h.Append(Sample{Time: base, Stars: 2})
	h.Append(Sample{Time: base.Add(-time.Hour), Stars: 3})
	h.Append(Sample{Time: base.Add(time.Hour), Stars: 4})

	if len(h.Samples) != 2 {
		t.Fatalf("len(Samples) = %d, want 2", len(h.Samples))
	}
	if h.Samples[1].Stars != 4 {
		t.Errorf("Samples[1].Stars = %d, want 4", h.Samples[1].Stars)
	}
}

func TestFromEventsCollapsesToDailySamples(t *testing.T) {
	day1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)
	events := []time.Time{
		day1.Add(9 * time.Hour),
		day1.Add(11 * time.Hour),
		day1.Add(23 * time.Hour),
		day2.Add(5 * time.Hour),
	}

	samples := FromEvents(events)

	if len(samples) != 2 {
		t.Fatalf("len(samples) = %d, want 2", len(samples))
	}
	if samples[0].Stars != 3 {
		t.Errorf("samples[0].Stars = %d, want 3", samples[0].Stars)
	}
	if !samples[0].Time.Equal(day1.Add(23 * time.Hour)) {
		t.Errorf("samples[0].Time = %v, want last event of day 1", samples[0].Time)
	}
	if samples[1].Stars != 4 {
		t.Errorf("samples[1].Stars = %d, want 4", samples[1].Stars)
	}
}

func TestFromEventsEmpty(t *testing.T) {
	if got := FromEvents(nil); len(got) != 0 {
		t.Errorf("FromEvents(nil) = %v, want empty", got)
	}
}
