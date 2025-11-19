package handlers

import (
	"bot_api_microservice/database"
	"bot_api_microservice/storage"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetExpenses(c *gin.Context) {
	user_id, err := strconv.Atoi(c.Param("user_id"))
	description := c.Param("description")
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user_id"})
	}

	expenses, err := storage.GetExpensesByUser(database.ConnectDB(), user_id, description)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error getting expenses"})
	}
	c.JSON(200, expenses)
}
