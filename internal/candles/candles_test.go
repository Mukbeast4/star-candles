package candles

import (
	"testing"
	"time"

	"star-candles/internal/history"
)

func day(d int, hour int) time.Time {
	return time.Date(2026, 7, d, hour, 0, 0, 0, time.UTC)
}

func TestFromSamplesEmpty(t *testing.T) {
	if got := FromSamples(nil, Daily); got != nil {
		t.Errorf("FromSamples(nil, Daily) = %v, want nil", got)
	}
}

func TestFromSamplesSingleDay(t *testing.T) {
	samples := []history.Sample{
		{Time: day(1, 9), Stars: 10},
		{Time: day(1, 12), Stars: 14},
		{Time: day(1, 15), Stars: 12},
		{Time: day(1, 20), Stars: 13},
	}

	cs := FromSamples(samples, Daily)

	if len(cs) != 1 {
		t.Fatalf("len(candles) = %d, want 1", len(cs))
	}
	c := cs[0]
	if c.Open != 10 || c.High != 14 || c.Low != 10 || c.Close != 13 {
		t.Errorf("candle = %+v, want O10 H14 L10 C13", c)
	}
	if !c.Date.Equal(day(1, 0)) {
		t.Errorf("Date = %v, want %v", c.Date, day(1, 0))
	}
}

func TestFromSamplesOpenIsPreviousClose(t *testing.T) {
	samples := []history.Sample{
		{Time: day(1, 10), Stars: 10},
		{Time: day(3, 10), Stars: 25},
		{Time: day(3, 18), Stars: 22},
	}

	cs := FromSamples(samples, Daily)

	if len(cs) != 2 {
		t.Fatalf("len(candles) = %d, want 2", len(cs))
	}
	if cs[1].Open != 10 {
		t.Errorf("Open = %d, want previous close 10", cs[1].Open)
	}
	if cs[1].High != 25 || cs[1].Low != 10 || cs[1].Close != 22 {
		t.Errorf("candle = %+v, want H25 L10 C22", cs[1])
	}
}

func TestFromSamplesBearishCandle(t *testing.T) {
	samples := []history.Sample{
		{Time: day(1, 10), Stars: 100},
		{Time: day(2, 8), Stars: 97},
		{Time: day(2, 20), Stars: 95},
	}

	cs := FromSamples(samples, Daily)

	if len(cs) != 2 {
		t.Fatalf("len(candles) = %d, want 2", len(cs))
	}
	c := cs[1]
	if c.Open != 100 || c.High != 100 || c.Low != 95 || c.Close != 95 {
		t.Errorf("candle = %+v, want O100 H100 L95 C95", c)
	}
	if c.Close >= c.Open {
		t.Error("candle should be bearish")
	}
}

func TestFromSamplesUnsortedInput(t *testing.T) {
	samples := []history.Sample{
		{Time: day(2, 10), Stars: 20},
		{Time: day(1, 10), Stars: 10},
	}

	cs := FromSamples(samples, Daily)

	if len(cs) != 2 {
		t.Fatalf("len(candles) = %d, want 2", len(cs))
	}
	if cs[0].Close != 10 || cs[1].Close != 20 {
		t.Errorf("candles out of order: %+v", cs)
	}
}

func TestFromSamplesMonthly(t *testing.T) {
	samples := []history.Sample{
		{Time: time.Date(2026, 5, 3, 10, 0, 0, 0, time.UTC), Stars: 10},
		{Time: time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC), Stars: 30},
		{Time: time.Date(2026, 6, 2, 10, 0, 0, 0, time.UTC), Stars: 28},
		{Time: time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC), Stars: 45},
	}

	cs := FromSamples(samples, Monthly)

	if len(cs) != 2 {
		t.Fatalf("len(candles) = %d, want 2", len(cs))
	}
	if !cs[0].Date.Equal(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Date = %v, want first of May", cs[0].Date)
	}
	if cs[0].Open != 10 || cs[0].Close != 30 {
		t.Errorf("May candle = %+v, want O10 C30", cs[0])
	}
	if cs[1].Open != 30 || cs[1].Low != 28 || cs[1].High != 45 || cs[1].Close != 45 {
		t.Errorf("June candle = %+v, want O30 H45 L28 C45", cs[1])
	}
}

func TestFromSamplesYearly(t *testing.T) {
	samples := []history.Sample{
		{Time: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), Stars: 100},
		{Time: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC), Stars: 300},
		{Time: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Stars: 280},
	}

	cs := FromSamples(samples, Yearly)

	if len(cs) != 2 {
		t.Fatalf("len(candles) = %d, want 2", len(cs))
	}
	if !cs[0].Date.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("Date = %v, want first of 2025", cs[0].Date)
	}
	if cs[1].Open != 300 || cs[1].Close != 280 {
		t.Errorf("2026 candle = %+v, want O300 C280", cs[1])
	}
}

func TestPeriodString(t *testing.T) {
	tests := []struct {
		period Period
		want   string
	}{
		{Daily, "daily"},
		{Monthly, "monthly"},
		{Yearly, "yearly"},
	}
	for _, tt := range tests {
		if got := tt.period.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
