#!/usr/bin/env python3
"""
Spectechle NLP Service
Flask API server for article classification and summarization
"""

import os
import logging
from datetime import datetime
from flask import Flask, request, jsonify
from flask_cors import CORS
from dotenv import load_dotenv

from classifier.text_classifier import TextClassifier
from summarizer.text_summarizer import TextSummarizer

# Load environment variables
load_dotenv('.env')

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Global model instances
classifier = None
summarizer = None

def initialize_models():
    """Initialize NLP models safely."""
    global summarizer, classifier
    logger.info("Initializing NLP models...")
    try:
        summarizer = TextSummarizer()
        classifier = TextClassifier()
        logger.info("✅ NLP models initialized successfully.")
    except Exception as e:
        logger.error(f"❌ Failed to initialize models: {e}")

# Initialize models on startup
initialize_models()

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    return jsonify({
        'status': 'healthy',
        'service': 'spectechle-nlp',
        'timestamp': datetime.now().isoformat(),
        'version': '1.0.0',
        'models_loaded': True
    })

@app.route('/classify', methods=['POST'])
def classify_text():
    """Classify text into tech categories"""
    try:
        data = request.get_json()
        
        if not data or 'text' not in data:
            return jsonify({
                'success': False,
                'error': 'Text field is required'
            }), 400
        
        text = data['text']
        options = data.get('options', {})
        
        if not text.strip():
            return jsonify({
                'success': False,
                'error': 'Text cannot be empty'
            }), 400
        
        logger.info(f"Classifying text: {text[:100]}...")
        
        # Perform classification
        result = classifier.classify(text, options)
        
        return jsonify({
            'success': True,
            'data': result
        })
        
    except Exception as e:
        logger.error(f"Classification error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/summarize', methods=['POST'])
def summarize_text():
    """Summarize text using BART-CNN"""
    try:
        data = request.get_json()
        
        if not data:
            logger.error("No JSON data received")
            return jsonify({
                'success': False,
                'error': 'No JSON data received'
            }), 400
        
        if 'text' not in data:
            logger.error("Missing 'text' field in request")
            return jsonify({
                'success': False,
                'error': 'Text field is required'
            }), 400
        
        text = data['text']
        options = data.get('options', {})
        
        if not text or not isinstance(text, str):
            logger.error(f"Invalid text field: {type(text)}")
            return jsonify({
                'success': False,
                'error': 'Text must be a non-empty string'
            }), 400
        
        if not text.strip():
            logger.error("Empty text after stripping")
            return jsonify({
                'success': False,
                'error': 'Text cannot be empty'
            }), 400
        
        logger.info(f"Summarizing text: {text[:100]}...")
        
        # Perform summarization
        max_length = options.get('max_length', 150)
        min_length = options.get('min_length', 40)
        result = summarizer.summarize(text, max_length, min_length)
        
        return jsonify({
            'success': True,
            'data': result
        })
        
    except Exception as e:
        logger.error(f"Summarization error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/batch_summarize', methods=['POST'])
def batch_summarize_texts():
    """Batch summarization of multiple texts"""
    try:
        data = request.get_json()
        
        if not data or 'texts' not in data:
            return jsonify({
                'success': False,
                'error': 'Texts field is required'
            }), 400
        
        texts = data['texts']
        options = data.get('options', {})
        
        if not isinstance(texts, list) or len(texts) == 0:
            return jsonify({
                'success': False,
                'error': 'Texts must be a non-empty list'
            }), 400
        
        logger.info(f"Batch summarizing {len(texts)} texts...")
        
        # Perform batch summarization
        results = summarizer.batch_summarize(texts, options)
        
        return jsonify({
            'success': True,
            'data': {
                'results': results,
                'total': len(results)
            }
        })
        
    except Exception as e:
        logger.error(f"Batch summarization error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/extract_keywords', methods=['POST'])
def extract_keywords():
    """Extract keywords from text"""
    try:
        data = request.get_json()
        
        if not data or 'text' not in data:
            return jsonify({
                'success': False,
                'error': 'Text field is required'
            }), 400
        
        text = data['text']
        options = data.get('options', {})
        
        if not text.strip():
            return jsonify({
                'success': False,
                'error': 'Text cannot be empty'
            }), 400
        
        logger.info(f"Extracting keywords from text: {text[:100]}...")
        
        # Extract keywords
        keywords = classifier.extract_keywords(text, options)
        
        return jsonify({
            'success': True,
            'data': {
                'keywords': keywords
            }
        })
        
    except Exception as e:
        logger.error(f"Keyword extraction error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.route('/process_article', methods=['POST'])
def process_article():
    """Process article with classification and summarization"""
    try:
        data = request.get_json()
        
        if not data or 'text' not in data:
            return jsonify({
                'success': False,
                'error': 'Text field is required'
            }), 400
        
        text = data['text']
        title = data.get('title', '')
        options = data.get('options', {})
        
        if not text.strip():
            return jsonify({
                'success': False,
                'error': 'Text cannot be empty'
            }), 400
        
        logger.info(f"Processing article: {title[:50]}...")
        
        # Combine title and text for better classification
        full_text = f"{title}\n\n{text}" if title else text
        
        # Perform classification
        classification = classifier.classify(full_text, options)
        
        # Perform summarization
        max_length = options.get('max_length', 150)
        min_length = options.get('min_length', 40)
        summary = summarizer.summarize(text, max_length, min_length)
        
        # Extract keywords
        keywords = classifier.extract_keywords(text, options)
        
        return jsonify({
            'success': True,
            'data': {
                'classification': classification,
                'summary': summary,
                'keywords': keywords
            }
        })
        
    except Exception as e:
        logger.error(f"Article processing error: {e}")
        return jsonify({
            'success': False,
            'error': str(e)
        }), 500

@app.errorhandler(404)
def not_found(error):
    return jsonify({
        'success': False,
        'error': 'Endpoint not found'
    }), 404

@app.errorhandler(500)
def internal_error(error):
    return jsonify({
        'success': False,
        'error': 'Internal server error'
    }), 500

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 5000))
    debug = os.environ.get('FLASK_ENV') == 'development'
    
    logger.info(f"🚀 Starting Spectechle NLP service on port {port}")
    logger.info(f"📊 API available at http://localhost:{port}")
    
    app.run(
        host='0.0.0.0',
        port=port,
        debug=debug,
        threaded=True
    )

