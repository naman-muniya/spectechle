# Spectechle - Project highlights

## Overview

Spectechle is a tech news aggregator with AI-powered summarization, keyword extraction, and a clean user interface. The goal is to deliver relevant content with clear feedback and robust error handling.

Tech stack: Go, Python, JavaScript, SQLite/PostgreSQL, Docker, HuggingFace Transformers (BART-CNN)

## Capabilities

### AI/ML
- BART-CNN summarization via HuggingFace Transformers
- Keyword extraction with word-boundary matching and technical term heuristics
- Content cleaning to exclude sidebars and unrelated sections
- Paywall-aware fallbacks that summarize from available title/preview

### Architecture
- Go backend API with concurrent scraping (Colly)
- Python NLP service (Flask) for summarization and text processing
- REST communication between services with structured error handling
- SQLite for development, PostgreSQL for production
- Dockerized services for local and production environments

### Frontend
- Go templates with vanilla JavaScript
- Loading overlays and in-modal progress indicators
- Modal article view with content, categories, and keywords
- Defensive UI against missing or empty content

### Reliability and performance
- Input validation across endpoints
- Timeouts for external calls and summarization
- Caching strategy for repeat results and model artifacts
- Connection pooling and concurrent operations

## Selected technical solutions

1) Keyword accuracy without false matches
- Problem: Partial matches (e.g., "seamlessly" triggering "ml")
- Solution: Word-boundary regex patterns and technical-term filters

2) Paywalled or limited content
- Problem: Hidden body content while still needing a summary
- Solution: Use available title and preview text, with clear UI messaging

3) Loading and user feedback
- Problem: No visual cue during summarization or modal load
- Solution: Overlay loading state before modal open; in-modal spinner for summarization

4) Content contamination
- Problem: Keywords pulled from unrelated sidebars or comment sections
- Solution: Content cleaning with sidebar/ad patterns and HTML removal

## Results
- Summarization requests protected by timeouts and abort handling
- Keyword extraction focused on main article content
- Improved user experience for slow or empty content cases
- Architecture ready for horizontal scaling and service separation
