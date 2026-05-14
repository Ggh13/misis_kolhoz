from typing import List, Optional

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from embedder.emb import build_event_payload, build_product_payload
from extractor import build_default_extractor

load_dotenv()

app = FastAPI()
extractor = build_default_extractor()


class ExtractRequest(BaseModel):
	product_description: str
	farmer_description: str


class ProductData(BaseModel):
	product_type: Optional[str] = None
	geo: Optional[str] = None
	production_method: Optional[List[str]] = None
	usage_vibes: Optional[List[str]] = None


class FarmerData(BaseModel):
	location: Optional[List[str]] = None
	assets: Optional[List[str]] = None
	farm_values: Optional[List[str]] = None
	services: Optional[List[str]] = None
	brand_story: Optional[str] = None


class EmbedProductRequest(BaseModel):
	product_features: ProductData
	farmer_features: Optional[FarmerData] = None


class EmbedEventRequest(BaseModel):
	event_name: str
	event_description: str


class FullPipelineRequest(BaseModel):
	product_description: str
	farmer_description: str
	event_name: str
	event_description: str


@app.post("/extract")
def extract_features(data: ExtractRequest):
	product_text = data.product_description.strip()
	farmer_text = data.farmer_description.strip()

	product_features = (
		extractor.extract_product_features(product_text) if product_text else {}
	)
	farmer_features = (
		extractor.extract_farm_features(farmer_text) if farmer_text else {}
	)

	return {
		"product_features": product_features,
		"farmer_features": farmer_features,
	}


@app.post("/embed/product")
def embed_product(data: EmbedProductRequest):
	product_features = data.product_features.dict(exclude_none=True)
	farmer_features = (
		data.farmer_features.dict(exclude_none=True) if data.farmer_features else {}
	)
	return build_product_payload(product_features, farmer_features)


@app.post("/embed/event")
def embed_event(data: EmbedEventRequest):
	try:
		return build_event_payload(data.event_name, data.event_description)
	except ValueError:
		raise HTTPException(status_code=400, detail="Injection detected")


@app.post("/pipeline/full")
def process_full_pipeline(data: FullPipelineRequest):
	product_text = data.product_description.strip()
	farmer_text = data.farmer_description.strip()

	product_features = (
		extractor.extract_product_features(product_text) if product_text else {}
	)
	farmer_features = (
		extractor.extract_farm_features(farmer_text) if farmer_text else {}
	)

	product_payload = build_product_payload(product_features, farmer_features)
	try:
		event_payload = build_event_payload(data.event_name, data.event_description)
	except ValueError:
		raise HTTPException(status_code=400, detail="Injection detected")

	return {
		"product_features": product_features,
		"farmer_features": farmer_features,
		"combined_text": product_payload["combined_text"],
		"product_embedding": product_payload["embedding"],
		"event": event_payload,
	}


if __name__ == "__main__":
	import uvicorn

	uvicorn.run(app, host="0.0.0.0", port=8000)