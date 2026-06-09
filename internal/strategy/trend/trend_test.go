package trend

import (
	"testing"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestEvaluateVolumeBreakout(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := makeTrendBars("T", start)
	benchmark := makeFlatBars("B", start, len(bars))

	result := Evaluator{}.Evaluate(bars, benchmark)
	if !result.Established {
		t.Fatalf("expected established trend, got %#v", result)
	}
	if result.Trigger != domain.TriggerVolumeBreakout || !result.TriggerConfirmed {
		t.Fatalf("expected volume breakout, got %s confirmed=%t", result.Trigger, result.TriggerConfirmed)
	}
}

func TestEvaluatePullbackConfirm(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := makePullbackBars("T", start)
	benchmark := makeFlatBars("B", start, len(bars))

	result := Evaluator{}.Evaluate(bars, benchmark)
	if !result.Established {
		t.Fatalf("expected established trend, got %#v", result)
	}
	if result.Trigger != domain.TriggerPullbackConfirm || !result.TriggerConfirmed {
		t.Fatalf("expected pullback confirmation, got %s confirmed=%t", result.Trigger, result.TriggerConfirmed)
	}
}

func TestEvaluateOverheatedWatch(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := makeTrendBars("T", start)
	last := len(bars) - 1
	bars[last].Close = bars[last].Close * 1.25
	benchmark := makeFlatBars("B", start, len(bars))

	result := Evaluator{}.Evaluate(bars, benchmark)
	if !result.Overheated {
		t.Fatalf("expected overheated signal")
	}
}

func makeTrendBars(symbol string, start time.Time) []domain.Bar {
	bars := make([]domain.Bar, 0, 130)
	price := 100.0
	for i := 0; i < 130; i++ {
		close := price * 1.002
		amount := 1000000.0
		if i == 129 {
			close = price * 1.08
			amount = 2300000
		}
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     start.AddDate(0, 0, i),
			Open:     price,
			High:     close * 1.02,
			Low:      price * 0.99,
			Close:    close,
			Volume:   amount / close,
			Amount:   amount,
			Turnover: 0.03,
		})
		price = close
	}
	return bars
}

func makeFlatBars(symbol string, start time.Time, n int) []domain.Bar {
	bars := make([]domain.Bar, 0, n)
	for i := 0; i < n; i++ {
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     start.AddDate(0, 0, i),
			Open:     100,
			High:     101,
			Low:      99,
			Close:    100,
			Volume:   1000000,
			Amount:   100000000,
			Turnover: 0.02,
		})
	}
	return bars
}

func makePullbackBars(symbol string, start time.Time) []domain.Bar {
	bars := make([]domain.Bar, 0, 130)
	price := 100.0
	for i := 0; i < 130; i++ {
		close := price * 1.002
		amount := 1000000.0
		if i == 116 {
			close = price * 1.09
			amount = 2400000
		}
		if i > 116 {
			close = price * 0.999
			amount = 780000
		}
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     start.AddDate(0, 0, i),
			Open:     price,
			High:     close * 1.02,
			Low:      price * 0.99,
			Close:    close,
			Volume:   amount / close,
			Amount:   amount,
			Turnover: 0.025,
		})
		price = close
	}
	return bars
}
