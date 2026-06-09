package trend

import (
	"fmt"
	"math"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/metrics"
)

type Evaluator struct{}

func (Evaluator) Evaluate(bars, benchmark []domain.Bar) domain.TrendResult {
	if len(bars) < 120 || len(benchmark) < 60 {
		return domain.TrendResult{
			Score:   0,
			Status:  "insufficient data",
			Reasons: []string{"needs at least 120 bars for right-side trend confirmation"},
		}
	}

	last := bars[len(bars)-1]
	prev := bars[len(bars)-2]
	ma20, _ := metrics.SMA(bars, 20)
	ma60, _ := metrics.SMA(bars, 60)
	ma20Prev, _ := metrics.SMABefore(bars, 20, 5)
	ret60, _ := metrics.Return(bars, 60)
	bench60, _ := metrics.Return(benchmark, 60)
	rel60 := ret60 - bench60
	amountRatio, ok := metrics.LatestAmountRatio(bars, 20)
	if !ok {
		amountRatio = 1
	}
	distanceFromMA20 := last.Close/ma20 - 1

	priceAboveMA60 := last.Close > ma60
	ma20Up := ma20 > ma20Prev
	relativeStrong := rel60 > 0
	established := priceAboveMA60 && ma20Up && relativeStrong

	trigger, confirmed := detectTrigger(bars, ma20, ma60, amountRatio)
	ret5, _ := metrics.Return(bars, 5)
	overheated := distanceFromMA20 > 0.12 || ret5 > 0.18
	tradingRestricted := last.Paused || last.LimitUp

	score := 0.0
	if priceAboveMA60 {
		score += 30
	}
	if ma20Up {
		score += 20
	}
	if relativeStrong {
		score += 20
	}
	score += domain.Normalize(rel60, -0.08, 0.18) * 0.15
	if confirmed {
		score += 15
	}
	if overheated {
		score -= 18
	}
	if tradingRestricted {
		score -= 25
	}
	score = domain.Clamp(score, 0, 100)

	status := "trend broken"
	if established {
		status = "trend established"
	}
	if confirmed {
		status = string(trigger)
	}
	if overheated {
		status = "overheated watch"
	}
	if tradingRestricted {
		status = "restricted watch"
	}

	reasons := []string{
		fmt.Sprintf("close %.2f vs MA60 %.2f", last.Close, ma60),
		fmt.Sprintf("MA20 slope %s", direction(ma20-ma20Prev)),
		fmt.Sprintf("60d relative return %.1f%%", rel60*100),
		fmt.Sprintf("distance from MA20 %.1f%%", distanceFromMA20*100),
	}
	if confirmed {
		reasons = append(reasons, fmt.Sprintf("right-side trigger: %s", trigger))
	}
	if overheated {
		reasons = append(reasons, "extended from MA20 or recent 5-day surge; keep on watch")
	}
	if last.LimitUp {
		reasons = append(reasons, "latest bar is limit-up; do not chase as active recommendation")
	}
	if last.Paused {
		reasons = append(reasons, "latest bar is paused")
	}
	if last.Close < prev.Close && amountRatio > 1.6 {
		reasons = append(reasons, "heavy-volume pullback needs confirmation")
	}

	return domain.TrendResult{
		Score:             score,
		Established:       established,
		Trigger:           trigger,
		TriggerConfirmed:  confirmed,
		RelativeReturn60:  rel60,
		DistanceFromMA20:  distanceFromMA20,
		Overheated:        overheated,
		TradingRestricted: tradingRestricted,
		ObservationPrice:  math.Max(ma20, ma60),
		Status:            status,
		Reasons:           reasons,
	}
}

func detectTrigger(bars []domain.Bar, ma20, ma60, amountRatio float64) (domain.TrendTrigger, bool) {
	last := bars[len(bars)-1]
	prev := bars[len(bars)-2]
	priorHigh, highOK := metrics.MaxClose(bars, 60, 1)
	if highOK && last.Close > priorHigh && last.Close > last.Open && amountRatio >= 1.45 {
		return domain.TriggerVolumeBreakout, true
	}

	recentBreakout := false
	for i := len(bars) - 15; i < len(bars)-1; i++ {
		if i < 61 {
			continue
		}
		high, ok := metrics.MaxClose(bars[:i+1], 60, 1)
		if ok && bars[i].Close > high {
			recentBreakout = true
			break
		}
	}
	if recentBreakout && last.Close >= ma20*0.98 && last.Close >= ma60 && last.Amount < averageAmountLast(bars, 20)*1.05 {
		return domain.TriggerPullbackConfirm, true
	}
	if recentBreakout && prev.Close >= ma20*0.97 && last.Close > prev.Close && amountRatio >= 1.2 {
		return domain.TriggerReacceleration, true
	}
	return domain.TriggerNone, false
}

func averageAmountLast(bars []domain.Bar, n int) float64 {
	avg, ok := metrics.AverageAmount(bars, n)
	if !ok {
		return 0
	}
	return avg
}

func direction(v float64) string {
	if v > 0 {
		return "up"
	}
	if v < 0 {
		return "down"
	}
	return "flat"
}
