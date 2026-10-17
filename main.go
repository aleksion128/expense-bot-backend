package main

import (
	"bot_api_microservice/database"
	"bot_api_microservice/handlers"
	"bot_api_microservice/middlewares"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Database connection is nil")
	}
	defer db.Close()

	migDir := envOr("MIGRATIONS_DIR", "migrations")
	if abs, err := filepath.Abs(migDir); err == nil {
		migDir = abs
	}
	if err := database.RunMigrations(db, migDir); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	r := gin.Default()
	r.Use(middlewares.CORS())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.Use(middlewares.SetDB(db), middlewares.DemoAuth())
	{
		api.GET("/accounts", handlers.ListAccounts)
		api.POST("/accounts", handlers.CreateAccount)
		api.PUT("/accounts/:id", handlers.UpdateAccount)

		api.GET("/transactions", handlers.ListTransactions)
		api.POST("/transactions", handlers.CreateTransaction)

		api.POST("/import", handlers.ImportStatement)
		api.GET("/imports", handlers.ListImports)

		api.GET("/analytics", handlers.GetAnalytics)
		api.GET("/forecast", handlers.GetForecast)
		api.POST("/scenarios", handlers.PostScenario)

		api.GET("/budget", handlers.ListBudget)
		api.POST("/budget", handlers.CreateBudget)
		api.PUT("/budget/:id", handlers.UpdateBudget)
		api.DELETE("/budget/:id", handlers.DeleteBudget)

		api.GET("/goals", handlers.ListGoalsHandler)
		api.POST("/goals", handlers.CreateGoalHandler)
		api.PUT("/goals/:id", handlers.UpdateGoalHandler)
		api.DELETE("/goals/:id", handlers.DeleteGoalHandler)

		api.GET("/categories", handlers.ListCategories)
	}

	// Serve frontend in production if built
	staticDir := envOr("STATIC_DIR", "frontend/dist")
	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		r.Static("/assets", filepath.Join(staticDir, "assets"))
		r.StaticFile("/", filepath.Join(staticDir, "index.html"))
		r.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(staticDir, "index.html"))
		})
	}

	addr := envOr("ADDR", ":8443")
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func envOr(k, v string) string {
	if x := os.Getenv(k); x != "" {
		return x
	}
	return v
}
