package database

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"spectechle/backend/models"

	"encoding/json"
	"fmt"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection
type DB struct {
	*sql.DB
}

// InitDB initializes the database connection and creates tables
func InitDB() (*DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "spectechle.db"
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	log.Println("✅ Database initialized successfully")
	return &DB{db}, nil
}

// createTables creates all necessary tables
func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			url TEXT UNIQUE NOT NULL,
			content TEXT,
			summary TEXT,
			category TEXT,
			source TEXT,
			author TEXT,
			published_at DATETIME,
			scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			keywords TEXT,
			read_time INTEGER DEFAULT 0,
			score REAL DEFAULT 0.0,
			mode TEXT DEFAULT 'news'
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS search_sessions (
			id TEXT PRIMARY KEY,
			query TEXT NOT NULL,
			mode TEXT NOT NULL,
			status TEXT DEFAULT 'processing',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			results_count INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS scrape_jobs (
			id TEXT PRIMARY KEY,
			urls TEXT NOT NULL,
			mode TEXT NOT NULL,
			status TEXT DEFAULT 'processing',
			progress INTEGER DEFAULT 0,
			total INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			error TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS search_cache (
			id TEXT PRIMARY KEY,
			query TEXT NOT NULL,
			mode TEXT NOT NULL,
			categories TEXT,
			article_ids TEXT NOT NULL,
			result_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_search_cache_query ON search_cache (query, mode)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Insert default categories
	return insertDefaultCategories(db)
}

// insertDefaultCategories inserts default tech categories
func insertDefaultCategories(db *sql.DB) error {
	categories := []models.Category{
		{Name: "Artificial Intelligence", Description: "AI, ML, Deep Learning"},
		{Name: "Cloud Computing", Description: "AWS, Azure, GCP, Serverless"},
		{Name: "Cybersecurity", Description: "Security, Privacy, Threats"},
		{Name: "Data Science", Description: "Big Data, Analytics, Visualization"},
		{Name: "DevOps", Description: "CI/CD, Infrastructure, Automation"},
		{Name: "Web Development", Description: "Frontend, Backend, Full-stack"},
		{Name: "Mobile Development", Description: "iOS, Android, React Native"},
		{Name: "Blockchain", Description: "Cryptocurrency, Smart Contracts"},
		{Name: "IoT", Description: "Internet of Things, Sensors"},
		{Name: "Quantum Computing", Description: "Quantum algorithms, Qubits"},
	}

	for _, cat := range categories {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO categories (name, description) 
			VALUES (?, ?)
		`, cat.Name, cat.Description)
		if err != nil {
			return err
		}
	}

	return nil
}

// ArticleRepository handles article database operations
type ArticleRepository struct {
	db *DB
}

// NewArticleRepository creates a new article repository
func NewArticleRepository(db *DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// CreateArticle inserts a new article
func (r *ArticleRepository) CreateArticle(article *models.Article) error {
	// Check if article already exists
	exists, err := r.ArticleExists(article.URL)
	if err != nil {
		return err
	}
	
	if exists {
		// Article already exists, skip insertion
		log.Printf("⏭️  Article already exists, skipping: %s", article.Title)
		return nil
	}
	
	query := `
		INSERT INTO articles (title, url, content, summary, category, source, author, published_at, keywords, read_time, score, mode)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query,
		article.Title, article.URL, article.Content, article.Summary,
		article.Category, article.Source, article.Author, article.PublishedAt,
		article.Keywords, article.ReadTime, article.Score, article.Mode)
	
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	article.ID = id
	return nil
}

// UpdateArticle updates an existing article with new data
func (r *ArticleRepository) UpdateArticle(article *models.Article) error {
	query := `
		UPDATE articles 
		SET title = ?, url = ?, content = ?, summary = ?, category = ?, 
		    source = ?, author = ?, published_at = ?, keywords = ?, 
		    read_time = ?, score = ?, mode = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query,
		article.Title, article.URL, article.Content, article.Summary,
		article.Category, article.Source, article.Author, article.PublishedAt,
		article.Keywords, article.ReadTime, article.Score, article.Mode,
		article.ID)
	
	if err != nil {
		return err
	}

	return nil
}

// ArticleExists checks if an article with the given URL already exists
func (r *ArticleRepository) ArticleExists(url string) (bool, error) {
	var exists int
	query := `SELECT 1 FROM articles WHERE url = ? LIMIT 1`
	
	err := r.db.QueryRow(query, url).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	
	return true, nil
}

// GetArticle retrieves an article by ID
func (r *ArticleRepository) GetArticle(id int64) (*models.Article, error) {
	article := &models.Article{}
	query := `SELECT * FROM articles WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(
		&article.ID, &article.Title, &article.URL, &article.Content,
		&article.Summary, &article.Category, &article.Source, &article.Author,
		&article.PublishedAt, &article.ScrapedAt, &article.Keywords,
		&article.ReadTime, &article.Score, &article.Mode)
	
	if err != nil {
		return nil, err
	}

	return article, nil
}

// SearchArticles searches articles based on criteria
func (r *ArticleRepository) SearchArticles(query string, mode string, categories []string, limit int) ([]models.Article, error) {
	baseQuery := `
		SELECT * FROM articles 
		WHERE (title LIKE ? OR content LIKE ? OR keywords LIKE ?)
		AND mode = ?
	`
	
	args := []interface{}{"%" + query + "%", "%" + query + "%", "%" + query + "%", mode}
	
	if len(categories) > 0 {
		baseQuery += " AND category IN ("
		for i, cat := range categories {
			if i > 0 {
				baseQuery += ","
			}
			baseQuery += "?"
			args = append(args, cat)
		}
		baseQuery += ")"
	}
	
	baseQuery += " ORDER BY score DESC, scraped_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return []models.Article{}, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(
			&article.ID, &article.Title, &article.URL, &article.Content,
			&article.Summary, &article.Category, &article.Source, &article.Author,
			&article.PublishedAt, &article.ScrapedAt, &article.Keywords,
			&article.ReadTime, &article.Score, &article.Mode)
		
		if err != nil {
			return []models.Article{}, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetCategories retrieves all categories with article counts
func (r *ArticleRepository) GetCategories() ([]models.Category, error) {
	query := `
		SELECT c.id, c.name, c.description, COUNT(a.id) as count
		FROM categories c
		LEFT JOIN articles a ON c.name = a.category
		GROUP BY c.id, c.name, c.description
		ORDER BY count DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.Count)
		if err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// DeleteArticle deletes an article by ID
func (r *ArticleRepository) DeleteArticle(id int64) error {
	query := `DELETE FROM articles WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// GetCachedSearch retrieves cached search results
func (r *ArticleRepository) GetCachedSearch(query, mode string, categories []string) ([]models.Article, error) {
	// Create cache key from query parameters
	cacheKey := r.createCacheKey(query, mode, categories)
	
	var articleIDs string
	var expiresAt sql.NullTime
	
	sqlQuery := `SELECT article_ids, expires_at FROM search_cache WHERE id = ? AND (expires_at IS NULL OR expires_at > datetime('now'))`
	err := r.db.QueryRow(sqlQuery, cacheKey).Scan(&articleIDs, &expiresAt)
	
	if err == sql.ErrNoRows {
		return nil, nil // No cache hit
	}
	if err != nil {
		return nil, err
	}
	
	// Parse article IDs and fetch articles
	ids := strings.Split(articleIDs, ",")
	if len(ids) == 0 {
		return nil, nil
	}
	
	// Build query to fetch articles by IDs
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
	
	articleQuery := fmt.Sprintf(`SELECT * FROM articles WHERE id IN (%s) ORDER BY score DESC`, placeholders)
	
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	
	rows, err := r.db.Query(articleQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var articles []models.Article
	for rows.Next() {
		var article models.Article
		err := rows.Scan(
			&article.ID, &article.Title, &article.URL, &article.Content,
			&article.Summary, &article.Category, &article.Source, &article.Author,
			&article.PublishedAt, &article.ScrapedAt, &article.Keywords,
			&article.ReadTime, &article.Score, &article.Mode)
		
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	
	return articles, nil
}

// CacheSearch stores search results in cache
func (r *ArticleRepository) CacheSearch(query, mode string, categories []string, articles []models.Article, ttlHours int) error {
	cacheKey := r.createCacheKey(query, mode, categories)
	
	// Convert article IDs to comma-separated string
	var ids []string
	for _, article := range articles {
		ids = append(ids, fmt.Sprintf("%d", article.ID))
	}
	articleIDs := strings.Join(ids, ",")
	
	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(ttlHours) * time.Hour)
	
	// Convert categories to JSON string
	categoriesJSON := "[]"
	if len(categories) > 0 {
		if jsonBytes, err := json.Marshal(categories); err == nil {
			categoriesJSON = string(jsonBytes)
		}
	}
	
	sqlQuery := `
		INSERT OR REPLACE INTO search_cache (id, query, mode, categories, article_ids, result_count, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(sqlQuery, cacheKey, query, mode, categoriesJSON, articleIDs, len(articles), expiresAt)
	return err
}

// createCacheKey creates a unique cache key for search parameters
func (r *ArticleRepository) createCacheKey(query, mode string, categories []string) string {
	// Sort categories for consistent cache keys
	sort.Strings(categories)
	
	// Create a hash of the search parameters
	key := fmt.Sprintf("%s:%s:%s", query, mode, strings.Join(categories, ","))
	
	// Simple hash function (in production, use a proper hash)
	hash := 0
	for _, char := range key {
		hash = ((hash << 5) - hash) + int(char)
		hash = hash & hash // Convert to 32-bit integer
	}
	
	return fmt.Sprintf("search_%d", hash)
}

// CleanExpiredCache removes expired cache entries
func (r *ArticleRepository) CleanExpiredCache() error {
	query := `DELETE FROM search_cache WHERE expires_at < datetime('now')`
	_, err := r.db.Exec(query)
	return err
}

// GetCacheStats returns cache statistics
func (r *ArticleRepository) GetCacheStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Total cache entries
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM search_cache`).Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total_entries"] = total
	
	// Expired entries
	var expired int
	err = r.db.QueryRow(`SELECT COUNT(*) FROM search_cache WHERE expires_at < datetime('now')`).Scan(&expired)
	if err != nil {
		return nil, err
	}
	stats["expired_entries"] = expired
	
	// Cache hit rate (would need to track this separately)
	stats["cache_hit_rate"] = "N/A" // TODO: Implement hit rate tracking
	
	return stats, nil
}


