import json
import os
import time
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

from dotenv import load_dotenv
from openai import OpenAI, RateLimitError, APIError


PRODUCT_SYSTEM_PROMPT = """
Ты — AI-ассистент по извлечению признаков из описаний фермерских товаров.
Твоя задача — прочитать текст и извлечь ключевые маркетинговые смыслы в формате JSON.
Используй строго следующие ключи:
- "product_type": что это за продукт (строка)
- "geo": регион производства, если указан (строка или null)
- "production_method": особенности производства (массив строк)
- "usage_vibes": ситуации использования, польза, эмоции, применение (массив строк)

Отвечай ТОЛЬКО валидным JSON.
"""

FARM_SYSTEM_PROMPT = """
Ты — AI-маркетолог. Твоя задача — проанализировать описание фермерского хозяйства и
извлечь ключевые данные для построения бренда в формате JSON.
Используй строго следующие ключи:
- "location": точное местоположение (массив строк: регион, район, деревня).
- "assets": что еще есть на ферме, кроме основного товара (массив строк: другие культуры, животные, объекты).
- "farm_values": ценности и подход к работе (массив строк: экологичность, традиции и т.д.).
- "services": дополнительные услуги (массив строк: экскурсии, туризм, мастер-классы).
- "brand_story": саммари описания в 1-2 предложениях от первого лица (строка).

Отвечай ТОЛЬКО валидным JSON.
"""

@dataclass(frozen=True)
class GroqConfig:
    api_key: str
    base_url: str = "https://api.groq.com/openai/v1"
    model: str = "llama-3.3-70b-versatile"
    temperature: float = 0.0

class FeatureExtractor:
    def __init__(self, config: GroqConfig) -> None:
        self._client = OpenAI(base_url=config.base_url, api_key=config.api_key)
        self._model = config.model
        self._temperature = config.temperature

    def extract_product_features(self, text: str) -> Dict[str, Any]:
        return self._extract_json(text=text, system_prompt=PRODUCT_SYSTEM_PROMPT)

    def extract_farm_features(self, text: str) -> Dict[str, Any]:
        return self._extract_json(text=text, system_prompt=FARM_SYSTEM_PROMPT)

    def get_combined_text(self, product_features: Dict[str, Any], farm_features: Dict[str, Any]) -> str:

        prod_type = product_features.get("product_type", "товар")
        methods = ", ".join(product_features.get("production_method", []))
        vibes = ", ".join(product_features.get("usage_vibes", []))
        
        loc = ", ".join(farm_features.get("location", []))
        assets = ", ".join(farm_features.get("assets", []))
        values = ", ".join(farm_features.get("farm_values", []))
        
        combined = (
            f"Продукт: {prod_type}. Особенности производства: {methods}. "
            f"Применение и эффект: {vibes}. "
            f"Локация фермы: {loc}. Дополнительно на ферме: {assets}. "
            f"Ценности: {values}."
        )
        return combined

    def _extract_json(self, text: str, system_prompt: str) -> Dict[str, Any]:
        try:
            response = self._client.chat.completions.create(
                model=self._model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": text},
                ],
                temperature=self._temperature,
                response_format={"type": "json_object"},
            )
            content = response.choices[0].message.content or "{}"
            return json.loads(content)
        except RateLimitError:
            time.sleep(10)
            return self._extract_json(text, system_prompt)
        except (json.JSONDecodeError, APIError) as e:
            return {}

def build_default_extractor() -> FeatureExtractor:
    load_dotenv()
    api_key = os.getenv("GROQ_API_KEY")
    return FeatureExtractor(GroqConfig(api_key=api_key))