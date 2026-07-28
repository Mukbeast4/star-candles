package candles

import (
	"sort"
	"time"

	"star-candles/internal/history"
)

type Period int

const (
	Daily Period = iota
	Monthly
	Yearly
)

func (p Period) String() string {
	switch p {
	case Monthly:
		return "monthly"
	case Yearly:
		return "yearly"
	default:
		return "daily"
	}
}

func (p Period) bucket(t time.Time) time.Time {
	u := t.UTC()
	switch p {
	case Monthly:
		return time.Date(u.Year(), u.Month(), 1, 0, 0, 0, 0, time.UTC)
	case Yearly:
		return time.Date(u.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		return u.Truncate(24 * time.Hour)
	}
}

type Candle struct {
	Date  time.Time
	Open  int
	High  int
	Low   int
	Close int
}

func FromSamples(samples []history.Sample, period Period) []Candle {
	if len(samples) == 0 {
		return nil
	}

	sorted := make([]history.Sample, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	var out []Candle
	for _, s := range sorted {
		day := period.bucket(s.Time)
		if len(out) == 0 || !out[len(out)-1].Date.Equal(day) {
			open := s.Stars
			if len(out) > 0 {
				open = out[len(out)-1].Close
			}
			out = append(out, Candle{
				Date:  day,
				Open:  open,
				High:  max(open, s.Stars),
				Low:   min(open, s.Stars),
				Close: s.Stars,
			})
			continue
		}

		c := &out[len(out)-1]
		c.Close = s.Stars
		c.High = max(c.High, s.Stars)
		c.Low = min(c.Low, s.Stars)
	}
	return out
}
