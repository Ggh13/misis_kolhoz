import math
import os
import random
from typing import Any, Dict

import psycopg
import requests
from pgvector.psycopg import register_vector
from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

try:
    from agents.marketing_graph import build_marketing_graph
except ModuleNotFoundError:
    from marketing_graph import build_marketing_graph

load_dotenv()

app = FastAPI()
workflow = build_marketing_graph()


class RunRequest(BaseModel):
    match_id: str


class RawRunRequest(BaseModel):
    farmer: Dict[str, Any]
    product: Dict[str, Any]
    event: Dict[str, Any]
    match: Dict[str, Any]


class DynamicRunRequest(BaseModel):
    product_id: int
    date_from: str
    date_to: str
    top_k: int = 1


class ImageRequest(BaseModel):
    prompt: str


def _sanitize(obj: Any) -> Any:
    if isinstance(obj, float):
        if math.isnan(obj) or math.isinf(obj):
            return None
        return obj
    if isinstance(obj, dict):
        return {k: _sanitize(v) for k, v in obj.items()}
    if isinstance(obj, list):
        return [_sanitize(v) for v in obj]
    return obj


def _get_db_url() -> str:
    url = os.getenv("DATABASE_URL")
    if not url:
        raise RuntimeError("DATABASE_URL is required")
    return url


def _get_pixazo_key() -> str:
    key = os.getenv("PIXAZO_SUBSCRIPTION_KEY")
    if not key:
        raise RuntimeError("PIXAZO_SUBSCRIPTION_KEY is required")
    return key


def _get_pixazo_url() -> str:
    return os.getenv("PIXAZO_URL", "https://gateway.pixazo.ai/flux-1-schnell/v1/getData")


def _fetch_context(match_id: str) -> Dict[str, Any]:
    query = """
        SELECT
            m.id AS match_id,
            m.score AS match_score,
            p.id AS product_id,
            p.farmer_id AS farmer_id,
            p.description AS product_description,
            p.tags_json AS product_tags,
            f.description AS farmer_description,
            f.tags_json AS farmer_tags,
            f.brand_story AS brand_story,
            e.id AS event_id,
            e.name AS event_name,
            e.description AS event_description,
            e.date_start AS event_start,
            e.date_end AS event_end
        FROM matches m
        JOIN products p ON p.id = m.product_id
        JOIN farmers f ON f.id = p.farmer_id
        JOIN events e ON e.id = m.event_id
        WHERE m.id = %s
    """

    with psycopg.connect(_get_db_url()) as conn:
        with conn.cursor() as cur:
            cur.execute(query, (match_id,))
            row = cur.fetchone()

    if not row:
        raise HTTPException(status_code=404, detail="match not found")

    (
        match_id,
        match_score,
        product_id,
        farmer_id,
        product_description,
        product_tags,
        farmer_description,
        farmer_tags,
        brand_story,
        event_id,
        event_name,
        event_description,
        event_start,
        event_end,
    ) = row

    return {
        "farmer": {
            "id": farmer_id,
            "description": farmer_description,
            "tags": farmer_tags,
            "brand_story": brand_story,
        },
        "product": {
            "id": product_id,
            "farmer_id": farmer_id,
            "description": product_description,
            "tags": product_tags,
        },
        "event": {
            "id": event_id,
            "name": event_name,
            "description": event_description,
            "date_start": event_start,
            "date_end": event_end,
        },
        "match": {
            "id": match_id,
            "score": match_score,
        },
    }


def _fetch_dynamic_context(
    product_id: int, date_from: str, date_to: str, top_k: int
) -> Dict[str, Any]:
    product_query = """
        SELECT
            product_id,
            farmer_id,
            product_name,
            category,
            farmer_description,
            product_description,
            embedding
        FROM product_embeddings
        WHERE product_id = %s
    """

    event_query = """
        SELECT
            id,
            event_date,
            holiday_info,
            category,
            about,
            food_customs,
            embedding <=> %s AS distance
        FROM event_embeddings
        WHERE event_date::date BETWEEN %s::date AND %s::date
        ORDER BY distance ASC
        LIMIT %s
    """

    with psycopg.connect(_get_db_url()) as conn:
        register_vector(conn)
        with conn.cursor() as cur:
            cur.execute(product_query, (product_id,))
            product_row = cur.fetchone()
            if not product_row:
                raise HTTPException(status_code=404, detail="product not found")

            (
                product_id,
                farmer_id,
                product_name,
                category,
                farmer_description,
                product_description,
                embedding,
            ) = product_row

            cur.execute(event_query, (embedding, date_from, date_to, top_k))
            event_rows = cur.fetchall()

    if not event_rows:
        raise HTTPException(status_code=404, detail="no events in date range")

    event_row = event_rows[0]
    (
        event_id,
        event_date,
        holiday_info,
        event_category,
        about,
        food_customs,
        distance,
    ) = event_row

    return {
        "farmer": {
            "id": farmer_id,
            "description": farmer_description,
        },
        "product": {
            "id": product_id,
            "farmer_id": farmer_id,
            "name": product_name,
            "category": category,
            "description": product_description,
        },
        "event": {
            "id": event_id,
            "name": holiday_info,
            "description": about,
            "food_customs": food_customs,
            "date": event_date,
            "category": event_category,
        },
        "match": {
            "id": f"product:{product_id}|event:{event_id}",
            "score": float(distance),
        },
    }


def _run_workflow(state: Dict[str, Any]) -> Dict[str, Any]:
    result = _sanitize(workflow.invoke(state))
    return _normalize_plan(result)


def _normalize_plan(result: Dict[str, Any]) -> Dict[str, Any]:
    plan = result.get("plan")
    if not isinstance(plan, dict):
        return result

    nested = plan.get("plan")
    if isinstance(nested, dict):
        plan = nested
        result["plan"] = plan

    promotions = plan.get("promotions")
    if isinstance(promotions, list):
        return result

    if isinstance(promotions, dict):
        items = promotions.get("items")
        if isinstance(items, list):
            plan["promotions"] = items
            return result

    if isinstance(promotions, str):
        lines = [line.strip().lstrip("•-") for line in promotions.splitlines()]
        clean_lines = [line.strip() for line in lines if line.strip()]
        if clean_lines:
            plan["promotions"] = [
                {"product": "", "promo": line, "reason": ""} for line in clean_lines
            ]
            return result

    fallback = plan.get("recommendations") or plan.get("promo_recommendations")
    if isinstance(fallback, list):
        plan["promotions"] = fallback

    return result


@app.post("/agents/run")
def run_agents(data: RunRequest):
    context = _fetch_context(data.match_id)
    state = {
        **_sanitize(context),
        "plan": {},
        "content": {},
        "validator_notes": [],
        "plan_approved": False,
        "retry_count": 0,
    }
    return _run_workflow(state)


@app.post("/agents/run_raw")
def run_agents_raw(data: RawRunRequest):
    state = {
        "farmer": data.farmer,
        "product": data.product,
        "event": data.event,
        "match": data.match,
        "plan": {},
        "content": {},
        "validator_notes": [],
        "plan_approved": False,
        "retry_count": 0,
    }
    return _run_workflow(state)


@app.post("/agents/generate_image")
def generate_image(data: ImageRequest):
    prompt = data.prompt.strip()
    if not prompt:
        raise HTTPException(status_code=400, detail="prompt is required")

    try:
        key = _get_pixazo_key()
    except RuntimeError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc

    payload = {
        "prompt": prompt,
        "num_steps": 4,
        "seed": random.randint(1, 10_000_000),
        "height": 512,
        "width": 512,
    }
    headers = {
        "Content-Type": "application/json",
        "Cache-Control": "no-cache",
        "Ocp-Apim-Subscription-Key": key,
    }

    try:
        response = requests.post(_get_pixazo_url(), json=payload, headers=headers, timeout=60)
    except requests.RequestException as exc:
        raise HTTPException(status_code=502, detail="image generation request failed") from exc

    if response.status_code >= 400:
        raise HTTPException(status_code=response.status_code, detail=response.text)

    data = response.json()
    output_url = data.get("output")
    if not output_url:
        raise HTTPException(status_code=502, detail="image generation returned empty output")

    return {"image_url": output_url, "prompt": prompt}


@app.post("/agents/run_dynamic")
def run_agents_dynamic(data: DynamicRunRequest):
    context = _fetch_dynamic_context(
        data.product_id, data.date_from, data.date_to, data.top_k
    )
    state = {
        **_sanitize(context),
        "plan": {},
        "content": {},
        "validator_notes": [],
        "plan_approved": False,
        "retry_count": 0,
    }
    return _run_workflow(state)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8010)
