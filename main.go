package main

import (
	"bot_api_microservice/database"
	"bot_api_microservice/handlers"
	"bot_api_microservice/middlewares"
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
	apiExpenses.Use(middlewares.SetDB(DB))
	{
		apiExpenses.POST("/add/:telegram_id", handlers.AddExpense)
		apiExpenses.GET("/find_by_descr/:telegram_id/:description", handlers.GetExpenses)
		apiExpenses.GET("/today/:telegram_id", handlers.GetLastFiveExpenses)
		apiExpenses.POST("/update/:telegram_id/:expense_id", handlers.UpdateExpense)
		apiExpenses.POST("/delete/:telegram_id/:expense_id", handlers.DeleteExpense)
	}
	apiIncomes := r.Group("api/v1/incomes")
	apiIncomes.Use(middlewares.SetDB(DB))
	{
		apiIncomes.POST("/add", handlers.AddIncome)
		apiIncomes.GET("/get_all", handlers.GetIncomes)
		apiIncomes.POST("/update", handlers.UpdateIncome)
		apiIncomes.POST("/delete", handlers.DeleteIncome)
	}
	apiGoals := r.Group("api/v1/goals")
	apiGoals.Use(middlewares.SetDB(DB))
	{
		apiGoals.POST("/add/", handlers.AddGoal)
		apiGoals.GET("/get_all", handlers.GetGoals)
		apiGoals.POST("/update", handlers.UpdateGoal)
		apiGoals.POST("/delete", handlers.DeleteGoal)
	}

	r.Run(":8443")
}
