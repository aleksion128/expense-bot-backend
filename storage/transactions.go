package storage

import (
	"bot_api_microservice/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type TxFilter struct {
	From      *time.Time
	To        *time.Time
	Type      *models.TxType
	AccountID *int
	Limit     int
	Offset    int
}

func ListTransactions(db *sql.DB, userID int, f TxFilter) ([]models.Transaction, error) {
	var (
		conds = []string{"user_id=$1"}
		args  = []any{userID}
		n     = 2
	)
	if f.From != nil {
		conds = append(conds, fmt.Sprintf("occurred_at >= $%d", n))
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		conds = append(conds, fmt.Sprintf("occurred_at < $%d", n))
		args = append(args, *f.To)
		n++
	}
	if f.Type != nil {
		conds = append(conds, fmt.Sprintf("type = $%d", n))
		args = append(args, *f.Type)
		n++
	}
	if f.AccountID != nil {
		conds = append(conds, fmt.Sprintf("account_id = $%d", n))
		args = append(args, *f.AccountID)
		n++
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 500
	}
	if limit > 20000 {
		limit = 20000
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}
	q := fmt.Sprintf(`
		SELECT id, user_id, account_id, transfer_account_id, type, amount, description, category,
		       occurred_at, external_id, import_batch_id, created_at
		FROM transactions WHERE %s
		ORDER BY occurred_at DESC, id DESC
		LIMIT %d OFFSET %d`, strings.Join(conds, " AND "), limit, offset)

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTransactions(rows)
}

func CreateTransaction(db *sql.DB, t models.Transaction) (models.Transaction, error) {
	tx, err := db.Begin()
	if err != nil {
		return t, err
	}
	defer func() { _ = tx.Rollback() }()

	err = tx.QueryRow(`
		INSERT INTO transactions
			(user_id, account_id, transfer_account_id, type, amount, description, category, occurred_at, external_id, import_batch_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, user_id, account_id, transfer_account_id, type, amount, description, category, occurred_at, external_id, import_batch_id, created_at`,
		t.UserID, t.AccountID, t.TransferAccountID, t.Type, t.Amount, t.Description, t.Category, t.OccurredAt, t.ExternalID, t.ImportBatchID,
	).Scan(&t.ID, &t.UserID, &t.AccountID, &t.TransferAccountID, &t.Type, &t.Amount, &t.Description, &t.Category, &t.OccurredAt, &t.ExternalID, &t.ImportBatchID, &t.CreatedAt)
	if err != nil {
		return t, err
	}

	if err := applyBalance(tx, t); err != nil {
		return t, err
	}
	if err := tx.Commit(); err != nil {
		return t, err
	}
	return t, nil
}

func DeleteTransactions(db *sql.DB, userID int, ids []int) (int, error) {
	unique := make([]int, 0, len(ids))
	seen := map[int]struct{}{}
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return 0, nil
	}
	if len(unique) > 1000 {
		return 0, fmt.Errorf("too many ids")
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	deleted := 0
	for _, id := range unique {
		t, err := lockTransaction(tx, userID, id)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return 0, err
		}
		if err := revertBalance(tx, t); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(`DELETE FROM transactions WHERE id=$1 AND user_id=$2`, id, userID); err != nil {
			return 0, err
		}
		deleted++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return deleted, nil
}

func lockTransaction(tx *sql.Tx, userID, id int) (models.Transaction, error) {
	var t models.Transaction
	err := tx.QueryRow(`
		SELECT id, user_id, account_id, transfer_account_id, type, amount, description, category,
		       occurred_at, external_id, import_batch_id, created_at
		FROM transactions WHERE id=$1 AND user_id=$2 FOR UPDATE`, id, userID).Scan(
		&t.ID, &t.UserID, &t.AccountID, &t.TransferAccountID, &t.Type, &t.Amount,
		&t.Description, &t.Category, &t.OccurredAt, &t.ExternalID, &t.ImportBatchID, &t.CreatedAt,
	)
	return t, err
}

func revertBalance(tx *sql.Tx, t models.Transaction) error {
	switch t.Type {
	case models.TxIncome:
		_, err := tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND user_id=$3`, t.Amount, t.AccountID, t.UserID)
		return err
	case models.TxExpense:
		_, err := tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2 AND user_id=$3`, t.Amount, t.AccountID, t.UserID)
		return err
	case models.TxTransfer:
		if t.TransferAccountID == nil {
			return fmt.Errorf("transfer_account_id required")
		}
		if _, err := tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2 AND user_id=$3`, t.Amount, t.AccountID, t.UserID); err != nil {
			return err
		}
		_, err := tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND user_id=$3`, t.Amount, *t.TransferAccountID, t.UserID)
		return err
	default:
		return fmt.Errorf("unknown tx type: %s", t.Type)
	}
}

func applyBalance(tx *sql.Tx, t models.Transaction) error {
	switch t.Type {
	case models.TxIncome:
		_, err := tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2`, t.Amount, t.AccountID)
		return err
	case models.TxExpense:
		_, err := tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id=$2`, t.Amount, t.AccountID)
		return err
	case models.TxTransfer:
		if t.TransferAccountID == nil {
			return fmt.Errorf("transfer_account_id required")
		}
		if _, err := tx.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id=$2`, t.Amount, t.AccountID); err != nil {
			return err
		}
		_, err := tx.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id=$2`, t.Amount, *t.TransferAccountID)
		return err
	default:
		return fmt.Errorf("unknown tx type: %s", t.Type)
	}
}

func InsertImportedTransaction(tx *sql.Tx, t models.Transaction) (bool, error) {
	res, err := tx.Exec(`
		INSERT INTO transactions
			(user_id, account_id, transfer_account_id, type, amount, description, category, occurred_at, external_id, import_batch_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (user_id, external_id) DO NOTHING`,
		t.UserID, t.AccountID, t.TransferAccountID, t.Type, t.Amount, t.Description, t.Category, t.OccurredAt, t.ExternalID, t.ImportBatchID,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, nil
	}
	if err := applyBalance(tx, t); err != nil {
		return false, err
	}
	return true, nil
}

func SumByType(db *sql.DB, userID int, txType models.TxType, from, to time.Time) (int64, error) {
	var sum sql.NullInt64
	err := db.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM transactions
		WHERE user_id=$1 AND type=$2 AND occurred_at >= $3 AND occurred_at < $4`,
		userID, txType, from, to,
	).Scan(&sum)
	return sum.Int64, err
}

func SumByCategory(db *sql.DB, userID int, txType models.TxType, from, to time.Time) (map[string]int64, error) {
	rows, err := db.Query(`
		SELECT category, COALESCE(SUM(amount),0) FROM transactions
		WHERE user_id=$1 AND type=$2 AND occurred_at >= $3 AND occurred_at < $4
		GROUP BY category ORDER BY 2 DESC`,
		userID, txType, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var cat string
		var sum int64
		if err := rows.Scan(&cat, &sum); err != nil {
			return nil, err
		}
		out[cat] = sum
	}
	return out, rows.Err()
}

func scanTransactions(rows *sql.Rows) ([]models.Transaction, error) {
	var out []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.AccountID, &t.TransferAccountID, &t.Type, &t.Amount,
			&t.Description, &t.Category, &t.OccurredAt, &t.ExternalID, &t.ImportBatchID, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
