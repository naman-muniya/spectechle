#!/usr/bin/env python3
"""
Script to download and test BART-CNN model from Hugging Face
Useful utility for testing model loading and debugging issues
"""

import logging
import torch
from transformers import BartTokenizer, BartForConditionalGeneration

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

def download_bart_cnn():
    """Download BART-CNN model from Hugging Face and test it."""
    try:
        logger.info("🔄 Downloading BART-CNN model from Hugging Face...")
        logger.info("This may take a few minutes on first run...")
        
        # Load the model and tokenizer
        model = BartForConditionalGeneration.from_pretrained('facebook/bart-large-cnn')
        tokenizer = BartTokenizer.from_pretrained('facebook/bart-large-cnn')
        
        logger.info("✅ BART-CNN model downloaded successfully!")
        logger.info("Model is ready to use.")
        
        # Test the model with a sample text
        test_text = """The Amazon rainforest, often referred to as the "lungs of the Earth," is the largest tropical rainforest in the world, spanning over 5.5 million square kilometers across nine South American countries, with the majority located in Brazil. This vast forest is home to an unparalleled diversity of species, including over 400 billion individual trees representing around 16,000 species, countless insects, amphibians, birds, and mammals. Beyond its biodiversity, the Amazon plays a crucial role in regulating the global climate by absorbing massive amounts of carbon dioxide, producing oxygen, and influencing weather patterns across the planet. However, the rainforest faces unprecedented threats due to human activities. Deforestation, largely driven by logging, agriculture, and cattle ranching, has accelerated in recent decades, resulting in the loss of millions of hectares of forest each year. Fires, often set deliberately to clear land, exacerbate the problem, releasing vast amounts of carbon into the atmosphere. Climate change further compounds these issues, as rising temperatures and altered rainfall patterns threaten the delicate ecological balance. Indigenous communities that have lived in harmony with the forest for centuries are also at risk, facing displacement, loss of livelihoods, and cultural erosion. Conservation efforts, including protected areas, sustainable management practices, and international agreements, aim to preserve this critical ecosystem. Nevertheless, global cooperation, stronger policies, and immediate action are essential to prevent irreversible damage to the Amazon and to safeguard its invaluable environmental, cultural, and economic contributions."""
        
        logger.info("🧪 Testing model with sample text...")
        input_ids = tokenizer(test_text, return_tensors='pt').input_ids
        output_ids = model.generate(
            input_ids, 
            num_beams=6, 
            length_penalty=1.5,
            max_length=120,
            min_length=60,
            no_repeat_ngram_size=3,
            early_stopping=True
        )
        summary = tokenizer.decode(output_ids[0], skip_special_tokens=True)
        logger.info(f"✅ Test summary: {summary}")
        
        logger.info("🎉 Model download and test completed successfully!")
        return True
        
    except Exception as e:
        logger.error(f"❌ Failed to download BART-CNN model: {e}")
        return False

def test_model_performance():
    """Test model performance with different text lengths."""
    try:
        logger.info("📊 Testing model performance...")
        
        # Load model
        model = BartForConditionalGeneration.from_pretrained('facebook/bart-large-cnn')
        tokenizer = BartTokenizer.from_pretrained('facebook/bart-large-cnn')
        
        # Test texts of different lengths
        test_texts = [
            "Short text for testing.",
            "This is a medium length text that should work well with the BART-CNN model for summarization testing.",
            """This is a longer text that contains multiple sentences and should provide a good test of the BART-CNN model's summarization capabilities. The model should be able to extract the key points and create a coherent summary that captures the main ideas while maintaining readability and accuracy."""
        ]
        
        for i, text in enumerate(test_texts, 1):
            logger.info(f"Testing text {i} ({len(text)} characters)...")
            input_ids = tokenizer(text, return_tensors='pt').input_ids
            output_ids = model.generate(
                input_ids,
                num_beams=4,
                length_penalty=2.0,
                max_length=100,
                min_length=30,
                no_repeat_ngram_size=3
            )
            summary = tokenizer.decode(output_ids[0], skip_special_tokens=True)
            logger.info(f"Summary {i}: {summary}")
        
        logger.info("✅ Performance test completed!")
        return True
        
    except Exception as e:
        logger.error(f"❌ Performance test failed: {e}")
        return False

if __name__ == "__main__":
    logger.info("🚀 Starting BART-CNN model download and test...")
    
    # Download and test model
    success = download_bart_cnn()
    
    if success:
        # Run performance test
        test_model_performance()
        logger.info("🎯 All tests completed successfully!")
    else:
        logger.error("💥 Model download failed. Check your internet connection and try again.")
