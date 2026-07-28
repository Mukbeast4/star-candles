package svg

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"star-candles/internal/candles"
)

func fixture() []candles.Candle {
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	return []candles.Candle{
		{Date: base, Open: 100, High: 120, Low: 98, Close: 115},
		{Date: base.AddDate(0, 0, 1), Open: 115, High: 118, Low: 105, Close: 108},
		{Date: base.AddDate(0, 0, 2), Open: 108, High: 140, Low: 108, Close: 138},
	}
}

func assertWellFormed(t *testing.T, doc string) {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(doc))
	for {
		_, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			t.Fatalf("invalid XML: %v", err)
		}
	}
}

func TestRenderWellFormed(t *testing.T) {
	for _, tt := range []struct {
		name  string
		theme Theme
	}{
		{"light", Light()},
		{"dark", Dark()},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := Render("owner/repo", fixture(), candles.Daily, tt.theme)
			assertWellFormed(t, doc)
			if !strings.Contains(doc, tt.theme.Surface) {
				t.Error("missing surface color")
			}
			if !strings.Contains(doc, tt.theme.Up) || !strings.Contains(doc, tt.theme.Down) {
				t.Error("missing up or down candle color")
			}
		})
	}
}

func TestRenderEscapesRepoName(t *testing.T) {
	doc := Render(`a<b>&"c'`, fixture(), candles.Daily, Light())
	assertWellFormed(t, doc)
	if strings.Contains(doc, "a<b>") {
		t.Error("repo name not escaped")
	}
	if !strings.Contains(doc, "a&lt;b&gt;&amp;&quot;c&apos;") {
		t.Error("expected escaped repo name")
	}
}

func TestRenderEmptyHistory(t *testing.T) {
	doc := Render("owner/repo", nil, candles.Daily, Light())
	assertWellFormed(t, doc)
	if !strings.Contains(doc, "No star history yet") {
		t.Error("missing empty state message")
	}
}

func TestRenderHeaderValueAndDelta(t *testing.T) {
	doc := Render("owner/repo", fixture(), candles.Daily, Light())
	if !strings.Contains(doc, ">138</tspan>") {
		t.Error("missing current star count in header")
	}
	if !strings.Contains(doc, "+30") {
		t.Error("missing positive delta in header")
	}
	if !strings.Contains(doc, Light().DeltaUp) {
		t.Error("delta should use the up delta color")
	}
}

func TestRenderNegativeDelta(t *testing.T) {
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	cs := []candles.Candle{
		{Date: base, Open: 100, High: 100, Low: 95, Close: 96},
	}
	doc := Render("owner/repo", cs, candles.Daily, Light())
	if !strings.Contains(doc, "-4") {
		t.Error("missing negative delta")
	}
}

func TestRenderFlatHistory(t *testing.T) {
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	cs := []candles.Candle{
		{Date: base, Open: 0, High: 0, Low: 0, Close: 0},
	}
	doc := Render("owner/repo", cs, candles.Daily, Light())
	assertWellFormed(t, doc)
}

func TestCommaFormatting(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1284, "1,284"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		if got := comma(tt.in); got != tt.want {
			t.Errorf("comma(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNiceStep(t *testing.T) {
	tests := []struct {
		in   float64
		want int
	}{
		{0.4, 1},
		{1.2, 2},
		{4.2, 5},
		{42, 50},
		{160, 200},
		{700, 1000},
	}
	for _, tt := range tests {
		if got := niceStep(tt.in); got != tt.want {
			t.Errorf("niceStep(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestRenderPeriodLabels(t *testing.T) {
	tests := []struct {
		period     candles.Period
		wantTitle  string
		wantSuffix string
	}{
		{candles.Daily, "Daily", "today"},
		{candles.Monthly, "Monthly", "this month"},
		{candles.Yearly, "Yearly", "this year"},
	}
	for _, tt := range tests {
		t.Run(tt.wantTitle, func(t *testing.T) {
			doc := Render("owner/repo", fixture(), tt.period, Light())
			assertWellFormed(t, doc)
			if !strings.Contains(doc, "owner/repo · "+tt.wantTitle) {
				t.Errorf("missing period title %q in header", tt.wantTitle)
			}
			if !strings.Contains(doc, tt.wantSuffix) {
				t.Errorf("missing delta suffix %q", tt.wantSuffix)
			}
		})
	}
}
