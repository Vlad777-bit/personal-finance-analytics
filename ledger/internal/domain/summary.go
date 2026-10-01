package domain

import "time"

type Summary struct {
	UserID     string
	From       time.Time
	To         time.Time
	TotalSpent int64
	Categories []CategorySummary
}

type CategorySummary struct {
	Category         string
	Spent            int64
	BudgetLimit      int64
	BudgetConfigured bool
	Remaining        int64
	BudgetExceeded   bool
}

type CategoryReportData struct {
	Category         string
	Spent            int64
	BudgetLimit      int64
	BudgetConfigured bool
}

func BuildSummary(
	userID string,
	from time.Time,
	to time.Time,
	data []CategoryReportData,
) Summary {
	categories := make([]CategorySummary, 0, len(data))
	var totalSpent int64

	for _, categoryData := range data {
		category := CategorySummary{
			Category:         categoryData.Category,
			Spent:            categoryData.Spent,
			BudgetLimit:      categoryData.BudgetLimit,
			BudgetConfigured: categoryData.BudgetConfigured,
		}
		if category.BudgetConfigured {
			category.Remaining = category.BudgetLimit - category.Spent
			category.BudgetExceeded = category.Spent > category.BudgetLimit
		}

		totalSpent += category.Spent
		categories = append(categories, category)
	}

	return Summary{
		UserID:     userID,
		From:       from,
		To:         to,
		TotalSpent: totalSpent,
		Categories: categories,
	}
}
