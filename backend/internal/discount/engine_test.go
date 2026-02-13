package discount

import (
	"testing"

	"github.com/kcgsperera/texa-multi-pos/backend/internal/models"
)

func TestCalculateDiscountBreakdown(t *testing.T) {
	capVal := 15.0
	tests := []struct {
		name        string
		order       float64
		item        float64
		rules       []models.DiscountRule
		loyalty     float64
		expectTotal float64
	}{
		{name: "percentage rule", order: 100, rules: []models.DiscountRule{{ID: "r1", Name: "10%", Type: "PERCENTAGE", Value: 10, IsActive: true}}, expectTotal: 10},
		{name: "flat plus loyalty", order: 100, rules: []models.DiscountRule{{ID: "r1", Name: "Flat", Type: "FLAT", Value: 5, IsActive: true}}, loyalty: 10, expectTotal: 15},
		{name: "max cap", order: 200, rules: []models.DiscountRule{{ID: "r1", Name: "20%", Type: "PERCENTAGE", Value: 20, MaxDiscountAmount: &capVal, IsActive: true}}, expectTotal: 15},
		{name: "never exceed order", order: 10, item: 8, loyalty: 10, expectTotal: 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Calculate(tc.order, tc.item, tc.rules, tc.loyalty)
			if got.Total != tc.expectTotal {
				t.Fatalf("expected %v, got %v", tc.expectTotal, got.Total)
			}
		})
	}
}
