package importsvc

import (
	"os"
	"testing"
)

func TestParseTinkoffSample(t *testing.T) {
	data, err := os.ReadFile("../../sample_data/tinkoff_sample.csv")
	if err != nil {
		t.Fatal(err)
	}
	res, err := ParseFile("tinkoff_sample.csv", data, "tinkoff")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 6 {
		t.Fatalf("rows=%d", len(res.Rows))
	}
	// salary positive
	var foundSalary bool
	for _, r := range res.Rows {
		if r.Amount == 15000000 {
			foundSalary = true
		}
	}
	if !foundSalary {
		t.Fatal("salary row not found")
	}
}

func TestParseAmount(t *testing.T) {
	v, err := parseAmount("-1 250,50")
	if err != nil {
		t.Fatal(err)
	}
	if v != -125050 {
		t.Fatalf("got %d", v)
	}
}
