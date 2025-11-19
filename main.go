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

	r.GET("/expenses/:user_id/:description", handlers.GetExpenses)

}
