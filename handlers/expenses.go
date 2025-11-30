package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetExpenses(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	telegram_id := c.Param("telegram_id")
	description := c.Param("description")
	expenses, err := storage.GetExpensesByUserAndDescription(sqlDB, telegram_id, description)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error getting expenses"})
		return
	}
	c.JSON(200, expenses)
}

// Adding expense using json
func AddExpense(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var userExpense models.Expense
	if err := c.ShouldBindBodyWithJSON(&userExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	id, err := storage.AddExpense(sqlDB, userExpense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"expense_id": id})
}

func DeleteExpense(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	telegram_id := c.Param("telegram_id")
	expense_id, err := strconv.Atoi(c.Param("expense_id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid expense id"})
		return
	}
	if err := storage.DeleteExpense(sqlDB, telegram_id, expense_id); err != nil {
		c.JSON(500, gin.H{"error": "Error deleting expense"})
		return
	}
	c.JSON(200, gin.H{"message": "Expense deleted successfully"})
}

func GetLastFiveExpenses(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	telegram_id := c.Param("telegram_id")
	expenses, err := storage.GetLastFiveExpenses(sqlDB, telegram_id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, expenses)
}

func UpdateExpense(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var userExpense models.Expense
	if err := c.ShouldBindBodyWithJSON(&userExpense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := storage.UpdateExpenseByUser(sqlDB, userExpense)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"expense_id": id})
}
