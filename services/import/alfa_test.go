package importsvc

import (
	"os"
	"testing"
)

func TestParseAlfaSampleSkipsPreamble(t *testing.T) {
	data, err := os.ReadFile("../../sample_data/alfa_sample.csv")
	if err != nil {
		t.Fatal(err)
	}
	res, err := ParseFile("alfa_sample.csv", data, "auto")
	if err != nil {
		t.Fatal(err)
	}
	if res.Bank != "alfa" {
		t.Fatalf("bank=%s", res.Bank)
	}
	// HOLD row must be skipped → 5 ops
	if len(res.Rows) != 5 {
		t.Fatalf("rows=%d %#v", len(res.Rows), res.Rows)
	}
	var salary, shop bool
	for _, r := range res.Rows {
		if r.Amount == 15000000 {
			salary = true
			if r.Category != "Зарплата" {
				t.Fatalf("salary category=%q", r.Category)
			}
		}
		if r.Amount == -125050 {
			shop = true
		}
	}
	if !salary || !shop {
		t.Fatalf("expected salary and shop rows, salary=%v shop=%v", salary, shop)
	}
}

func TestParseAlfaExplicitBank(t *testing.T) {
	data, err := os.ReadFile("../../sample_data/alfa_sample.csv")
	if err != nil {
		t.Fatal(err)
	}
	res, err := ParseFile("statement.csv", data, "alfa")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 5 {
		t.Fatalf("rows=%d", len(res.Rows))
	}
}

func TestFindHeaderAmongPreamble(t *testing.T) {
	text := "a\nb\nc\nДата операции;Сумма в валюте счета;Описание\n01.01.2026;-10;x\n"
	rows, err := parseAlfaCSV(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Amount != -1000 {
		t.Fatalf("%#v", rows)
	}
}
