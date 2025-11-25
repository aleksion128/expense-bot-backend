package main

import (
	"bot_api_microservice/database"
	"bot_api_microservice/handlers"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connecton to database
	DB := database.ConnectDB()
	if DB == nil {
		log.Fatal("Database connection is nil")
	}
	defer DB.Close()

	// Initialize gin
	r := gin.Default()

	apiExpenses := r.Group("api/v1/expenses")
	{
		apiExpenses.POST("/add/:telegram_id", handlers.AddExpense)
		apiExpenses.GET("/find_by_descr/:telegram_id/:description", handlers.GetExpenses)
		apiExpenses.GET("/today/:telegram_id", handlers.GetLastFiveExpenses)
	}

	r.Run(":8443")
}
