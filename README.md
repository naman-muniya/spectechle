# Spectechle - Smart Tech Search Engine & Summarizer

A sophisticated tech news aggregator with AI-powered summarization, intelligent keyword extraction, and beautiful user experience.

## Demo
https://github.com/user-attachments/assets/6099bc9b-d0d2-47dd-a553-e520c1d03e7d


## 🚀 Features

### **🎯 Smart Search & Discovery**
- **Dual Search Modes**: Tech news and Research papers (arXiv, IEEE)
- **Real-time Scraping**: Live article collection from multiple sources
- **Category Filtering**: AI, Cloud, Security, Data Science, DevOps, and more
- **Live Progress Tracking**: Real-time search progress with beautiful UI

### **🧠 AI-Powered Intelligence**
- **BART-CNN Summarization**: High-quality article summaries using local AI models
- **Smart Classification**: Automatic categorization of tech content
- **Intelligent Keyword Extraction**: Context-aware keyword detection with sidebar filtering
- **Paywall Handling**: Graceful handling of subscription-required content

### **✨ Beautiful User Experience**
- **Instant Loading States**: Professional loading overlays with AI-themed messaging
- **Smooth Animations**: Fade-in, scale, and transition effects
- **Responsive Design**: Works perfectly on all devices
- **Error Recovery**: Graceful fallbacks and user-friendly error messages

### **🔧 Robust Architecture**
- **Microservices**: Separate Go backend and Python NLP services
- **Concurrent Processing**: Parallel scraping and AI processing
- **Smart Caching**: Intelligent result caching for performance
- **Timeout Handling**: Prevents hanging with proper timeouts

## 🏗️ Architecture

```
Spectechle/
├── backend/          # Go API server + scraper
├── nlp/             # Python NLP pipeline + BART-CNN summarizer
├── frontend/        # Modern web UI with beautiful animations
├── docs/           # Comprehensive documentation
└── tests/          # Unit and integration tests
```

### Tech Stack

- **Backend**: Go (Fast API server, concurrent scraping)
- **NLP Engine**: Python (BART-CNN, Transformers, classification)
- **Database**: SQLite (development) / PostgreSQL (production)
- **Communication**: REST API between services
- **Frontend**: Modern HTML5 + CSS3 + Vanilla JavaScript
- **AI Models**: HuggingFace Transformers (BART-CNN for summarization)

## 🛠️ Quick Start

See **[SETUP.md](docs/SETUP.md)** for detailed installation and setup instructions.

### Quick Commands

```bash
# Clone and setup
git clone <your-repo>
cd spectechle

# Start services
cd nlp && python -m venv venv && source venv/bin/activate && pip install -r requirements.txt && python app.py
# In new terminal: cd backend && go mod tidy && go run main.go

# Or use Docker
docker-compose up --build
```

**Access**: http://localhost:8080

## 🎨 Key Features in Detail

### **Smart Keyword Extraction**
- **Precise Pattern Matching**: Uses regex with word boundaries to avoid false matches
- **Technical Term Recognition**: Identifies tech-specific terms and patterns
- **Content Cleaning**: Removes sidebar content and advertisements
- **Paywall Handling**: Extracts keywords from available preview content

### **Beautiful Loading Experience**
- **Instant Feedback**: Loading overlay appears immediately on article click
- **AI-Themed Messaging**: "🤖 Preparing AI-powered insights..."
- **Smooth Animations**: Professional fade-in and scale effects
- **Timeout Protection**: 30-second timeout prevents hanging

### **Intelligent Content Processing**
- **Sidebar Filtering**: Removes navigation, ads, and related articles
- **Main Content Focus**: Prioritizes the first 2000 characters of article body
- **Subscription Detection**: Identifies paywalled content automatically
- **Graceful Fallbacks**: Shows appropriate messages for different content states

## 📁 Project Structure

```
Spectechle/
├── backend/          # Go API server + scraper
│   ├── main.go       # API server entry point
│   ├── scraper/      # Web scraping logic with concurrent processing
│   ├── api/          # REST API handlers with error handling
│   ├── models/       # Data structures and validation
│   └── database/     # Database operations with caching
├── nlp/              # Python NLP pipeline + BART-CNN summarizer
│   ├── app.py        # Flask API server with comprehensive error handling
│   ├── classifier/   # Article classification with tech categories
│   ├── summarizer/   # BART-CNN summarization with input validation
│   └── models/       # ML model loading and inference
├── frontend/         # Modern web UI with beautiful animations
│   ├── templates/    # Modern HTML templates
│   ├── static/css/   # Beautiful CSS with animations and responsive design
│   └── static/js/    # Vanilla JavaScript with error handling and loading states
├── docs/             # Comprehensive documentation
│   ├── SETUP.md      # Installation and setup guide
│   ├── ARCHITECTURE.md # Technical architecture details
│   └── HIGHLIGHTS.md # Project highlights and technical achievements
└── tests/            # Unit and integration tests
```

## 📚 Documentation

- **[SETUP.md](docs/SETUP.md)** - Quick start guide and deployment instructions
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Technical architecture and system design
- **[HIGHLIGHTS.md](docs/HIGHLIGHTS.md)** - Project highlights and technical achievements

## 🧪 Testing

```bash
# Go tests
cd backend
go test ./...

# Python tests
cd nlp
python -m pytest tests/

# Run all tests
cd tests
python -m pytest python/
cd ../backend
go test ./...
```

## 📊 Performance & Scalability

- **Concurrent Scraping**: Go goroutines for parallel data collection
- **Smart Caching**: Intelligent result caching for performance
- **Connection Pooling**: Database connection optimization
- **Load Balancing**: Ready for horizontal scaling
- **Containerization**: Docker for easy deployment
- **Health Checks**: Built-in monitoring and metrics

## 🎯 Recent Improvements

### **v2.1 - Multi-Source Scraping Enhancement**
- ✅ Enhanced scraper with support for 10+ tech news sources
- ✅ Improved URL extraction for TechRadar, Forbes, ZDNet, CNET, InfoQ, and more
- ✅ Better article detection and filtering algorithms
- ✅ Robust error handling for different site structures
- ✅ Comprehensive logging for debugging and monitoring

### **v2.0 - Enhanced User Experience**
- ✅ Beautiful loading overlays with AI-themed messaging
- ✅ Smart keyword extraction with sidebar filtering
- ✅ Paywall detection and graceful handling
- ✅ Comprehensive error handling and recovery
- ✅ Smooth animations and transitions
- ✅ Responsive design improvements
- ✅ Timeout protection for all operations
- ✅ Professional error messages and fallbacks

### **v1.0 - Core Features**
- ✅ BART-CNN summarization pipeline
- ✅ Real-time article scraping
- ✅ Tech category classification
- ✅ Modern web interface
- ✅ Microservices architecture
- ✅ Docker containerization

## 📝 Development Roadmap

- [x] Project scaffold and architecture
- [x] Basic Go API server
- [x] Python NLP service with BART-CNN
- [x] Web scraping implementation
- [x] Article classification
- [x] Summarization pipeline
- [x] Frontend UI with loading states
- [x] Smart keyword extraction
- [x] Paywall handling
- [x] Error handling and recovery
- [x] Performance optimization
- [x] Comprehensive testing
- [x] Documentation and deployment guide
- [x] Docker containerization
- [x] Beautiful animations and UX
- [x] Multi-source scraping enhancement
- [x] Improved article extraction algorithms

---

**Built with modern software engineering practices and a focus on user experience.**

