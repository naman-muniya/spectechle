# Spectechle setup guide

## Quick start

### Prerequisites
- Go 1.21+
- Python 3.9+
- Git

### 1) Clone and setup
```bash
git clone <your-repo-url>
cd spectechle
```

### 2) Start Python NLP service
```bash
cd nlp
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
python app.py
```

### 3) Start Go backend (new terminal)
```bash
cd backend
go mod tidy
go run main.go
```

### 4) Access the application
- Web UI: http://localhost:8080
- API: http://localhost:8080/api
- NLP service: http://localhost:5000

## Docker setup (optional)

### Single command
```bash
docker-compose up --build
```

### Access
- Web UI: http://localhost:8080
- All services are pre-configured

## Configuration

### Environment variables
```bash
# Backend
PORT=8080
NLP_SERVICE_URL=http://localhost:5000

# NLP service
FLASK_ENV=development
NLP_PORT=5000
```

### Customization
- Search sources: edit `backend/api/handler.go`
- Categories: edit `frontend/templates/index.html`
- Keywords: update `frontend/static/js/app.js`

## Performance tips

### Production
1. Use PostgreSQL
2. Add Redis caching
3. Enable HTTPS (nginx/Traefik)
4. Add monitoring (Prometheus/Grafana)

### Development
1. Use SQLite
2. Enable debug logs
3. Prefer smaller models for speed

## Troubleshooting

### NLP service not starting
```bash
python --version  # 3.9+
pip install -r requirements.txt --force-reinstall
```

### Go backend issues
```bash
go mod tidy
go mod download
```

### Port conflicts
```bash
lsof -i :8080
lsof -i :5000
kill -9 <PID>
```

### Logs
- Backend: terminal output
- NLP service: terminal output
- Frontend: browser console (F12)
