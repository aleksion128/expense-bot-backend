package categorize

import (
	"bot_api_microservice/models"
	"testing"
)

func TestCategorize(t *testing.T) {
	cats := []models.Category{
		{Name: "Продукты", Kind: "expense", Keywords: []string{"пятерочка", "магнит"}},
		{Name: "Зарплата", Kind: "income", Keywords: []string{"зарплата"}},
	}
	if got := Categorize("Пятёрочка Москва", models.TxExpense, cats); got != "Продукты" {
		t.Fatalf("got %s", got)
	}
	if got := Categorize("Магнит у дома", models.TxExpense, cats); got != "Продукты" {
		t.Fatalf("got %s", got)
	}
	if got := Categorize("Зарплата март", models.TxIncome, cats); got != "Зарплата" {
		t.Fatalf("got %s", got)
	}
	if got := Categorize("x", models.TxTransfer, cats); got != "Перевод себе" {
		t.Fatalf("got %s", got)
	}
}

func TestDetectTransfer(t *testing.T) {
	if !DetectTransfer("Перевод между своими счетами", nil) {
		t.Fatal("expected transfer")
	}
}
