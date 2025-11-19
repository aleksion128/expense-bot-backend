package models

type Expense struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Date        string `json:"date"`
}
