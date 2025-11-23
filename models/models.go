package models

import "time"

type Expense struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Date        time.Time `json:"created_at"`
}
