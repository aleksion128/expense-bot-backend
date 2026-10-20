package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/storage"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func ListTransactions(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	f := storage.TxFilter{}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Offset = n
		}
	}
	if v := c.Query("type"); v != "" {
		t := models.TxType(v)
		f.Type = &t
	}
	if v := c.Query("account_id"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.AccountID = &n
		}
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.From = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			f.To = &t
		}
	}
	items, err := storage.ListTransactions(db, uid, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []models.Transaction{}
	}
	c.JSON(http.StatusOK, items)
}

func CreateTransaction(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var t models.Transaction
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	t.UserID = uid
	if t.OccurredAt.IsZero() {
		t.OccurredAt = time.Now()
	}
	if t.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive (kopecks)"})
		return
	}
	if t.Category == "" {
		t.Category = "Другое"
	}
	created, err := storage.CreateTransaction(db, t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func DeleteTransaction(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	n, err := storage.DeleteTransactions(db, uid, []int{id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "операция не найдена"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": n})
}

func DeleteTransactionsBulk(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var body struct {
		IDs []int `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if len(body.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	n, err := storage.DeleteTransactions(db, uid, body.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "deleted": n})
}
