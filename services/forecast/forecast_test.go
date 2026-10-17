package forecast

import (
	"bot_api_microservice/models"
	"testing"
	"time"
)

func TestMonthlyOccurrences(t *testing.T) {
	day := 5
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	n := monthlyOccurrences(&day, from, to)
	if n != 3 {
		t.Fatalf("expected 3, got %d", n)
	}
}

func TestComputeForecast(t *testing.T) {
	day := 1
	items := []models.BudgetItem{
		{Kind: models.BudgetPlannedIncome, Amount: 100_000_00, DayOfMonth: &day, IsActive: true},
		{Kind: models.BudgetRecurringExpense, Amount: 40_000_00, DayOfMonth: &day, IsActive: true},
	}
	res := Compute(50_000_00, items, []int{1}, nil)
	if len(res.Points) != 1 {
		t.Fatal("expected 1 point")
	}
	// ~1 month: +100k -40k = +60k → balance 110k
	if res.Points[0].Balance != 110_000_00 {
		t.Fatalf("balance=%d", res.Points[0].Balance)
	}
}

func TestScenario(t *testing.T) {
	day := 1
	items := []models.BudgetItem{
		{Kind: models.BudgetPlannedIncome, Amount: 100_000_00, DayOfMonth: &day, IsActive: true},
		{Kind: models.BudgetRecurringExpense, Amount: 40_000_00, DayOfMonth: &day, IsActive: true},
	}
	res := Scenario(50_000_00, items, models.ScenarioRequest{
		Name:   "Объектив",
		Amount: 20_000_00,
	})
	if res.DeltaByHorizon["1m"] != -20_000_00 {
		t.Fatalf("delta=%v", res.DeltaByHorizon)
	}
}
