import logging
import os
import re
from typing import Dict, Any, List
from transformers import BartTokenizer, BartForConditionalGeneration
import torch

logger = logging.getLogger(__name__)

class TextSummarizer:
    """
    A text summarizer that uses BART-CNN model locally.
    """

    def __init__(self):
        self.model = None
        self.tokenizer = None
        self._initialize_summarizer()

    def _initialize_summarizer(self):
        """Initialize the BART-CNN model and tokenizer."""
        try:
            logger.info("🔄 Downloading BART-CNN model from Hugging Face...")
            logger.info("This may take a few minutes on first run...")
            
            # Load the model and tokenizer with proper error handling
            self.tokenizer = BartTokenizer.from_pretrained('facebook/bart-large-cnn')
            self.model = BartForConditionalGeneration.from_pretrained('facebook/bart-large-cnn')
            
            # Set model to evaluation mode
            self.model.eval()
            
            logger.info("✅ BART-CNN model and tokenizer initialized successfully!")
            
        except Exception as e:
            logger.error(f"❌ Failed to initialize BART-CNN model: {e}")
            logger.warning("Falling back to simple text truncation.")
            self.model = None
            self.tokenizer = None

    def summarize(self, text: str, max_length: int = 150, min_length: int = 40) -> Dict[str, Any]:
        """
        Generates a summary for the given text using BART-CNN.
        """
        cleaned_text = self._clean_text(text)
        if not cleaned_text:
            logger.warning("Empty text provided, returning empty response")
            return self._format_response("The provided text was empty after cleaning.", "empty_text")

        # Use BART-CNN if available
        if self.model and self.tokenizer:
            try:
                logger.info("🎯 Using BART-CNN model for summarization")
                return self._bart_summarize(cleaned_text, max_length, min_length)
            except Exception as e:
                logger.error(f"❌ BART-CNN summarization failed: {e}")
                logger.warning("🔄 Falling back to simple truncation")
        
        # Fallback to simple truncation
        logger.warning("⚠️ Using fallback summarization (BART-CNN not available)")
        return self._fallback_summarize(cleaned_text)

    def batch_summarize(self, texts: List[str], options: Dict[str, Any] = None) -> List[Dict[str, Any]]:
        """
        Batch summarization of multiple texts.
        """
        if options is None:
            options = {}
        
        max_length = options.get('max_length', 150)
        min_length = options.get('min_length', 40)
        
        results = []
        for text in texts:
            try:
                result = self.summarize(text, max_length, min_length)
                results.append(result)
            except Exception as e:
                logger.error(f"❌ Failed to summarize text in batch: {e}")
                # Add fallback result
                results.append(self._format_response(
                    f"Summarization failed: {str(e)}", 
                    "error"
                ))
        
        return results

    def _bart_summarize(self, text: str, max_length: int, min_length: int) -> Dict[str, Any]:
        """Uses the BART-CNN model to generate a summary."""
        try:
            # Ensure model is in evaluation mode
            self.model.eval()
            
            # First, check the token length without truncation
            tokenized = self.tokenizer(text, return_tensors='pt', add_special_tokens=True)
            input_length = tokenized['input_ids'].shape[1]
            
            logger.info(f"📏 Input text tokenized to {input_length} tokens")
            
            # If input is too long, truncate it intelligently
            if input_length > 1024:
                logger.warning(f"⚠️ Input too long ({input_length} tokens), truncating to 1024 tokens")
                # Truncate text to fit within BART's limit
                # Estimate characters per token (roughly 4 chars per token)
                max_chars = int(1024 * 4 * 0.8)  # Conservative estimate
                if len(text) > max_chars:
                    text = text[:max_chars] + "..."
                    logger.info(f"📝 Truncated text to {len(text)} characters")
            
            # Tokenize input with proper truncation
            inputs = self.tokenizer(
                text, 
                return_tensors='pt', 
                max_length=1024,  # BART's maximum input length
                truncation=True,
                padding=True,
                add_special_tokens=True
            )
            
            # Verify the final input length
            final_length = inputs['input_ids'].shape[1]
            logger.info(f"✅ Final input length: {final_length} tokens")
            
            # Check if input is empty or too short
            if final_length == 0:
                logger.warning("Empty input after tokenization")
                return self._format_response("Input text was too short to summarize.", "short_text")
            
            # Generate summary with proper error handling
            with torch.no_grad():
                output_ids = self.model.generate(
                    inputs['input_ids'],
                    attention_mask=inputs['attention_mask'],
                    num_beams=4,  # Reduced from 6 to avoid memory issues
                    length_penalty=1.0,  # Reduced penalty
                    max_length=max_length,
                    min_length=min_length,
                    no_repeat_ngram_size=2,  # Reduced from 3
                    early_stopping=True,
                    pad_token_id=self.tokenizer.eos_token_id
                )
            
            # Decode the output
            summary = self.tokenizer.decode(output_ids[0], skip_special_tokens=True)
            
            # Validate summary
            if not summary or summary.strip() == "":
                logger.warning("Generated summary is empty")
                return self._format_response("Could not generate a meaningful summary.", "empty_summary")
            
            return self._format_response(summary, "bart_cnn")
            
        except Exception as e:
            logger.error(f"❌ BART summarization error: {e}")
            raise e

    def _fallback_summarize(self, text: str) -> Dict[str, Any]:
        """A simple truncation fallback."""
        try:
            words = text.split()
            if len(words) > 100:
                summary = ' '.join(words[:100]) + '...'
            else:
                summary = text
            return self._format_response(summary, "fallback")
        except Exception as e:
            logger.error(f"❌ Fallback summarization error: {e}")
            return self._format_response("Summarization failed.", "error")

    def _clean_text(self, text: str) -> str:
        """Basic text cleaning with length management."""
        try:
            if not text or not isinstance(text, str):
                return ""
            
            # Remove extra whitespace
            text = re.sub(r'\s+', ' ', text).strip()
            
            # Remove URLs
            text = re.sub(r'http\S+', '', text)
            
            # Remove special characters that might cause issues
            text = re.sub(r'[^\w\s\.\,\!\?\;\:\-\(\)]', '', text)
            
            # If text is extremely long, truncate it early to prevent tokenization issues
            # BART typically handles ~4000 characters well (roughly 1000 tokens)
            max_chars = 4000
            if len(text) > max_chars:
                logger.info(f"📝 Text too long ({len(text)} chars), truncating to {max_chars} chars")
                # Try to truncate at sentence boundaries
                sentences = text.split('.')
                truncated_text = ""
                for sentence in sentences:
                    if len(truncated_text + sentence + '.') <= max_chars:
                        truncated_text += sentence + '.'
                    else:
                        break
                
                if not truncated_text:
                    # If no complete sentences fit, just truncate
                    truncated_text = text[:max_chars] + "..."
                
                text = truncated_text
                logger.info(f"✅ Truncated text to {len(text)} characters")
            
            return text
        except Exception as e:
            logger.error(f"❌ Text cleaning error: {e}")
            return ""

    def _format_response(self, summary: str, method: str) -> Dict[str, Any]:
        """Helper to create a consistent response format."""
        try:
            return {
                "summary": summary,
                "method": method,
                "summary_length": len(summary.split()) if summary else 0
            }
        except Exception as e:
            logger.error(f"❌ Response formatting error: {e}")
            return {
                "summary": "Error occurred during processing",
                "method": "error",
                "summary_length": 0
            }
