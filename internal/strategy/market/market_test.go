package market

import (
	"testing"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestEvaluateRiskStates(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := map[string][]domain.Bar{
		"000300": marketBars("000300", start, 130, 100, 0.002),
		"000905": marketBars("000905", start, 130, 100, 0.002),
		"399006": marketBars("399006", start, 130, 100, 0.002),
		"000688": marketBars("000688", start, 130, 100, 0.002),
	}
	riskOn := Evaluator{}.Evaluate(domain.DataSet{
		Bars: bars,
		Breadth: domain.MarketBreadth{
			Advancers:      3200,
			Decliners:      1800,
			LimitDownCount: 10,
			TotalAmount:    11000,
			AvgAmount20:    10000,
			TechReturn60:   0.12,
			BroadReturn60:  0.05,
		},
	})
	if riskOn.State != domain.RiskOn {
		t.Fatalf("expected Risk-On, got %s", riskOn.State)
	}

	riskOff := Evaluator{}.Evaluate(domain.DataSet{
		Bars: map[string][]domain.Bar{
			"000300": marketBars("000300", start, 130, 100, -0.002),
			"000905": marketBars("000905", start, 130, 100, -0.002),
			"399006": marketBars("399006", start, 130, 100, -0.002),
			"000688": marketBars("000688", start, 130, 100, -0.002),
		},
		Breadth: domain.MarketBreadth{
			Advancers:      900,
			Decliners:      4200,
			LimitDownCount: 100,
			TotalAmount:    8500,
			AvgAmount20:    10000,
			TechReturn60:   -0.10,
			BroadReturn60:  0.02,
		},
	})
	if riskOff.State != domain.RiskOff {
		t.Fatalf("expected Risk-Off, got %s", riskOff.State)
	}
}

func marketBars(symbol string, start time.Time, n int, base, drift float64) []domain.Bar {
	bars := make([]domain.Bar, 0, n)
	price := base
	for i := 0; i < n; i++ {
		close := price * (1 + drift)
		bars = append(bars, domain.Bar{
			Symbol: symbol,
			Date:   start.AddDate(0, 0, i),
			Open:   price,
			High:   close * 1.01,
			Low:    close * 0.99,
			Close:  close,
			Amount: 1000000,
			Volume: 10000,
		})
		price = close
	}
	return bars
}
