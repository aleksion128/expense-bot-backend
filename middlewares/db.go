package middlewares

import (
	"bot_api_microservice/database"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func SetDB(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.ConnectDB()
		c.Set("db", db)
		c.Next()
	}
}

func GetDB(c *gin.Context) *sql.DB {
	db, ok := c.Get("db")
	if !ok {
		return nil
	}
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return nil
	}
	return sqlDB
}
