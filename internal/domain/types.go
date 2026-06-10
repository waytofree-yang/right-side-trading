package domain

import (
	"fmt"
	"math"
	"time"
)

// AssetKind 表示证券类型。
// 当前系统只把 ETF 和个股作为可推荐标的，指数主要用作市场/行业基准。
type AssetKind string

const (
	// AssetETF 表示交易型开放式指数基金，第一版主要推荐对象。
	AssetETF AssetKind = "ETF"
	// AssetStock 表示 A 股个股，当前可参与评分，但会经过更严格的基本面护栏。
	AssetStock AssetKind = "Stock"
	// AssetIndex 表示指数，只作为行情基准或市场状态判断对象，不直接推荐。
	AssetIndex AssetKind = "Index"
)

// Security 是一个证券或指数的静态信息。
// 它描述“这个标的是什么、属于哪个行业、应该和谁比较”。
type Security struct {
	// Symbol 是证券代码或内部代码，例如 515980、000300。
	Symbol string
	// Name 是展示名称，用于报告输出。
	Name string
	// Kind 区分 ETF、个股和指数。
	Kind AssetKind
	// Sector 是策略使用的行业/主题分组，例如 AI、Chip、Software。
	Sector string
	// Chain 是 AI 产业链环节描述，例如 AI infrastructure、AI chips。
	Chain string
	// BenchmarkSymbol 是该标的做相对强弱比较时使用的基准代码。
	// 个股通常对比所属主题 ETF；ETF 通常对比宽基指数。
	BenchmarkSymbol string
}

// IsETF 判断标的是否为 ETF。
func (s Security) IsETF() bool {
	return s.Kind == AssetETF
}

// IsTradableCandidate 判断标的是否可以进入推荐候选池。
// 指数只做基准，不作为买入/观察建议输出。
func (s Security) IsTradableCandidate() bool {
	return s.Kind == AssetETF || s.Kind == AssetStock
}

// Bar 是单个交易日的行情数据。
// 策略模块用它计算均线、相对收益、放量、OBV、CMF 等技术/资金代理指标。
type Bar struct {
	// Symbol 对应 Security.Symbol。
	Symbol string
	// Date 是交易日期。
	Date time.Time
	// Open/High/Low/Close 是日线 OHLC 价格。
	Open  float64
	High  float64
	Low   float64
	Close float64
	// Volume 是成交量。单位由数据源决定，但同一数据集内必须一致。
	Volume float64
	// Amount 是成交额，用于放量倍数、市场成交额趋势等判断。
	Amount float64
	// Turnover 是换手率，用于衡量资金活跃度。
	Turnover float64
	// LimitUp 表示当天是否涨停。涨停不追高，只能降为观察。
	LimitUp bool
	// LimitDown 表示当天是否跌停。跌停会影响可交易性和风险判断。
	LimitDown bool
	// Paused 表示当天是否停牌。停牌标的不应给积极推荐。
	Paused bool
	// AdjFactor 是复权因子预留字段，当前 MVP 尚未主动使用。
	AdjFactor float64
}

// Fundamental 是个股基本面快照。
// 它在当前系统里主要做“质量护栏”，不是买点触发器。
type Fundamental struct {
	// Symbol 对应 Security.Symbol。
	Symbol string
	// RevenueGrowth 是营收增长率，例如 0.20 表示增长 20%。
	RevenueGrowth float64
	// ProfitGrowth 是净利润增长率。
	ProfitGrowth float64
	// ROE 是净资产收益率。
	ROE float64
	// GrossMargin 是毛利率，用于衡量业务质量和竞争壁垒。
	GrossMargin float64
	// OperatingCashflowRatio 是经营现金流质量指标。
	OperatingCashflowRatio float64
	// RDRatio 是研发投入强度，AI/科技股会重点参考。
	RDRatio float64
	// AIRelevance 是 AI 产业相关度，范围通常为 0 到 1。
	AIRelevance float64
	// ST 表示是否 ST，触发后会直接阻断基本面护栏。
	ST bool
	// DelistingRisk 表示是否存在退市风险。
	DelistingRisk bool
	// ListedDays 是上市天数，过短会降低策略可靠性。
	ListedDays int
	// LossDeteriorating 表示亏损是否继续恶化。
	LossDeteriorating bool
	// CashflowPoor 表示经营现金流是否明显偏弱。
	CashflowPoor bool
	// GoodwillReceivableRisk 表示商誉或应收账款风险是否偏高。
	GoodwillReceivableRisk bool
	// RecentReportSlowdown 表示近一期财报是否明显失速。
	RecentReportSlowdown bool
	// ConsecutiveLosses 表示是否连续亏损。
	ConsecutiveLosses bool
}

// MarketBreadth 是市场广度和总量信息。
// 它补充指数价格无法表达的市场内部强弱，例如涨跌家数和跌停数量。
type MarketBreadth struct {
	// Date 是广度数据日期。
	Date time.Time
	// Advancers 是上涨家数。
	Advancers int
	// Decliners 是下跌家数。
	Decliners int
	// LimitDownCount 是跌停家数，用于识别系统性风险。
	LimitDownCount int
	// TotalAmount 是全市场或目标市场总成交额。
	TotalAmount float64
	// AvgAmount20 是 20 日平均成交额，用于计算市场成交额放大/萎缩。
	AvgAmount20 float64
	// TechReturn60 是科技方向 60 日收益率。
	TechReturn60 float64
	// BroadReturn60 是宽基市场 60 日收益率。
	BroadReturn60 float64
}

// DataSet 是一次评分所需的完整输入快照。
// 策略引擎只依赖这个结构，不关心数据来自内置样例、CSV 还是未来的数据服务。
type DataSet struct {
	// Securities 是证券静态信息列表。
	Securities []Security
	// Bars 按 Symbol 存放历史 K 线。
	Bars map[string][]Bar
	// Fundamentals 按 Symbol 存放基本面快照。
	Fundamentals map[string]Fundamental
	// Breadth 是市场广度快照。
	Breadth MarketBreadth
}

// MarketState 是市场风险状态。
// 它是总闸门：Risk-Off 会阻止系统给出积极推荐。
type MarketState string

const (
	// RiskOn 表示市场环境支持积极右侧交易。
	RiskOn MarketState = "Risk-On"
	// Neutral 表示市场环境中性，推荐需要更谨慎。
	Neutral MarketState = "Neutral"
	// RiskOff 表示市场风险偏高，只输出观察或回避。
	RiskOff MarketState = "Risk-Off"
)

// MarketResult 是市场状态模块的输出。
type MarketResult struct {
	// State 是最终市场状态。
	State MarketState
	// Score 是 0 到 100 的市场环境评分。
	Score float64
	// AboveMA120 表示核心指数中站上 120 日均线的数量。
	AboveMA120 int
	// Breadth 是上涨家数占比。
	Breadth float64
	// VolumeRatio 是当前总成交额相对 20 日平均成交额的倍数。
	VolumeRatio float64
	// TechRelative 是科技方向相对宽基市场的 60 日超额收益。
	TechRelative float64
	// Reasons 是用于报告解释的文字原因。
	Reasons []string
}

// SectorResult 是行业/主题相对强度模块的输出。
type SectorResult struct {
	// Sector 是行业或主题名称。
	Sector string
	// Symbol 是该行业代理 ETF 或指数代码。
	Symbol string
	// Name 是代理标的名称。
	Name string
	// Score 是行业强度总分。
	Score float64
	// Rank 是行业强度排名。
	Rank int
	// RelativeReturn20 是 20 日相对基准收益。
	RelativeReturn20 float64
	// RelativeReturn60 是 60 日相对基准收益。
	RelativeReturn60 float64
	// TrendScore 是均线趋势分。
	TrendScore float64
	// AmountScore 是成交额放大分。
	AmountScore float64
	// ResilienceScore 是回撤韧性分。
	ResilienceScore float64
	// Allowed 表示该行业是否进入 Top 3，可以参与积极推荐。
	Allowed bool
	// Reasons 是行业评分解释。
	Reasons []string
}

// TrendTrigger 表示右侧交易触发形态。
type TrendTrigger string

const (
	// TriggerNone 表示没有完成右侧触发。
	TriggerNone TrendTrigger = "None"
	// TriggerVolumeBreakout 表示放量突破平台或阶段高点。
	TriggerVolumeBreakout TrendTrigger = "VolumeBreakout"
	// TriggerPullbackConfirm 表示突破后缩量回踩且不破关键均线。
	TriggerPullbackConfirm TrendTrigger = "PullbackConfirm"
	// TriggerReacceleration 表示回踩后重新放量上行。
	TriggerReacceleration TrendTrigger = "Reacceleration"
)

// TrendResult 是个股/ETF 趋势和右侧触发模块的输出。
type TrendResult struct {
	// Score 是趋势模块评分。
	Score float64
	// Established 表示右侧趋势是否成立：价格站上关键均线、短均线向上、相对收益为正。
	Established bool
	// Trigger 是识别到的右侧触发形态。
	Trigger TrendTrigger
	// TriggerConfirmed 表示触发形态是否已经确认。
	TriggerConfirmed bool
	// RelativeReturn60 是标的相对基准的 60 日收益。
	RelativeReturn60 float64
	// DistanceFromMA20 是当前价格距离 20 日均线的偏离度。
	DistanceFromMA20 float64
	// Overheated 表示短期涨幅过大或远离均线，不能追高。
	Overheated bool
	// TradingRestricted 表示停牌、涨停等不适合追入的交易状态。
	TradingRestricted bool
	// ObservationPrice 是报告里的观察价位，通常取关键均线附近。
	ObservationPrice float64
	// Status 是面向报告的简短状态描述。
	Status string
	// Reasons 是趋势判断解释。
	Reasons []string
}

// VolumeResult 是量价资金代理模块的输出。
// 第一版不依赖 Level2，使用成交额、换手率、OBV、CMF 等公开数据近似判断资金行为。
type VolumeResult struct {
	// Score 是资金/量价模块评分。
	Score float64
	// AmountRatio 是最新成交额相对 20 日均值的倍数。
	AmountRatio float64
	// TurnoverPercentile 是当前换手率在历史窗口中的分位。
	TurnoverPercentile float64
	// OBVUp 表示 OBV 是否向上，用于观察资金持续性。
	OBVUp bool
	// CMF 是 Chaikin Money Flow 指标，近似衡量资金流强弱。
	CMF float64
	// UpAmountShare 是上涨日成交额占比。
	UpAmountShare float64
	// Accumulation 表示疑似吸筹/承接良好。
	Accumulation bool
	// Distribution 表示疑似放量派发或巨量滞涨风险。
	Distribution bool
	// Status 是面向报告的量价状态描述。
	Status string
	// Reasons 是量价判断解释。
	Reasons []string
}

// FundamentalResult 是基本面护栏模块的输出。
type FundamentalResult struct {
	// Score 是基本面质量分。
	Score float64
	// Pass 表示是否通过质量护栏。
	Pass bool
	// Summary 是报告里展示的基本面摘要。
	Summary string
	// Reasons 是基本面评分或剔除原因。
	Reasons []string
}

// Grade 是最终推荐等级。
type Grade string

const (
	// GradeA 表示强推荐：市场、行业、趋势、触发和资金都确认。
	GradeA Grade = "A"
	// GradeB 表示可关注推荐：趋势成立，但触发或风险收益还不够完美。
	GradeB Grade = "B"
	// GradeWatch 表示观察：有亮点但还不能积极推荐。
	GradeWatch Grade = "Watch"
	// GradeAvoid 表示回避：行业弱、趋势坏、资金风险或基本面风险较明显。
	GradeAvoid Grade = "Avoid"
)

// Recommendation 是单个标的最终输出给报告的完整推荐结果。
type Recommendation struct {
	// Security 是被评分的 ETF 或个股。
	Security Security
	// Grade 是最终推荐等级。
	Grade Grade
	// TotalScore 是 0 到 100 的综合分。
	TotalScore float64
	// SectorScore 是行业强度加权后的贡献分。
	SectorScore float64
	// TrendScore 是趋势加权后的贡献分。
	TrendScore float64
	// TriggerScore 是右侧触发加权后的贡献分。
	TriggerScore float64
	// VolumeScore 是资金量价加权后的贡献分。
	VolumeScore float64
	// FundamentalScore 是基本面加权后的贡献分。
	FundamentalScore float64
	// Sector/Trend/Volume/Fundamental 保留各子模块原始结果，便于报告解释和调试。
	Sector      SectorResult
	Trend       TrendResult
	Volume      VolumeResult
	Fundamental FundamentalResult
	// Risks 是导致降级或回避的风险列表。
	Risks []string
	// Reasons 是最终推荐理由列表。
	Reasons []string
	// ObservationPrice 是观察价位，通常用于回踩或风控参考。
	ObservationPrice float64
}

// ScoreText 将综合分格式化成一位小数，便于报告输出。
func (r Recommendation) ScoreText() string {
	return fmt.Sprintf("%.1f", r.TotalScore)
}

// Clamp 将数值限制在 [min, max] 区间。
// NaN 和 Inf 会被当作异常输入处理，直接返回 min。
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

// Normalize 将原始指标按线性区间归一化到 0 到 100。
// 小于 min 的值记为 0，大于 max 的值记为 100。
func Normalize(v, min, max float64) float64 {
	if max <= min {
		return 0
	}
	return Clamp((v-min)/(max-min)*100, 0, 100)
}

// LastBar 返回最后一根 K 线。
// 第二个返回值表示输入切片是否非空。
func LastBar(bars []Bar) (Bar, bool) {
	if len(bars) == 0 {
		return Bar{}, false
	}
	return bars[len(bars)-1], true
}
