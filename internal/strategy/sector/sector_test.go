package sector

import (
	"testing"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestRankAllowsOnlyTopThreeSectors(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ds := domain.DataSet{
		Securities: []domain.Security{
			{Symbol: "000300", Name: "CSI300", Kind: domain.AssetIndex, Sector: "Broad"},
			{Symbol: "A", Name: "A ETF", Kind: domain.AssetETF, Sector: "AI"},
			{Symbol: "B", Name: "B ETF", Kind: domain.AssetETF, Sector: "Chip"},
			{Symbol: "C", Name: "C ETF", Kind: domain.AssetETF, Sector: "Cloud"},
			{Symbol: "D", Name: "D ETF", Kind: domain.AssetETF, Sector: "Robot"},
		},
		Bars: map[string][]domain.Bar{
			"000300": testBars("000300", start, 130, 100, 0.0002),
			"A":      testBars("A", start, 130, 100, 0.0030),
			"B":      testBars("B", start, 130, 100, 0.0025),
			"C":      testBars("C", start, 130, 100, 0.0020),
			"D":      testBars("D", start, 130, 100, 0.0004),
		},
	}
	results := Evaluator{TopN: 3}.Rank(ds)
	if len(results) != 4 {
		t.Fatalf("expected 4 sector results, got %d", len(results))
	}
	for i := 0; i < 3; i++ {
		if !results[i].Allowed {
			t.Fatalf("rank %d should be allowed", i+1)
		}
	}
	if results[3].Allowed {
		t.Fatalf("rank 4 should not be allowed")
	}
}

func testBars(symbol string, start time.Time, n int, base, drift float64) []domain.Bar {
	bars := make([]domain.Bar, 0, n)
	price := base
	for i := 0; i < n; i++ {
		close := price * (1 + drift)
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     start.AddDate(0, 0, i),
			Open:     price,
			High:     close * 1.01,
			Low:      price * 0.99,
			Close:    close,
			Volume:   1000000 + float64(i)*1000,
			Amount:   close * (1000000 + float64(i)*1000),
			Turnover: 0.02,
		})
		price = close
	}
	return bars
}
