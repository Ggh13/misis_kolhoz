import json
import os
import time
from dataclasses import dataclass
from typing import Any, Dict, List, Optional

import pandas as pd
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
        return get_combined_text(product_features, farm_features)

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


def _normalize_text(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, float) and pd.isna(value):
        return ""
    return str(value).strip()


def extract_from_table_xlsx(table_path: str, extractor: FeatureExtractor) -> List[Dict[str, Any]]:
    df = pd.read_excel(table_path)
    results: List[Dict[str, Any]] = []

    for _, row in df.iterrows():
        farmer_text = _normalize_text(row.get("farmer_description", ""))
        product_text = _normalize_text(row.get("product_description", ""))

        farmer_features = (
            extractor.extract_farm_features(farmer_text) if farmer_text else {}
        )
        product_features = (
            extractor.extract_product_features(product_text) if product_text else {}
        )

        results.append(
            {
                "organization_id": row.get("organization_id"),
                "product_id": row.get("product_id"),
                "farmer_features": farmer_features,
                "product_features": product_features,
            }
        )

    return results


def extract_for_product_id(
    table_path: str, product_id: Any, extractor: FeatureExtractor
) -> Dict[str, Any]:
    df = pd.read_excel(table_path)
    matched = df[df["product_id"] == product_id]
    if matched.empty:
        raise ValueError(f"product_id not found: {product_id}")

    row = matched.iloc[0]
    farmer_text = _normalize_text(row.get("farmer_description", ""))
    product_text = _normalize_text(row.get("product_description", ""))

    farmer_features = (
        extractor.extract_farm_features(farmer_text) if farmer_text else {}
    )
    product_features = (
        extractor.extract_product_features(product_text) if product_text else {}
    )

    return {
        "organization_id": row.get("organization_id"),
        "product_id": row.get("product_id"),
        "farmer_features": farmer_features,
        "product_features": product_features,
    }


def get_combined_text(
    product_features: Dict[str, Any], farm_features: Dict[str, Any]
) -> str:
    def _stringify(value: Any) -> str:
        if not value:
            return ""
        if isinstance(value, list):
            items = [str(item).strip() for item in value if str(item).strip()]
            return "; ".join(items)
        return str(value).strip()

    farmer_block = "; ".join(
        filter(
            None,
            [
                _stringify(farm_features.get("location")),
                _stringify(farm_features.get("assets")),
                _stringify(farm_features.get("farm_values")),
                _stringify(farm_features.get("services")),
                _stringify(farm_features.get("brand_story")),
            ],
        )
    )

    product_block = "; ".join(
        filter(
            None,
            [
                _stringify(product_features.get("product_type")),
                _stringify(product_features.get("geo")),
                _stringify(product_features.get("production_method")),
                _stringify(product_features.get("usage_vibes")),
            ],
        )
    )

    return f"FARMER: {farmer_block}\nPRODUCT: {product_block}".strip()

def build_default_extractor() -> FeatureExtractor:
    load_dotenv()
    api_key = os.getenv("GROQ_API_KEY")
    return FeatureExtractor(GroqConfig(api_key=api_key))


__all__ = [
    "FeatureExtractor",
    "GroqConfig",
    "build_default_extractor",
    "extract_from_table_xlsx",
    "extract_for_product_id",
    "get_combined_text",
]