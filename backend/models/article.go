package models

import (
	"time"
)

// Article represents a scraped article or research paper
type Article struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	URL         string    `json:"url" db:"url"`
	Content     string    `json:"content" db:"content"`
	Summary     string    `json:"summary" db:"summary"`
	Category    string    `json:"category" db:"category"`
	Source      string    `json:"source" db:"source"`
	Author      string    `json:"author" db:"author"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	ScrapedAt   time.Time `json:"scraped_at" db:"scraped_at"`
	Keywords    string    `json:"keywords" db:"keywords"`
	ReadTime    int       `json:"read_time" db:"read_time"`
	Score       float64   `json:"score" db:"score"`
	Mode        string    `json:"mode" db:"mode"` // "news" or "research"
}

// SearchRequest represents a search query from the frontend
type SearchRequest struct {
	Query       string   `json:"query" binding:"required"`
	Mode        string   `json:"mode" binding:"required"` // "news" or "research"
	Categories  []string `json:"categories"`
	Limit       int      `json:"limit"`
	Sources     []string `json:"sources"`
	DateFrom    string   `json:"date_from"`
	DateTo      string   `json:"date_to"`
}

// SearchResponse represents the response from a search operation
type SearchResponse struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Mode      string    `json:"mode"`
	Articles  []Article `json:"articles"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "processing", "completed", "failed"
	CacheHit  bool      `json:"cache_hit,omitempty"` // Whether results came from cache
}

// ScrapeRequest represents a scraping request
type ScrapeRequest struct {
	URLs       []string `json:"urls" binding:"required"`
	Mode       string   `json:"mode" binding:"required"`
	Categories []string `json:"categories"`
}

// ScrapeResponse represents the response from a scraping operation
type ScrapeResponse struct {
	ID        string    `json:"id"`
	URLs      []string  `json:"urls"`
	Status    string    `json:"status"` // "processing", "completed", "failed"
	Progress  int       `json:"progress"`
	Total     int       `json:"total"`
	CreatedAt time.Time `json:"created_at"`
	Results   []Article `json:"results,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Category represents an article category
type Category struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Count       int    `json:"count" db:"count"`
}

// NLPRequest represents a request to the Python NLP service
type NLPRequest struct {
	Text     string   `json:"text"`
	Task     string   `json:"task"` // "classify", "summarize", "extract_keywords"
	Options  NLPOptions `json:"options,omitempty"`
}

// NLPResponse represents a response from the Python NLP service
type NLPResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// NLPOptions represents options for NLP processing
type NLPOptions struct {
	MaxLength    int     `json:"max_length,omitempty"`
	MinLength    int     `json:"min_length,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
	TopK         int     `json:"top_k,omitempty"`
	TopP         float64 `json:"top_p,omitempty"`
}

// ClassificationResult represents the result of article classification
type ClassificationResult struct {
	Category    string  `json:"category"`
	Confidence  float64 `json:"confidence"`
	Categories  []CategoryScore `json:"categories"`
}

// CategoryScore represents a category with its confidence score
type CategoryScore struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"`
}

// SummarizationResult represents the result of text summarization
type SummarizationResult struct {
	Summary    string  `json:"summary"`
	WordCount  int     `json:"word_count"`
	CompressionRatio float64 `json:"compression_ratio"`
}


