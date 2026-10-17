package storage

import (
	"bot_api_microservice/models"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

func ListBudgetItems(db *sql.DB, userID int) ([]models.BudgetItem, error) {
	rows, err := db.Query(`
		SELECT id, user_id, kind, name, amount, category, day_of_month, due_date, is_active, created_at
		FROM budget_items WHERE user_id=$1 AND is_active=TRUE ORDER BY kind, id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.BudgetItem
	for rows.Next() {
		var b models.BudgetItem
		var due sql.NullTime
		if err := rows.Scan(&b.ID, &b.UserID, &b.Kind, &b.Name, &b.Amount, &b.Category, &b.DayOfMonth, &due, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		if due.Valid {
			t := due.Time
			b.DueDate = &t
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func CreateBudgetItem(db *sql.DB, b models.BudgetItem) (models.BudgetItem, error) {
	var due any
	if b.DueDate != nil {
		due = b.DueDate.Format("2006-01-02")
	}
	var dueOut sql.NullTime
	err := db.QueryRow(`
		INSERT INTO budget_items (user_id, kind, name, amount, category, day_of_month, due_date)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, user_id, kind, name, amount, category, day_of_month, due_date, is_active, created_at`,
		b.UserID, b.Kind, b.Name, b.Amount, b.Category, b.DayOfMonth, due,
	).Scan(&b.ID, &b.UserID, &b.Kind, &b.Name, &b.Amount, &b.Category, &b.DayOfMonth, &dueOut, &b.IsActive, &b.CreatedAt)
	if err != nil {
		return b, err
	}
	if dueOut.Valid {
		t := dueOut.Time
		b.DueDate = &t
	}
	return b, nil
}

func UpdateBudgetItem(db *sql.DB, b models.BudgetItem) (models.BudgetItem, error) {
	var due any
	if b.DueDate != nil {
		due = b.DueDate.Format("2006-01-02")
	}
	var dueOut sql.NullTime
	err := db.QueryRow(`
		UPDATE budget_items SET kind=$1, name=$2, amount=$3, category=$4, day_of_month=$5, due_date=$6, is_active=$7
		WHERE id=$8 AND user_id=$9
		RETURNING id, user_id, kind, name, amount, category, day_of_month, due_date, is_active, created_at`,
		b.Kind, b.Name, b.Amount, b.Category, b.DayOfMonth, due, b.IsActive, b.ID, b.UserID,
	).Scan(&b.ID, &b.UserID, &b.Kind, &b.Name, &b.Amount, &b.Category, &b.DayOfMonth, &dueOut, &b.IsActive, &b.CreatedAt)
	if err != nil {
		return b, err
	}
	if dueOut.Valid {
		t := dueOut.Time
		b.DueDate = &t
	} else {
		b.DueDate = nil
	}
	return b, nil
}

func DeleteBudgetItem(db *sql.DB, userID, id int) error {
	_, err := db.Exec(`UPDATE budget_items SET is_active=FALSE WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func ListGoals(db *sql.DB, userID int) ([]models.Goal, error) {
	rows, err := db.Query(`
		SELECT id, user_id, name, target_amount, current_amount, deadline, created_at
		FROM goals WHERE user_id=$1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Goal
	for rows.Next() {
		var g models.Goal
		var deadline sql.NullTime
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &deadline, &g.CreatedAt); err != nil {
			return nil, err
		}
		if deadline.Valid {
			t := deadline.Time
			g.Deadline = &t
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func CreateGoal(db *sql.DB, g models.Goal) (models.Goal, error) {
	var deadline any
	if g.Deadline != nil {
		deadline = g.Deadline.Format("2006-01-02")
	}
	var dl sql.NullTime
	err := db.QueryRow(`
		INSERT INTO goals (user_id, name, target_amount, current_amount, deadline)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, user_id, name, target_amount, current_amount, deadline, created_at`,
		g.UserID, g.Name, g.TargetAmount, g.CurrentAmount, deadline,
	).Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &dl, &g.CreatedAt)
	if err != nil {
		return g, err
	}
	if dl.Valid {
		t := dl.Time
		g.Deadline = &t
	}
	return g, nil
}

func UpdateGoal(db *sql.DB, g models.Goal) (models.Goal, error) {
	var deadline any
	if g.Deadline != nil {
		deadline = g.Deadline.Format("2006-01-02")
	}
	var dl sql.NullTime
	err := db.QueryRow(`
		UPDATE goals SET name=$1, target_amount=$2, current_amount=$3, deadline=$4
		WHERE id=$5 AND user_id=$6
		RETURNING id, user_id, name, target_amount, current_amount, deadline, created_at`,
		g.Name, g.TargetAmount, g.CurrentAmount, deadline, g.ID, g.UserID,
	).Scan(&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &dl, &g.CreatedAt)
	if err != nil {
		return g, err
	}
	if dl.Valid {
		t := dl.Time
		g.Deadline = &t
	} else {
		g.Deadline = nil
	}
	return g, nil
}

func DeleteGoal(db *sql.DB, userID, id int) error {
	_, err := db.Exec(`DELETE FROM goals WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func CreateImportBatch(db *sql.DB, b models.ImportBatch) (models.ImportBatch, error) {
	err := db.QueryRow(`
		INSERT INTO import_batches (user_id, account_id, filename, bank, status, rows_total, rows_imported, rows_skipped)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, user_id, account_id, filename, bank, status, rows_total, rows_imported, rows_skipped, created_at`,
		b.UserID, b.AccountID, b.Filename, b.Bank, b.Status, b.RowsTotal, b.RowsImported, b.RowsSkipped,
	).Scan(&b.ID, &b.UserID, &b.AccountID, &b.Filename, &b.Bank, &b.Status, &b.RowsTotal, &b.RowsImported, &b.RowsSkipped, &b.CreatedAt)
	return b, err
}

func UpdateImportBatchStats(tx *sql.Tx, id, total, imported, skipped int, status string) error {
	_, err := tx.Exec(`
		UPDATE import_batches SET rows_total=$1, rows_imported=$2, rows_skipped=$3, status=$4 WHERE id=$5`,
		total, imported, skipped, status, id)
	return err
}

func ListImportBatches(db *sql.DB, userID int) ([]models.ImportBatch, error) {
	rows, err := db.Query(`
		SELECT id, user_id, account_id, filename, bank, status, rows_total, rows_imported, rows_skipped, created_at
		FROM import_batches WHERE user_id=$1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ImportBatch
	for rows.Next() {
		var b models.ImportBatch
		if err := rows.Scan(&b.ID, &b.UserID, &b.AccountID, &b.Filename, &b.Bank, &b.Status, &b.RowsTotal, &b.RowsImported, &b.RowsSkipped, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func ListCategories(db *sql.DB, userID int) ([]models.Category, error) {
	rows, err := db.Query(`
		SELECT id, user_id, name, kind, icon, keywords
		FROM categories
		WHERE user_id IS NULL OR user_id=$1
		ORDER BY kind, name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Category
	for rows.Next() {
		var c models.Category
		var keywords pq.StringArray
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Kind, &c.Icon, &keywords); err != nil {
			return nil, err
		}
		c.Keywords = []string(keywords)
		out = append(out, c)
	}
	return out, rows.Err()
}

// helpers for optional date parsing from JSON strings
func ParseDatePtr(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}
