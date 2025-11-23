package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() *sql.DB {
	connstr := fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PASSWORD"))
	db, err := sql.Open("postgres", connstr)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	return db
}
