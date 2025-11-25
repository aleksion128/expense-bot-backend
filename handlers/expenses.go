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
	telegram_id := c.Param("telegram_id")
	description := c.Param("description")

	expenses, err := storage.GetExpensesByUserAndDescription(database.ConnectDB(), telegram_id, description)
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

func DeleteExpense(c *gin.Context) {
	telegram_id := c.Param("telegram_id")
	expense_id, err := strconv.Atoi(c.Param("expense_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid expense id"})
		return
	}
	if err := storage.DeleteExpense(database.ConnectDB(), telegram_id, expense_id); err != nil {
		c.JSON(500, gin.H{"error": "Error deleting expense"})
		return
	}
	c.JSON(200, gin.H{"message": "Expense deleted successfully"})
}

func GetLastFiveExpenses(c *gin.Context) {
	telegram_id := c.Param("telegram_id")
	expenses, err := storage.GetLastFiveExpenses(database.ConnectDB(), telegram_id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error getting expenses"})
		return
	}
	c.JSON(200, expenses)
}
