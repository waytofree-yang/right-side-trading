package fundamental

import (
	"testing"

	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestEvaluateRejectsQualityRisks(t *testing.T) {
	result := Evaluator{}.Evaluate(
		domain.Security{Symbol: "BAD", Kind: domain.AssetStock},
		domain.Fundamental{Symbol: "BAD", ST: true, ListedDays: 300},
		true,
	)
	if result.Pass {
		t.Fatalf("expected quality gate failure")
	}
}

func TestEvaluateETFDefaultsToPass(t *testing.T) {
	result := Evaluator{}.Evaluate(domain.Security{Symbol: "ETF", Kind: domain.AssetETF}, domain.Fundamental{}, false)
	if !result.Pass {
		t.Fatalf("expected ETF to pass neutral fundamental gate")
	}
}
