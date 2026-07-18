package core

import "fmt"

func FormatBudget(cents int64) string {
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("%d.%02d", dollars, remainder)
}

func FormatBudgetTier(cents int64, isCustom bool) string {
	if isCustom {
		return fmt.Sprintf("Custom ($%s)", FormatBudget(cents))
	}

	switch cents {
	case 250000:
		return "Small ($1,000 - $5,000)"
	case 750000:
		return "Starter ($5,000 - $10,000)"
	case 1750000:
		return "Growth ($10,000 - $25,000)"
	case 5000000:
		return "Scale ($25,000 - $75,000)"
	case 11250000:
		return "Enterprise ($75,000 - $150,000)"
	case 20000000:
		return "Premium ($150,000+)"
	default:
		return fmt.Sprintf("Custom ($%s)", FormatBudget(cents))
	}
}
