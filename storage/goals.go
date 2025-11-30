package storage

import (
	"bot_api_microservice/models"
	"database/sql"
)

func GetGoals(db *sql.DB, user_id int) ([]models.Goal, error) {
	var goals []models.Goal
	rows, err := db.Query("SELECT * FROM goals WHERE user_id = $1", user_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var goal models.Goal
		var description string
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.Amount, description, &goal.Month, &goal.Year)
		if err != nil {
			return nil, err
		}
		goal.Description, err = Decrypt(description)
		if err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, nil
}

func AddGoal(db *sql.DB, goal models.Goal) error {
	description, err := Encrypt(goal.Description)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO goals (user_id, amount, description, month, year) VALUES ($1, $2, $3, $4, $5)", goal.UserID, goal.Amount, description, goal.Month, goal.Year)
	return err
}

func UpdateGoal(db *sql.DB, goal models.Goal) error {
	description, err := Encrypt(goal.Description)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE goals SET user_id=$1, amount=$2, description=$3, month=$4, year=$5 WHERE id=$6`, goal.UserID, goal.Amount, description, goal.Month, goal.Year, goal.ID)
	return err
}

func DeleteGoal(db *sql.DB, goal models.Goal) error {
	_, err := db.Exec("DELETE FROM goals WHERE id=$1", goal.ID)
	return err
}
