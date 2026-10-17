package forecast

import (
	"bot_api_microservice/models"
	"fmt"
	"time"
)

type extra struct {
	Amount int64
	When   time.Time
}

func Compute(currentBalance int64, items []models.BudgetItem, horizons []int, extraExpense *extra) models.ForecastResult {
	if len(horizons) == 0 {
		horizons = []int{1, 3, 6}
	}
	now := time.Now()
	points := make([]models.ForecastPoint, 0, len(horizons))
	for _, m := range horizons {
		end := now.AddDate(0, m, 0)
		income, expense := project(items, now, end)
		if extraExpense != nil && !extraExpense.When.After(end) {
			expense += extraExpense.Amount
		}
		points = append(points, models.ForecastPoint{
			MonthsAhead: m,
			Income:      income,
			Expense:     expense,
			Balance:     currentBalance + income - expense,
		})
	}
	return models.ForecastResult{
		CurrentBalance: currentBalance,
		Points:         points,
	}
}

func Scenario(currentBalance int64, items []models.BudgetItem, req models.ScenarioRequest) models.ScenarioResult {
	horizons := req.MonthsAhead
	if len(horizons) == 0 {
		horizons = []int{1, 3, 6}
	}
	when := time.Now()
	if req.When != nil {
		when = *req.When
	}
	baseline := Compute(currentBalance, items, horizons, nil)
	with := Compute(currentBalance, items, horizons, &extra{Amount: req.Amount, When: when})
	delta := map[string]int64{}
	for i, p := range with.Points {
		delta[monthsKey(p.MonthsAhead)] = p.Balance - baseline.Points[i].Balance
	}
	return models.ScenarioResult{
		Name:           req.Name,
		Amount:         req.Amount,
		Baseline:       baseline,
		WithPurchase:   with,
		DeltaByHorizon: delta,
	}
}

func project(items []models.BudgetItem, from, to time.Time) (income, expense int64) {
	for _, it := range items {
		if !it.IsActive {
			continue
		}
		switch it.Kind {
		case models.BudgetPlannedIncome:
			income += monthlyOccurrences(it.DayOfMonth, from, to) * it.Amount
		case models.BudgetRecurringExpense:
			expense += monthlyOccurrences(it.DayOfMonth, from, to) * it.Amount
		case models.BudgetLargePurchase:
			if it.DueDate != nil {
				d := *it.DueDate
				if !d.Before(from) && d.Before(to) {
					expense += it.Amount
				}
			}
		}
	}
	return
}

func monthlyOccurrences(dayOfMonth *int, from, to time.Time) int64 {
	day := 1
	if dayOfMonth != nil && *dayOfMonth >= 1 && *dayOfMonth <= 28 {
		day = *dayOfMonth
	}
	var count int64
	cursor := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
	for cursor.Before(to) {
		cand := time.Date(cursor.Year(), cursor.Month(), day, 12, 0, 0, 0, from.Location())
		if !cand.Before(from) && cand.Before(to) {
			count++
		}
		cursor = cursor.AddDate(0, 1, 0)
		if count > 120 {
			break
		}
	}
	return count
}

func monthsKey(m int) string {
	return fmt.Sprintf("%dm", m)
}
