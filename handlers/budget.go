package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListBudget(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	items, err := storage.ListBudgetItems(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.BudgetItem{}
	}
	c.JSON(http.StatusOK, items)
}

func CreateBudget(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var b models.BudgetItem
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	b.UserID = uid
	created, err := storage.CreateBudgetItem(db, b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func UpdateBudget(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var b models.BudgetItem
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	b.ID = id
	b.UserID = uid
	updated, err := storage.UpdateBudgetItem(db, b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteBudget(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := storage.DeleteBudgetItem(db, uid, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListGoalsHandler(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	items, err := storage.ListGoals(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.Goal{}
	}
	c.JSON(http.StatusOK, items)
}

func CreateGoalHandler(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var g models.Goal
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	g.UserID = uid
	created, err := storage.CreateGoal(db, g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func UpdateGoalHandler(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var g models.Goal
	if err := c.ShouldBindJSON(&g); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	g.ID = id
	g.UserID = uid
	updated, err := storage.UpdateGoal(db, g)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func DeleteGoalHandler(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := storage.DeleteGoal(db, uid, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ListCategories(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	items, err := storage.ListCategories(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.Category{}
	}
	c.JSON(http.StatusOK, items)
}
