package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddGoal(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var goal models.Goal
	if err := c.ShouldBindBodyWithJSON(&goal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON"})
		return
	}
	err := storage.AddGoal(sqlDB, goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add goal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Goal added successfully"})
}

func UpdateGoal(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var goal models.Goal
	err := c.ShouldBindBodyWithJSON(&goal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check JSON body"})
		return
	}
	err = storage.UpdateGoal(sqlDB, goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update goal"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Goal updated successfully"})
}

func GetGoals(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	telegram_id := c.Param("telegram_id")
	user_id, err := storage.GetUserIDByTelegramID(sqlDB, telegram_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}
	goals, err := storage.GetGoals(sqlDB, user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get goals"})
		return
	}
	c.JSON(http.StatusOK, goals)
}

func DeleteGoal(c *gin.Context) {
	sqlDB := middlewares.GetDB(c)
	if sqlDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}
	var goal models.Goal
	err := c.ShouldBindBodyWithJSON(&goal)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Check JSON body"})
		return
	}
	err = storage.DeleteGoal(sqlDB, goal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete goal"})
		return
	}
}
