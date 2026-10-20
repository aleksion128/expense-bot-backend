package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
	var body struct {
		Name     string             `json:"name"`
		Type     models.AccountType `json:"type"`
		Balance  *int64             `json:"balance"`
		Currency string             `json:"currency"`
		BankName *string            `json:"bank_name"`
		IsActive *bool              `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	switch body.Type {
	case "", models.AccountCard, models.AccountSavings, models.AccountBrokerage, models.AccountCash:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account type"})
		return
	}
	existing, err := storage.GetAccount(db, uid, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "счёт не найден"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	a := *existing
	if strings.TrimSpace(body.Name) != "" {
		a.Name = strings.TrimSpace(body.Name)
	}
	if body.Type != "" {
		a.Type = body.Type
	}
	if body.Balance != nil {
		a.Balance = *body.Balance
	}
	if body.Currency != "" {
		a.Currency = body.Currency
	}
	if body.BankName != nil {
		bank := strings.TrimSpace(*body.BankName)
		if bank == "" {
			a.BankName = nil
		} else {
			a.BankName = &bank
		}
	}
	if body.IsActive != nil {
		a.IsActive = *body.IsActive
	}
	updated, err := storage.UpdateAccount(db, a)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
