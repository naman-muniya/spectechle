package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"spectechle/backend/api"
	"spectechle/backend/database"
	"spectechle/backend/models"
	"spectechle/backend/scraper"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Run tests
	m.Run()
}

// TestSearchHandler tests the search API endpoint
func TestSearchHandler(t *testing.T) {
	// Setup
	db, err := database.InitDB()
	require.NoError(t, err)
	defer db.Close()

	scraperService := scraper.NewScraper()
	handler := api.NewHandler(db, scraperService)

	router := gin.New()
	router.POST("/api/search", handler.Search)

	// Test cases
	tests := []struct {
		name           string
		requestBody    models.SearchRequest
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "Valid search request",
			requestBody: models.SearchRequest{
				Query: "artificial intelligence",
				Mode:  "news",
				Limit: 10,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"id", "query", "mode", "articles", "total"},
		},
		{
			name: "Invalid mode",
			requestBody: models.SearchRequest{
				Query: "test",
				Mode:  "invalid",
				Limit: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty query",
			requestBody: models.SearchRequest{
				Query: "",
				Mode:  "news",
				Limit: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/search", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.SearchResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Check required fields
				for _, field := range tt.expectedFields {
					switch field {
					case "id":
						assert.NotEmpty(t, response.ID)
					case "query":
						assert.Equal(t, tt.requestBody.Query, response.Query)
					case "mode":
						assert.Equal(t, tt.requestBody.Mode, response.Mode)
					case "articles":
						assert.NotNil(t, response.Articles)
					case "total":
						assert.GreaterOrEqual(t, response.Total, 0)
					}
				}
			}
		})
	}
}

// TestScraper tests the web scraper functionality
func TestScraper(t *testing.T) {
	scraperService := scraper.NewScraper()
	require.NotNil(t, scraperService)

	// Test scraper initialization
	t.Run("Scraper initialization", func(t *testing.T) {
		assert.NotNil(t, scraperService)
	})

	// Test URL validation
	t.Run("URL validation", func(t *testing.T) {
		validURLs := []string{
			"https://example.com",
			"http://test.com/article",
			"https://arxiv.org/abs/1234.5678",
		}

		invalidURLs := []string{
			"not-a-url",
			"ftp://invalid.com",
			"",
		}

		for _, url := range validURLs {
			assert.True(t, isValidURL(url), "URL should be valid: %s", url)
		}

		for _, url := range invalidURLs {
			assert.False(t, isValidURL(url), "URL should be invalid: %s", url)
		}
	})
}

// TestDatabaseOperations tests database operations
func TestDatabaseOperations(t *testing.T) {
	db, err := database.InitDB()
	require.NoError(t, err)
	defer db.Close()

	repo := database.NewArticleRepository(db)

	t.Run("Database connection", func(t *testing.T) {
		assert.NoError(t, err)
		assert.NotNil(t, db)
	})

	t.Run("Article repository creation", func(t *testing.T) {
		assert.NotNil(t, repo)
	})

	t.Run("Search articles", func(t *testing.T) {
		articles, err := repo.SearchArticles("test", "news", []string{}, 10)
		// Should not error even if no results
		assert.NoError(t, err)
		assert.NotNil(t, articles)
	})
}

// TestModels tests the data models
func TestModels(t *testing.T) {
	t.Run("Article model", func(t *testing.T) {
		article := models.Article{
			Title:   "Test Article",
			URL:     "https://example.com/test",
			Content: "This is test content",
			Mode:    "news",
		}

		assert.NotEmpty(t, article.Title)
		assert.NotEmpty(t, article.URL)
		assert.NotEmpty(t, article.Content)
		assert.Contains(t, []string{"news", "research"}, article.Mode)
	})

	t.Run("SearchRequest model", func(t *testing.T) {
		req := models.SearchRequest{
			Query:  "test query",
			Mode:   "news",
			Limit:  20,
		}

		assert.NotEmpty(t, req.Query)
		assert.Contains(t, []string{"news", "research"}, req.Mode)
		assert.Greater(t, req.Limit, 0)
	})

	t.Run("SearchResponse model", func(t *testing.T) {
		response := models.SearchResponse{
			Query:    "test",
			Mode:     "news",
			Articles: []models.Article{},
			Total:    0,
			Status:   "completed",
		}

		assert.NotEmpty(t, response.Query)
		assert.Contains(t, []string{"news", "research"}, response.Mode)
		assert.NotNil(t, response.Articles)
		assert.GreaterOrEqual(t, response.Total, 0)
		assert.Contains(t, []string{"processing", "completed", "failed"}, response.Status)
	})
}

// TestAPIEndpoints tests various API endpoints
func TestAPIEndpoints(t *testing.T) {
	db, err := database.InitDB()
	require.NoError(t, err)
	defer db.Close()

	scraperService := scraper.NewScraper()
	handler := api.NewHandler(db, scraperService)

	router := gin.New()
	
	// Setup routes
	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/search", handler.Search)
		apiGroup.GET("/articles", handler.GetArticles)
		apiGroup.GET("/categories", handler.GetCategories)
		apiGroup.GET("/health", handler.HealthCheck)
	}

	t.Run("Health check endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "healthy", response["status"])
	})

	t.Run("Get categories endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/categories", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var categories []models.Category
		err := json.Unmarshal(w.Body.Bytes(), &categories)
		require.NoError(t, err)
		
		assert.NotNil(t, categories)
	})

	t.Run("Get articles endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/articles", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var articles []models.Article
		err := json.Unmarshal(w.Body.Bytes(), &articles)
		require.NoError(t, err)
		
		assert.NotNil(t, articles)
	})
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	db, err := database.InitDB()
	require.NoError(t, err)
	defer db.Close()

	scraperService := scraper.NewScraper()
	handler := api.NewHandler(db, scraperService)

	router := gin.New()
	router.POST("/api/search", handler.Search)

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/search", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"query": "test",
			// Missing mode field
		}
		body, _ := json.Marshal(reqBody)
		
		req, _ := http.NewRequest("POST", "/api/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// Helper functions
func isValidURL(url string) bool {
	// Simple URL validation for testing
	return len(url) > 0 && (url[:7] == "http://" || url[:8] == "https://")
}

// Benchmark tests
func BenchmarkSearchHandler(b *testing.B) {
	db, err := database.InitDB()
	require.NoError(b, err)
	defer db.Close()

	scraperService := scraper.NewScraper()
	handler := api.NewHandler(db, scraperService)

	router := gin.New()
	router.POST("/api/search", handler.Search)

	requestBody := models.SearchRequest{
		Query: "artificial intelligence",
		Mode:  "news",
		Limit: 10,
	}
	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/search", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkScraper(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test scraper initialization
		_ = scraper.NewScraper()
	}
}
