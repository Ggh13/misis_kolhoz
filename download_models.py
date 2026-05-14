"""Pre-download HF models during Docker build."""
import os
from transformers import AutoTokenizer, AutoModel, AutoModelForSequenceClassification

hf_token = os.getenv("HF_TOKEN") or None
print(f"Downloading E5 model (token={'set' if hf_token else 'not set'})...")
AutoTokenizer.from_pretrained("intfloat/multilingual-e5-small", token=hf_token)
AutoModel.from_pretrained("intfloat/multilingual-e5-small", token=hf_token)
print("E5 ready")
print("Downloading injection model...")
AutoTokenizer.from_pretrained("protectai/deberta-v3-base-prompt-injection-v2")
AutoModelForSequenceClassification.from_pretrained("protectai/deberta-v3-base-prompt-injection-v2")
print("All models ready")