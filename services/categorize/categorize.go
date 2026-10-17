package categorize

import (
	"bot_api_microservice/models"
	"strings"
)

func Categorize(description string, txType models.TxType, cats []models.Category) string {
	if txType == models.TxTransfer {
		return "Перевод себе"
	}
	desc := normalize(description)
	kind := string(txType)
	bestName := defaultName(txType)
	bestScore := 0
	for _, c := range cats {
		if c.Kind != kind {
			continue
		}
		for _, kw := range c.Keywords {
			kw = normalize(kw)
			if kw == "" {
				continue
			}
			if strings.Contains(desc, kw) {
				score := len([]rune(kw))
				if score > bestScore {
					bestScore = score
					bestName = c.Name
				}
			}
		}
	}
	return bestName
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "ё", "е")
	return s
}

func DetectTransfer(description string, ownAccountHints []string) bool {
	desc := normalize(description)
	markers := []string{
		"перевод между счетами",
		"между своими счетами",
		"перевод себе",
		"внутренний перевод",
		"own account",
		"me2me",
	}
	for _, m := range markers {
		if strings.Contains(desc, m) {
			return true
		}
	}
	for _, hint := range ownAccountHints {
		h := normalize(hint)
		if len(h) >= 4 && strings.Contains(desc, h) {
			return true
		}
	}
	return false
}

func defaultName(txType models.TxType) string {
	switch txType {
	case models.TxIncome:
		return "Прочий доход"
	case models.TxTransfer:
		return "Перевод себе"
	default:
		return "Другое"
	}
}
