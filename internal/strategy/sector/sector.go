package sector

import (
	"fmt"
	"sort"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/metrics"
)

type Evaluator struct {
	BenchmarkSymbol string
	TopN            int
}

func (e Evaluator) Rank(ds domain.DataSet) []domain.SectorResult {
	benchmarkSymbol := e.BenchmarkSymbol
	if benchmarkSymbol == "" {
		benchmarkSymbol = "000300"
	}
	topN := e.TopN
	if topN <= 0 {
		topN = 3
	}

	benchmark := ds.Bars[benchmarkSymbol]
	results := make([]domain.SectorResult, 0)
	for _, security := range ds.Securities {
		if !security.IsETF() {
			continue
		}
		bars := ds.Bars[security.Symbol]
		result, ok := EvaluateSecurity(security, bars, benchmark)
		if ok {
			results = append(results, result)
		}
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Sector < results[j].Sector
		}
		return results[i].Score > results[j].Score
	})
	for i := range results {
		results[i].Rank = i + 1
		results[i].Allowed = i < topN
		if results[i].Allowed {
			results[i].Reasons = append(results[i].Reasons, fmt.Sprintf("sector is top %d by relative strength", topN))
		}
	}
	return results
}

func EvaluateSecurity(security domain.Security, bars, benchmark []domain.Bar) (domain.SectorResult, bool) {
	if len(bars) < 120 || len(benchmark) < 120 {
		return domain.SectorResult{}, false
	}
	r20, _ := metrics.Return(bars, 20)
	r60, _ := metrics.Return(bars, 60)
	b20, _ := metrics.Return(benchmark, 20)
	b60, _ := metrics.Return(benchmark, 60)
	rel20 := r20 - b20
	rel60 := r60 - b60

	last, _ := domain.LastBar(bars)
	ma20, _ := metrics.SMA(bars, 20)
	ma60, _ := metrics.SMA(bars, 60)
	ma20Prev, _ := metrics.SMABefore(bars, 20, 5)
	trendScore := 0.0
	if last.Close > ma60 {
		trendScore += 45
	}
	if ma20 > ma60 {
		trendScore += 30
	}
	if ma20 > ma20Prev {
		trendScore += 25
	}

	amountRatio, ok := metrics.LatestAmountRatio(bars, 20)
	if !ok {
		amountRatio = 1
	}
	amountScore := domain.Normalize(amountRatio, 0.7, 1.8)
	maxDD, ok := metrics.MaxDrawdown(bars, 60)
	if !ok {
		maxDD = -0.2
	}
	resilienceScore := 100 - domain.Normalize(-maxDD, 0.06, 0.24)
	score := domain.Normalize(rel60, -0.12, 0.25)*0.35 +
		domain.Normalize(rel20, -0.08, 0.16)*0.25 +
		trendScore*0.20 +
		amountScore*0.10 +
		resilienceScore*0.10

	reasons := []string{
		fmt.Sprintf("60d relative return %.1f%%", rel60*100),
		fmt.Sprintf("20d relative return %.1f%%", rel20*100),
		fmt.Sprintf("amount ratio %.2fx", amountRatio),
		fmt.Sprintf("60d max drawdown %.1f%%", maxDD*100),
	}

	return domain.SectorResult{
		Sector:           security.Sector,
		Symbol:           security.Symbol,
		Name:             security.Name,
		Score:            domain.Clamp(score, 0, 100),
		RelativeReturn20: rel20,
		RelativeReturn60: rel60,
		TrendScore:       trendScore,
		AmountScore:      amountScore,
		ResilienceScore:  resilienceScore,
		Reasons:          reasons,
	}, true
}
