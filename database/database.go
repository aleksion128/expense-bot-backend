package database

import (
	"database/sql"
	"log"
)

var DB *sql.DB

func ConnectDB() *sql.DB {
	connstr := "user=postgres dbname=postgres password=postgres sslmode=disable"
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	return db
}
