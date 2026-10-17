package handlers

import (
	"bot_api_microservice/middlewares"
	"bot_api_microservice/models"
	"bot_api_microservice/services/forecast"
	"bot_api_microservice/storage"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAnalytics(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	to := from.AddDate(0, 1, 0)
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t
		}
	}

	accounts, err := storage.ListAccounts(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	total, _ := storage.TotalBalance(db, uid)
	income, _ := storage.SumByType(db, uid, models.TxIncome, from, to)
	expense, _ := storage.SumByType(db, uid, models.TxExpense, from, to)
	byExp, _ := storage.SumByCategory(db, uid, models.TxExpense, from, to)
	byInc, _ := storage.SumByCategory(db, uid, models.TxIncome, from, to)
	if byExp == nil {
		byExp = map[string]int64{}
	}
	if byInc == nil {
		byInc = map[string]int64{}
	}
	if accounts == nil {
		accounts = []models.Account{}
	}

	c.JSON(http.StatusOK, models.AnalyticsSummary{
		TotalBalance:      total,
		IncomePeriod:      income,
		ExpensePeriod:     expense,
		NetPeriod:         income - expense,
		ByCategoryExpense: byExp,
		ByCategoryIncome:  byInc,
		Accounts:          accounts,
	})
}

func GetForecast(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	total, err := storage.TotalBalance(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items, err := storage.ListBudgetItems(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, forecast.Compute(total, items, []int{1, 3, 6}, nil))
}

func PostScenario(c *gin.Context) {
	db := middlewares.GetDB(c)
	uid := middlewares.GetUserID(c)
	var req models.ScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive (kopecks)"})
		return
	}
	if req.Name == "" {
		req.Name = "Покупка"
	}
	total, err := storage.TotalBalance(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items, err := storage.ListBudgetItems(db, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, forecast.Scenario(total, items, req))
}
