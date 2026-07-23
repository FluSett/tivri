package core

import "fmt"

func FormatBudget(usd int64) string {
	if usd <= 0 {
		return "$0"
	}

	str := fmt.Sprintf("%d", usd)
	n := len(str)
	if n <= 3 {
		return "$" + str
	}

	var result []byte
	for i, c := range str {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return "$" + string(result)
}

func FormatBudgetTier(usd int64, _ bool) string {
	return FormatBudget(usd)
}

