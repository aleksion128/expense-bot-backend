package models

import "time"

type Expense struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Category    string    `json:"category"`
	Date        time.Time `json:"created_at"`
}

type User struct {
	ID         int       `json:"id"`
	TelegramID string    `json:"telegram_id"`
	Username   string    `json:"username"`
	CreatedAt  time.Time `json:"created_at"`
}

type Income struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Month       int    `json:"month"`
	Year        int    `json:"year"`
}

type Goal struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Month       int    `json:"month"`
	Year        int    `json:"year"`
}
