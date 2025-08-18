package main

import (
	"log"
	"net/http"
	"os"

	"spectechle/backend/api"
	"spectechle/backend/database"
	"spectechle/backend/scraper"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default configuration")
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize scraper
	scraperService := scraper.NewScraper()

	// Initialize API handlers
	apiHandler := api.NewHandler(db, scraperService)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	})

	// Apply CORS middleware
	router.Use(func(c *gin.Context) {
		corsMiddleware.ServeHTTP(c.Writer, c.Request, func(w http.ResponseWriter, r *http.Request) {
			c.Next()
		})
	})

	// Serve static files for frontend
	router.Static("/static", "../frontend/static")
	router.LoadHTMLGlob("../frontend/templates/*")

	// API routes
	apiGroup := router.Group("/api")
	{
		// Search endpoints
		apiGroup.POST("/search", apiHandler.Search)
		apiGroup.GET("/search/:id", apiHandler.GetSearchResult)
		
		// Scraping endpoints
		apiGroup.POST("/scrape", apiHandler.ScrapeContent)
		apiGroup.GET("/scrape/status/:id", apiHandler.GetScrapeStatus)
		apiGroup.GET("/scrape/jobs", apiHandler.GetActiveScrapingJobs)
		
		// Article management
		apiGroup.GET("/articles", apiHandler.GetArticles)
		apiGroup.GET("/articles/:id", apiHandler.GetArticle)
		apiGroup.POST("/articles/:id/process-nlp", apiHandler.ProcessArticleWithNLP)
		apiGroup.DELETE("/articles/:id", apiHandler.DeleteArticle)
		
		// Categories
		apiGroup.GET("/categories", apiHandler.GetCategories)
		
		// Health check
		apiGroup.GET("/health", apiHandler.HealthCheck)
		
		// Test endpoints
		apiGroup.GET("/test-urls", apiHandler.TestSearchURLs)
	}

	// Frontend routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Spectechle - Smart Tech Search",
		})
	})

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Starting Spectechle backend server on port %s", port)
	log.Printf("📊 API available at http://localhost:%s/api", port)
	log.Printf("🌐 Web UI available at http://localhost:%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}


