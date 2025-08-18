/**
 * Spectechle Frontend JavaScript
 * Handles search functionality, API calls, and UI interactions
 */

class SpectechleApp {
    constructor() {
        this.apiBaseUrl = '/api';
        this.nlpApiUrl = 'http://localhost:5000';
        this.currentSearchId = null;
        this.selectedCategories = new Set();
        this.currentMode = 'news';
        
        this.initializeEventListeners();
        this.initializeUI();
    }

    initializeEventListeners() {
        // Search functionality
        const searchBtn = document.getElementById('searchBtn');
        const searchInput = document.getElementById('searchInput');
        
        searchBtn.addEventListener('click', () => this.performSearch());
        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.performSearch();
            }
        });

        // Stop polling when page is unloaded
        window.addEventListener('beforeunload', () => {
            if (this.pollingInterval) {
                clearInterval(this.pollingInterval);
            }
        });

        // Keyboard handlers
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeModal();
            }
        });

        // Mode selection
        const modeButtons = document.querySelectorAll('.mode-btn');
        modeButtons.forEach(btn => {
            btn.addEventListener('click', () => {
                this.setActiveMode(btn.dataset.mode);
            });
        });

        // Category selection
        const categoryChips = document.querySelectorAll('.category-chip');
        categoryChips.forEach(chip => {
            chip.addEventListener('click', () => {
                this.toggleCategory(chip);
            });
        });

        // Modal functionality
        const modalClose = document.getElementById('modalClose');
        const modal = document.getElementById('articleModal');
        const loadingOverlay = document.getElementById('articleLoadingOverlay');
        
        modalClose.addEventListener('click', () => this.closeModal());
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                this.closeModal();
            }
        });
        
        // Close loading overlay if clicked outside
        loadingOverlay.addEventListener('click', (e) => {
            if (e.target === loadingOverlay) {
                loadingOverlay.classList.add('hidden');
            }
        });

        // Retry button
        const retryBtn = document.getElementById('retryBtn');
        retryBtn.addEventListener('click', () => this.performSearch());


    }

    initializeUI() {
        // Set default mode
        this.setActiveMode('news');
        
        // Show initial state
        this.showSearchState();
    }

    setActiveMode(mode) {
        this.currentMode = mode;
        
        // Update UI
        document.querySelectorAll('.mode-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-mode="${mode}"]`).classList.add('active');
        
        // Update placeholder text
        const searchInput = document.getElementById('searchInput');
        if (mode === 'research') {
            searchInput.placeholder = 'Search for research papers, academic articles...';
        } else {
            searchInput.placeholder = 'Search for tech articles, blogs, or news...';
        }
    }

    toggleCategory(chip) {
        const category = chip.dataset.category;
        
        if (this.selectedCategories.has(category)) {
            this.selectedCategories.delete(category);
            chip.classList.remove('selected');
        } else {
            this.selectedCategories.add(category);
            chip.classList.add('selected');
        }
    }

    async performSearch() {
        const query = document.getElementById('searchInput').value.trim();
        
        if (!query) {
            this.showError('Please enter a search query');
            return;
        }

        // Stop any existing polling
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
        }

        this.showLoadingState();

        try {
            const searchData = {
                query: query,
                mode: this.currentMode,
                categories: Array.from(this.selectedCategories),
                limit: 20
            };

            const response = await fetch(`${this.apiBaseUrl}/search`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(searchData)
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            this.currentSearchId = result.id;
            
            // Handle the initial search response
            const initialCount = (result.articles && result.articles.length) ? result.articles.length : 0;
            
            // If we have articles, display them
            if (result.articles && result.articles.length > 0) {
                this.updateProgressBar(initialCount, 20);
                this.displayResults(result.articles, result.status === 'searching');
                
                // Start polling if still searching
                if (result.status === 'searching' && initialCount < 20) {
                    this.startPollingForUpdates(result.id, query);
                }
            } else {
                // No articles yet - show searching state instead of no results
                if (result.status === 'searching') {
                    this.showSearchingState(query);
                    this.startPollingForUpdates(result.id, query);
                } else {
                    // If status is not searching and no articles, then show no results
                    this.showNoResults(query);
                }
            }

        } catch (error) {
            console.error('Search error:', error);
            this.showError('Failed to perform search. Please try again.');
        }
    }

    async displayResults(articles, inProgress = false) {
        const resultsContainer = document.getElementById('resultsContainer');
        const resultCount = document.getElementById('resultCount');
        
        resultCount.textContent = articles.length;
        
        // Clear previous results
        resultsContainer.innerHTML = '';
        
        // Process each article
        for (const article of articles) {
            const articleCard = await this.createArticleCard(article);
            resultsContainer.appendChild(articleCard);
        }
        
        // When still searching, keep the progress bar visible
        if (inProgress) {
            document.getElementById('loadingState').classList.add('hidden');
            document.getElementById('resultsSection').classList.remove('hidden');
            document.getElementById('errorState').classList.add('hidden');
        } else {
            this.showResultsState();
        }
    }

    async createArticleCard(article) {
        const card = document.createElement('div');
        card.className = 'article-card';
        
        // Format date
        const date = new Date(article.published_at || article.scraped_at);
        const formattedDate = date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });

        // Create categories HTML
        const categoriesHtml = article.categories ? 
            article.categories.map(cat => `<span class="category-tag">${cat}</span>`).join('') : '';

        // Create score HTML
        const scoreHtml = article.score ? 
            `<div class="article-score">
                <i class="fas fa-star"></i>
                <span>${article.score.toFixed(1)}</span>
            </div>` : '';

        card.innerHTML = `
            <div class="article-header">
                <div>
                    <h3 class="article-title">${this.escapeHtml(article.title)}</h3>
                    <div class="article-meta">
                        <span class="source">${this.escapeHtml(article.source)}</span>
                        <span class="date">${formattedDate}</span>
                        <span class="read-time">
                            <i class="fas fa-clock"></i>
                            ${article.read_time || '5'} min read
                        </span>
                    </div>
                </div>
                ${scoreHtml}
            </div>
            <div class="article-categories">
                ${categoriesHtml}
            </div>
            <div class="article-summary">
                ${this.escapeHtml(article.summary || this.getArticlePreview(article))}
            </div>
        `;

        // Add click event to show modal with loading state
        card.addEventListener('click', () => {
            this.showArticleModalWithLoading(article);
        });

        return card;
    }

    async showArticleModalWithLoading(article) {
        console.log('🔄 Starting article modal with loading...', article);
        
        // Show loading overlay immediately
        const loadingOverlay = document.getElementById('articleLoadingOverlay');
        console.log('Loading overlay element:', loadingOverlay);
        
        if (!loadingOverlay) {
            console.error('❌ Loading overlay element not found!');
            // Fallback to direct modal opening
            await this.showArticleModal(article);
            return;
        }
        
        console.log('✅ Showing loading overlay...');
        loadingOverlay.classList.remove('hidden');
        loadingOverlay.style.display = 'flex';
        
        try {
            // Small delay to show loading state (makes it feel more responsive)
            await new Promise(resolve => setTimeout(resolve, 500));
            
            // Show the actual modal
            console.log('🔄 Calling showArticleModal...');
            await this.showArticleModal(article);
            console.log('✅ showArticleModal completed successfully');
        } catch (error) {
            console.error('❌ Error showing article modal:', error);
            console.error('Error stack:', error.stack);
            
            // Show error in modal instead of hiding overlay
            const modal = document.getElementById('articleModal');
            const modalTitle = document.getElementById('modalTitle');
            const modalSummary = document.getElementById('modalSummary');
            
            if (modal && modalTitle && modalSummary) {
                modalTitle.textContent = 'Error Opening Article';
                modalSummary.innerHTML = `
                    <div class="article-content-empty">
                        <i class="fas fa-exclamation-triangle"></i>
                        <h4>Failed to Load Article</h4>
                        <p>There was an error opening this article. Please try again.</p>
                        <p><small>Error: ${error.message}</small></p>
                    </div>
                `;
                modal.classList.remove('hidden');
                document.body.style.overflow = 'hidden';
            }
        } finally {
            // Hide loading overlay
            console.log('✅ Hiding loading overlay...');
            loadingOverlay.classList.add('hidden');
            loadingOverlay.style.display = 'none';
        }
    }

    async showArticleModal(article) {
        console.log('🔄 showArticleModal called with article:', article);
        
        // Validate article data
        if (!article) {
            throw new Error('No article data provided');
        }
        
        const modal = document.getElementById('articleModal');
        const modalTitle = document.getElementById('modalTitle');
        const modalSource = document.getElementById('modalSource');
        const modalDate = document.getElementById('modalDate');
        const modalReadTime = document.getElementById('modalReadTime');
        const modalCategories = document.getElementById('modalCategories');
        const modalSummary = document.getElementById('modalSummary');
        const modalSummaryLoading = document.getElementById('modalSummaryLoading');
        const modalKeywords = document.getElementById('modalKeywords');
        const modalUrl = document.getElementById('modalUrl');

        // Check if all required elements exist
        if (!modal || !modalTitle || !modalSummary) {
            throw new Error('Required modal elements not found');
        }

        // Debug: Log article structure
        console.log('Article data:', {
            title: article.title,
            hasContent: !!(article.Content || article.content),
            contentLength: (article.Content || article.content || '').length,
            hasSummary: !!article.summary,
            summaryLength: (article.summary || '').length
        });

        // Check if article has any content at all
        const content = article.Content || article.content || '';
        const hasAnyContent = content && content.trim().length > 0;
        const hasSummary = article.summary && article.summary.trim().length > 0;

        // Set basic info
        modalTitle.textContent = article.title || 'Untitled Article';
        modalSource.textContent = article.source || 'Unknown Source';
        modalDate.textContent = new Date(article.published_at || article.scraped_at || Date.now()).toLocaleDateString();
        modalReadTime.textContent = `${article.read_time || '5'} min read`;
        modalUrl.href = article.url || '#';

        // Set categories
        try {
            let categories = [];
            if (article.categories) {
                if (Array.isArray(article.categories)) {
                    categories = article.categories;
                } else if (typeof article.categories === 'string') {
                    // If categories is a comma-separated string, split it
                    categories = article.categories.split(',').map(c => c.trim()).filter(c => c.length > 0);
                }
            }
            
            if (categories.length > 0) {
                modalCategories.innerHTML = categories
                    .map(cat => `<span class="category-tag">${this.escapeHtml(cat)}</span>`)
                    .join('');
            } else {
                modalCategories.innerHTML = '<span class="category-tag">General</span>';
            }
        } catch (error) {
            console.error('Error processing categories:', error);
            modalCategories.innerHTML = '<span class="category-tag">General</span>';
        }

        // Handle summary section
        if (hasSummary) {
            // Article already has a summary
            modalSummary.textContent = article.summary;
            modalSummaryLoading.classList.add('hidden');
        } else if (hasAnyContent) {
            // Show loading state for summarization
            modalSummary.textContent = '';
            modalSummaryLoading.classList.remove('hidden');
            
            try {
                console.log('Processing article with NLP on-demand:', {
                    articleId: article.id,
                    hasContent: !!content,
                    contentLength: content.length
                });
                
                // Call our backend endpoint for on-demand NLP processing
                const controller = new AbortController();
                const timeoutId = setTimeout(() => controller.abort(), 60000); // 60 second timeout
                
                const nlpResponse = await fetch(`${this.apiBaseUrl}/articles/${article.id}/process-nlp`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    signal: controller.signal
                });
                
                clearTimeout(timeoutId);
                
                console.log('NLP processing response:', {
                    status: nlpResponse.status,
                    statusText: nlpResponse.statusText,
                    ok: nlpResponse.ok
                });
                
                // Hide loading state
                modalSummaryLoading.classList.add('hidden');
                
                if (nlpResponse.ok) {
                    const nlpData = await nlpResponse.json();
                    console.log('NLP processing data:', nlpData);
                    
                    if (nlpData.article && nlpData.article.summary) {
                        modalSummary.textContent = nlpData.article.summary;
                        
                        // Update categories if available
                        if (nlpData.article.category) {
                            const categories = nlpData.article.category.split(',').map(c => c.trim()).filter(c => c.length > 0);
                            if (categories.length > 0) {
                                modalCategories.innerHTML = categories
                                    .map(cat => `<span class="category-tag">${this.escapeHtml(cat)}</span>`)
                                    .join('');
                            }
                        }
                        
                        // Update keywords if available
                        if (nlpData.article.keywords) {
                            const keywords = nlpData.article.keywords.split(',').map(k => k.trim()).filter(k => k.length > 0);
                            if (keywords.length > 0) {
                                modalKeywords.innerHTML = keywords
                                    .map(keyword => `<span class="keyword-tag">${this.escapeHtml(keyword)}</span>`)
                                    .join('');
                            }
                        }
                    } else {
                        modalSummary.textContent = 'Unable to generate summary at this time.';
                    }
                } else {
                    console.error('NLP processing error:', nlpResponse.status, nlpResponse.statusText);
                    const errorText = await nlpResponse.text();
                    console.error('Error response body:', errorText);
                    modalSummary.textContent = content.substring(0, 200) + '...';
                }
            } catch (error) {
                console.error('NLP processing error:', error);
                modalSummaryLoading.classList.add('hidden');
                
                if (error.name === 'AbortError') {
                    modalSummary.textContent = 'NLP processing timed out. Showing article preview instead.';
                } else {
                    modalSummary.textContent = content.substring(0, 200) + '...';
                }
            }
        } else {
            // No content available (likely paywalled)
            modalSummaryLoading.classList.add('hidden');
            modalSummary.innerHTML = `
                <div class="article-content-empty">
                    <i class="fas fa-lock"></i>
                    <h4>Subscription Required</h4>
                    <p>This article requires a subscription to view the full content. Click the link below to read the article and subscribe if interested.</p>
                    <p><small>Keywords extracted from available preview content.</small></p>
                </div>
            `;
        }

        // Set keywords
        try {
            // Handle different keyword formats
            let keywords = [];
            if (article.keywords) {
                if (Array.isArray(article.keywords)) {
                    keywords = article.keywords;
                } else if (typeof article.keywords === 'string') {
                    // If keywords is a comma-separated string, split it
                    keywords = article.keywords.split(',').map(k => k.trim()).filter(k => k.length > 0);
                }
            }
            
            if (keywords.length > 0) {
                modalKeywords.innerHTML = keywords
                    .slice(0, 10)
                    .map(keyword => `<span class="keyword-tag">${this.escapeHtml(keyword)}</span>`)
                    .join('');
            } else {
                // Extract keywords from content if not available
                const extractedKeywords = this.extractKeywordsFromContent(content);
                if (extractedKeywords.length > 0) {
                    modalKeywords.innerHTML = extractedKeywords
                        .slice(0, 8)
                        .map(keyword => `<span class="keyword-tag">${this.escapeHtml(keyword)}</span>`)
                        .join('');
                } else {
                    modalKeywords.innerHTML = '<span class="keyword-tag">No keywords available</span>';
                }
            }
        } catch (error) {
            console.error('Error processing keywords:', error);
            modalKeywords.innerHTML = '<span class="keyword-tag">No keywords available</span>';
        }

        // Show modal
        modal.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
        
        console.log('✅ Modal opened successfully');
    }

    closeModal() {
        const modal = document.getElementById('articleModal');
        const loadingOverlay = document.getElementById('articleLoadingOverlay');
        
        modal.classList.add('hidden');
        loadingOverlay.classList.add('hidden');
        document.body.style.overflow = 'auto';
    }

    showLoadingState() {
        console.log('Showing loading state...');
        document.getElementById('loadingState').classList.remove('hidden');
        document.getElementById('resultsSection').classList.add('hidden');
        document.getElementById('errorState').classList.add('hidden');
        
        // Show progress bar independently
        this.showProgressBar();
    }

    showProgressBar() {
        const progressContainer = document.getElementById('progressContainer');
        console.log('Progress container:', progressContainer);
        if (progressContainer) {
            progressContainer.classList.remove('hidden');
            progressContainer.style.display = 'block';
            progressContainer.style.opacity = '1';
            console.log('Progress container shown');
        } else {
            console.error('Progress container not found!');
        }
        this.resetProgressBar();
    }

    hideProgressBar() {
        const progressContainer = document.getElementById('progressContainer');
        if (progressContainer) {
            progressContainer.classList.add('hidden');
            progressContainer.style.display = 'none';
        }
    }

    showResultsState() {
        document.getElementById('loadingState').classList.add('hidden');
        document.getElementById('resultsSection').classList.remove('hidden');
        document.getElementById('errorState').classList.add('hidden');
        
        // Hide progress bar
        this.hideProgressBar();
    }

    showSearchState() {
        document.getElementById('loadingState').classList.add('hidden');
        document.getElementById('resultsSection').classList.add('hidden');
        document.getElementById('errorState').classList.add('hidden');
    }

    showError(message) {
        document.getElementById('loadingState').classList.add('hidden');
        document.getElementById('resultsSection').classList.add('hidden');
        document.getElementById('errorState').classList.remove('hidden');
        document.getElementById('errorMessage').textContent = message;
    }

    showSearchingState(query) {
        // Show loading state with progress bar
        document.getElementById('loadingState').classList.remove('hidden');
        document.getElementById('resultsSection').classList.add('hidden');
        document.getElementById('errorState').classList.add('hidden');
        
        // Show progress bar
        this.showProgressBar();
        
        // Update progress text to show we're searching
        const progressText = document.getElementById('progressText');
        if (progressText) {
            progressText.textContent = `🔍 Searching for "${query}"...`;
        }
    }

    showNoResults(query) {
        document.getElementById('loadingState').classList.add('hidden');
        document.getElementById('resultsSection').classList.remove('hidden');
        document.getElementById('errorState').classList.add('hidden');
        
        const resultsContainer = document.getElementById('resultsContainer');
        const resultCount = document.getElementById('resultCount');
        
        resultCount.textContent = '0';
        resultsContainer.innerHTML = `
            <div class="no-results">
                <i class="fas fa-search"></i>
                <h3>No results found</h3>
                <p>No articles found for "${query}". Try different keywords or categories.</p>
            </div>
        `;
    }

    startPollingForUpdates(searchId, query) {
        // Clear any existing polling
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
        }

        let pollCount = 0;
        const maxPolls = 30; // Poll for up to 5 minutes (30 * 10 seconds)
        const targetArticles = 20; // Target number of articles
        
        this.pollingInterval = setInterval(async () => {
            pollCount++;
            
            try {
                // Search again with the same parameters to get updated results
                const searchData = {
                    query: query,
                    mode: this.currentMode,
                    categories: Array.from(this.selectedCategories),
                    limit: targetArticles
                };

                const response = await fetch(`${this.apiBaseUrl}/search`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(searchData)
                });

                if (response.ok) {
                    const result = await response.json();
                    
                    // Update progress bar with current article count
                    this.updateProgressBar(result.articles.length, targetArticles);
                    
                    // Update results; keep progress bar if still searching
                    if (result.articles && result.articles.length > 0) {
                        this.displayResults(result.articles, result.status === 'searching');
                    } else if (result.status === 'searching') {
                        // Still searching but no results yet - keep showing searching state
                        this.showSearchingState(query);
                    }
                    
                    // Stop polling if completed, target reached, or max polls reached
                    if (result.status === 'completed' ||
                        result.articles.length >= targetArticles ||
                        pollCount >= maxPolls) {
                        clearInterval(this.pollingInterval);
                        this.pollingInterval = null;
                        
                        // Show completion message
                        this.showProgressCompletion(result.articles.length, targetArticles);
                        
                        // If no articles found after completion, show no results
                        if (result.articles.length === 0) {
                            this.showNoResults(query);
                        }
                    }
                }
            } catch (error) {
                console.error('Polling error:', error);
                // Stop polling on error but don't immediately show error state
                // Let the user retry manually
                clearInterval(this.pollingInterval);
                this.pollingInterval = null;
                
                // Only show error if we have no results at all
                const resultsContainer = document.getElementById('resultsContainer');
                if (resultsContainer && resultsContainer.children.length === 0) {
                    this.showProgressError();
                }
            }
        }, 10000); // Poll every 10 seconds
    }

    resetProgressBar() {
        const progressBar = document.getElementById('searchProgressBar');
        const progressText = document.getElementById('progressText');
        
        console.log('Resetting progress bar...', { progressBar, progressText });
        
        if (progressBar) {
            progressBar.style.width = '10%'; // Start with 10% to make it clearly visible
            progressBar.style.backgroundColor = '#ffc107'; // Start with yellow (searching)
            console.log('Progress bar reset to 10%');
        }
        
        if (progressText) {
            progressText.textContent = '🔍 Starting search...';
            console.log('Progress text reset');
        }
    }

    updateProgressBar(currentArticles, targetArticles) {
        const progressBar = document.getElementById('searchProgressBar');
        const progressText = document.getElementById('progressText');
        
        console.log('Updating progress bar:', { currentArticles, targetArticles, progressBar, progressText });
        
        if (!progressBar || !progressText) {
            console.error('Progress bar elements not found!');
            return;
        }
        
        // Calculate progress percentage purely based on article count
        const totalProgress = Math.min((currentArticles / targetArticles) * 100, 100);
        
        // Update progress bar
        progressBar.style.width = `${totalProgress}%`;
        
        // Update text with live information purely as N/20
        if (currentArticles >= targetArticles) {
            progressText.textContent = `✅ ${currentArticles}/${targetArticles}`;
            progressBar.style.backgroundColor = '#28a745'; // Green
        } else if (currentArticles > 0) {
            progressText.textContent = `${currentArticles}/${targetArticles}`;
            progressBar.style.backgroundColor = '#17a2b8'; // Blue
        } else {
            // When we have 0 articles, show searching state instead of error
            progressText.textContent = `🔍 Searching... (0/${targetArticles})`;
            progressBar.style.backgroundColor = '#ffc107'; // Yellow (searching)
        }
    }

    showProgressCompletion(currentArticles, targetArticles) {
        const progressBar = document.getElementById('searchProgressBar');
        const progressText = document.getElementById('progressText');
        
        if (!progressBar || !progressText) return;
        
        progressBar.style.width = '100%';
        progressBar.style.backgroundColor = '#28a745'; // Green
        
        if (currentArticles >= targetArticles) {
            progressText.textContent = `🎉 Search completed! Found ${currentArticles} articles.`;
        } else {
            progressText.textContent = `✅ Search completed! Found ${currentArticles} articles (target: ${targetArticles}).`;
        }
        
        // Hide progress bar after 3 seconds
        setTimeout(() => {
            this.hideProgressBar();
        }, 3000);
    }

    showProgressError() {
        const progressBar = document.getElementById('searchProgressBar');
        const progressText = document.getElementById('progressText');
        
        if (!progressBar || !progressText) return;
        
        progressBar.style.backgroundColor = '#dc3545'; // Red
        progressText.textContent = '❌ Search encountered an error. Please try again.';
        
        // Also show the error state
        this.showError('Search encountered an error. Please try again.');
    }

    showLiveUpdateNotification(count, isFinal = false) {
        // Create or update notification
        let notification = document.getElementById('liveUpdateNotification');
        
        if (!notification) {
            notification = document.createElement('div');
            notification.id = 'liveUpdateNotification';
            notification.className = 'live-update-notification';
            document.body.appendChild(notification);
        }
        
        if (isFinal) {
            notification.textContent = 'Search completed! All results loaded.';
            notification.className = 'live-update-notification final';
        } else {
            notification.textContent = `Live update: Found ${count} articles so far...`;
            notification.className = 'live-update-notification';
        }
        
        // Auto-hide after 3 seconds
        setTimeout(() => {
            if (notification && notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 3000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Helper method to get article preview text
    getArticlePreview(article) {
        const content = article.Content || article.content || '';
        if (content && content.trim().length > 0) {
            return content.substring(0, 200) + '...';
        } else if (article.summary && article.summary.trim().length > 0) {
            return article.summary;
        } else {
            return 'No preview available for this article.';
        }
    }

    // Extract keywords from content
    extractKeywordsFromContent(content) {
        if (!content || typeof content !== 'string') {
            return [];
        }

        // Clean the content to focus on main article body
        const cleanedContent = this.cleanArticleContent(content);
        
        // If content is too short (likely paywalled), extract from title/description
        if (cleanedContent.length < 200) {
            return this.extractKeywordsFromShortContent(content);
        }

        // Tech keywords with context patterns to avoid false matches
        const techKeywordPatterns = [
            // AI/ML - require context
            { keyword: 'artificial intelligence', pattern: /\b(artificial intelligence|ai)\b/gi },
            { keyword: 'machine learning', pattern: /\b(machine learning|ml)\b/gi },
            { keyword: 'deep learning', pattern: /\bdeep learning\b/gi },
            { keyword: 'neural network', pattern: /\bneural networks?\b/gi },
            { keyword: 'algorithm', pattern: /\balgorithms?\b/gi },
            
            // Cloud & Infrastructure
            { keyword: 'aws', pattern: /\b(aws|amazon web services)\b/gi },
            { keyword: 'azure', pattern: /\b(azure|microsoft azure)\b/gi },
            { keyword: 'gcp', pattern: /\b(gcp|google cloud)\b/gi },
            { keyword: 'kubernetes', pattern: /\bkubernetes\b/gi },
            { keyword: 'docker', pattern: /\bdocker\b/gi },
            { keyword: 'serverless', pattern: /\bserverless\b/gi },
            
            // Programming & Development
            { keyword: 'javascript', pattern: /\b(javascript|js)\b/gi },
            { keyword: 'python', pattern: /\bpython\b/gi },
            { keyword: 'java', pattern: /\bjava\b/gi },
            { keyword: 'react', pattern: /\breact\b/gi },
            { keyword: 'vue', pattern: /\bvue\b/gi },
            { keyword: 'angular', pattern: /\bangular\b/gi },
            { keyword: 'node.js', pattern: /\b(node\.js|nodejs)\b/gi },
            
            // Databases
            { keyword: 'database', pattern: /\bdatabases?\b/gi },
            { keyword: 'sql', pattern: /\bsql\b/gi },
            { keyword: 'mongodb', pattern: /\bmongodb\b/gi },
            { keyword: 'postgresql', pattern: /\bpostgresql\b/gi },
            
            // Security
            { keyword: 'cybersecurity', pattern: /\bcybersecurity\b/gi },
            { keyword: 'security', pattern: /\bsecurity\b/gi },
            { keyword: 'encryption', pattern: /\bencryption\b/gi },
            
            // DevOps
            { keyword: 'devops', pattern: /\bdevops\b/gi },
            { keyword: 'ci/cd', pattern: /\b(ci\/cd|continuous integration|continuous deployment)\b/gi },
            { keyword: 'git', pattern: /\bgit\b/gi },
            { keyword: 'github', pattern: /\bgithub\b/gi },
            
            // Mobile & IoT
            { keyword: 'ios', pattern: /\bios\b/gi },
            { keyword: 'android', pattern: /\bandroid\b/gi },
            { keyword: 'iot', pattern: /\b(iot|internet of things)\b/gi },
            
            // Emerging Tech
            { keyword: 'blockchain', pattern: /\bblockchain\b/gi },
            { keyword: 'bitcoin', pattern: /\bbitcoin\b/gi },
            { keyword: 'ethereum', pattern: /\bethereum\b/gi },
            { keyword: '5g', pattern: /\b5g\b/gi },
            
            // Business Tech
            { keyword: 'saas', pattern: /\bsaas\b/gi },
            { keyword: 'fintech', pattern: /\bfintech\b/gi },
            { keyword: 'e-commerce', pattern: /\b(e-commerce|ecommerce)\b/gi }
        ];

        const text = cleanedContent.toLowerCase();
        const foundKeywords = [];

        // Find tech keywords using proper word boundaries
        for (const { keyword, pattern } of techKeywordPatterns) {
            const matches = text.match(pattern);
            if (matches && matches.length > 0) {
                foundKeywords.push(keyword);
            }
        }

        // Extract meaningful technical terms (not just any word)
        const technicalTerms = this.extractTechnicalTerms(text);
        
        // Combine and prioritize
        const allKeywords = [...new Set([...foundKeywords, ...technicalTerms])];
        
        return allKeywords.slice(0, 8); // Return top 8 keywords
    }

    // Extract meaningful technical terms
    extractTechnicalTerms(text) {
        // Common technical suffixes and patterns
        const technicalSuffixes = ['ing', 'tion', 'sion', 'ment', 'ness', 'ity', 'al', 'ic', 'ous', 'ive'];
        const technicalPrefixes = ['micro', 'macro', 'hyper', 'super', 'ultra', 'multi', 'bi', 'tri', 'auto', 'semi'];
        
        // Technical word patterns
        const technicalPatterns = [
            /\b[a-z]+(?:ing|tion|sion|ment|ness|ity|al|ic|ous|ive)\b/g,
            /\b(?:micro|macro|hyper|super|ultra|multi|bi|tri|auto|semi)[a-z]+\b/g,
            /\b[a-z]{2,}(?:ware|tech|net|web|app|api|sdk|ui|ux|db|sql|nosql)\b/g,
            /\b[a-z]+(?:system|platform|framework|library|tool|service|api|sdk)\b/g
        ];

        const words = text.match(/\b[a-zA-Z]{4,}\b/g) || [];
        const wordCount = {};
        const stopWords = new Set([
            'the', 'and', 'for', 'are', 'but', 'not', 'you', 'all', 'can', 'had', 'her', 'was', 'one', 'our', 'out', 
            'day', 'get', 'has', 'him', 'his', 'how', 'man', 'new', 'now', 'old', 'see', 'two', 'way', 'who', 'boy', 
            'did', 'its', 'let', 'put', 'say', 'she', 'too', 'use', 'will', 'with', 'this', 'that', 'they', 'have', 
            'been', 'from', 'their', 'time', 'would', 'there', 'could', 'very', 'after', 'words', 'about', 'many', 
            'then', 'them', 'these', 'so', 'some', 'her', 'would', 'make', 'like', 'into', 'him', 'time', 'has', 
            'two', 'more', 'go', 'no', 'way', 'could', 'my', 'than', 'first', 'been', 'call', 'who', 'its', 'now', 
            'find', 'long', 'down', 'day', 'did', 'get', 'come', 'made', 'may', 'part', 'over', 'new', 'sound', 
            'take', 'only', 'little', 'work', 'know', 'place', 'year', 'live', 'me', 'back', 'give', 'most', 'very', 
            'after', 'thing', 'our', 'just', 'name', 'good', 'sentence', 'man', 'think', 'say', 'great', 'where', 
            'help', 'through', 'much', 'before', 'line', 'right', 'too', 'mean', 'old', 'any', 'same', 'tell', 
            'boy', 'follow', 'came', 'want', 'show', 'also', 'around', 'form', 'three', 'small', 'set', 'put', 
            'end', 'does', 'another', 'well', 'large', 'must', 'big', 'even', 'such', 'because', 'turn', 'here', 
            'why', 'ask', 'went', 'men', 'read', 'need', 'land', 'different', 'home', 'us', 'move', 'try', 'kind', 
            'hand', 'picture', 'again', 'change', 'off', 'play', 'spell', 'air', 'away', 'animal', 'house', 'point', 
            'page', 'letter', 'mother', 'answer', 'found', 'study', 'still', 'learn', 'should', 'America', 'world'
        ]);

        words.forEach(word => {
            const lowerWord = word.toLowerCase();
            
            // Skip stop words and very common words
            if (stopWords.has(lowerWord) || lowerWord.length < 4) {
                return;
            }

            // Check if it's a technical term
            let isTechnical = false;
            
            // Check technical patterns
            for (const pattern of technicalPatterns) {
                if (pattern.test(lowerWord)) {
                    isTechnical = true;
                    break;
                }
            }
            
            // Check if it has technical characteristics
            if (!isTechnical) {
                // Has technical suffix
                if (technicalSuffixes.some(suffix => lowerWord.endsWith(suffix))) {
                    isTechnical = true;
                }
                // Has technical prefix
                if (technicalPrefixes.some(prefix => lowerWord.startsWith(prefix))) {
                    isTechnical = true;
                }
                // Contains technical substrings
                if (lowerWord.includes('tech') || lowerWord.includes('data') || 
                    lowerWord.includes('soft') || lowerWord.includes('hard') ||
                    lowerWord.includes('net') || lowerWord.includes('web')) {
                    isTechnical = true;
                }
            }

            if (isTechnical) {
                wordCount[lowerWord] = (wordCount[lowerWord] || 0) + 1;
            }
        });

        // Get top technical terms by frequency
        const topTerms = Object.entries(wordCount)
            .sort(([,a], [,b]) => b - a)
            .slice(0, 4)
            .map(([word]) => word);

                 return topTerms;
     }

    // Clean article content to focus on main body
    cleanArticleContent(content) {
        if (!content || typeof content !== 'string') {
            return '';
        }

        let cleanedContent = content;

        // Remove common sidebar/advertisement patterns
        const sidebarPatterns = [
            /related articles?.*?$/gim,
            /recommended.*?$/gim,
            /popular.*?$/gim,
            /trending.*?$/gim,
            /sponsored.*?$/gim,
            /advertisement.*?$/gim,
            /subscribe.*?$/gim,
            /newsletter.*?$/gim,
            /sign up.*?$/gim,
            /read more.*?$/gim,
            /continue reading.*?$/gim,
            /share this.*?$/gim,
            /follow us.*?$/gim,
            /social media.*?$/gim,
            /comments.*?$/gim,
            /leave a comment.*?$/gim,
            /most read.*?$/gim,
            /top stories.*?$/gim,
            /breaking news.*?$/gim,
            /latest.*?$/gim
        ];

        // Remove sidebar content
        sidebarPatterns.forEach(pattern => {
            cleanedContent = cleanedContent.replace(pattern, '');
        });

        // Remove HTML-like content that might be navigation/sidebar
        cleanedContent = cleanedContent.replace(/<[^>]*>/g, ' ');
        
        // Remove excessive whitespace
        cleanedContent = cleanedContent.replace(/\s+/g, ' ').trim();
        
        // Take only the first 2000 characters (main article content)
        if (cleanedContent.length > 2000) {
            cleanedContent = cleanedContent.substring(0, 2000);
        }

        return cleanedContent;
    }

    // Extract keywords from short content (paywalled articles)
    extractKeywordsFromShortContent(content) {
        if (!content || typeof content !== 'string') {
            return [];
        }

        // For short content, focus on title-like patterns and tech terms
        const shortContentKeywords = [
            // Common tech terms that might appear in titles
            { keyword: 'ai', pattern: /\b(ai|artificial intelligence)\b/gi },
            { keyword: 'machine learning', pattern: /\b(machine learning|ml)\b/gi },
            { keyword: 'blockchain', pattern: /\bblockchain\b/gi },
            { keyword: 'cryptocurrency', pattern: /\b(cryptocurrency|crypto)\b/gi },
            { keyword: 'cybersecurity', pattern: /\bcybersecurity\b/gi },
            { keyword: 'cloud', pattern: /\bcloud\b/gi },
            { keyword: 'startup', pattern: /\bstartup\b/gi },
            { keyword: 'funding', pattern: /\bfunding\b/gi },
            { keyword: 'acquisition', pattern: /\bacquisition\b/gi },
            { keyword: 'ipo', pattern: /\bipo\b/gi },
            { keyword: 'venture capital', pattern: /\b(venture capital|vc)\b/gi },
            { keyword: 'tech', pattern: /\btech\b/gi },
            { keyword: 'innovation', pattern: /\binnovation\b/gi },
            { keyword: 'digital', pattern: /\bdigital\b/gi },
            { keyword: 'platform', pattern: /\bplatform\b/gi },
            { keyword: 'app', pattern: /\bapp\b/gi },
            { keyword: 'mobile', pattern: /\bmobile\b/gi },
            { keyword: 'web', pattern: /\bweb\b/gi },
            { keyword: 'data', pattern: /\bdata\b/gi },
            { keyword: 'software', pattern: /\bsoftware\b/gi },
            { keyword: 'hardware', pattern: /\bhardware\b/gi },
            { keyword: 'api', pattern: /\bapi\b/gi },
            { keyword: 'saas', pattern: /\bsaas\b/gi },
            { keyword: 'fintech', pattern: /\bfintech\b/gi },
            { keyword: 'healthtech', pattern: /\bhealthtech\b/gi },
            { keyword: 'edtech', pattern: /\bedtech\b/gi }
        ];

        const text = content.toLowerCase();
        const foundKeywords = [];

        // Find keywords in short content
        for (const { keyword, pattern } of shortContentKeywords) {
            const matches = text.match(pattern);
            if (matches && matches.length > 0) {
                foundKeywords.push(keyword);
            }
        }

        return foundKeywords.slice(0, 6); // Return fewer keywords for short content
    }



    // Utility method to format text
    formatText(text, maxLength = 200) {
        if (!text) return '';
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    console.log('🚀 Spectechle app initializing...');
    window.spectechleApp = new SpectechleApp();
    console.log('✅ Spectechle app initialized!');
    
    // Test if JavaScript is working
    console.log('🧪 JavaScript is working! Test button should be functional.');
});

// Add some utility functions for debugging
window.debugSpectechle = {
    getCurrentState: () => {
        return {
            mode: window.spectechleApp.currentMode,
            categories: Array.from(window.spectechleApp.selectedCategories),
            searchId: window.spectechleApp.currentSearchId
        };
    },
    
    testSearch: async (query = 'artificial intelligence') => {
        document.getElementById('searchInput').value = query;
        await window.spectechleApp.performSearch();
    }
};
