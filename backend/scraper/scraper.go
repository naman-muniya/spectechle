package scraper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"spectechle/backend/models"

	"github.com/gocolly/colly/v2"
)

// NewsSource represents a news source configuration
type NewsSource struct {
	Name        string   `json:"name"`
	Domain      string   `json:"domain"`
	URLPatterns []string `json:"url_patterns"`
	Selectors   struct {
		Title       []string `json:"title"`
		Content     []string `json:"content"`
		Author      []string `json:"author"`
		Date        []string `json:"date"`
		Keywords    []string `json:"keywords"`
		Description []string `json:"description"`
	} `json:"selectors"`
	APIConfig *APIConfig `json:"api_config,omitempty"`
}

// APIConfig represents API configuration for news sources
type APIConfig struct {
	Endpoint string            `json:"endpoint"`
	APIKey   string            `json:"api_key"`
	Headers  map[string]string `json:"headers"`
	Params   map[string]string `json:"params"`
}

// Scraper handles web scraping operations
type Scraper struct {
	collector *colly.Collector
	sources   map[string]*NewsSource
	apiClient *http.Client
}

// NewScraper creates a new scraper instance
func NewScraper() *Scraper {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowURLRevisit(),
	)

	// Set request delay to be respectful but faster for better performance
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 500 * time.Millisecond, // Reduced from 2 seconds to 500ms
		Parallelism: 5, // Allow 5 concurrent requests per domain
	})

	scraper := &Scraper{
		collector: c,
		sources:   make(map[string]*NewsSource),
		apiClient: &http.Client{
			Timeout: 15 * time.Second, // Reduced timeout for faster failure detection
		},
	}

	// Initialize news sources
	scraper.initializeNewsSources()

	return scraper
}

// initializeNewsSources sets up the news source configurations
func (s *Scraper) initializeNewsSources() {
	// TechCrunch
	s.sources["techcrunch.com"] = &NewsSource{
		Name:   "TechCrunch",
		Domain: "techcrunch.com",
		URLPatterns: []string{
			`/20\d{2}/`, `/post/`, `/article/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".article__title", ".post-title"},
			Content:     []string{".article-content", ".post-content", ".entry-content"},
			Author:      []string{".author", ".byline", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// Wired
	s.sources["wired.com"] = &NewsSource{
		Name:   "Wired",
		Domain: "wired.com",
		URLPatterns: []string{
			`/story/`, `/20\d{2}/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".headline", ".article__title"},
			Content:     []string{".article__content", ".content", ".story-body"},
			Author:      []string{".byline", ".author", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// Ars Technica
	s.sources["arstechnica.com"] = &NewsSource{
		Name:   "Ars Technica",
		Domain: "arstechnica.com",
		URLPatterns: []string{
			`/20\d{2}/`, `/tech-policy/`, `/science/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".title", ".headline"},
			Content:     []string{".article-content", ".entry-content", ".post-content"},
			Author:      []string{".author", ".byline", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// The Verge
	s.sources["theverge.com"] = &NewsSource{
		Name:   "The Verge",
		Domain: "theverge.com",
		URLPatterns: []string{
			`/20\d{2}/`, `/article/`, `/review/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".c-entry-title", ".headline"},
			Content:     []string{".c-entry-content", ".article-content", ".content"},
			Author:      []string{".c-byline", ".author", "[rel='author']"},
			Date:        []string{"time", ".c-byline__item", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// VentureBeat
	s.sources["venturebeat.com"] = &NewsSource{
		Name:   "VentureBeat",
		Domain: "venturebeat.com",
		URLPatterns: []string{
			`/20\d{2}/`, `/article/`, `/news/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".entry-title", ".headline"},
			Content:     []string{".article-content", ".entry-content", ".post-content"},
			Author:      []string{".author", ".byline", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// MIT Technology Review
	s.sources["technologyreview.com"] = &NewsSource{
		Name:   "MIT Technology Review",
		Domain: "technologyreview.com",
		URLPatterns: []string{
			`/20\d{2}/`, `/article/`, `/story/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".article__title", ".headline"},
			Content:     []string{".article__content", ".content", ".story-body"},
			Author:      []string{".byline", ".author", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// IEEE Spectrum
	s.sources["ieee.org"] = &NewsSource{
		Name:   "IEEE Spectrum",
		Domain: "ieee.org",
		URLPatterns: []string{
			`/spectrum/`, `/20\d{2}/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".title", ".headline"},
			Content:     []string{".article-content", ".entry-content", ".post-content"},
			Author:      []string{".author", ".byline", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}

	// ACM
	s.sources["acm.org"] = &NewsSource{
		Name:   "ACM",
		Domain: "acm.org",
		URLPatterns: []string{
			`/publications/`, `/20\d{2}/`,
		},
		Selectors: struct {
			Title       []string `json:"title"`
			Content     []string `json:"content"`
			Author      []string `json:"author"`
			Date        []string `json:"date"`
			Keywords    []string `json:"keywords"`
			Description []string `json:"description"`
		}{
			Title:       []string{"h1", ".title", ".headline"},
			Content:     []string{".article-content", ".entry-content", ".post-content"},
			Author:      []string{".author", ".byline", "[rel='author']"},
			Date:        []string{"time", ".published-date", "[property='article:published_time']"},
			Keywords:    []string{"[name='keywords']", "[property='article:tag']"},
			Description: []string{"[property='og:description']", "[name='description']"},
		},
	}
}

// ScrapeURL scrapes content from a given URL with enhanced extraction
func (s *Scraper) ScrapeURL(url, mode string) (*models.Article, error) {
	// Normalize the URL
	normalizedURL := s.normalizeURL(url)
	
	article := &models.Article{
		URL:       normalizedURL,
		Mode:      mode,
		ScrapedAt: time.Now(),
	}

	// Check if we have a configured source for this domain
	source, hasSource := s.getSourceByURL(normalizedURL)
	// Clone collector per scrape to avoid handler leakage across concurrent requests
	c := s.collector.Clone()

	// Configure collector for this specific URL
	c.OnHTML("html", func(e *colly.HTMLElement) {
		if hasSource {
			// Use source-specific extraction
			article.Title = s.extractWithSelectors(e, source.Selectors.Title)
			article.Content = s.extractWithSelectors(e, source.Selectors.Content)
			article.Author = s.extractWithSelectors(e, source.Selectors.Author)
			article.Keywords = s.extractKeywordsWithSelectors(e, source.Selectors.Keywords)
		} else {
			// Fallback to generic extraction
			article.Title = s.extractTitle(e)
			if mode == "research" {
				article.Content = s.extractResearchContent(e)
			} else {
				article.Content = s.extractNewsContent(e)
			}
			article.Author = s.extractAuthor(e)
			article.Keywords = s.extractKeywords(e)
		}
		
		// Common metadata extraction
		article.Source = s.extractSource(normalizedURL)
		article.PublishedAt = s.extractPublishedDate(e)
		article.ReadTime = s.calculateReadTime(article.Content)
		article.Score = s.calculateScore(article)
		article.Category = s.classifyArticle(article.Title, article.Content)
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error scraping %s: %v", normalizedURL, err)
	})

	// Visit the URL
	if err := c.Visit(normalizedURL); err != nil {
		return nil, err
	}

	return article, nil
}

// extractWithSelectors extracts content using multiple selectors
func (s *Scraper) extractWithSelectors(e *colly.HTMLElement, selectors []string) string {
	for _, selector := range selectors {
		if content := e.ChildText(selector); content != "" {
			cleaned := strings.TrimSpace(content)
			if len(cleaned) > 10 {
				return cleaned
			}
		}
	}
	return ""
}

// extractKeywordsWithSelectors extracts keywords using specific selectors
func (s *Scraper) extractKeywordsWithSelectors(e *colly.HTMLElement, selectors []string) string {
	for _, selector := range selectors {
		if keywords := e.ChildAttr(selector, "content"); keywords != "" {
			return strings.TrimSpace(keywords)
		}
	}
	return ""
}

// classifyArticle classifies the article based on content
func (s *Scraper) classifyArticle(title, content string) string {
	text := strings.ToLower(title + " " + content)
	
	categories := map[string][]string{
		"Artificial Intelligence": {"ai", "artificial intelligence", "machine learning", "ml", "neural network", "deep learning"},
		"Cloud Computing": {"cloud", "aws", "azure", "gcp", "kubernetes", "docker", "microservices"},
		"Cybersecurity": {"security", "cybersecurity", "hack", "vulnerability", "breach", "encryption"},
		"Data Science": {"data science", "analytics", "big data", "data mining", "statistics"},
		"Blockchain": {"blockchain", "bitcoin", "ethereum", "crypto", "cryptocurrency", "defi"},
		"Quantum Computing": {"quantum", "qubit", "quantum computing", "quantum algorithm"},
		"IoT": {"iot", "internet of things", "sensor", "smart device"},
		"Web Development": {"web", "frontend", "backend", "javascript", "react", "vue"},
		"Mobile Development": {"mobile", "ios", "android", "app", "smartphone"},
		"DevOps": {"devops", "ci/cd", "automation", "deployment"},
	}
	
	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				return category
			}
		}
	}
	
	return "Technology"
}

// ScrapeSearchResults scrapes search result pages and extracts article URLs
func (s *Scraper) ScrapeSearchResults(searchURL, mode string, limit int) ([]string, error) {
	var articleURLs []string

	// Clone collector to avoid handler reuse across searches
	c := s.collector.Clone()

	// Configure collector for search results
	c.OnHTML("html", func(e *colly.HTMLElement) {
		if mode == "research" {
			// Extract arXiv paper URLs
			e.ForEach("a[href*='/abs/']", func(_ int, el *colly.HTMLElement) {
				href := el.Attr("href")
				if strings.Contains(href, "/abs/") {
					// Fix URL construction - ensure proper format
					if strings.HasPrefix(href, "/abs/") {
						articleURLs = append(articleURLs, "https://arxiv.org"+href)
					} else if strings.Contains(href, "arxiv.org/abs/") {
						articleURLs = append(articleURLs, href)
					}
				}
			})
		} else {
			// Extract news article URLs with better logic
			s.extractNewsURLs(e, searchURL, &articleURLs)
		}
	})
	
	// Visit the search URL
	if err := c.Visit(searchURL); err != nil {
		return nil, err
	}
	
	log.Printf("🔍 Found %d article URLs from %s", len(articleURLs), searchURL)
	
	// Limit results per site to ensure better distribution
	articleURLs = s.limitResultsPerSite(articleURLs, limit)
	
	// Final limit check
	if len(articleURLs) > limit {
		articleURLs = articleURLs[:limit]
	}
	
	log.Printf("📊 Final article URLs after limiting: %d from %s", len(articleURLs), searchURL)
	
	return articleURLs, nil
}

// limitResultsPerSite limits the number of results per site to ensure better distribution
func (s *Scraper) limitResultsPerSite(urls []string, totalLimit int) []string {
	siteCounts := make(map[string]int)
	maxPerSite := totalLimit / 4 // Distribute evenly across sites
	
	var result []string
	
	for _, url := range urls {
		domain := s.extractDomain(url)
		if siteCounts[domain] < maxPerSite {
			result = append(result, url)
			siteCounts[domain]++
		}
	}
	
	return result
}

// extractNewsURLs extracts news article URLs with improved logic
func (s *Scraper) extractNewsURLs(e *colly.HTMLElement, searchURL string, articleURLs *[]string) {
	// Get the domain to understand which site we're scraping
	domain := s.extractDomain(searchURL)
	
	log.Printf("🎯 Extracting URLs from domain: %s", domain)
	
	switch domain {
	case "news.google.com":
		s.extractGoogleNewsURLs(e, articleURLs)
	case "techcrunch.com":
		s.extractTechCrunchURLs(e, articleURLs)
	case "wired.com":
		s.extractWiredURLs(e, articleURLs)
	case "arstechnica.com":
		s.extractArsTechnicaURLs(e, articleURLs)
	case "theverge.com":
		s.extractTheVergeURLs(e, articleURLs)
	case "zdnet.com":
		s.extractZDNetURLs(e, articleURLs)
	case "cnet.com":
		s.extractCNETURLs(e, articleURLs)
	case "infoq.com":
		s.extractInfoQURLs(e, articleURLs)
	case "techradar.com":
		s.extractTechRadarURLs(e, articleURLs)
	case "forbes.com":
		s.extractForbesURLs(e, articleURLs)
	case "sitepoint.com":
		s.extractSitePointURLs(e, articleURLs)
	default:
		log.Printf("⚠️  Using generic extraction for domain: %s", domain)
		s.extractGenericNewsURLs(e, searchURL, articleURLs)
	}
}

// extractGoogleNewsURLs extracts URLs from Google News
func (s *Scraper) extractGoogleNewsURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("article a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// Google News URLs are relative and need to be converted
		if strings.HasPrefix(href, "./") && len(title) > 20 {
			// Convert Google News URL to actual article URL
			actualURL := s.convertGoogleNewsURL(href)
			if actualURL != "" {
				*articleURLs = append(*articleURLs, actualURL)
			}
		}
	})
}

// extractTechCrunchURLs extracts URLs from TechCrunch
func (s *Scraper) extractTechCrunchURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href*='/202']", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		if len(title) > 20 && strings.Contains(href, "/202") {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://techcrunch.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractWiredURLs extracts URLs from Wired
func (s *Scraper) extractWiredURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href*='/story/']", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		if len(title) > 20 && strings.Contains(href, "/story/") {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.wired.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractArsTechnicaURLs extracts URLs from Ars Technica
func (s *Scraper) extractArsTechnicaURLs(e *colly.HTMLElement, articleURLs *[]string) {
	// Match broader patterns due to client-side search UI (hash params, etc.)
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		text := strings.TrimSpace(el.Text)
		if len(text) < 10 {
			return
		}
		if strings.Contains(href, "/202") || strings.Contains(href, "/science/") || strings.Contains(href, "/tech-policy/") || strings.Contains(href, "/information-technology/") {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://arstechnica.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractTheVergeURLs extracts URLs from The Verge
func (s *Scraper) extractTheVergeURLs(e *colly.HTMLElement, articleURLs *[]string) {
	// The Verge often uses non-year paths; include article and reviews
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		text := strings.TrimSpace(el.Text)
		if len(text) < 10 {
			return
		}
		if strings.Contains(href, "/202") || strings.Contains(href, "/article/") || strings.Contains(href, "/reviews/") {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.theverge.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractZDNetURLs extracts URLs from ZDNet
func (s *Scraper) extractZDNetURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// ZDNet articles typically have paths like /article/, /news/, /review/, etc.
		if len(title) > 20 && (strings.Contains(href, "/article/") || 
			strings.Contains(href, "/news/") || 
			strings.Contains(href, "/review/") ||
			strings.Contains(href, "/202")) {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.zdnet.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractCNETURLs extracts URLs from CNET
func (s *Scraper) extractCNETURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// CNET articles typically have paths like /news/, /reviews/, /how-to/, etc.
		if len(title) > 20 && (strings.Contains(href, "/news/") || 
			strings.Contains(href, "/reviews/") || 
			strings.Contains(href, "/how-to/") ||
			strings.Contains(href, "/202")) {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.cnet.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractInfoQURLs extracts URLs from InfoQ
func (s *Scraper) extractInfoQURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// InfoQ articles typically have paths like /news/, /articles/, etc.
		if len(title) > 20 && (strings.Contains(href, "/news/") || 
			strings.Contains(href, "/articles/") ||
			strings.Contains(href, "/202")) {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.infoq.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractSitePointURLs extracts URLs from SitePoint
func (s *Scraper) extractSitePointURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href*='/']", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		if len(title) > 20 && !strings.Contains(href, "#") && !strings.Contains(href, "javascript:") {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.sitepoint.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractTechRadarURLs extracts URLs from TechRadar
func (s *Scraper) extractTechRadarURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// TechRadar articles typically have paths like /news/, /reviews/, /how-to/, etc.
		if len(title) > 20 && (strings.Contains(href, "/news/") || 
			strings.Contains(href, "/reviews/") || 
			strings.Contains(href, "/how-to/") || 
			strings.Contains(href, "/features/") ||
			strings.Contains(href, "/202")) {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.techradar.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractForbesURLs extracts URLs from Forbes
func (s *Scraper) extractForbesURLs(e *colly.HTMLElement, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// Forbes articles typically have paths with /sites/ or year patterns
		if len(title) > 20 && (strings.Contains(href, "/sites/") || 
			strings.Contains(href, "/202") ||
			strings.Contains(href, "/article/")) {
			if strings.HasPrefix(href, "/") {
				*articleURLs = append(*articleURLs, "https://www.forbes.com"+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// extractGenericNewsURLs extracts URLs from generic news sites
func (s *Scraper) extractGenericNewsURLs(e *colly.HTMLElement, searchURL string, articleURLs *[]string) {
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		title := strings.TrimSpace(el.Text)
		
		// Better filtering for article links
		if len(title) > 20 && s.isArticleURL(href) {
			// Ensure full URL
			if strings.HasPrefix(href, "/") {
				baseURL := "https://" + s.extractDomain(searchURL)
				*articleURLs = append(*articleURLs, baseURL+href)
			} else if strings.HasPrefix(href, "http") {
				*articleURLs = append(*articleURLs, href)
			}
		}
	})
}

// isArticleURL checks if a URL looks like an article URL
func (s *Scraper) isArticleURL(href string) bool {
	// Skip non-article URLs
	skipPatterns := []string{
		"/about", "/privacy", "/terms", "/contact", "/subscribe", "/newsletter",
		"/advertise", "/careers", "/jobs", "/premium", "/library", "/cookie",
		"/editorial", "/standards", "/values", "/policy", "/legal",
		"/search", "/tag", "/category", "/author", "/user", "/profile",
		"/login", "/register", "/signup", "/signin", "/logout",
	}
	
	for _, pattern := range skipPatterns {
		if strings.Contains(href, pattern) {
			return false
		}
	}
	
	// Look for article patterns
	articlePatterns := []string{
		"/article/", "/story/", "/post/", "/news/", "/202", "/2023", "/2024", "/2025",
		"/tech/", "/blockchain/", "/crypto/", "/ai/", "/machine-learning/",
		"/artificial-intelligence/", "/cloud/", "/security/", "/cybersecurity/",
		"/reviews/", "/how-to/", "/features/", "/analysis/", "/opinion/",
		"/tutorial/", "/guide/", "/tips/", "/best-practices/", "/trends/",
		"/research/", "/study/", "/report/", "/survey/", "/interview/",
	}
	
	for _, pattern := range articlePatterns {
		if strings.Contains(href, pattern) {
			return true
		}
	}
	return false
}

// extractDomain extracts the domain from a URL
func (s *Scraper) extractDomain(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	
	// Get domain part
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		domain := parts[0]
		
		// Check if domain looks valid (contains a dot)
		if !strings.Contains(domain, ".") {
			return ""
		}
		
		// Remove www prefix for consistent matching
		domain = strings.TrimPrefix(domain, "www.")
		
		return domain
	}
	return ""
}

// normalizeURL normalizes a URL for consistent processing
func (s *Scraper) normalizeURL(url string) string {
	// Ensure URL has protocol
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")
	
	return url
}

// getSourceByURL gets the appropriate news source configuration for a URL
func (s *Scraper) getSourceByURL(url string) (*NewsSource, bool) {
	domain := s.extractDomain(url)
	source, exists := s.sources[domain]
	return source, exists
}

// convertGoogleNewsURL converts Google News URL to actual article URL
func (s *Scraper) convertGoogleNewsURL(googleURL string) string {
	// Google News URLs are in format ./articles/...
	// We need to extract the actual URL from the redirect
	// For now, return empty string as this requires special handling
	return ""
}

// extractTitle extracts the article title
func (s *Scraper) extractTitle(e *colly.HTMLElement) string {
	// Try multiple selectors for title
	selectors := []string{
		"h1",
		"title",
		"[property='og:title']",
		"[name='twitter:title']",
		".title",
		".headline",
		"#title",
		"#headline",
	}

	for _, selector := range selectors {
		if title := e.ChildText(selector); title != "" {
			return strings.TrimSpace(title)
		}
	}

	// Fallback to page title
	return strings.TrimSpace(e.ChildText("title"))
}

// extractNewsContent extracts content from news articles
func (s *Scraper) extractNewsContent(e *colly.HTMLElement) string {
	// Get the domain to use site-specific selectors
	url := e.Request.URL.String()
	domain := s.extractDomain(url)
	
	// Site-specific content selectors
	var selectors []string
	
	switch domain {
	case "techcrunch.com":
		selectors = []string{
			".article-content",
			".post-content",
			"article .content",
			".entry-content",
		}
	case "wired.com":
		selectors = []string{
			".article__content",
			".content",
			"article .body",
			".story-body",
		}
	case "arstechnica.com":
		selectors = []string{
			".article-content",
			".entry-content",
			"article .content",
			".post-content",
		}
	case "theverge.com":
		selectors = []string{
			".c-entry-content",
			".article-content",
			".content",
			"article .body",
		}
	case "venturebeat.com":
		selectors = []string{
			".article-content",
			".entry-content",
			"article .content",
			".post-content",
		}
	default:
		// Generic selectors for other sites
		selectors = []string{
			"article",
			".article-content",
			".post-content",
			".entry-content",
			".content",
			"main",
			"[role='main']",
			".story-body",
			".article-body",
			".post-body",
		}
	}

	// Try site-specific selectors first
	for _, selector := range selectors {
		content := e.ChildText(selector)
		if len(content) > 200 {
			return s.cleanContent(content)
		}
	}

	// Fallback: extract all paragraph text with better filtering
	var paragraphs []string
	e.ForEach("p", func(_ int, el *colly.HTMLElement) {
		text := strings.TrimSpace(el.Text)
		// Filter out navigation, ads, and short text
		if len(text) > 100 && !s.isNavigationText(text) {
			paragraphs = append(paragraphs, text)
		}
	})

	content := strings.Join(paragraphs, "\n\n")
	
	// Limit content length to avoid overly long articles
	if len(content) > 5000 {
		content = content[:5000] + "..."
	}
	
	return s.cleanContent(content)
}

// isNavigationText checks if text looks like navigation/ad content
func (s *Scraper) isNavigationText(text string) bool {
	navigationPatterns := []string{
		"subscribe", "newsletter", "sign up", "follow us", "share this",
		"advertisement", "sponsored", "cookie", "privacy policy",
		"terms of service", "contact us", "about us", "home",
		"search", "menu", "navigation", "footer", "header",
	}
	
	textLower := strings.ToLower(text)
	for _, pattern := range navigationPatterns {
		if strings.Contains(textLower, pattern) {
			return true
		}
	}
	return false
}

// extractResearchContent extracts content from research papers
func (s *Scraper) extractResearchContent(e *colly.HTMLElement) string {
	// Specific selectors for research papers (arXiv, IEEE, etc.)
	selectors := []string{
		".abstract",
		"#abstract",
		".paper-abstract",
		".research-content",
		".paper-content",
		"main",
		"article",
	}

	for _, selector := range selectors {
		content := e.ChildText(selector)
		if len(content) > 100 {
			return s.cleanContent(content)
		}
	}

	// Fallback to news content extraction
	return s.extractNewsContent(e)
}

// extractAuthor extracts the article author
func (s *Scraper) extractAuthor(e *colly.HTMLElement) string {
	selectors := []string{
		"[rel='author']",
		".author",
		".byline",
		"[property='article:author']",
		"[name='author']",
		".writer",
		".reporter",
	}

	for _, selector := range selectors {
		if author := e.ChildText(selector); author != "" {
			return strings.TrimSpace(author)
		}
	}

	return "Unknown Author"
}

// extractSource extracts the source from URL
func (s *Scraper) extractSource(url string) string {
	// Extract domain from URL
	if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			domain := parts[2]
			
			// Remove www prefix for cleaner source names
			domain = strings.TrimPrefix(domain, "www.")
			
			return domain
		}
	}
	return "Unknown Source"
}

// extractPublishedDate extracts the publication date
func (s *Scraper) extractPublishedDate(e *colly.HTMLElement) time.Time {
	selectors := []string{
		"[property='article:published_time']",
		"[name='publish_date']",
		".published-date",
		".date",
		".timestamp",
		"time",
	}

	for _, selector := range selectors {
		if dateStr := e.ChildAttr(selector, "content"); dateStr != "" {
			if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
				return t
			}
		}
		if dateStr := e.ChildText(selector); dateStr != "" {
			// Try common date formats
			formats := []string{
				"2006-01-02",
				"2006-01-02T15:04:05Z",
				"2006-01-02 15:04:05",
				"January 2, 2006",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, dateStr); err == nil {
					return t
				}
			}
		}
	}

	return time.Now() // Fallback to current time
}

// extractKeywords extracts keywords from meta tags
func (s *Scraper) extractKeywords(e *colly.HTMLElement) string {
	keywords := e.ChildAttr("[name='keywords']", "content")
	if keywords == "" {
		keywords = e.ChildAttr("[property='article:tag']", "content")
	}
	return keywords
}

// cleanContent cleans and formats the extracted content
func (s *Scraper) cleanContent(content string) string {
	// Remove extra whitespace
	content = strings.TrimSpace(content)
	
	// Replace multiple newlines with double newlines
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	
	// Remove common unwanted text
	unwanted := []string{
		"Advertisement",
		"Subscribe",
		"Sign up",
		"Follow us",
		"Share this",
		"Related articles",
	}

	for _, text := range unwanted {
		content = strings.ReplaceAll(content, text, "")
	}

	return content
}

// calculateReadTime estimates reading time in minutes
func (s *Scraper) calculateReadTime(content string) int {
	words := len(strings.Fields(content))
	// Average reading speed: 200 words per minute
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

// calculateScore calculates a relevance score for the article
func (s *Scraper) calculateScore(article *models.Article) float64 {
	score := 0.0
	
	// Content length score
	if len(article.Content) > 500 {
		score += 0.3
	} else if len(article.Content) > 200 {
		score += 0.2
	}
	
	// Title quality score
	if len(article.Title) > 10 && len(article.Title) < 100 {
		score += 0.2
	}
	
	// Author score
	if article.Author != "Unknown Author" {
		score += 0.1
	}
	
	// Source credibility score
	credibleSources := []string{
		"techcrunch.com",
		"wired.com",
		"arstechnica.com",
		"theverge.com",
		"arxiv.org",
		"ieee.org",
		"acm.org",
	}
	
	for _, source := range credibleSources {
		if strings.Contains(article.Source, source) {
			score += 0.2
			break
		}
	}
	
	// Recency score
	daysSincePublished := time.Since(article.PublishedAt).Hours() / 24
	if daysSincePublished < 7 {
		score += 0.2
	} else if daysSincePublished < 30 {
		score += 0.1
	}
	
	return score
}

// ScrapeMultipleURLs scrapes multiple URLs concurrently
func (s *Scraper) ScrapeMultipleURLs(urls []string, mode string) ([]*models.Article, error) {
	articles := make([]*models.Article, 0, len(urls))
	errors := make([]error, 0)
	
	// Use a channel to collect results
	resultChan := make(chan *models.Article, len(urls))
	errorChan := make(chan error, len(urls))
	
	// Start scraping goroutines
	for _, url := range urls {
		go func(u string) {
			article, err := s.ScrapeURL(u, mode)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- article
		}(url)
	}
	
	// Collect results
	for i := 0; i < len(urls); i++ {
		select {
		case article := <-resultChan:
			articles = append(articles, article)
		case err := <-errorChan:
			errors = append(errors, err)
		}
	}
	
	// Log errors but don't fail the entire operation
	if len(errors) > 0 {
		log.Printf("Encountered %d errors during scraping", len(errors))
		for _, err := range errors {
			log.Printf("Scraping error: %v", err)
		}
	}
	
	return articles, nil
}

// ScrapeWithAPI scrapes articles using news APIs
func (s *Scraper) ScrapeWithAPI(query string, limit int) ([]*models.Article, error) {
	var articles []*models.Article
	
	// Try multiple news APIs
	apis := []string{"newsapi", "guardian", "nyt"}
	
	for _, api := range apis {
		if apiArticles, err := s.scrapeFromAPI(api, query, limit); err == nil {
			articles = append(articles, apiArticles...)
			if len(articles) >= limit {
				break
			}
		}
	}
	
	// Limit results
	if len(articles) > limit {
		articles = articles[:limit]
	}
	
	return articles, nil
}

// scrapeFromAPI scrapes from a specific news API
func (s *Scraper) scrapeFromAPI(apiName, query string, limit int) ([]*models.Article, error) {
	switch apiName {
	case "newsapi":
		return s.scrapeNewsAPI(query, limit)
	case "guardian":
		return s.scrapeGuardianAPI(query, limit)
	case "nyt":
		return s.scrapeNYTAPI(query, limit)
	default:
		return nil, fmt.Errorf("unsupported API: %s", apiName)
	}
}

// scrapeNewsAPI scrapes from NewsAPI
func (s *Scraper) scrapeNewsAPI(_ string, _ int) ([]*models.Article, error) {
	// This would require an API key from newsapi.org
	// For now, return empty results
	return []*models.Article{}, nil
}

// scrapeGuardianAPI scrapes from The Guardian API
func (s *Scraper) scrapeGuardianAPI(_ string, _ int) ([]*models.Article, error) {
	// This would require an API key from open-platform.theguardian.com
	// For now, return empty results
	return []*models.Article{}, nil
}

// scrapeNYTAPI scrapes from New York Times API
func (s *Scraper) scrapeNYTAPI(_ string, _ int) ([]*models.Article, error) {
	// This would require an API key from developer.nytimes.com
	// For now, return empty results
	return []*models.Article{}, nil
}

// extractArticleURLsWithRegex extracts article URLs using regex patterns
func (s *Scraper) extractArticleURLsWithRegex(html string, patterns []string) []string {
	var urls []string
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(html, -1)
		urls = append(urls, matches...)
	}
	
	return urls
}

// parseJSONResponse parses JSON response from API
func (s *Scraper) parseJSONResponse(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// makeAPIRequest makes an HTTP request to an API
func (s *Scraper) makeAPIRequest(url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := s.apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return ioutil.ReadAll(resp.Body)
}

// TestSearchURLs tests which search URLs are working
func (s *Scraper) TestSearchURLs() {
	testQuery := "nlp"
	newsQuery := strings.ReplaceAll(testQuery, " ", "+")

	searchURLs := []struct {
		name string
		url  string
	}{
		{"TechCrunch", fmt.Sprintf("https://techcrunch.com/search/%s/", newsQuery)},
		{"ZDNet", fmt.Sprintf("https://www.zdnet.com/search/?searchQuery=%s", newsQuery)},
		{"CNET", fmt.Sprintf("https://www.cnet.com/search/?query=%s", newsQuery)},
		{"The Verge", fmt.Sprintf("https://www.theverge.com/search?q=%s", newsQuery)},
		{"Ars Technica", fmt.Sprintf("https://arstechnica.com/search/?q=%s", newsQuery)},
		{"InfoQ", fmt.Sprintf("https://www.infoq.com/search.action?queryString=%s", newsQuery)},
		{"TechRadar", fmt.Sprintf("https://www.techradar.com/search?searchTerm=%s", newsQuery)},
		{"Forbes", fmt.Sprintf("https://www.forbes.com/search/?q=%s", newsQuery)},
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	fmt.Println("Testing search URLs...")
	fmt.Println("========================")

	for _, site := range searchURLs {
		fmt.Printf("Testing %s: %s\n", site.name, site.url)
		
		resp, err := client.Get(site.url)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", site.name, err)
			continue
		}
		
		defer resp.Body.Close()
		
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ %s: Error reading body - %v\n", site.name, err)
			continue
		}
		
		bodyStr := string(body)
		
		// Check if we got a valid response
		if resp.StatusCode == 200 {
			// Check if the page contains search results (basic heuristics)
			if strings.Contains(bodyStr, "search") || strings.Contains(bodyStr, "result") || 
			   strings.Contains(bodyStr, "article") || strings.Contains(bodyStr, "news") ||
			   len(bodyStr) > 5000 { // Reasonable page size
				fmt.Printf("✅ %s: Working (Status: %d, Size: %d bytes)\n", site.name, resp.StatusCode, len(bodyStr))
			} else {
				fmt.Printf("⚠️  %s: Status OK but may not have results (Status: %d, Size: %d bytes)\n", site.name, resp.StatusCode, len(bodyStr))
			}
		} else {
			fmt.Printf("❌ %s: Status %d (Size: %d bytes)\n", site.name, resp.StatusCode, len(bodyStr))
		}
		
		fmt.Println()
	}
}


