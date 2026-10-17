package middlewares

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const DemoUserID = 1

func SetDB(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

// DemoAuth sets user_id from X-User-Id header or falls back to demo user.
func DemoAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := DemoUserID
		if raw := c.GetHeader("X-User-Id"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				uid = parsed
			}
		}
		c.Set("user_id", uid)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int {
	v, ok := c.Get("user_id")
	if !ok {
		return DemoUserID
	}
	id, ok := v.(int)
	if !ok {
		return DemoUserID
	}
	return id
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,X-User-Id")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
