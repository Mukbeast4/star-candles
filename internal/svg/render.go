package svg

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"star-candles/internal/candles"
)

const (
	width        = 880
	height       = 440
	padding      = 24
	headerHeight = 64
	axisBand     = 26
	labelGutter  = 64
	fontStack    = "system-ui, -apple-system, 'Segoe UI', sans-serif"
)

type Theme struct {
	Surface   string
	Primary   string
	Secondary string
	Muted     string
	Grid      string
	Border    string
	Up        string
	Down      string
	DeltaUp   string
	DeltaDown string
}

func Light() Theme {
	return Theme{
		Surface:   "#fcfcfb",
		Primary:   "#0b0b0b",
		Secondary: "#52514e",
		Muted:     "#898781",
		Grid:      "#e1e0d9",
		Border:    "rgba(11,11,11,0.10)",
		Up:        "#199e70",
		Down:      "#d03b3b",
		DeltaUp:   "#006300",
		DeltaDown: "#d03b3b",
	}
}

func Dark() Theme {
	return Theme{
		Surface:   "#1a1a19",
		Primary:   "#ffffff",
		Secondary: "#c3c2b7",
		Muted:     "#898781",
		Grid:      "#2c2c2a",
		Border:    "rgba(255,255,255,0.10)",
		Up:        "#199e70",
		Down:      "#d03b3b",
		DeltaUp:   "#0ca30c",
		DeltaDown: "#d03b3b",
	}
}

func Render(repo string, cs []candles.Candle, period candles.Period, theme Theme) string {
	var b strings.Builder

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s candlestick chart of GitHub stars for %s">`,
		width, height, width, height, periodTitle(period), escape(repo))
	b.WriteString("\n")
	fmt.Fprintf(&b, `<rect width="%d" height="%d" rx="8" fill="%s"/>`, width, height, theme.Surface)
	b.WriteString("\n")

	if len(cs) == 0 {
		renderEmpty(&b, repo, period, theme)
	} else {
		renderChart(&b, repo, cs, period, theme)
	}

	fmt.Fprintf(&b, `<rect x="0.5" y="0.5" width="%d" height="%d" rx="8" fill="none" stroke="%s"/>`, width-1, height-1, theme.Border)
	b.WriteString("\n</svg>\n")
	return b.String()
}

func renderEmpty(b *strings.Builder, repo string, period candles.Period, theme Theme) {
	writeHeader(b, repo, period, 0, 0, 0, theme)
	fmt.Fprintf(b, `<text x="%d" y="%d" text-anchor="middle" font-family="%s" font-size="13" fill="%s">No star history yet</text>`,
		width/2, height/2+20, fontStack, theme.Muted)
	b.WriteString("\n")
}

func renderChart(b *strings.Builder, repo string, cs []candles.Candle, period candles.Period, theme Theme) {
	plotX0 := float64(padding)
	plotX1 := float64(width - padding - labelGutter)
	plotY0 := float64(padding + headerHeight)
	plotY1 := float64(height - padding - axisBand)
	plotW := plotX1 - plotX0
	plotH := plotY1 - plotY0

	lo, hi := domain(cs)
	scaleY := func(v float64) float64 {
		return plotY1 - (v-lo)/(hi-lo)*plotH
	}

	last := cs[len(cs)-1]
	writeHeader(b, repo, period, last.Close, last.Close-last.Open, last.Open, theme)

	step := niceStep((hi - lo) / 5)
	for tick := int(math.Ceil(lo/float64(step))) * step; float64(tick) <= hi; tick += step {
		y := scaleY(float64(tick))
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="1"/>`,
			plotX0, y, plotX1, y, theme.Grid)
		b.WriteString("\n")
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" dy="0.35em" font-family="%s" font-size="11" fill="%s" style="font-variant-numeric: tabular-nums">%s</text>`,
			plotX1+10, y, fontStack, theme.Muted, comma(tick))
		b.WriteString("\n")
	}

	slot := plotW / float64(len(cs))
	bodyW := math.Min(slot-2, 24)
	if bodyW < 1 {
		bodyW = 1
	}

	labelStep := int(math.Ceil(float64(len(cs)) / 6))
	dateFormat := dateLayout(period, last.Date.Sub(cs[0].Date))

	for i, c := range cs {
		cx := plotX0 + slot*(float64(i)+0.5)
		color := theme.Up
		if c.Close < c.Open {
			color = theme.Down
		}

		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="1"/>`,
			cx, scaleY(float64(c.High)), cx, scaleY(float64(c.Low)), color)
		b.WriteString("\n")

		top := scaleY(float64(max(c.Open, c.Close)))
		bodyH := scaleY(float64(min(c.Open, c.Close))) - top
		if bodyH < 1.5 {
			top -= (1.5 - bodyH) / 2
			bodyH = 1.5
		}

		x := cx - bodyW/2
		if c.Close >= c.Open && bodyW >= 2.5 {
			fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="1" fill="%s" stroke="%s" stroke-width="1.5"/>`,
				x+0.75, top+0.75, bodyW-1.5, math.Max(bodyH-1.5, 0.5), theme.Surface, color)
		} else {
			fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="1" fill="%s"/>`,
				x, top, bodyW, bodyH, color)
		}
		b.WriteString("\n")

		if i%labelStep == 0 {
			fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-family="%s" font-size="11" fill="%s">%s</text>`,
				cx, plotY1+18, fontStack, theme.Muted, c.Date.Format(dateFormat))
			b.WriteString("\n")
		}
	}
}

func writeHeader(b *strings.Builder, repo string, period candles.Period, value, delta, base int, theme Theme) {
	fmt.Fprintf(b, `<text x="%d" y="%d" font-family="%s" font-size="13" fill="%s">%s · %s</text>`,
		padding, padding+18, fontStack, theme.Secondary, escape(repo), periodTitle(period))
	b.WriteString("\n")

	fmt.Fprintf(b, `<text x="%d" y="%d" font-family="%s">`, padding, padding+50, fontStack)
	fmt.Fprintf(b, `<tspan font-size="28" font-weight="600" fill="%s">%s</tspan>`, theme.Primary, comma(value))
	fmt.Fprintf(b, `<tspan dx="8" font-size="13" fill="%s">stars</tspan>`, theme.Secondary)
	if delta != 0 {
		color := theme.DeltaUp
		sign := "+"
		if delta < 0 {
			color = theme.DeltaDown
			sign = "-"
		}
		text := fmt.Sprintf("%s%s %s", sign, comma(abs(delta)), deltaSuffix(period))
		if base > 0 {
			text = fmt.Sprintf("%s%s (%s%.1f%%) %s", sign, comma(abs(delta)), sign, math.Abs(float64(delta))*100/float64(base), deltaSuffix(period))
		}
		fmt.Fprintf(b, `<tspan dx="12" font-size="13" fill="%s">%s</tspan>`, color, text)
	}
	b.WriteString(`</text>`)
	b.WriteString("\n")
}

func periodTitle(p candles.Period) string {
	switch p {
	case candles.Monthly:
		return "Monthly"
	case candles.Yearly:
		return "Yearly"
	default:
		return "Daily"
	}
}

func deltaSuffix(p candles.Period) string {
	switch p {
	case candles.Monthly:
		return "this month"
	case candles.Yearly:
		return "this year"
	default:
		return "today"
	}
}

func dateLayout(p candles.Period, span time.Duration) string {
	switch p {
	case candles.Monthly:
		return "Jan 2006"
	case candles.Yearly:
		return "2006"
	default:
		if span > 300*24*time.Hour {
			return "Jan 2006"
		}
		return "Jan 2"
	}
}

func domain(cs []candles.Candle) (float64, float64) {
	lo := cs[0].Low
	hi := cs[0].High
	for _, c := range cs {
		lo = min(lo, c.Low)
		hi = max(hi, c.High)
	}

	pad := math.Max(1, float64(hi-lo)*0.08)
	dLo := math.Max(0, float64(lo)-pad)
	dHi := float64(hi) + pad
	if dHi <= dLo {
		dHi = dLo + 1
	}
	return dLo, dHi
}

func niceStep(raw float64) int {
	if raw <= 1 {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, m := range []float64{1, 2, 5, 10} {
		if m*mag >= raw {
			return int(m * mag)
		}
	}
	return int(10 * mag)
}

func comma(n int) string {
	s := strconv.Itoa(abs(n))
	if len(s) > 3 {
		var parts []string
		for len(s) > 3 {
			parts = append([]string{s[len(s)-3:]}, parts...)
			s = s[:len(s)-3]
		}
		s = s + "," + strings.Join(parts, ",")
	}
	if n < 0 {
		return "-" + s
	}
	return s
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

var xmlReplacer = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&apos;",
)

func escape(s string) string {
	return xmlReplacer.Replace(s)
}
