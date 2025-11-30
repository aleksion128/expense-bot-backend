package storage

import (
	"bot_api_microservice/models"
	"database/sql"
)

// Add income for user
func AddIncome(db *sql.DB, income models.Income) int {
	description, err := Encrypt(income.Description)
	if err != nil {
		return -1
	}
	rows, err := db.Query(`INSERT INTO incomes (user_id, amount, description, month, year) VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		income.UserID,
		income.Amount,
		description,
		income.Month,
		income.Year,
	)
	if err != nil {
		return -1
	}
	var id int
	if err := rows.Scan(&id); err != nil {
		return -1
	}
	defer rows.Close()
	return id
}

// Delete income for user
func DeleteIncome(db *sql.DB, income models.Income) error {
	description, err := Encrypt(income.Description)
	if err != nil {
		return err
	}
	rows, err := db.Query(`DELETE FROM incomes WHERE user_id=$1 AND month=$2 AND year=$3 AND description=$4`, income.UserID, income.Month, income.Year, description)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}

// Get incomes for user
func GetIncomes(db *sql.DB, income models.Income) ([]models.Income, error) {
	rows, err := db.Query(`SELECT * FROM incomes WHERE user_id=$1 AND month=$2 AND year=$3`, income.UserID, income.Month, income.Year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var incomesArr []models.Income
	for rows.Next() {
		var incomeInRow models.Income
		if err := rows.Scan(&incomeInRow); err != nil {
			return nil, err
		}
		description, err := Decrypt(incomeInRow.Description)
		if err != nil {
			return nil, err
		}
		incomeInRow.Description = description
		incomesArr = append(incomesArr, incomeInRow)
	}
	return incomesArr, nil
}

// Update income for user
func UpdateIncome(db *sql.DB, income models.Income) error {
	description, err := Encrypt(income.Description)
	if err != nil {
		return err
	}
	rows, err := db.Query(`UPDATE incomes SET amount=$1, description=$2 WHERE user_id=$3 AND month=$4 AND year=$5`,
		income.Amount,
		description,
		income.UserID,
		income.Month,
		income.Year,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	return nil
}
