package scoring

import (
	"fmt"
	"sort"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/fundamental"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/market"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/sector"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/trend"
	"github.com/waytofree-yang/right-side-trading/internal/strategy/volume"
)

type Engine struct {
	Market      market.Evaluator
	Sector      sector.Evaluator
	Trend       trend.Evaluator
	Volume      volume.Evaluator
	Fundamental fundamental.Evaluator
}

type Result struct {
	Market          domain.MarketResult
	Sectors         []domain.SectorResult
	Recommendations []domain.Recommendation
}

func NewEngine() Engine {
	return Engine{
		Market:      market.Evaluator{},
		Sector:      sector.Evaluator{},
		Trend:       trend.Evaluator{},
		Volume:      volume.Evaluator{},
		Fundamental: fundamental.Evaluator{},
	}
}

func (e Engine) Score(ds domain.DataSet) Result {
	marketResult := e.Market.Evaluate(ds)
	sectorResults := e.Sector.Rank(ds)
	sectorByName := make(map[string]domain.SectorResult, len(sectorResults))
	for _, result := range sectorResults {
		sectorByName[result.Sector] = result
	}

	recommendations := make([]domain.Recommendation, 0)
	for _, security := range ds.Securities {
		if !security.IsTradableCandidate() {
			continue
		}
		bars := ds.Bars[security.Symbol]
		if len(bars) == 0 {
			continue
		}
		sectorResult, ok := sectorByName[security.Sector]
		if !ok && security.IsETF() {
			sectorResult, ok = sector.EvaluateSecurity(security, bars, ds.Bars["000300"])
		}
		if !ok {
			sectorResult = domain.SectorResult{
				Sector:  security.Sector,
				Symbol:  security.Symbol,
				Name:    security.Name,
				Score:   35,
				Allowed: false,
				Reasons: []string{"sector proxy is missing; candidate is downgraded"},
			}
		}

		benchmarkSymbol := security.BenchmarkSymbol
		if benchmarkSymbol == "" {
			benchmarkSymbol = sectorResult.Symbol
		}
		benchmarkBars := ds.Bars[benchmarkSymbol]
		if len(benchmarkBars) == 0 {
			benchmarkBars = ds.Bars["000300"]
		}
		trendResult := e.Trend.Evaluate(bars, benchmarkBars)
		volumeResult := e.Volume.Evaluate(bars)
		fundamentalSnapshot, hasFundamental := ds.Fundamentals[security.Symbol]
		fundamentalResult := e.Fundamental.Evaluate(security, fundamentalSnapshot, hasFundamental)

		rec := Combine(security, marketResult, sectorResult, trendResult, volumeResult, fundamentalResult)
		recommendations = append(recommendations, rec)
	}

	sort.SliceStable(recommendations, func(i, j int) bool {
		if recommendations[i].Grade == recommendations[j].Grade {
			return recommendations[i].TotalScore > recommendations[j].TotalScore
		}
		return gradeRank(recommendations[i].Grade) < gradeRank(recommendations[j].Grade)
	})

	return Result{
		Market:          marketResult,
		Sectors:         sectorResults,
		Recommendations: recommendations,
	}
}

func Combine(
	security domain.Security,
	marketResult domain.MarketResult,
	sectorResult domain.SectorResult,
	trendResult domain.TrendResult,
	volumeResult domain.VolumeResult,
	fundamentalResult domain.FundamentalResult,
) domain.Recommendation {
	sectorScore := sectorResult.Score * 0.25
	trendScore := trendResult.Score * 0.25
	triggerRaw := 30.0
	if trendResult.TriggerConfirmed {
		triggerRaw = 100
	} else if trendResult.Established {
		triggerRaw = 55
	}
	if trendResult.Overheated || trendResult.TradingRestricted {
		triggerRaw -= 35
	}
	triggerRaw = domain.Clamp(triggerRaw, 0, 100)
	triggerScore := triggerRaw * 0.20
	volumeScore := volumeResult.Score * 0.15
	fundamentalScore := fundamentalResult.Score * 0.15
	total := sectorScore + trendScore + triggerScore + volumeScore + fundamentalScore

	risks := make([]string, 0)
	reasons := make([]string, 0)
	reasons = append(reasons, fmt.Sprintf("sector strength %.1f/100", sectorResult.Score))
	reasons = append(reasons, fmt.Sprintf("trend %s", trendResult.Status))
	reasons = append(reasons, fmt.Sprintf("volume %s", volumeResult.Status))
	reasons = append(reasons, fundamentalResult.Summary)

	if marketResult.State == domain.RiskOff {
		risks = append(risks, "market is Risk-Off")
	}
	if !sectorResult.Allowed {
		risks = append(risks, "sector is outside top relative-strength groups")
	}
	if !trendResult.Established {
		risks = append(risks, "right-side trend is not established")
	}
	if !trendResult.TriggerConfirmed {
		risks = append(risks, "right-side trigger is incomplete")
	}
	if trendResult.Overheated {
		risks = append(risks, "short-term extension is too high")
	}
	if trendResult.TradingRestricted {
		risks = append(risks, "latest bar is not suitable for chasing")
	}
	if volumeResult.Distribution {
		risks = append(risks, "heavy-volume distribution signal")
	}
	if !fundamentalResult.Pass {
		risks = append(risks, "fundamental quality gate failed")
	}

	grade := domain.GradeWatch
	switch {
	case !fundamentalResult.Pass || volumeResult.Distribution || (!security.IsETF() && !sectorResult.Allowed && !trendResult.Established):
		grade = domain.GradeAvoid
	case marketResult.State == domain.RiskOff:
		grade = domain.GradeWatch
	case trendResult.TradingRestricted || trendResult.Overheated:
		grade = domain.GradeWatch
	case total >= 80 && sectorResult.Allowed && trendResult.Established && trendResult.TriggerConfirmed && volumeResult.Accumulation:
		grade = domain.GradeA
	case total >= 70 && sectorResult.Allowed && trendResult.Established:
		grade = domain.GradeB
	case !sectorResult.Allowed || !trendResult.Established:
		grade = domain.GradeAvoid
	default:
		grade = domain.GradeWatch
	}

	if grade == domain.GradeA {
		reasons = append(reasons, "market is not Risk-Off and trend/volume are confirmed")
	}
	if grade == domain.GradeB {
		reasons = append(reasons, "setup is valid but waits for cleaner trigger or better risk-reward")
	}

	return domain.Recommendation{
		Security:         security,
		Grade:            grade,
		TotalScore:       domain.Clamp(total, 0, 100),
		SectorScore:      sectorScore,
		TrendScore:       trendScore,
		TriggerScore:     triggerScore,
		VolumeScore:      volumeScore,
		FundamentalScore: fundamentalScore,
		Sector:           sectorResult,
		Trend:            trendResult,
		Volume:           volumeResult,
		Fundamental:      fundamentalResult,
		Risks:            risks,
		Reasons:          reasons,
		ObservationPrice: trendResult.ObservationPrice,
	}
}

func gradeRank(grade domain.Grade) int {
	switch grade {
	case domain.GradeA:
		return 0
	case domain.GradeB:
		return 1
	case domain.GradeWatch:
		return 2
	default:
		return 3
	}
}
