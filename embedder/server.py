import os
import torch
import torch.nn.functional as F
from typing import Dict, Any, List, Optional
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from dotenv import load_dotenv
from openai import OpenAI
from transformers import AutoModel, AutoTokenizer, AutoModelForSequenceClassification, pipeline
from deep_translator import GoogleTranslator

load_dotenv()

app = FastAPI()

translator = GoogleTranslator(source="auto", target="en")

injection_model_name = 'protectai/deberta-v3-base-prompt-injection-v2'
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
e5_tokenizer = AutoTokenizer.from_pretrained("intfloat/multilingual-e5-small", token=hf_token)
model = AutoModel.from_pretrained("intfloat/multilingual-e5-small", token=hf_token)

torch.backends.quantized.engine = "qnnpack"
model_quant = torch.quantization.quantize_dynamic(
    model, {torch.nn.Linear}, dtype=torch.qint8
)

client = OpenAI(api_key=os.getenv("OPENAI_API_KEY"), base_url=os.getenv("BASE_URL"))

class EventRequest(BaseModel):
    event_name: str
    event_description: str

class ProductData(BaseModel):
    product_type: Optional[str] = None
    geo: Optional[str] = None
    production_method: Optional[List[str]] = None
    usage_vibes: Optional[List[str]] = None

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
    scores_dict = {item['label']: item['score'] for item in all_scores}
    return scores_dict.get('INJECTION', 0) >= threshold

def expanded_event_description(name: str, description: str) -> str:
    try:
        response = client.chat.completions.create(
            model=os.getenv("API_MODEL", "llama-3.3-70b-versatile"),
            messages=[
                {"role": "system", "content": "Ты эксперт по анализу контента. Твоя задача — выделять семантическое ядро."},
                {"role": "user", "content": f"Название: {name}. Описание: {description}. Составь список из 10-15 ключевых слов, ассоциаций и предметов через запятую. Только слова, без пояснений."}
            ],
            temperature=0.3,
            max_tokens=150,
        )
        return response.choices[0].message.content.strip().rstrip(".")
    except Exception:
        return f"{name}, {description}"

@app.post("/process/event")
def process_event(data: EventRequest):
    if is_injection(data.event_name) or is_injection(data.event_description):
        raise HTTPException(status_code=400, detail="Injection detected")
    
    combined_context = expanded_event_description(data.event_name, data.event_description)
    emb = e5_embs(combined_context, "passage: ")
    
    return {
        "event_name": data.event_name,
        "processed_context": combined_context,
        "embedding": emb
    }

@app.post("/process/product")
def process_product(data: ProductData):
    output = {}
    if data.product_type:
        output["product_type"] = {"text": data.product_type, "embedding": e5_embs(data.product_type, "query: ")}
    if data.geo:
        output["geo"] = {"text": data.geo, "embedding": e5_embs(data.geo, "query: ")}
    if data.production_method:
        output["production_method"] = [{"text": i, "embedding": e5_embs(i, "query: ")} for i in data.production_method]
    if data.usage_vibes:
        output["usage_vibes"] = [{"text": i, "embedding": e5_embs(i, "query: ")} for i in data.usage_vibes]
    return output

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)