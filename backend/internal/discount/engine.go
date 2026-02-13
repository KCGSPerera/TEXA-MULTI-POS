package discount

import "github.com/kcgsperera/texa-multi-pos/backend/internal/models"

type Breakdown struct {
	Applied []models.AppliedDiscount
	Total   float64
}

func Calculate(orderAmount float64, itemLevelTotal float64, rules []models.DiscountRule, loyaltyRedeem float64) Breakdown {
	applied := make([]models.AppliedDiscount, 0)
	total := 0.0

	if itemLevelTotal > 0 {
		applied = append(applied, models.AppliedDiscount{Name: "Item Discounts", Type: "ITEM", Amount: itemLevelTotal})
		total += itemLevelTotal
	}

	for _, rule := range rules {
		if !rule.IsActive || orderAmount < rule.MinOrderAmount {
			continue
		}
		amount := 0.0
		switch rule.Type {
		case "PERCENTAGE", "LOYALTY", "PROMO":
			amount = orderAmount * (rule.Value / 100)
		case "FLAT":
			amount = rule.Value
		default:
			continue
		}
		if rule.MaxDiscountAmount != nil && amount > *rule.MaxDiscountAmount {
			amount = *rule.MaxDiscountAmount
		}
		if amount < 0 {
			amount = 0
		}
		id := rule.ID
		applied = append(applied, models.AppliedDiscount{DiscountRuleID: &id, Name: rule.Name, Type: rule.Type, Amount: amount})
		total += amount
	}

	if loyaltyRedeem > 0 {
		applied = append(applied, models.AppliedDiscount{Name: "Loyalty Redemption", Type: "LOYALTY_REDEEM", Amount: loyaltyRedeem})
		total += loyaltyRedeem
	}

	if total > orderAmount {
		total = orderAmount
	}

	return Breakdown{Applied: applied, Total: total}
}
