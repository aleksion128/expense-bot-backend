package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddIncome(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var income models.Income
	if err := c.ShouldBindBodyWithJSON(&income); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	err := storage.AddIncome(sqlDB, income)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add income to database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Income added successfully"})
}

func UpdateIncome(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var income models.Income
	err := c.ShouldBindBodyWithJSON(&income)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check JSON body"})
		return
	}
	err = storage.UpdateIncome(sqlDB, income)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update income"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Income updated succesfully"})
}

func GetIncomes(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var income models.Income
	err := c.ShouldBindBodyWithJSON(&income)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check JSON body"})
		return
	}
	incomes, err := storage.GetIncomes(sqlDB, income)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get incomes"})
		return
	}
	c.JSON(http.StatusOK, incomes)
}

func DeleteIncome(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var income models.Income
	err := c.ShouldBindBodyWithJSON(&income)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check JSON body"})
		return
	}
	err = storage.DeleteIncome(sqlDB, income)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete income"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Income deleted succesfully"})
}
