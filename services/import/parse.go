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
	"unicode"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type ParsedRow struct {
	OccurredAt  time.Time
	Amount      int64 // signed: negative = expense, positive = income
	Description string
	Category    string
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
		text, err := xlsxToDelimited(data, ';')
		if err != nil {
			return ParseResult{}, err
		}
		rows, err := parseByBank(text, bank)
		return ParseResult{Bank: normalizeBankName(bank), Rows: rows}, err
	}

	text, err := decodeText(data)
	if err != nil {
		return ParseResult{}, err
	}
	rows, err := parseByBank(text, bank)
	return ParseResult{Bank: normalizeBankName(bank), Rows: rows}, err
}

func parseByBank(text, bank string) ([]ParsedRow, error) {
	switch bank {
	case "tinkoff", "tbank":
		return parseTinkoffCSV(text)
	case "alfa", "alfabank", "альфа", "альфа-банк", "альфабанк":
		return parseAlfaCSV(text)
	case "sber", "sberbank":
		return parseTableCSV(text, "sber")
	default:
		// Alfa-like files with preamble often come as "auto/generic"
		if looksLikeAlfa(text) {
			return parseAlfaCSV(text)
		}
		return parseTableCSV(text, "generic")
	}
}

func normalizeBankName(bank string) string {
	switch strings.ToLower(bank) {
	case "tinkoff", "tbank":
		return "tinkoff"
	case "alfa", "alfabank", "альфа", "альфа-банк", "альфабанк":
		return "alfa"
	case "sber", "sberbank":
		return "sber"
	default:
		return bank
	}
}

func detectBank(filename string, data []byte) string {
	sample := data
	if len(sample) > 4000 {
		sample = sample[:4000]
	}
	s := strings.ToLower(filename + " " + string(sample))
	s = strings.ReplaceAll(s, "\u00a0", " ")
	switch {
	case strings.Contains(s, "альфа") || strings.Contains(s, "alfa") || strings.Contains(s, "сумма в валюте счета") || strings.Contains(s, "дата проводки"):
		return "alfa"
	case strings.Contains(s, "tinkoff") || strings.Contains(s, "тинькофф") || strings.Contains(s, "t-bank") || strings.Contains(s, "номер карты"):
		return "tinkoff"
	case strings.Contains(s, "sber") || strings.Contains(s, "сбер"):
		return "sber"
	default:
		return "generic"
	}
}

func looksLikeAlfa(text string) bool {
	low := strings.ToLower(strings.ReplaceAll(text, "\u00a0", " "))
	return strings.Contains(low, "сумма в валюте счета") ||
		(strings.Contains(low, "дата операции") && strings.Contains(low, "дата проводки") && strings.Contains(low, "категория"))
}

func decodeText(data []byte) (string, error) {
	if utf8.Valid(data) {
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
	headerIdx := findHeaderIndex(records, "дата операции", "сумма операции")
	if headerIdx < 0 {
		headerIdx = 0
	}
	if headerIdx >= len(records) {
		return nil, fmt.Errorf("empty tinkoff csv")
	}
	header := normalizeHeader(records[headerIdx])
	idx := headerIndexFunc(header)
	iDate := idx("дата операции", "дата")
	iAmount := idx("сумма операции", "сумма")
	iDesc := idx("описание", "описание операции")
	iStatus := idx("статус")
	iCard := idx("номер карты")
	if iDate < 0 || iAmount < 0 {
		return parseTableCSV(text, "tinkoff")
	}
	return parseRows(records[headerIdx+1:], "tinkoff", iDate, iAmount, iDesc, iStatus, -1, iCard)
}

// parseAlfaCSV handles Alfa exports where account/person preamble comes first,
// and the real table starts later (often ~row 20) with:
// Дата операции | Дата проводки | Код | Категория | Описание | Сумма в валюте счета | Статус
func parseAlfaCSV(text string) ([]ParsedRow, error) {
	records, err := readLooseTable(text)
	if err != nil {
		return nil, err
	}
	headerIdx := findHeaderIndex(records, "дата операции", "сумма в валюте счета")
	if headerIdx < 0 {
		headerIdx = findHeaderIndex(records, "дата операции", "сумма")
	}
	if headerIdx < 0 {
		return nil, fmt.Errorf("alfa: cannot find table header (ожидали строку с «Дата операции» и «Сумма в валюте счета»)")
	}
	header := normalizeHeader(records[headerIdx])
	idx := headerIndexFunc(header)
	iDate := idx("дата операции", "дата")
	iAmount := idx("сумма в валюте счета", "сумма в валюте счёта", "сумма операции", "сумма")
	iDesc := idx("описание", "описание операции", "назначение")
	iStatus := idx("статус")
	iCat := idx("категория")
	iCode := idx("код")
	iPosting := idx("дата проводки")
	if iDate < 0 || iAmount < 0 {
		return nil, fmt.Errorf("alfa: header found but date/amount columns missing: %v", records[headerIdx])
	}

	var out []ParsedRow
	for _, rec := range records[headerIdx+1:] {
		if isEmptyRecord(rec) {
			continue
		}
		get := func(i int) string {
			if i >= 0 && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		// stop on totals / footer
		joined := strings.ToLower(strings.Join(rec, " "))
		if strings.Contains(joined, "итого") || strings.Contains(joined, "всего по") {
			break
		}
		status := get(iStatus)
		if status != "" && !isSuccessStatus(status) {
			continue
		}
		occurred, err := parseRUDate(get(iDate))
		if err != nil {
			// try posting date as fallback
			occurred, err = parseRUDate(get(iPosting))
			if err != nil {
				continue
			}
		}
		amount, err := parseAmount(get(iAmount))
		if err != nil || amount == 0 {
			continue
		}
		desc := get(iDesc)
		cat := get(iCat)
		if desc == "" {
			desc = cat
		}
		ext := hashID("alfa", get(iDate), get(iAmount), desc, get(iCode), status)
		out = append(out, ParsedRow{
			OccurredAt:  occurred,
			Amount:      amount,
			Description: desc,
			Category:    cat,
			Status:      status,
			ExternalID:  ext,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("alfa: header found at row %d, but no operations parsed", headerIdx+1)
	}
	return out, nil
}

func parseTableCSV(text, bank string) ([]ParsedRow, error) {
	records, err := readLooseTable(text)
	if err != nil {
		return nil, err
	}
	headerIdx := findHeaderIndex(records, "дата операции", "сумма")
	if headerIdx < 0 {
		headerIdx = findHeaderIndex(records, "date", "amount")
	}
	if headerIdx < 0 {
		headerIdx = 0
	}
	if headerIdx >= len(records) {
		return nil, fmt.Errorf("empty csv")
	}
	header := normalizeHeader(records[headerIdx])
	idx := headerIndexFunc(header)
	iDate := idx("date", "дата", "дата операции", "дата проводки", "operation date")
	iAmount := idx("amount", "сумма", "сумма операции", "сумма в валюте счета", "сумма в валюте счёта", "sum")
	iDesc := idx("description", "описание", "назначение", "details", "comment", "комментарий")
	iStatus := idx("статус", "status")
	iCat := idx("категория", "category")
	if iDate < 0 || iAmount < 0 {
		return nil, fmt.Errorf("cannot detect date/amount columns; got header: %v", records[headerIdx])
	}
	return parseRows(records[headerIdx+1:], bank, iDate, iAmount, iDesc, iStatus, iCat, -1)
}

func parseRows(records [][]string, bank string, iDate, iAmount, iDesc, iStatus, iCat, iExtra int) ([]ParsedRow, error) {
	var out []ParsedRow
	for _, rec := range records {
		if isEmptyRecord(rec) {
			continue
		}
		get := func(i int) string {
			if i >= 0 && i < len(rec) {
				return strings.TrimSpace(rec[i])
			}
			return ""
		}
		status := get(iStatus)
		if status != "" && !isSuccessStatus(status) {
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
		cat := get(iCat)
		ext := hashID(bank, get(iDate), get(iAmount), desc, get(iExtra), status)
		out = append(out, ParsedRow{
			OccurredAt:  occurred,
			Amount:      amount,
			Description: desc,
			Category:    cat,
			Status:      status,
			ExternalID:  ext,
		})
	}
	return out, nil
}

func isSuccessStatus(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	s = strings.ReplaceAll(s, "ё", "е")
	switch s {
	case "ok", "успешно", "исполнено", "выполнен", "выполнена", "completed", "done", "проведено", "проведена":
		return true
	default:
		// Alfa sometimes leaves status empty for posted ops
		return s == ""
	}
}

func findHeaderIndex(records [][]string, required ...string) int {
	need := make([]string, len(required))
	for i, r := range required {
		need[i] = normalizeKey(r)
	}
	limit := len(records)
	if limit > 80 {
		limit = 80
	}
	for i := 0; i < limit; i++ {
		header := normalizeHeader(records[i])
		ok := true
		for _, n := range need {
			if !headerHas(header, n) {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	// softer: all required as substrings across joined header cells
	for i := 0; i < limit; i++ {
		joined := normalizeKey(strings.Join(records[i], " "))
		ok := true
		for _, n := range need {
			if !strings.Contains(joined, n) {
				ok = false
				break
			}
		}
		if ok {
			return i
		}
	}
	return -1
}

func headerHas(header map[string]int, want string) bool {
	if _, ok := header[want]; ok {
		return true
	}
	for k := range header {
		if strings.Contains(k, want) || strings.Contains(want, k) {
			return true
		}
	}
	return false
}

func headerIndexFunc(header map[string]int) func(names ...string) int {
	return func(names ...string) int {
		for _, n := range names {
			n = normalizeKey(n)
			if i, ok := header[n]; ok {
				return i
			}
			for k, i := range header {
				if k == n || strings.Contains(k, n) {
					return i
				}
			}
		}
		return -1
	}
}

func xlsxToDelimited(data []byte, sep rune) (string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("xlsx has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for i, row := range rows {
		for j, cell := range row {
			if j > 0 {
				b.WriteRune(sep)
			}
			b.WriteString(cell)
		}
		if i < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func readLooseTable(text string) ([][]string, error) {
	// Prefer the separator that best matches the eventual header line.
	candidates := []rune{'\t', ';', ','}
	var best [][]string
	bestScore := -1
	var bestErr error
	for _, sep := range candidates {
		records, err := readCSV(text, sep)
		if err != nil {
			bestErr = err
			continue
		}
		score := 0
		idx := findHeaderIndex(records, "дата операции")
		if idx >= 0 {
			score = 100 + len(records[idx])
			h := normalizeHeader(records[idx])
			if headerHas(h, normalizeKey("сумма в валюте счета")) || headerHas(h, "сумма") {
				score += 50
			}
		} else if len(records) > 0 {
			score = len(records[0])
		}
		if score > bestScore {
			bestScore = score
			best = records
		}
	}
	if best == nil {
		if bestErr != nil {
			return nil, bestErr
		}
		return nil, fmt.Errorf("cannot parse table")
	}
	return best, nil
}

func readCSV(text string, sep rune) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(text))
	r.Comma = sep
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	return r.ReadAll()
}

func normalizeHeader(row []string) map[string]int {
	out := map[string]int{}
	for i, h := range row {
		key := normalizeKey(h)
		if key == "" {
			continue
		}
		// first wins — keep leftmost meaningful column
		if _, exists := out[key]; !exists {
			out[key] = i
		}
	}
	return out
}

func normalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "\ufeff")
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "ё", "е")
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func isEmptyRecord(rec []string) bool {
	for _, c := range rec {
		if strings.TrimSpace(c) != "" {
			return false
		}
	}
	return true
}

func parseRUDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
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
