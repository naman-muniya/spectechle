package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"spectechle/backend/database"
	"spectechle/backend/models"
	"spectechle/backend/scraper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ScrapingJob represents an ongoing scraping job
type ScrapingJob struct {
	ID        string
	Query     string
	Mode      string
	Status    string // "running", "completed", "failed"
	StartTime time.Time
	Progress  int
	Total     int
	mu        sync.RWMutex
}

// Handler manages API endpoints
type Handler struct {
	db           *database.DB
	scraper      *scraper.Scraper
	repo         *database.ArticleRepository
	scrapingJobs map[string]*ScrapingJob // Track ongoing scraping jobs
	jobMutex     sync.RWMutex
}

// NewHandler creates a new API handler
func NewHandler(db *database.DB, scraper *scraper.Scraper) *Handler {
	return &Handler{
		db:           db,
		scraper:      scraper,
		repo:         database.NewArticleRepository(db),
		scrapingJobs: make(map[string]*ScrapingJob),
	}
}

// getOrCreateScrapingJob gets an existing job or creates a new one
func (h *Handler) getOrCreateScrapingJob(query, mode string) *ScrapingJob {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()
	
	// Create a unique job key
	jobKey := fmt.Sprintf("%s:%s", query, mode)
	
	// Check if job already exists and is still running
	if job, exists := h.scrapingJobs[jobKey]; exists {
		job.mu.RLock()
		status := job.Status
		job.mu.RUnlock()
		
		if status == "running" {
			log.Printf("🔄 Reusing existing scraping job for query: %s", query)
			return job
		}
	}
	
	// Create new job
	job := &ScrapingJob{
		ID:        uuid.New().String(),
		Query:     query,
		Mode:      mode,
		Status:    "running",
		StartTime: time.Now(),
		Progress:  0,
		Total:     0,
	}
	
	h.scrapingJobs[jobKey] = job
	log.Printf("🆕 Created new scraping job for query: %s", query)
	
	return job
}

// cleanupCompletedJobs removes completed jobs older than 1 hour
func (h *Handler) cleanupCompletedJobs() {
	h.jobMutex.Lock()
	defer h.jobMutex.Unlock()
	
	cutoff := time.Now().Add(-1 * time.Hour)
	for key, job := range h.scrapingJobs {
		job.mu.RLock()
		status := job.Status
		startTime := job.StartTime
		job.mu.RUnlock()
		
		if (status == "completed" || status == "failed") && startTime.Before(cutoff) {
			delete(h.scrapingJobs, key)
			log.Printf("🧹 Cleaned up completed job: %s", job.Query)
		}
	}
}

// Search handles search requests
func (h *Handler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate mode
	if req.Mode != "news" && req.Mode != "research" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'news' or 'research'"})
		return
	}

	// Set default limit with maximum cap
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50 // Cap maximum results to prevent resource abuse
	}

	// Try cache first for better performance
	cachedArticles, err := h.repo.GetCachedSearch(req.Query, req.Mode, req.Categories)
	if err != nil {
		log.Printf("Cache lookup error: %v", err)
	}
	
	var articles []models.Article
	cacheHit := false
	
	if len(cachedArticles) > 0 {
		articles = cachedArticles
		cacheHit = true
		log.Printf("Cache hit: found %d articles for query '%s'", len(articles), req.Query)
	} else {
		// Search existing articles in database
		articles, err = h.repo.SearchArticles(req.Query, req.Mode, req.Categories, req.Limit)
		if err != nil {
			log.Printf("Error searching articles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search articles"})
			return
		}
		log.Printf("Database search: found %d articles for query '%s'", len(articles), req.Query)
	}

	// Check if we have enough results
	hasEnoughResults := len(articles) >= req.Limit
	
	// If not enough results, trigger real-time scraping with limits
	if !hasEnoughResults && len(articles) < 10 { // Only scrape if we have very few results
		log.Printf("Limited results (%d < %d), triggering controlled real-time scraping", len(articles), req.Limit)
		
		// Get or create scraping job with timeout protection
		job := h.getOrCreateScrapingJob(req.Query, req.Mode)
		
		// Check if job is already running for too long (prevent infinite loops)
		job.mu.RLock()
		jobAge := time.Since(job.StartTime)
		job.mu.RUnlock()
		
		if jobAge < 5*time.Minute { // Only start new scraping if job is less than 5 minutes old
			// Start real-time scraping in background with limited scope
			go h.triggerRealTimeScraping(job, req.Query, req.Mode, req.Categories, req.Limit-len(articles))
		} else {
			log.Printf("⚠️ Skipping scraping - job already running for %v", jobAge)
		}
	} else if !cacheHit {
		// Cache the results for future searches (TTL: 2 hours for news, 24 hours for research)
		ttlHours := 2
		if req.Mode == "research" {
			ttlHours = 24
		}
		go func() {
			if err := h.repo.CacheSearch(req.Query, req.Mode, req.Categories, articles, ttlHours); err != nil {
				log.Printf("Failed to cache search results: %v", err)
			}
		}()
	}
	
	// Determine status based on results
	status := "completed"
	if !hasEnoughResults && len(articles) < 10 {
		status = "searching"
	}
	
	// Return results with appropriate status
	response := models.SearchResponse{
		ID:        uuid.New().String(),
		Query:     req.Query,
		Mode:      req.Mode,
		Articles:  articles,
		Total:     len(articles),
		CreatedAt: time.Now(),
		Status:    status,
		CacheHit:  cacheHit,
	}
	c.JSON(http.StatusOK, response)
}

// GetSearchResult retrieves a specific search result
func (h *Handler) GetSearchResult(c *gin.Context) {
	searchID := c.Param("id")
	if searchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search ID is required"})
		return
	}

	// TODO: Implement search session tracking
	c.JSON(http.StatusOK, gin.H{
		"id":     searchID,
		"status": "completed",
		"message": "Search result retrieval not yet implemented",
	})
}

// ScrapeContent handles content scraping requests
func (h *Handler) ScrapeContent(c *gin.Context) {
	var req models.ScrapeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate mode
	if req.Mode != "news" && req.Mode != "research" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be 'news' or 'research'"})
		return
	}

	// Create scrape job
	jobID := uuid.New().String()
	response := models.ScrapeResponse{
		ID:        jobID,
		URLs:      req.URLs,
		Status:    "processing",
		Progress:  0,
		Total:     len(req.URLs),
		CreatedAt: time.Now(),
	}

	// Start scraping in background
	go h.processScrapingJob(jobID, req)

	c.JSON(http.StatusAccepted, response)
}

// processScrapingJob processes a scraping job in the background
func (h *Handler) processScrapingJob(jobID string, req models.ScrapeRequest) {
	log.Printf("Starting scraping job %s for %d URLs", jobID, len(req.URLs))

	var results []models.Article
	for i, url := range req.URLs {
		log.Printf("Scraping URL %d/%d: %s", i+1, len(req.URLs), url)

		// Scrape the URL
		article, err := h.scraper.ScrapeURL(url, req.Mode)
		if err != nil {
			log.Printf("Error scraping %s: %v", url, err)
			continue
		}

		// TODO: Send to Python NLP service for classification and summarization
		// For now, use basic classification
		article.Category = h.basicClassify(article.Title + " " + article.Content)

		// Save to database
		if err := h.repo.CreateArticle(article); err != nil {
			log.Printf("Error saving article: %v", err)
			continue
		}

		results = append(results, *article)
	}

	log.Printf("Completed scraping job %s with %d results", jobID, len(results))
}

// triggerRealTimeScraping triggers real-time scraping for search queries
func (h *Handler) triggerRealTimeScraping(job *ScrapingJob, query, mode string, categories []string, limit int) {
	log.Printf("Starting real-time scraping for query: %s, mode: %s, limit: %d", query, mode, limit)
	
	// Add timeout protection
	timeout := time.After(3 * time.Minute) // 3-minute timeout
	done := make(chan bool, 1)
	
	go func() {
		defer func() {
			done <- true
			// Mark job as completed
			job.mu.Lock()
			job.Status = "completed"
			job.mu.Unlock()
		}()
		
		// Generate search URLs based on query and mode
		searchURLs := h.generateSearchURLs(query, mode, limit)
		
		if len(searchURLs) == 0 {
			log.Printf("No search URLs generated for query: %s", query)
			return
		}
		
		// Collect all article URLs from search results
		var allArticleURLs []string
		seenURLs := make(map[string]bool) // Track seen URLs to avoid duplicates
		
		// Scrape search results to get article URLs with timeout
		for _, searchURL := range searchURLs {
			select {
			case <-timeout:
				log.Printf("⚠️ Scraping timeout reached for query: %s", query)
				return
			default:
				articleURLs, err := h.scraper.ScrapeSearchResults(searchURL, mode, limit)
				if err != nil {
					log.Printf("Error scraping search results from %s: %v", searchURL, err)
					continue
				}
				
				// Deduplicate URLs
				for _, articleURL := range articleURLs {
					if !seenURLs[articleURL] {
						seenURLs[articleURL] = true
						allArticleURLs = append(allArticleURLs, articleURL)
					}
				}
			}
		}
		
		// Limit the number of URLs to process
		if len(allArticleURLs) > limit*2 {
			allArticleURLs = allArticleURLs[:limit*2]
		}
		
		log.Printf("Found %d unique articles to scrape", len(allArticleURLs))
		
		// Process articles with rate limiting
		processedCount := 0
		for _, articleURL := range allArticleURLs {
			select {
			case <-timeout:
				log.Printf("⚠️ Article processing timeout reached")
				return
			default:
				// Rate limiting - pause between requests
				if processedCount > 0 {
					time.Sleep(500 * time.Millisecond) // 500ms delay between requests
				}
				
				// Scrape the article
				article, err := h.scraper.ScrapeURL(articleURL, mode)
				if err != nil {
					log.Printf("Error scraping %s: %v", articleURL, err)
					continue
				}
				
				// Save to database (without NLP processing - will be done on-demand)
				if err := h.repo.CreateArticle(article); err != nil {
					log.Printf("Error saving article: %v", err)
					continue
				}
				
				log.Printf("📄 Saved article: %s (NLP processing will be done on-demand)", article.Title)
				
				processedCount++
				
				// Update job progress
				job.mu.Lock()
				job.Progress = processedCount
				job.Total = len(allArticleURLs)
				job.mu.Unlock()
				
				// Stop if we've processed enough articles
				if processedCount >= limit {
					break
				}
			}
		}
		
		log.Printf("✅ Completed scraping job for query '%s': %d articles processed", query, processedCount)
	}()
	
	// Wait for completion or timeout
	select {
	case <-done:
		log.Printf("✅ Scraping job completed for query: %s", query)
	case <-timeout:
		log.Printf("⚠️ Scraping job timed out for query: %s", query)
		job.mu.Lock()
		job.Status = "timeout"
		job.mu.Unlock()
	}
}



// generateSearchURLs generates URLs to scrape based on search query
func (h *Handler) generateSearchURLs(query, mode string, limit int) []string {
	var urls []string
	
	if mode == "research" {
		// Generate arXiv search URLs
		arxivQuery := strings.ReplaceAll(query, " ", "+")
		urls = append(urls, fmt.Sprintf("https://arxiv.org/search/?query=%s&searchtype=all&source=header", arxivQuery))
		
		// Add more research sources
		urls = append(urls, fmt.Sprintf("https://scholar.google.com/scholar?q=%s", arxivQuery))
		
	} else {
		// Generate news search URLs - expanded sources for better coverage
		newsQuery := strings.ReplaceAll(query, " ", "+")

		candidates := []string{
			// High-Quality Tech News Sites (Prioritized by performance)
			fmt.Sprintf("https://techcrunch.com/search/%s/", newsQuery),
			fmt.Sprintf("https://www.zdnet.com/search/?searchQuery=%s", newsQuery),
			fmt.Sprintf("https://www.cnet.com/search/?query=%s", newsQuery),
			
			// Additional Reliable Tech Sources
			fmt.Sprintf("https://www.theverge.com/search?q=%s", newsQuery),
			fmt.Sprintf("https://arstechnica.com/search/?q=%s", newsQuery),
			
			// Developer/Programming Sources
			fmt.Sprintf("https://www.infoq.com/search.action?queryString=%s", newsQuery),
			
			// Additional Tech Sources
			fmt.Sprintf("https://www.techradar.com/search?searchTerm=%s", newsQuery),
			
			// Business Tech Sources
			fmt.Sprintf("https://www.forbes.com/search/?q=%s", newsQuery),
		}

		// Shuffle all sources to distribute load evenly across all sites
		rand.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
		
		// Limit to top 8 sources to prevent overwhelming
		if len(candidates) > 8 {
			candidates = candidates[:8]
		}
		
		urls = append(urls, candidates...)
	}
	
	// De-duplicate and cap to limit after shuffling
	if limit > 0 && len(urls) > 0 {
		seen := make(map[string]struct{}, len(urls))
		uniq := make([]string, 0, len(urls))
		for _, u := range urls {
			if _, ok := seen[u]; ok {
				continue
			}
			seen[u] = struct{}{}
			uniq = append(uniq, u)
			if len(uniq) == limit {
				break
			}
		}
		urls = uniq
	}
	
	return urls
}

// processArticleWithNLP sends article to NLP service for processing
func (h *Handler) processArticleWithNLP(article *models.Article) {
	// Filter out low-quality articles
	if h.isLowQualityArticle(article) {
		log.Printf("🚫 Skipping low-quality article: %s", article.Title)
		return
	}
	
	// Send to NLP service for classification and summarization
	if err := h.callNLPService(article); err != nil {
		log.Printf("❌ NLP service error for %s: %v", article.Title, err)
		// Fallback to basic classification
		article.Category = h.basicClassify(article.Title + " " + article.Content)
		return
	}
	
	log.Printf("✅ Processed article with NLP: %s", article.Title)
}

// callNLPService sends article to Python NLP service for processing
func (h *Handler) callNLPService(article *models.Article) error {
	// Prepare the request for full article processing
	request := models.NLPRequest{
		Text: article.Content,
		Task: "process_article",
		Options: models.NLPOptions{
			MaxLength:   150,
			MinLength:   50,
			Temperature: 0.7,
		},
	}
	
	// Add title to the request if available
	if article.Title != "" {
		request.Text = article.Title + "\n\n" + article.Content
	}
	
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal NLP request: %v", err)
	}
	
	// Call the NLP service
	resp, err := http.Post("http://localhost:5000/process_article", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to call NLP service: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("NLP service returned status %d", resp.StatusCode)
	}
	
	// Parse the response
	var nlpResponse models.NLPResponse
	if err := json.NewDecoder(resp.Body).Decode(&nlpResponse); err != nil {
		return fmt.Errorf("failed to decode NLP response: %v", err)
	}
	
	if !nlpResponse.Success {
		return fmt.Errorf("NLP service error: %s", nlpResponse.Error)
	}
	
	// Extract results from the response
	if data, ok := nlpResponse.Data.(map[string]interface{}); ok {
		// Update article with classification
		if classification, ok := data["classification"].(map[string]interface{}); ok {
			if category, ok := classification["category"].(string); ok {
				article.Category = category
			}
		}
		
		// Update article with summary
		if summary, ok := data["summary"].(map[string]interface{}); ok {
			if summaryText, ok := summary["summary"].(string); ok {
				article.Summary = summaryText
			}
		}
		
		// Update article with keywords
		if keywords, ok := data["keywords"].([]interface{}); ok {
			keywordStrings := make([]string, len(keywords))
			for i, kw := range keywords {
				if keyword, ok := kw.(string); ok {
					keywordStrings[i] = keyword
				}
			}
			article.Keywords = strings.Join(keywordStrings, ", ")
		}
	}
	
	return nil
}

// isLowQualityArticle checks if an article should be filtered out
func (h *Handler) isLowQualityArticle(article *models.Article) bool {
	// Filter out arXiv papers with very long titles (usually technical papers)
	if strings.Contains(article.Source, "arxiv.org") {
		title := article.Title
		// Skip if title contains too many technical terms or is too long
		if len(title) > 200 || 
		   strings.Contains(title, "Bibliographic and Citation Tools") ||
		   strings.Contains(title, "Code, Data and Media") ||
		   strings.Contains(title, "DemosRecommenders") {
			return true
		}
	}
	
	// Filter out articles with very short content
	if len(article.Content) < 100 {
		return true
	}
	
	// Filter out articles with suspicious titles
	suspiciousPatterns := []string{
		"Title:",
		"Bibliographic",
		"Citation Tools",
		"Code, Data and Media",
		"DemosRecommenders",
		"Search Tools",
		"arXivLabs",
		"About ",
		"Editorial Values",
		"Editorial Standards",
		"Privacy Policy",
		"Terms of Service",
		"Cookie Policy",
		"Contact Us",
		"Subscribe",
		"Newsletter",
		"Advertise",
		"Careers",
		"Jobs",
		"Premium",
		"Library",
	}
	
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(article.Title, pattern) {
			return true
		}
	}
	
	return false
}

// basicClassify provides basic classification (placeholder for NLP service)
func (h *Handler) basicClassify(text string) string {
	// Simple keyword-based classification
	keywords := map[string][]string{
		"Artificial Intelligence": {"ai", "machine learning", "deep learning", "neural", "algorithm"},
		"Cloud Computing":        {"cloud", "aws", "azure", "gcp", "serverless", "kubernetes"},
		"Cybersecurity":          {"security", "cyber", "hack", "vulnerability", "encryption"},
		"Data Science":           {"data", "analytics", "big data", "visualization", "statistics"},
		"DevOps":                 {"devops", "ci/cd", "pipeline", "deployment", "infrastructure"},
	}

	for category, words := range keywords {
		for _, word := range words {
			if contains(text, word) {
				return category
			}
		}
	}

	return "Technology"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr ||
		     containsSubstring(s, substr)))
}

// containsSubstring is a simple substring check
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetActiveScrapingJobs returns all active scraping jobs
func (h *Handler) GetActiveScrapingJobs(c *gin.Context) {
	h.jobMutex.RLock()
	defer h.jobMutex.RUnlock()
	
	var activeJobs []gin.H
	for _, job := range h.scrapingJobs {
		job.mu.RLock()
		if job.Status == "running" {
			activeJobs = append(activeJobs, gin.H{
				"id":        job.ID,
				"query":     job.Query,
				"mode":      job.Mode,
				"status":    job.Status,
				"progress":  job.Progress,
				"total":     job.Total,
				"startTime": job.StartTime,
			})
		}
		job.mu.RUnlock()
	}
	
	c.JSON(http.StatusOK, gin.H{
		"activeJobs": activeJobs,
		"count":      len(activeJobs),
	})
}

// GetScrapeStatus retrieves the status of a scraping job
func (h *Handler) GetScrapeStatus(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Job ID is required"})
		return
	}

	// Find the job by ID
	h.jobMutex.RLock()
	var targetJob *ScrapingJob
	for _, job := range h.scrapingJobs {
		if job.ID == jobID {
			targetJob = job
			break
		}
	}
	h.jobMutex.RUnlock()

	if targetJob == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
		return
	}

	// Get job status
	targetJob.mu.RLock()
	status := gin.H{
		"id":        targetJob.ID,
		"query":     targetJob.Query,
		"mode":      targetJob.Mode,
		"status":    targetJob.Status,
		"progress":  targetJob.Progress,
		"total":     targetJob.Total,
		"startTime": targetJob.StartTime,
	}
	targetJob.mu.RUnlock()

	c.JSON(http.StatusOK, status)
}

// GetArticles retrieves articles with optional filtering
func (h *Handler) GetArticles(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	mode := c.DefaultQuery("mode", "news")
	category := c.Query("category")

	// Get articles from database
	var categories []string
	if category != "" {
		categories = []string{category}
	}
	
	articles, err := h.repo.SearchArticles("", mode, categories, limit)
	if err != nil {
		log.Printf("❌ Error retrieving articles: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"total":    len(articles),
		"limit":    limit,
		"mode":     mode,
		"category": category,
	})
}

// GetArticle retrieves a specific article by ID
func (h *Handler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	article, err := h.repo.GetArticle(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, article)
}

// ProcessArticleWithNLP processes an article with NLP on-demand
func (h *Handler) ProcessArticleWithNLP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	log.Printf("🔄 On-demand NLP processing requested for article ID: %d", id)

	// Get the article from database
	article, err := h.repo.GetArticle(id)
	if err != nil {
		log.Printf("❌ Article not found for ID %d: %v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// Check if article already has NLP processing
	if article.Summary != "" && article.Category != "" && article.Keywords != "" {
		log.Printf("✅ Article %d already processed with NLP, returning cached results", id)
		c.JSON(http.StatusOK, gin.H{
			"message": "Article already processed with NLP",
			"article": article,
		})
		return
	}

	log.Printf("🤖 Starting NLP processing for article: %s", article.Title)

	// Process with NLP
	h.processArticleWithNLP(article)

	// Update the article in database with NLP results
	if err := h.repo.UpdateArticle(article); err != nil {
		log.Printf("❌ Error updating article with NLP results: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update article with NLP results"})
		return
	}

	log.Printf("✅ Successfully processed article %d with NLP", id)

	c.JSON(http.StatusOK, gin.H{
		"message": "Article processed with NLP successfully",
		"article": article,
	})
}

// DeleteArticle deletes an article by ID
func (h *Handler) DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid article ID"})
		return
	}

	if err := h.repo.DeleteArticle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Article deleted successfully"})
}

// GetCategories retrieves all categories
func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.repo.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// HealthCheck provides a health check endpoint
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "spectechle-backend",
		"version":   "1.0.0",
	})
}

// TestSearchURLs tests which search URLs are working
func (h *Handler) TestSearchURLs(c *gin.Context) {
	// Call the test function from scraper package
	h.scraper.TestSearchURLs()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "URL test completed, check server logs for results",
		"time":    time.Now().Format(time.RFC3339),
	})
}

