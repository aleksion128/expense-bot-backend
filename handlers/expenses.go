package handlers

import (
	"bot_api_microservice/database"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetExpenses(c *gin.Context) {
	user_id, err := strconv.Atoi(c.Param("user_id"))
	description := c.Param("description")
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user id"})
		return
	}

	expenses, err := storage.GetExpensesByUserAndDescription(database.ConnectDB(), user_id, description)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error getting expenses"})
		return
	}
	c.JSON(200, expenses)
}

// Adding expense using json
func AddExpense(c *gin.Context) {
	var userExpense models.Expense
	if err := c.ShouldBindBodyWithJSON(&userExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing json data while getting new expense"})
		return
	}
	if err := storage.AddExpense(database.ConnectDB(), userExpense); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding new expense"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Expense added successfully"})
}
