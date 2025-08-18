package scraper

import (
	"testing"
)

func TestExtractDomain(t *testing.T) {
	scraper := &Scraper{}
	
	tests := []struct {
		url    string
		expect string
	}{
		{"https://www.theverge.com/search?q=glasses", "theverge.com"},
		{"https://theverge.com/article/123", "theverge.com"},
		{"http://www.techcrunch.com/post/456", "techcrunch.com"},
		{"https://techcrunch.com/2024/01/01/article", "techcrunch.com"},
		{"https://www.wired.com/story/article", "wired.com"},
		{"https://wired.com/2024/article", "wired.com"},
		{"https://www.arstechnica.com/tech-policy/", "arstechnica.com"},
		{"https://arstechnica.com/science/", "arstechnica.com"},
		{"https://www.venturebeat.com/news/", "venturebeat.com"},
		{"https://venturebeat.com/2024/", "venturebeat.com"},
		{"https://www.technologyreview.com/article/", "technologyreview.com"},
		{"https://technologyreview.com/2024/", "technologyreview.com"},
		{"https://www.ieee.org/spectrum/", "ieee.org"},
		{"https://ieee.org/publications/", "ieee.org"},
		{"https://www.acm.org/publications/", "acm.org"},
		{"https://acm.org/2024/", "acm.org"},
		{"invalid-url", ""},
		{"", ""},
	}
	
	for _, test := range tests {
		result := scraper.extractDomain(test.url)
		if result != test.expect {
			t.Errorf("extractDomain(%s) = %s, want %s", test.url, result, test.expect)
		}
	}
}

func TestNormalizeURL(t *testing.T) {
	scraper := &Scraper{}
	
	tests := []struct {
		url    string
		expect string
	}{
		{"https://www.theverge.com/search?q=glasses", "https://www.theverge.com/search?q=glasses"},
		{"http://theverge.com/article/123", "http://theverge.com/article/123"},
		{"www.techcrunch.com/post/456", "https://www.techcrunch.com/post/456"},
		{"techcrunch.com/2024/01/01/article", "https://techcrunch.com/2024/01/01/article"},
		{"https://wired.com/story/article/", "https://wired.com/story/article"},
		{"https://arstechnica.com/tech-policy", "https://arstechnica.com/tech-policy"},
	}
	
	for _, test := range tests {
		result := scraper.normalizeURL(test.url)
		if result != test.expect {
			t.Errorf("normalizeURL(%s) = %s, want %s", test.url, result, test.expect)
		}
	}
}

func TestGetSourceByURL(t *testing.T) {
	scraper := NewScraper()
	
	tests := []struct {
		url        string
		expectName string
		expectFound bool
	}{
		{"https://www.theverge.com/search?q=glasses", "The Verge", true},
		{"https://theverge.com/article/123", "The Verge", true},
		{"https://www.techcrunch.com/post/456", "TechCrunch", true},
		{"https://techcrunch.com/2024/01/01/article", "TechCrunch", true},
		{"https://www.wired.com/story/article", "Wired", true},
		{"https://wired.com/2024/article", "Wired", true},
		{"https://www.arstechnica.com/tech-policy/", "Ars Technica", true},
		{"https://arstechnica.com/science/", "Ars Technica", true},
		{"https://www.venturebeat.com/news/", "VentureBeat", true},
		{"https://venturebeat.com/2024/", "VentureBeat", true},
		{"https://www.technologyreview.com/article/", "MIT Technology Review", true},
		{"https://technologyreview.com/2024/", "MIT Technology Review", true},
		{"https://www.ieee.org/spectrum/", "IEEE Spectrum", true},
		{"https://ieee.org/publications/", "IEEE Spectrum", true},
		{"https://www.acm.org/publications/", "ACM", true},
		{"https://acm.org/2024/", "ACM", true},
		{"https://unknown-site.com/article", "", false},
		{"https://www.unknown-site.com/article", "", false},
	}
	
	for _, test := range tests {
		source, found := scraper.getSourceByURL(test.url)
		if found != test.expectFound {
			t.Errorf("getSourceByURL(%s) found = %v, want %v", test.url, found, test.expectFound)
		}
		if found && source.Name != test.expectName {
			t.Errorf("getSourceByURL(%s) name = %s, want %s", test.url, source.Name, test.expectName)
		}
	}
}

func TestExtractSource(t *testing.T) {
	scraper := &Scraper{}
	
	tests := []struct {
		url    string
		expect string
	}{
		{"https://www.theverge.com/search?q=glasses", "theverge.com"},
		{"https://theverge.com/article/123", "theverge.com"},
		{"http://www.techcrunch.com/post/456", "techcrunch.com"},
		{"https://techcrunch.com/2024/01/01/article", "techcrunch.com"},
		{"https://www.wired.com/story/article", "wired.com"},
		{"https://wired.com/2024/article", "wired.com"},
		{"invalid-url", "Unknown Source"},
		{"", "Unknown Source"},
	}
	
	for _, test := range tests {
		result := scraper.extractSource(test.url)
		if result != test.expect {
			t.Errorf("extractSource(%s) = %s, want %s", test.url, result, test.expect)
		}
	}
}
