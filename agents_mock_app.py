from fastapi import FastAPI

app = FastAPI(title="Agents Mock Service")


@app.post("/agents/run_dynamic")
def run_dynamic(payload: dict):
    product_id = payload.get("product_id", 1)
    return {
        "farmer": {"id": 1001, "description": "mock farmer"},
        "product": {"id": product_id, "name": "Творог"},
        "event": {"id": 7, "name": "Пасха", "description": "mock event"},
        "match": {"id": f"product:{product_id}|event:7", "score": 0.12},
        "plan": {"hypothesis": "Запустить пасхальный набор"},
        "content": {
            "post_text": "Mock пост",
            "story_bullets": ["Mock сторис 1", "Mock сторис 2"],
            "push_title": "Mock push",
            "push_text": "Mock push text в диапазоне 100-120 символов для демо интерфейса и тестового сценария работы вкладки ML сервисов.",
        },
        "validator_notes": [],
        "plan_approved": True,
        "retry_count": 0,
    }
