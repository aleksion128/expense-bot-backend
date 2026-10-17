package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListAccounts(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	items, err := storage.ListAccounts(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.Account{}
	}
	c.JSON(http.StatusOK, items)
}

func CreateAccount(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var a models.Account
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	a.UserID = uid
	if a.Currency == "" {
		a.Currency = "RUB"
	}
	created, err := storage.CreateAccount(db, a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func UpdateAccount(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var a models.Account
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	a.ID = id
	a.UserID = uid
	updated, err := storage.UpdateAccount(db, a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
