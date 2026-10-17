package storage

import (
	"bot_api_microservice/models"
	"database/sql"
)

func ListAccounts(db *sql.DB, userID int) ([]models.Account, error) {
	rows, err := db.Query(`
		SELECT id, user_id, name, type, balance, currency, bank_name, is_active, created_at
		FROM accounts WHERE user_id=$1 AND is_active=TRUE ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.BankName, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func GetAccount(db *sql.DB, userID, id int) (*models.Account, error) {
	var a models.Account
	err := db.QueryRow(`
		SELECT id, user_id, name, type, balance, currency, bank_name, is_active, created_at
		FROM accounts WHERE id=$1 AND user_id=$2`, id, userID).
		Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.BankName, &a.IsActive, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateAccount(db *sql.DB, a models.Account) (models.Account, error) {
	err := db.QueryRow(`
		INSERT INTO accounts (user_id, name, type, balance, currency, bank_name)
		VALUES ($1,$2,$3,$4,COALESCE(NULLIF($5,''),'RUB'),$6)
		RETURNING id, user_id, name, type, balance, currency, bank_name, is_active, created_at`,
		a.UserID, a.Name, a.Type, a.Balance, a.Currency, a.BankName,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.BankName, &a.IsActive, &a.CreatedAt)
	return a, err
}

func UpdateAccount(db *sql.DB, a models.Account) (models.Account, error) {
	err := db.QueryRow(`
		UPDATE accounts SET name=$1, type=$2, balance=$3, currency=$4, bank_name=$5, is_active=$6
		WHERE id=$7 AND user_id=$8
		RETURNING id, user_id, name, type, balance, currency, bank_name, is_active, created_at`,
		a.Name, a.Type, a.Balance, a.Currency, a.BankName, a.IsActive, a.ID, a.UserID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.BankName, &a.IsActive, &a.CreatedAt)
	return a, err
}

func AdjustBalance(db *sql.DB, tx *sql.Tx, accountID int, delta int64) error {
	var exec interface {
		Exec(query string, args ...any) (sql.Result, error)
	} = db
	if tx != nil {
		exec = tx
	}
	_, err := exec.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2`, delta, accountID)
	return err
}

func TotalBalance(db *sql.DB, userID int) (int64, error) {
	var total sql.NullInt64
	err := db.QueryRow(`SELECT COALESCE(SUM(balance),0) FROM accounts WHERE user_id=$1 AND is_active=TRUE`, userID).Scan(&total)
	return total.Int64, err
}
