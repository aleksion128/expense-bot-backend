package storage

import (
	"bot_api_microservice/models"
	"crypto/aes"
	"database/sql"
	"os"
	"strings"
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

func GetExpensesByUser(db *sql.DB, user_id int, description string) ([]models.Expense, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT amount, description FROM expenses WHERE user_id=$1`, user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var expensesArr []models.Expense
	for rows.Next() {
		var amount int64
		var description string
		err = rows.Scan(&amount, &description)
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
		expensesArr = append(expensesArr, expense)
	}
	return expensesArr, nil
}

func FindExpensesByDescription(expenses []models.Expense, description string) ([]models.Expense, error) {
	var filteredExpenses []models.Expense
	for _, expense := range expenses {
		if strings.Contains(expense.Description, description) {
			filteredExpenses = append(filteredExpenses, expense)
		}

}
