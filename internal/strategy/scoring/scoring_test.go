package scoring

import (
	"testing"

	"github.com/waytofree-yang/right-side-trading/internal/data"
	"github.com/waytofree-yang/right-side-trading/internal/domain"
)

func TestRiskOffBlocksActiveRecommendations(t *testing.T) {
	ds, err := data.FixtureProvider{}.Load()
	if err != nil {
		t.Fatal(err)
	}
	ds.Breadth.Advancers = 900
	ds.Breadth.Decliners = 4300
	ds.Breadth.LimitDownCount = 120
	ds.Breadth.TechReturn60 = -0.12
	ds.Breadth.BroadReturn60 = 0.02

	result := NewEngine().Score(ds)
	if result.Market.State != domain.RiskOff {
		t.Fatalf("expected Risk-Off, got %s", result.Market.State)
	}
	for _, rec := range result.Recommendations {
		if rec.Grade == domain.GradeA || rec.Grade == domain.GradeB {
			t.Fatalf("Risk-Off should block active recommendations, got %s for %s", rec.Grade, rec.Security.Symbol)
		}
	}
}
