"""
Text Classifier
Classifies text into tech categories using ML models or keyword matching
"""

import re
import logging
from typing import Dict, List, Any, Optional
from collections import Counter

logger = logging.getLogger(__name__)

class TextClassifier:
    """Text classification for tech articles"""
    
    def __init__(self):
        self.categories = [
            'Artificial Intelligence',
            'Cloud Computing',
            'Cybersecurity',
            'Data Science',
            'DevOps',
            'Web Development',
            'Mobile Development',
            'Blockchain',
            'IoT',
            'Quantum Computing'
        ]
        
        # Keyword mappings for classification
        self.keyword_mappings = {
            'Artificial Intelligence': [
                'ai', 'artificial intelligence', 'machine learning', 'ml', 'deep learning',
                'neural network', 'neural networks', 'algorithm', 'algorithms', 'gpt',
                'chatgpt', 'llm', 'large language model', 'transformer', 'bert',
                'computer vision', 'nlp', 'natural language processing', 'reinforcement learning',
                'supervised learning', 'unsupervised learning', 'tensorflow', 'pytorch',
                'scikit-learn', 'opencv', 'keras', 'theano', 'caffe', 'mxnet'
            ],
            'Cloud Computing': [
                'cloud', 'aws', 'amazon web services', 'azure', 'google cloud', 'gcp',
                'serverless', 'lambda', 'kubernetes', 'k8s', 'docker', 'container',
                'microservices', 'api gateway', 'load balancer', 'auto scaling',
                'elastic compute', 'ec2', 's3', 'rds', 'dynamodb', 'lambda',
                'cloudformation', 'terraform', 'ansible', 'jenkins', 'gitlab ci',
                'github actions', 'circleci', 'travis ci'
            ],
            'Cybersecurity': [
                'security', 'cybersecurity', 'cyber security', 'hack', 'hacking',
                'vulnerability', 'vulnerabilities', 'exploit', 'malware', 'virus',
                'ransomware', 'phishing', 'ddos', 'firewall', 'encryption',
                'authentication', 'authorization', 'oauth', 'jwt', 'ssl', 'tls',
                'penetration testing', 'pen test', 'security audit', 'compliance',
                'gdpr', 'hipaa', 'sox', 'pci dss', 'zero trust', 'vpn'
            ],
            'Data Science': [
                'data science', 'data scientist', 'big data', 'analytics',
                'statistics', 'statistical', 'data analysis', 'data mining',
                'data visualization', 'dashboard', 'bi', 'business intelligence',
                'etl', 'data warehouse', 'data lake', 'hadoop', 'spark',
                'pandas', 'numpy', 'matplotlib', 'seaborn', 'plotly',
                'tableau', 'power bi', 'jupyter', 'notebook', 'r', 'python'
            ],
            'DevOps': [
                'devops', 'ci/cd', 'continuous integration', 'continuous deployment',
                'continuous delivery', 'pipeline', 'deployment', 'infrastructure',
                'infrastructure as code', 'iac', 'monitoring', 'logging',
                'prometheus', 'grafana', 'elk stack', 'elasticsearch', 'logstash',
                'kibana', 'nagios', 'zabbix', 'datadog', 'new relic', 'splunk'
            ],
            'Web Development': [
                'web development', 'frontend', 'backend', 'full stack', 'fullstack',
                'html', 'css', 'javascript', 'js', 'react', 'vue', 'angular',
                'node.js', 'nodejs', 'express', 'django', 'flask', 'spring',
                'php', 'laravel', 'wordpress', 'drupal', 'joomla', 'api',
                'rest api', 'graphql', 'websocket', 'http', 'https'
            ],
            'Mobile Development': [
                'mobile development', 'ios', 'android', 'react native',
                'flutter', 'xamarin', 'swift', 'kotlin', 'java', 'objective-c',
                'mobile app', 'app development', 'cross platform', 'hybrid app',
                'native app', 'app store', 'google play', 'xcode', 'android studio'
            ],
            'Blockchain': [
                'blockchain', 'bitcoin', 'ethereum', 'cryptocurrency', 'crypto',
                'smart contract', 'defi', 'decentralized finance', 'nft',
                'non-fungible token', 'web3', 'metamask', 'solidity', 'hyperledger',
                'consensus', 'mining', 'wallet', 'exchange', 'ico', 'token'
            ],
            'IoT': [
                'iot', 'internet of things', 'sensor', 'sensors', 'arduino',
                'raspberry pi', 'esp32', 'esp8266', 'mqtt', 'coap', 'zigbee',
                'bluetooth', 'wifi', 'cellular', '5g', 'edge computing',
                'smart home', 'wearable', 'connected device', 'embedded system'
            ],
            'Quantum Computing': [
                'quantum computing', 'quantum computer', 'qubit', 'qubits',
                'quantum algorithm', 'quantum supremacy', 'quantum entanglement',
                'superposition', 'quantum gate', 'ibm q', 'google quantum',
                'quantum error correction', 'quantum cryptography', 'qiskit',
                'cirq', 'quantum machine learning', 'quantum advantage'
            ]
        }
    
    def classify(self, text: str, options: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Classify text into tech categories
        
        Args:
            text: Input text to classify
            options: Classification options
            
        Returns:
            Classification result with category and confidence
        """
        if not text or not text.strip():
            return {
                'category': 'Unknown',
                'confidence': 0.0,
                'categories': []
            }
        
        # Keyword-based classification is the primary method for now
        return self._keyword_classify(text, options)
    
    def _keyword_classify(self, text: str, options: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Keyword-based classification"""
        text_lower = text.lower()
        
        # Count keyword matches for each category
        category_scores = {}
        
        for category, keywords in self.keyword_mappings.items():
            score = 0
            for keyword in keywords:
                # Count occurrences of keyword
                count = text_lower.count(keyword.lower())
                score += count
            
            if score > 0:
                category_scores[category] = score
        
        if not category_scores:
            return {
                'category': 'Technology',
                'confidence': 0.1,
                'categories': [{'category': 'Technology', 'confidence': 0.1}]
            }
        
        # Get top categories
        sorted_categories = sorted(
            category_scores.items(),
            key=lambda x: x[1],
            reverse=True
        )
        
        # Calculate confidence scores
        total_score = sum(category_scores.values())
        categories = []
        
        for category, score in sorted_categories[:3]:
            confidence = min(score / total_score, 0.95)  # Cap at 95%
            categories.append({
                'category': category,
                'confidence': confidence
            })
        
        return {
            'category': categories[0]['category'],
            'confidence': categories[0]['confidence'],
            'categories': categories
        }
    
    def extract_keywords(self, text: str, options: Optional[Dict[str, Any]] = None) -> List[str]:
        """
        Extract keywords from text
        
        Args:
            text: Input text
            options: Extraction options
            
        Returns:
            List of extracted keywords
        """
        if not text or not text.strip():
            return []
        
        # Rule-based extraction is the primary method for now
        return self._rule_extract_keywords(text, options)
    
    def _rule_extract_keywords(self, text: str, options: Optional[Dict[str, Any]] = None) -> List[str]:
        """Rule-based keyword extraction"""
        # Extract words that look like technical terms
        words = re.findall(r'\b[a-zA-Z][a-zA-Z0-9_]*\b', text.lower())
        
        # Filter for technical keywords
        technical_keywords = set()
        
        # Add all keywords from our mappings
        for keywords in self.keyword_mappings.values():
            technical_keywords.update(keywords)
        
        # Find matches
        found_keywords = []
        for word in words:
            if word in technical_keywords and len(word) > 2:
                found_keywords.append(word)
        
        # Count occurrences and get top keywords
        keyword_counts = Counter(found_keywords)
        top_keywords = [kw for kw, count in keyword_counts.most_common(10)]
        
        return top_keywords
    
    def get_categories(self) -> List[str]:
        """Get available categories"""
        return self.categories.copy()
    
    def get_keyword_mappings(self) -> Dict[str, List[str]]:
        """Get keyword mappings"""
        return self.keyword_mappings.copy()

