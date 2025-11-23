package main

import (
	"bot_api_microservice/database"
	"bot_api_microservice/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connecton to database
	db := database.ConnectDB()
	defer db.Close()

	// Initialize gin
	r := gin.Default()

	r.GET("/expenses/find_by_descr/:user_id/:description", handlers.GetExpenses)
	r.POST("/expenses/add/:user_id", handlers.AddExpense)

	r.Run(":8443")
}
