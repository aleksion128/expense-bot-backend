package importsvc

import (
	"bytes"
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type ParsedRow struct {
	OccurredAt  time.Time
	Amount      int64 // signed: negative = expense, positive = income
	Description string
	Status      string
	ExternalID  string
	Raw         map[string]string
}

type ParseResult struct {
	Bank string
	Rows []ParsedRow
}

func ParseFile(filename string, data []byte, bankHint string) (ParseResult, error) {
	lower := strings.ToLower(filename)
	bank := strings.ToLower(strings.TrimSpace(bankHint))
	if bank == "" || bank == "auto" {
		bank = detectBank(filename, data)
	}

	if strings.HasSuffix(lower, ".xlsx") || strings.HasSuffix(lower, ".xls") {
		rows, err := parseXLSX(data)
		return ParseResult{Bank: bank, Rows: rows}, err
	}

	text, err := decodeText(data)
	if err != nil {
		return ParseResult{}, err
	}

	switch bank {
	case "tinkoff", "tbank":
		rows, err := parseTinkoffCSV(text)
		return ParseResult{Bank: "tinkoff", Rows: rows}, err
	case "sber", "sberbank":
		rows, err := parseGenericCSV(text)
		return ParseResult{Bank: "sber", Rows: rows}, err
	default:
		rows, err := parseGenericCSV(text)
		if err != nil {
			return ParseResult{}, err
		}
		return ParseResult{Bank: "generic", Rows: rows}, nil
	}
}

func detectBank(filename string, data []byte) string {
	s := strings.ToLower(filename + " " + string(data[:min(len(data), 2000)]))
	switch {
	case strings.Contains(s, "tinkoff") || strings.Contains(s, "тинькофф") || strings.Contains(s, "t-bank") || strings.Contains(s, "номер карты"):
		return "tinkoff"
	case strings.Contains(s, "sber") || strings.Contains(s, "сбер"):
		return "sber"
	default:
		return "generic"
	}
}

func decodeText(data []byte) (string, error) {
	if utf8.Valid(data) {
		// strip BOM
		data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
		return string(data), nil
	}
	r := transform.NewReader(bytes.NewReader(data), charmap.Windows1251.NewDecoder())
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseTinkoffCSV(text string) ([]ParsedRow, error) {
	records, err := readCSV(text, ';')
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("empty tinkoff csv")
	}
	header := normalizeHeader(records[0])
	idx := func(names ...string) int {
		for _, n := range names {
			if i, ok := header[n]; ok {
				return i
			}
		}
		return -1
	}
	iDate := idx("дата операции", "дата")
	iAmount := idx("сумма операции", "сумма")
	iDesc := idx("описание", "описание операции")
	iStatus := idx("статус")
	if iDate < 0 || iAmount < 0 {
		return parseGenericCSV(text)
	}

	var out []ParsedRow
	for _, rec := range records[1:] {
		if len(rec) == 0 {
			continue
		}
		get := func(i int) string {
			if i >= 0 && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		status := get(iStatus)
		if status != "" && !strings.EqualFold(status, "OK") && !strings.EqualFold(status, "успешно") {
			continue
		}
		occurred, err := parseRUDate(get(iDate))
		if err != nil {
			continue
		}
		amount, err := parseAmount(get(iAmount))
		if err != nil || amount == 0 {
			continue
		}
		desc := get(iDesc)
		ext := hashID("tinkoff", get(iDate), get(iAmount), desc, get(idx("номер карты")))
		out = append(out, ParsedRow{
			OccurredAt:  occurred,
			Amount:      amount,
			Description: desc,
			Status:      status,
			ExternalID:  ext,
		})
	}
	return out, nil
}

func parseGenericCSV(text string) ([]ParsedRow, error) {
	sep := detectSep(text)
	records, err := readCSV(text, sep)
	if err != nil {
		return nil, err
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("empty csv")
	}
	header := normalizeHeader(records[0])
	idx := func(names ...string) int {
		for _, n := range names {
			if i, ok := header[n]; ok {
				return i
			}
		}
		return -1
	}
	iDate := idx("date", "дата", "дата операции", "дата проводки", "operation date")
	iAmount := idx("amount", "сумма", "сумма операции", "sum")
	iDesc := idx("description", "описание", "назначение", "details", "comment", "комментарий")
	if iDate < 0 || iAmount < 0 {
		return nil, fmt.Errorf("cannot detect date/amount columns; got header: %v", records[0])
	}

	var out []ParsedRow
	for _, rec := range records[1:] {
		get := func(i int) string {
			if i >= 0 && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		occurred, err := parseRUDate(get(iDate))
		if err != nil {
			continue
		}
		amount, err := parseAmount(get(iAmount))
		if err != nil || amount == 0 {
			continue
		}
		desc := get(iDesc)
		ext := hashID("generic", get(iDate), get(iAmount), desc)
		out = append(out, ParsedRow{
			OccurredAt:  occurred,
			Amount:      amount,
			Description: desc,
			ExternalID:  ext,
		})
	}
	return out, nil
}

func parseXLSX(data []byte) ([]ParsedRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("xlsx has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for i, row := range rows {
		for j, cell := range row {
			if j > 0 {
				b.WriteByte(';')
			}
			b.WriteString(cell)
		}
		if i < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	return parseGenericCSV(b.String())
}

func readCSV(text string, sep rune) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = sep
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

func detectSep(text string) rune {
	first := text
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		first = text[:i]
	}
	if strings.Count(first, ";") >= strings.Count(first, ",") {
		return ';'
	}
	return ','
}

func normalizeHeader(row []string) map[string]int {
	out := map[string]int{}
	for i, h := range row {
		key := strings.ToLower(strings.TrimSpace(h))
		key = strings.TrimPrefix(key, "\ufeff")
		out[key] = i
	}
	return out
}

func parseRUDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	layouts := []string{
		"02.01.2006 15:04:05",
		"02.01.2006 15:04",
		"02.01.2006",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("bad date: %s", s)
}

func parseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, "руб.", "")
	s = strings.ReplaceAll(s, "руб", "")
	s = strings.ReplaceAll(s, ",", ".")
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	cents := int64(f*100 + 0.5)
	if neg {
		cents = -cents
	}
	return cents, nil
}

func hashID(parts ...string) string {
	h := sha1.Sum([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
