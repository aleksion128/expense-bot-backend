package storage

import "database/sql"

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
