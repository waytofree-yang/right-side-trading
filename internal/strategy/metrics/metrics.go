package metrics

import (
	"math"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func SMA(bars []domain.Bar, n int) (float64, bool) {
	if n <= 0 || len(bars) < n {
		return 0, false
	}
	var sum float64
	for _, b := range bars[len(bars)-n:] {
		sum += b.Close
	}
	return sum / float64(n), true
}

func SMABefore(bars []domain.Bar, n int, offset int) (float64, bool) {
	if offset < 0 || n <= 0 || len(bars) < n+offset {
		return 0, false
	}
	end := len(bars) - offset
	start := end - n
	var sum float64
	for _, b := range bars[start:end] {
		sum += b.Close
	}
	return sum / float64(n), true
}

func AverageAmount(bars []domain.Bar, n int) (float64, bool) {
	if n <= 0 || len(bars) < n {
		return 0, false
	}
	var sum float64
	for _, b := range bars[len(bars)-n:] {
		sum += b.Amount
	}
	return sum / float64(n), true
}

func Return(bars []domain.Bar, n int) (float64, bool) {
	if n <= 0 || len(bars) <= n {
		return 0, false
	}
	start := bars[len(bars)-1-n].Close
	end := bars[len(bars)-1].Close
	if start <= 0 {
		return 0, false
	}
	return end/start - 1, true
}

func MaxClose(bars []domain.Bar, lookback int, excludeRecent int) (float64, bool) {
	if lookback <= 0 || excludeRecent < 0 || len(bars) < lookback+excludeRecent {
		return 0, false
	}
	end := len(bars) - excludeRecent
	start := end - lookback
	maxClose := bars[start].Close
	for _, b := range bars[start:end] {
		if b.Close > maxClose {
			maxClose = b.Close
		}
	}
	return maxClose, true
}

func MaxDrawdown(bars []domain.Bar, n int) (float64, bool) {
	if n <= 1 || len(bars) < n {
		return 0, false
	}
	window := bars[len(bars)-n:]
	peak := window[0].Close
	maxDD := 0.0
	for _, b := range window {
		if b.Close > peak {
			peak = b.Close
		}
		if peak > 0 {
			dd := b.Close/peak - 1
			if dd < maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD, true
}

func Percentile(values []float64, v float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var below float64
	for _, x := range values {
		if x <= v {
			below++
		}
	}
	return below / float64(len(values))
}

func LatestAmountRatio(bars []domain.Bar, n int) (float64, bool) {
	if len(bars) < n+1 {
		return 0, false
	}
	avg, ok := AverageAmount(bars[:len(bars)-1], n)
	if !ok || avg <= 0 {
		return 0, false
	}
	return bars[len(bars)-1].Amount / avg, true
}

func SafeRatio(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func Round1(v float64) float64 {
	return math.Round(v*10) / 10
}
