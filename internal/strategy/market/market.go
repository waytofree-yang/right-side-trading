package market

import (
	"fmt"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/metrics"
)

var DefaultIndexSymbols = []string{"000300", "000905", "399006", "000688"}

type Evaluator struct {
	IndexSymbols []string
}

func (e Evaluator) Evaluate(ds domain.DataSet) domain.MarketResult {
	indexSymbols := e.IndexSymbols
	if len(indexSymbols) == 0 {
		indexSymbols = DefaultIndexSymbols
	}

	var above int
	var checked int
	reasons := make([]string, 0, 6)
	for _, symbol := range indexSymbols {
		bars := ds.Bars[symbol]
		last, ok := domain.LastBar(bars)
		ma120, maOK := metrics.SMA(bars, 120)
		if !ok || !maOK {
			reasons = append(reasons, fmt.Sprintf("%s lacks 120-day data", symbol))
			continue
		}
		checked++
		if last.Close > ma120 {
			above++
		}
	}

	breadth := 0.5
	if total := ds.Breadth.Advancers + ds.Breadth.Decliners; total > 0 {
		breadth = float64(ds.Breadth.Advancers) / float64(total)
	}
	volumeRatio := 1.0
	if ds.Breadth.AvgAmount20 > 0 {
		volumeRatio = ds.Breadth.TotalAmount / ds.Breadth.AvgAmount20
	}
	techRelative := ds.Breadth.TechReturn60 - ds.Breadth.BroadReturn60

	aboveScore := 50.0
	if checked > 0 {
		aboveScore = float64(above) / float64(checked) * 100
	}
	breadthScore := domain.Normalize(breadth, 0.35, 0.65)
	volumeScore := domain.Normalize(volumeRatio, 0.75, 1.25)
	limitDownScore := 100 - domain.Normalize(float64(ds.Breadth.LimitDownCount), 20, 120)
	techScore := domain.Normalize(techRelative, -0.08, 0.12)
	score := aboveScore*0.35 + breadthScore*0.2 + volumeScore*0.15 + limitDownScore*0.15 + techScore*0.15

	state := domain.Neutral
	if checked > 0 && (above >= 3 && breadth >= 0.52 && volumeRatio >= 0.95 && ds.Breadth.LimitDownCount <= 30 && techRelative >= 0) {
		state = domain.RiskOn
	}
	if checked > 0 && (above <= 1 || (breadth < 0.40 && ds.Breadth.LimitDownCount > 80) || (techRelative < 0 && above <= 2)) {
		state = domain.RiskOff
	}

	reasons = append(reasons,
		fmt.Sprintf("%d/%d major indices are above MA120", above, checked),
		fmt.Sprintf("breadth %.0f%% advancers", breadth*100),
		fmt.Sprintf("market amount ratio %.2fx", volumeRatio),
		fmt.Sprintf("tech relative return %.1f%% over 60 days", techRelative*100),
	)
	if state == domain.RiskOff {
		reasons = append(reasons, "risk gate blocks active recommendations")
	}

	return domain.MarketResult{
		State:        state,
		Score:        domain.Clamp(score, 0, 100),
		AboveMA120:   above,
		Breadth:      breadth,
		VolumeRatio:  volumeRatio,
		TechRelative: techRelative,
		Reasons:      reasons,
	}
}
