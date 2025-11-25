package storage

import (
	"bot_api_microservice/database"
	"bot_api_microservice/models"
	"crypto/aes"
	"database/sql"
	"os"
	"strings"
	"time"
)

// Encryption and decryption descriptions
func Encrypt(text string) (string, error) {
	key := []byte(os.Getenv("KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, len(plaintext))

	block.Encrypt(ciphertext, plaintext)

	return string(ciphertext), nil
}

func Decrypt(text string) (string, error) {
	key := []byte(os.Getenv("KEY"))
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, len(plaintext))

	block.Decrypt(ciphertext, plaintext)

	return string(ciphertext), nil
}

// EXPENSES

// Find expenses by keyword
func GetExpensesByUserAndDescription(db *sql.DB, telegram_id string, description string) ([]models.Expense, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	user_id, err := GetUserIDByTelegramID(database.ConnectDB(), telegram_id)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT amount, description,category FROM expenses WHERE user_id=$1`, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var expensesArr []models.Expense
	for rows.Next() {
		var amount int64
		var description string
		var category string
		err = rows.Scan(&amount, &description, &category)
		if err != nil {
			return nil, err
		}
		descriptionDecr, err := Decrypt(description)
		if err != nil {
			return nil, err
		}
		var expense models.Expense
		expense.Amount = amount
		expense.Description = descriptionDecr
		expense.Category = category
		expensesArr = append(expensesArr, expense)
	}
	return FindExpensesByDescription(expensesArr, description)
}

func FindExpensesByDescription(expenses []models.Expense, description string) ([]models.Expense, error) {
	var filteredExpenses []models.Expense
	for _, expense := range expenses {
		if strings.Contains(expense.Description, description) {
			filteredExpenses = append(filteredExpenses, expense)
		}
	}
	return filteredExpenses, nil
}

// Get five last expenses for current day
func GetLastFiveExpenses(db *sql.DB, telegram_id string) ([]models.Expense, error) {
	user_id, err := GetUserIDByTelegramID(database.ConnectDB(), telegram_id)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT amount, description,category FROM expenses WHERE user_id=$1 AND created_at=$2 ORDER BY created_at DESC LIMIT 5`, user_id, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var expensesArr []models.Expense
	for rows.Next() {
		var amount int64
		var description string
		var category string
		err = rows.Scan(&amount, &description, &category)
		if err != nil {
			return nil, err
		}
		descriptionDecr, err := Decrypt(description)
		if err != nil {
			return nil, err
		}
		var expense models.Expense
		expense.Amount = amount
		expense.Description = descriptionDecr
		expense.Category = category
		expensesArr = append(expensesArr, expense)
	}
	return expensesArr, nil
}

func UpdateExpenseByUser(db *sql.DB, expense models.Expense) error {
	rows, err := db.Query(`UPDATE expenses SET (amount, description) VALUES ($1,$2) WHERE user_id=$3`, expense.Amount, expense.Description, expense.UserID)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func AddExpense(db *sql.DB, expense models.Expense) error {
	description, err := Encrypt(expense.Description)
	if err != nil {
		return err
	}
	rows, err := db.Query(`INSERT INTO expenses (user_id, amount, description,created_at) VALUES ($1,$2,$3,$4)`, expense.UserID, expense.Amount, description, expense.Date)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

func DeleteExpense(db *sql.DB, telegram_id string, expense_id int) error {
	user_id, err := GetUserIDByTelegramID(database.ConnectDB(), telegram_id)
	if err != nil {
		return err
	}
	rows, err := db.Query(`DELETE FROM expenses WHERE user_id=$1 AND id=$2`, user_id, expense_id)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

// USERS

func GetUserIDByTelegramID(db *sql.DB, telegram_id string) (int, error) {
	row := db.QueryRow(`SELECT id FROM users WHERE telegram_id=$1`, telegram_id)
	var user_id int
	err := row.Scan(&user_id)
	if err != nil {
		return -1, err
	}
	return user_id, nil
}
