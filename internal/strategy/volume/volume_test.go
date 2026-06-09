package volume

import (
	"testing"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestEvaluateDistributionRisk(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := makeVolumeBars("V", start, 60)
	last := len(bars) - 1
	bars[last].Close = bars[last-1].Close * 0.96
	bars[last].Amount = bars[last].Amount * 3
	bars[last].Volume = bars[last].Volume * 3

	result := Evaluator{}.Evaluate(bars)
	if !result.Distribution {
		t.Fatalf("expected distribution risk, got %#v", result)
	}
}

func TestEvaluateAccumulation(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := makeVolumeBars("V", start, 60)
	last := len(bars) - 1
	bars[last].Close = bars[last-1].Close * 1.04
	bars[last].Amount = bars[last].Amount * 2
	bars[last].Volume = bars[last].Volume * 2

	result := Evaluator{}.Evaluate(bars)
	if !result.Accumulation {
		t.Fatalf("expected accumulation, got %#v", result)
	}
	if result.AmountRatio < 1.5 {
		t.Fatalf("expected amount ratio expansion, got %.2f", result.AmountRatio)
	}
}

func makeVolumeBars(symbol string, start time.Time, n int) []domain.Bar {
	bars := make([]domain.Bar, 0, n)
	price := 100.0
	for i := 0; i < n; i++ {
		close := price * (1 + 0.001)
		amount := 1000000.0 + float64(i)*1000
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     start.AddDate(0, 0, i),
			Open:     price,
			High:     close * 1.01,
			Low:      price * 0.99,
			Close:    close,
			Volume:   amount / close,
			Amount:   amount,
			Turnover: 0.02 + float64(i%10)*0.001,
		})
		price = close
	}
	return bars
}
