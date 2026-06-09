package domain

import (
	"fmt"
	"math"
	"time"
)

type AssetKind string

const (
	AssetETF   AssetKind = "ETF"
	AssetStock AssetKind = "Stock"
	AssetIndex AssetKind = "Index"
)

type Security struct {
	Symbol          string
	Name            string
	Kind            AssetKind
	Sector          string
	Chain           string
	BenchmarkSymbol string
}

func (s Security) IsETF() bool {
	return s.Kind == AssetETF
}

func (s Security) IsTradableCandidate() bool {
	return s.Kind == AssetETF || s.Kind == AssetStock
}

type Bar struct {
	Symbol    string
	Date      time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
	Amount    float64
	Turnover  float64
	LimitUp   bool
	LimitDown bool
	Paused    bool
	AdjFactor float64
}

type Fundamental struct {
	Symbol                 string
	RevenueGrowth          float64
	ProfitGrowth           float64
	ROE                    float64
	GrossMargin            float64
	OperatingCashflowRatio float64
	RDRatio                float64
	AIRelevance            float64
	ST                     bool
	DelistingRisk          bool
	ListedDays             int
	LossDeteriorating      bool
	CashflowPoor           bool
	GoodwillReceivableRisk bool
	RecentReportSlowdown   bool
	ConsecutiveLosses      bool
}

type MarketBreadth struct {
	Date           time.Time
	Advancers      int
	Decliners      int
	LimitDownCount int
	TotalAmount    float64
	AvgAmount20    float64
	TechReturn60   float64
	BroadReturn60  float64
}

type DataSet struct {
	Securities   []Security
	Bars         map[string][]Bar
	Fundamentals map[string]Fundamental
	Breadth      MarketBreadth
}

type MarketState string

const (
	RiskOn  MarketState = "Risk-On"
	Neutral MarketState = "Neutral"
	RiskOff MarketState = "Risk-Off"
)

type MarketResult struct {
	State        MarketState
	Score        float64
	AboveMA120   int
	Breadth      float64
	VolumeRatio  float64
	TechRelative float64
	Reasons      []string
}

type SectorResult struct {
	Sector           string
	Symbol           string
	Name             string
	Score            float64
	Rank             int
	RelativeReturn20 float64
	RelativeReturn60 float64
	TrendScore       float64
	AmountScore      float64
	ResilienceScore  float64
	Allowed          bool
	Reasons          []string
}

type TrendTrigger string

const (
	TriggerNone            TrendTrigger = "None"
	TriggerVolumeBreakout  TrendTrigger = "VolumeBreakout"
	TriggerPullbackConfirm TrendTrigger = "PullbackConfirm"
	TriggerReacceleration  TrendTrigger = "Reacceleration"
)

type TrendResult struct {
	Score             float64
	Established       bool
	Trigger           TrendTrigger
	TriggerConfirmed  bool
	RelativeReturn60  float64
	DistanceFromMA20  float64
	Overheated        bool
	TradingRestricted bool
	ObservationPrice  float64
	Status            string
	Reasons           []string
}

type VolumeResult struct {
	Score              float64
	AmountRatio        float64
	TurnoverPercentile float64
	OBVUp              bool
	CMF                float64
	UpAmountShare      float64
	Accumulation       bool
	Distribution       bool
	Status             string
	Reasons            []string
}

type FundamentalResult struct {
	Score   float64
	Pass    bool
	Summary string
	Reasons []string
}

type Grade string

const (
	GradeA     Grade = "A"
	GradeB     Grade = "B"
	GradeWatch Grade = "Watch"
	GradeAvoid Grade = "Avoid"
)

type Recommendation struct {
	Security         Security
	Grade            Grade
	TotalScore       float64
	SectorScore      float64
	TrendScore       float64
	TriggerScore     float64
	VolumeScore      float64
	FundamentalScore float64
	Sector           SectorResult
	Trend            TrendResult
	Volume           VolumeResult
	Fundamental      FundamentalResult
	Risks            []string
	Reasons          []string
	ObservationPrice float64
}

func (r Recommendation) ScoreText() string {
	return fmt.Sprintf("%.1f", r.TotalScore)
}

func Clamp(v, min, max float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return min
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func Normalize(v, min, max float64) float64 {
	if max <= min {
		return 0
	}
	return Clamp((v-min)/(max-min)*100, 0, 100)
}

func LastBar(bars []Bar) (Bar, bool) {
	if len(bars) == 0 {
		return Bar{}, false
	}
	return bars[len(bars)-1], true
}
