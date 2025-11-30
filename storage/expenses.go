package storage

import (
	"bot_api_microservice/models"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"
)

// Encryption and decryption descriptions
func Encrypt(text string) (string, error) {
	keyHex := os.Getenv("KEY")
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(text), nil)
	combined := append(nonce, ciphertext...)
	return hex.EncodeToString(combined), nil
}

func Decrypt(text string) (string, error) {
	keyHex := os.Getenv("KEY")
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	combined, err := hex.DecodeString(text)
	if err != nil {
		return "", err
	}
	if len(combined) < aesgcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce := combined[:aesgcm.NonceSize()]
	ciphertext := combined[aesgcm.NonceSize():]
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EXPENSES

// Find expenses by keyword
func GetExpensesByUserAndDescription(db *sql.DB, telegram_id string, description string) ([]models.Expense, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	user_id, err := GetUserIDByTelegramID(db, telegram_id)
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
	user_id, err := GetUserIDByTelegramID(db, telegram_id)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT amount, description,category FROM expenses WHERE user_id=$1 AND
	 created_at>=$2 AND created_at<=$3 ORDER BY created_at DESC LIMIT 5`,
		user_id,
		time.Now().AddDate(0, 0, -1),
		time.Now(),
	)
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

func UpdateExpenseByUser(db *sql.DB, expense models.Expense) (int, error) {
	description, err := Encrypt(expense.Description)
	if err != nil {
		return -1, err
	}
	rows, err := db.Query(`UPDATE expenses SET amount=$1, description=$2, category=$3 WHERE user_id=$4 AND created_at=$5 RETURNING id`,
		expense.Amount,
		description,
		expense.Category,
		expense.UserID,
		expense.Date,
	)
	if err != nil {
		return -1, err
	}
	var id int
	rows.Scan(&id)
	defer rows.Close()
	return id, nil
}

func AddExpense(db *sql.DB, expense models.Expense) (int, error) {
	description, err := Encrypt(expense.Description)
	if err != nil {
		return -1, err
	}
	rows, err := db.Query(`INSERT INTO expenses (user_id, amount, description,created_at) VALUES ($1,$2,$3,$4) RETURNING id`, expense.UserID, expense.Amount, description, expense.Date)
	if err != nil {
		return -1, err
	}
	var id int
	rows.Scan(&id)
	defer rows.Close()
	return id, nil
}

func DeleteExpense(db *sql.DB, telegram_id string, expense_id int) error {
	user_id, err := GetUserIDByTelegramID(db, telegram_id)
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

