package data

import (
	"math"
	"time"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

type FixtureProvider struct{}

func (FixtureProvider) Load() (domain.DataSet, error) {
	securities := []domain.Security{
		{Symbol: "000300", Name: "CSI 300", Kind: domain.AssetIndex, Sector: "Broad"},
		{Symbol: "000905", Name: "CSI 500", Kind: domain.AssetIndex, Sector: "Broad"},
		{Symbol: "399006", Name: "ChiNext Index", Kind: domain.AssetIndex, Sector: "Growth"},
		{Symbol: "000688", Name: "STAR 50", Kind: domain.AssetIndex, Sector: "Growth"},
		{Symbol: "515980", Name: "AI Industry ETF", Kind: domain.AssetETF, Sector: "AI", Chain: "AI infrastructure", BenchmarkSymbol: "000300"},
		{Symbol: "512760", Name: "Chip ETF", Kind: domain.AssetETF, Sector: "Chip", Chain: "Semiconductor", BenchmarkSymbol: "000300"},
		{Symbol: "512480", Name: "Semiconductor ETF", Kind: domain.AssetETF, Sector: "Semiconductor", Chain: "Semiconductor", BenchmarkSymbol: "000300"},
		{Symbol: "159995", Name: "Computer ETF", Kind: domain.AssetETF, Sector: "Software", Chain: "AI software", BenchmarkSymbol: "000300"},
		{Symbol: "588000", Name: "STAR ETF", Kind: domain.AssetETF, Sector: "STAR", Chain: "Hard tech", BenchmarkSymbol: "000300"},
		{Symbol: "688256", Name: "AI Chip Leader", Kind: domain.AssetStock, Sector: "AI", Chain: "AI chips", BenchmarkSymbol: "515980"},
		{Symbol: "002230", Name: "AI Application Leader", Kind: domain.AssetStock, Sector: "Software", Chain: "AI application", BenchmarkSymbol: "159995"},
		{Symbol: "300308", Name: "Optical Module Leader", Kind: domain.AssetStock, Sector: "AI", Chain: "AI compute network", BenchmarkSymbol: "515980"},
	}

	start := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	bars := map[string][]domain.Bar{
		"000300": makeSeries("000300", start, 150, 3800, 0.0006, 0.012, 1.00, 0),
		"000905": makeSeries("000905", start, 150, 5600, 0.0008, 0.014, 1.02, 0),
		"399006": makeSeries("399006", start, 150, 2100, 0.0014, 0.018, 1.12, 0),
		"000688": makeSeries("000688", start, 150, 980, 0.0017, 0.019, 1.16, 0),
		"515980": makeSeries("515980", start, 150, 0.92, 0.0026, 0.021, 1.45, 1),
		"512760": makeSeries("512760", start, 150, 1.10, 0.0023, 0.020, 1.34, 1),
		"512480": makeSeries("512480", start, 150, 0.88, 0.0021, 0.020, 1.25, 0),
		"159995": makeSeries("159995", start, 150, 0.75, 0.0016, 0.018, 1.10, 0),
		"588000": makeSeries("588000", start, 150, 1.05, 0.0012, 0.018, 0.92, 0),
		"688256": makeSeries("688256", start, 150, 145, 0.0030, 0.032, 1.70, 2),
		"002230": makeSeries("002230", start, 150, 42, 0.0014, 0.026, 1.05, -1),
		"300308": makeSeries("300308", start, 150, 118, 0.0028, 0.030, 1.48, 1),
	}

	fundamentals := map[string]domain.Fundamental{
		"688256": {Symbol: "688256", RevenueGrowth: 0.38, ProfitGrowth: 0.52, ROE: 0.12, GrossMargin: 0.58, OperatingCashflowRatio: 0.16, RDRatio: 0.21, AIRelevance: 0.95, ListedDays: 2200},
		"002230": {Symbol: "002230", RevenueGrowth: 0.08, ProfitGrowth: -0.05, ROE: 0.06, GrossMargin: 0.41, OperatingCashflowRatio: 0.03, RDRatio: 0.15, AIRelevance: 0.76, ListedDays: 4000, RecentReportSlowdown: true},
		"300308": {Symbol: "300308", RevenueGrowth: 0.34, ProfitGrowth: 0.46, ROE: 0.18, GrossMargin: 0.31, OperatingCashflowRatio: 0.19, RDRatio: 0.11, AIRelevance: 0.88, ListedDays: 3200},
	}

	return domain.DataSet{
		Securities:   securities,
		Bars:         bars,
		Fundamentals: fundamentals,
		Breadth: domain.MarketBreadth{
			Date:           start.AddDate(0, 0, 149),
			Advancers:      3180,
			Decliners:      1850,
			LimitDownCount: 18,
			TotalAmount:    11200,
			AvgAmount20:    10100,
			TechReturn60:   0.18,
			BroadReturn60:  0.06,
		},
	}, nil
}

func makeSeries(symbol string, start time.Time, n int, base, drift, wave, amountScale float64, pattern int) []domain.Bar {
	bars := make([]domain.Bar, 0, n)
	price := base
	for i := 0; i < n; i++ {
		date := start.AddDate(0, 0, i)
		seasonal := math.Sin(float64(i)/7.0) * wave
		daily := drift + seasonal*0.12
		if i > 105 {
			daily += 0.0018 * float64(pattern)
		}
		if i == n-1 && pattern > 0 {
			daily += 0.028
		}
		if i == n-1 && pattern < 0 {
			daily -= 0.038
		}
		open := price * (1 + seasonal*0.05)
		close := price * (1 + daily)
		high := math.Max(open, close) * (1 + 0.006 + wave*0.08)
		low := math.Min(open, close) * (1 - 0.006 - wave*0.05)
		volume := (1000000 + float64(i%17)*35000) * amountScale
		if i == n-1 && pattern > 0 {
			volume *= 1.9
		}
		if i == n-1 && pattern < 0 {
			volume *= 1.8
		}
		amount := volume * close
		bars = append(bars, domain.Bar{
			Symbol:   symbol,
			Date:     date,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Volume:   volume,
			Amount:   amount,
			Turnover: 0.01 + float64(i%20)*0.001 + amountScale*0.002,
		})
		price = close
	}
	return bars
}
