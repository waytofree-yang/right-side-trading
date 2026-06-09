package volume

import (
	"fmt"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/metrics"
)

type Evaluator struct{}

func (Evaluator) Evaluate(bars []domain.Bar) domain.VolumeResult {
	if len(bars) < 40 {
		return domain.VolumeResult{
			Score:   0,
			Status:  "insufficient data",
			Reasons: []string{"needs at least 40 bars for volume proxy signals"},
		}
	}
	last := bars[len(bars)-1]
	prev := bars[len(bars)-2]
	amountRatio, ok := metrics.LatestAmountRatio(bars, 20)
	if !ok {
		amountRatio = 1
	}
	turnoverValues := make([]float64, 0, len(bars))
	for _, b := range bars {
		if b.Turnover > 0 {
			turnoverValues = append(turnoverValues, b.Turnover)
		}
	}
	turnoverPct := metrics.Percentile(turnoverValues, last.Turnover)
	obvUp := obvSlopeUp(bars, 20)
	cmf := chaikinMoneyFlow(bars, 20)
	upShare := upAmountShare(bars, 20)

	accumulation := (last.Close > prev.Close && amountRatio >= 1.25) || (upShare >= 0.56 && obvUp && cmf > 0)
	distribution := (last.Close < prev.Close && amountRatio >= 1.65) || (amountRatio >= 2.4 && last.Close <= prev.Close*1.01)

	score := domain.Normalize(amountRatio, 0.8, 1.9)*0.3 +
		turnoverPct*100*0.15 +
		boolScore(obvUp)*0.2 +
		domain.Normalize(cmf, -0.15, 0.25)*0.2 +
		domain.Normalize(upShare, 0.42, 0.65)*0.15
	if accumulation {
		score += 8
	}
	if distribution {
		score -= 28
	}
	score = domain.Clamp(score, 0, 100)

	status := "neutral volume"
	if accumulation {
		status = "accumulation"
	}
	if distribution {
		status = "distribution risk"
	}

	reasons := []string{
		fmt.Sprintf("amount ratio %.2fx", amountRatio),
		fmt.Sprintf("turnover percentile %.0f%%", turnoverPct*100),
		fmt.Sprintf("OBV trend %t", obvUp),
		fmt.Sprintf("CMF %.2f", cmf),
		fmt.Sprintf("up-day amount share %.0f%%", upShare*100),
	}
	if distribution {
		reasons = append(reasons, "heavy-volume down/flat bar weakens the right-side setup")
	}
	if accumulation {
		reasons = append(reasons, "volume expands on up days and contracts on pullbacks")
	}

	return domain.VolumeResult{
		Score:              score,
		AmountRatio:        amountRatio,
		TurnoverPercentile: turnoverPct,
		OBVUp:              obvUp,
		CMF:                cmf,
		UpAmountShare:      upShare,
		Accumulation:       accumulation,
		Distribution:       distribution,
		Status:             status,
		Reasons:            reasons,
	}
}

func boolScore(v bool) float64 {
	if v {
		return 100
	}
	return 0
}

func obvSlopeUp(bars []domain.Bar, n int) bool {
	if len(bars) < n+1 {
		return false
	}
	start := len(bars) - n
	var obv float64
	midValue := 0.0
	for i := start; i < len(bars); i++ {
		if bars[i].Close > bars[i-1].Close {
			obv += bars[i].Volume
		} else if bars[i].Close < bars[i-1].Close {
			obv -= bars[i].Volume
		}
		if i == start+n/2 {
			midValue = obv
		}
	}
	return obv > midValue
}

func chaikinMoneyFlow(bars []domain.Bar, n int) float64 {
	if len(bars) < n {
		return 0
	}
	var mfv, volume float64
	for _, b := range bars[len(bars)-n:] {
		if b.High <= b.Low {
			continue
		}
		multiplier := ((b.Close - b.Low) - (b.High - b.Close)) / (b.High - b.Low)
		mfv += multiplier * b.Volume
		volume += b.Volume
	}
	return metrics.SafeRatio(mfv, volume)
}

func upAmountShare(bars []domain.Bar, n int) float64 {
	if len(bars) < n+1 {
		return 0.5
	}
	var upAmount, totalAmount float64
	for i := len(bars) - n; i < len(bars); i++ {
		totalAmount += bars[i].Amount
		if bars[i].Close > bars[i-1].Close {
			upAmount += bars[i].Amount
		}
	}
	if totalAmount <= 0 {
		return 0.5
	}
	return upAmount / totalAmount
}
