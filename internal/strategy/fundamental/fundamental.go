package fundamental

import (
	"fmt"
	"strings"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

type Evaluator struct{}

func (Evaluator) Evaluate(security domain.Security, f domain.Fundamental, hasFundamental bool) domain.FundamentalResult {
	if security.IsETF() {
		return domain.FundamentalResult{
			Score:   75,
			Pass:    true,
			Summary: "ETF uses liquidity/trend quality; issuer fundamentals not scored",
			Reasons: []string{"ETF fundamental gate defaults to neutral-positive"},
		}
	}
	if !hasFundamental {
		return domain.FundamentalResult{
			Score:   45,
			Pass:    false,
			Summary: "missing fundamentals",
			Reasons: []string{"stock lacks fundamental snapshot and is kept on watch only"},
		}
	}

	risks := make([]string, 0)
	if f.ST {
		risks = append(risks, "ST")
	}
	if f.DelistingRisk {
		risks = append(risks, "delisting risk")
	}
	if f.ListedDays > 0 && f.ListedDays < 120 {
		risks = append(risks, "listed less than 120 days")
	}
	if f.LossDeteriorating || f.ConsecutiveLosses {
		risks = append(risks, "losses deteriorating")
	}
	if f.CashflowPoor {
		risks = append(risks, "poor operating cash flow")
	}
	if f.GoodwillReceivableRisk {
		risks = append(risks, "goodwill/receivable risk")
	}
	if f.RecentReportSlowdown {
		risks = append(risks, "recent report slowdown")
	}
	pass := len(risks) == 0

	score := 35.0
	score += domain.Normalize(f.RevenueGrowth, -0.10, 0.45) * 0.16
	score += domain.Normalize(f.ProfitGrowth, -0.20, 0.60) * 0.18
	score += domain.Normalize(f.ROE, 0.02, 0.20) * 0.18
	score += domain.Normalize(f.GrossMargin, 0.15, 0.60) * 0.12
	score += domain.Normalize(f.OperatingCashflowRatio, -0.10, 0.25) * 0.14
	score += domain.Normalize(f.RDRatio, 0.02, 0.18) * 0.10
	score += domain.Normalize(f.AIRelevance, 0.20, 1.00) * 0.12
	if !pass {
		score -= 35
	}
	score = domain.Clamp(score, 0, 100)

	reasons := []string{
		fmt.Sprintf("revenue growth %.1f%%", f.RevenueGrowth*100),
		fmt.Sprintf("profit growth %.1f%%", f.ProfitGrowth*100),
		fmt.Sprintf("ROE %.1f%%", f.ROE*100),
		fmt.Sprintf("AI relevance %.0f%%", f.AIRelevance*100),
	}
	if !pass {
		reasons = append(reasons, risks...)
	}

	summary := "quality gate passed"
	if !pass {
		summary = "quality gate failed: " + strings.Join(risks, ", ")
	}
	return domain.FundamentalResult{
		Score:   score,
		Pass:    pass,
		Summary: summary,
		Reasons: reasons,
	}
}
