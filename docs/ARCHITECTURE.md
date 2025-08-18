# Spectechle Architecture

## System overview

Spectechle is a microservices-based application that combines Go and Python to create a tech search engine and summarizer. The architecture prioritizes simplicity, maintainability, and performance.

## High-level architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Go Backend    │    │  Python NLP     │
│   (Templates)   │◄──►│   (API Server)  │◄──►│   (ML Pipeline) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   SQLite/DB     │
                       │   (Articles)    │
                       └─────────────────┘
```

## Component breakdown

### 1. Frontend layer
- Technology: Go HTML templates + vanilla JavaScript
- Purpose: User interface for search and results display
- Features:
  - Responsive layout
  - Search with loading states
  - Modal-based article details
  - Category filtering and mode selection

### 2. Go backend service
- Technology: Go 1.21+ (Gin)
- Purpose: API server, web scraping, orchestration
- Components:
  - REST API endpoints
  - Web scraper (Colly)
  - Database access
  - Communication with Python NLP service

### 3. Python NLP service
- Technology: Python 3.9+ (Flask)
- Purpose: Machine learning pipeline for text processing
- Components:
  - Text classification
  - Text summarization (BART-CNN)
  - Keyword extraction
  - Model management and caching

### 4. Data layer
- Technology: SQLite (development) / PostgreSQL (production)
- Purpose: Storage for articles and metadata
- Schema: Optimized for search and categorization

## Scalability design

### Horizontal scaling
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │───►│  Go Backend 1   │    │  Go Backend 2   │
│   (Nginx)       │    │   (Port 8080)   │    │   (Port 8081)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │  Python NLP 1   │    │  Python NLP 2   │
                       │   (Port 5000)   │    │   (Port 5001)   │
                       └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────────────────────────────┐
                       │           Shared Database               │
                       │        (PostgreSQL Cluster)            │
                       └─────────────────────────────────────────┘
```

### Caching strategy
- Redis for search results and model outputs
- Local caching for HuggingFace models
- CDN for static assets (CSS, JS)

### Database scaling
- Read replicas for search-heavy workloads
- Optional sharding by category/date
- Connection pooling

## Data flow

### Search flow
1. User submits a query on the frontend
2. Go backend receives the request
3. Database is queried for existing results
4. Optional real-time scraping if needed
5. NLP processing for classification and summarization
6. Response returned to client

### Scraping flow
1. URL discovery based on query
2. Content extraction using Colly
3. Text cleaning and normalization
4. NLP processing (classify/summarize)
5. Persist to database with metadata

## Security considerations

### API security
- Rate limiting
- CORS configuration
- Input validation
- Authentication (JWT) as a future enhancement

### Data security
- Parameterized queries
- XSS prevention via sanitization
- HTTPS for transport security
- Data privacy considerations

## Performance optimization

### Backend
- Concurrent scraping via goroutines
- Connection pooling (DB and HTTP)
- Caching hot paths
- Response compression (gzip)

### NLP
- Pre-loaded models
- Batch processing where applicable
- Optional GPU acceleration (CUDA)
- Optional model quantization

### Frontend
- Lazy loading
- Minified assets
- CDN delivery
- Progressive enhancement

## Configuration management

### Environment variables
```bash
# Backend
PORT=8080
DB_PATH=./data/spectechle.db
NLP_SERVICE_URL=http://localhost:5000
GIN_MODE=release

# NLP Service
FLASK_ENV=production
FLASK_DEBUG=false
NLP_PORT=5000
HF_HOME=./models
TRANSFORMERS_CACHE=./models

# Database
DB_TYPE=postgresql
DB_HOST=localhost
DB_PORT=5432
DB_NAME=spectechle
DB_USER=spectechle
DB_PASSWORD=secure_password
```

### Feature flags
- Enable/disable scraping
- Select ML model type
- Cache TTLs
- API rate limits

## Testing strategy

### Unit tests
- Go: standard testing + testify
- Python: unittest + pytest

### Integration tests
- End-to-end API tests
- Persistence tests
- NLP pipeline tests (sanity/consistency)

### Performance tests
- Load and stress testing
- Benchmark critical paths

## Monitoring and observability

### Metrics
- Application: latency, throughput, error rate
- System: CPU, memory, disk
- Business: search volume, article yield

### Logging
- Structured JSON logs
- Log levels (DEBUG, INFO, WARN, ERROR)
- Centralized aggregation (ELK or similar)

### Alerting
- Service health and uptime
- Latency/error thresholds
- Resource saturation

## Deployment architecture

### Development
```
┌─────────────────┐    ┌─────────────────┐
│   Local Go      │    │   Local Python  │
│   Backend       │◄──►│   NLP Service   │
└─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   SQLite DB     │
                       └─────────────────┘
```

### Production
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CDN           │    │   Load Balancer │    │   Auto Scaling  │
│   (Static)      │    │   (Nginx)       │    │   Group         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Go Backend    │    │   Python NLP    │
                       │   (Multiple)    │    │   (Multiple)    │
                       └─────────────────┘    └─────────────────┘
                              │                        │
                              ▼                        ▼
                       ┌─────────────────────────────────────────┐
                       │           Database Cluster              │
                       │        (PostgreSQL + Redis)            │
                       └─────────────────────────────────────────┘
```

## Future enhancements

### Scalability
- Kubernetes orchestration
- Service mesh (Istio)
- Event streaming (Kafka)
- Optional GraphQL gateway

### Features
- Authentication and user preferences
- Advanced search features
- Real-time updates (WebSocket)
- Mobile client

### AI/ML
- Domain-tuned models
- Content recommendation
- Sentiment analysis
- Topic modeling

## Technology stack summary

| Component | Technology | Purpose |
|-----------|------------|---------|
| Frontend | Go Templates + JavaScript | User interface |
| Backend | Go + Gin | API server and scraping |
| NLP | Python + Flask + Transformers | ML pipeline |
| Database | SQLite/PostgreSQL | Data storage |
| Caching | Redis | Performance |
| Containerization | Docker + Docker Compose | Deployment |
| Monitoring | Prometheus + Grafana | Observability |
| CI/CD | GitHub Actions | Automation |
