import os
from typing import Any, Dict, List

import torch
import torch.nn.functional as F
from deep_translator import GoogleTranslator
from dotenv import load_dotenv
from openai import OpenAI
from transformers import (
    AutoModel,
    AutoModelForSequenceClassification,
    AutoTokenizer,
    pipeline,
)

from extractor import get_combined_text

load_dotenv()

translator = GoogleTranslator(source="auto", target="en")

injection_model_name = "protectai/deberta-v3-base-prompt-injection-v2"
injection_tokenizer = AutoTokenizer.from_pretrained(injection_model_name)
injection_model = AutoModelForSequenceClassification.from_pretrained(injection_model_name)

classifier = pipeline(
    "text-classification",
    model=injection_model,
    tokenizer=injection_tokenizer,
    truncation=True,
    max_length=512,
    device=torch.device("cuda" if torch.cuda.is_available() else "cpu"),
)

hf_token = os.getenv("HF_TOKEN") or None
e5_tokenizer = AutoTokenizer.from_pretrained(
    "intfloat/multilingual-e5-small", token=hf_token
)
model = AutoModel.from_pretrained("intfloat/multilingual-e5-small", token=hf_token)

torch.backends.quantized.engine = "qnnpack"
model_quant = torch.quantization.quantize_dynamic(
    model, {torch.nn.Linear}, dtype=torch.qint8
)

client = OpenAI(
    base_url="https://api.groq.com/openai/v1",
    api_key=os.getenv("GROQ_API_KEY"),
)


def average_pool(last_hidden_states: torch.Tensor, attention_mask: torch.Tensor) -> torch.Tensor:
    last_hidden = last_hidden_states.masked_fill(~attention_mask[..., None].bool(), 0.0)
    return last_hidden.sum(dim=1) / attention_mask.sum(dim=1)[..., None]


def e5_embs(text: str, prefix: str) -> List[float]:
    if not text:
        return []
    fulltext = f"{prefix}{text}"
    tokens = e5_tokenizer(
        fulltext, max_length=512, padding=True, truncation=True, return_tensors="pt"
    )
    with torch.no_grad():
        outputs = model_quant(**tokens)
        embeddings = F.normalize(
            average_pool(outputs.last_hidden_state, tokens["attention_mask"]),
            p=2,
            dim=1,
        )
    return embeddings[0].tolist()


def is_injection(prompt: str) -> bool:
    threshold = 0.8
    translated_text = translator.translate(prompt)
    all_scores = classifier(translated_text, top_k=None)
    scores_dict = {item["label"]: item["score"] for item in all_scores}
    return scores_dict.get("INJECTION", 0) >= threshold


def expanded_event_description(name: str, description: str) -> str:
    try:
        response = client.chat.completions.create(
            model=os.getenv("API_MODEL", "llama-3.3-70b-versatile"),
            messages=[
                {
                    "role": "system",
                    "content": "Ты эксперт по анализу контента. Твоя задача — выделять семантическое ядро.",
                },
                {
                    "role": "user",
                    "content": (
                        f"Название: {name}. Описание: {description}. "
                        "Составь список из 10-15 ключевых слов, ассоциаций и предметов через запятую. "
                        "Только слова, без пояснений."
                    ),
                },
            ],
            temperature=0.3,
            max_tokens=150,
        )
        return response.choices[0].message.content.strip().rstrip(".")
    except Exception:
        return f"{name}, {description}"


def build_event_payload(event_name: str, event_description: str) -> Dict[str, Any]:
    if is_injection(event_name) or is_injection(event_description):
        raise ValueError("Injection detected")

    combined_context = expanded_event_description(event_name, event_description)
    emb = e5_embs(combined_context, "query: ")

    return {
        "event_name": event_name,
        "processed_context": combined_context,
        "embedding": emb,
    }


def build_product_payload(
    product_features: Dict[str, Any], farmer_features: Dict[str, Any]
) -> Dict[str, Any]:
    combined_text = get_combined_text(product_features, farmer_features)
    emb = e5_embs(combined_text, "passage: ")

    return {
        "combined_text": combined_text,
        "embedding": emb,
    }


__all__ = [
    "build_event_payload",
    "build_product_payload",
    "e5_embs",
    "expanded_event_description",
    "is_injection",
]
