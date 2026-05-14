import json
import os
from typing import Any, Dict, List, TypedDict

from langchain_openai import ChatOpenAI
from langgraph.graph import END, StateGraph


def _build_llm() -> ChatOpenAI:
    return ChatOpenAI(
        api_key=os.getenv("GROQ_API_KEY"),
        base_url=os.getenv("GROQ_BASE_URL", "https://api.groq.com/openai/v1"),
        model=os.getenv("AGENT_MODEL", "llama-3.3-70b-versatile"),
        temperature=0.2,
        model_kwargs={"response_format": {"type": "json_object"}},
    )


class MarketingState(TypedDict):
    farmer: Dict[str, Any]
    product: Dict[str, Any]
    event: Dict[str, Any]
    match: Dict[str, Any]
    plan: Dict[str, Any]
    content: Dict[str, Any]
    validator_notes: List[str]
    plan_approved: bool
    retry_count: int


def strategist_node(state: MarketingState) -> MarketingState:
    llm = _build_llm()

    farmer_tags = state.get("farmer", {}).get("tags")
    event_name = state.get("event", {}).get("name")
    event_start = state.get("event", {}).get("date_start")

    system = (
        "Ты стратег-маркетолог. Возвращай ТОЛЬКО валидный JSON и пиши по-русски. "
        "Сформируй план для связки товар+событие: даты, механика, гипотеза, "
        "целевая аудитория. "
        "Запланируй прогрев за 2-3 дня до event_start, если возможно. "
        "Максимизируй продажи через логичные ассоциации между товаром и событием."
    )
    user = json.dumps(
        {
            "farmer": state.get("farmer", {}),
            "product": state.get("product", {}),
            "event": state.get("event", {}),
            "match": state.get("match", {}),
            "hints": {
                "farmer_tags": farmer_tags,
                "event_name": event_name,
                "event_start": event_start,
            },
        },
        ensure_ascii=False,
    )

    response = llm.invoke(
        [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ]
    )

    try:
        plan = json.loads(response.content)
    except json.JSONDecodeError:
        plan = {"error": "invalid_json", "raw": response.content}

    return {
        **state,
        "plan": plan,
        "plan_approved": False,
        "validator_notes": [],
    }


def smm_node(state: MarketingState) -> MarketingState:
    llm = _build_llm()

    system = (
        "Ты SMM-лид. Возвращай ТОЛЬКО валидный JSON и пиши по-русски. "
        "Ключи: post_text, story_bullets, push_title, push_text. "
        "push_text должен быть 100-120 символов. story_bullets — список коротких фраз. "
        "Если есть validator_notes, исправь все указанные проблемы."
    )
    user = json.dumps(
        {
            "plan": state.get("plan", {}),
            "brand_story": state.get("farmer", {}).get("brand_story"),
            "validator_notes": state.get("validator_notes", []),
        },
        ensure_ascii=False,
    )

    response = llm.invoke(
        [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ]
    )

    try:
        content = json.loads(response.content)
    except json.JSONDecodeError:
        content = {"error": "invalid_json", "raw": response.content}

    return {
        **state,
        "content": content,
    }


def validator_node(state: MarketingState) -> MarketingState:
    notes: List[str] = []
    content = state.get("content", {})

    push_text = content.get("push_text", "") if isinstance(content, dict) else ""
    if not push_text:
        notes.append("push_text is missing")
    elif not (100 <= len(push_text) <= 120):
        notes.append("push_text must be 100-120 chars")

    for key in ("post_text", "story_bullets", "push_title"):
        if key not in content:
            notes.append(f"{key} is missing")

    retry_count = state.get("retry_count", 0) + 1
    if notes and retry_count > 3:
        notes.append("max_retries_exceeded")
        return {
            **state,
            "validator_notes": notes,
            "plan_approved": True,
            "retry_count": retry_count,
        }

    plan_approved = len(notes) == 0

    return {
        **state,
        "validator_notes": notes,
        "plan_approved": plan_approved,
        "retry_count": retry_count,
    }


def build_marketing_graph():
    graph = StateGraph(MarketingState)
    graph.add_node("strategist", strategist_node)
    graph.add_node("smm", smm_node)
    graph.add_node("validator", validator_node)

    graph.set_entry_point("strategist")
    graph.add_edge("strategist", "smm")
    graph.add_edge("smm", "validator")

    def _route(state: MarketingState) -> str:
        return END if state.get("plan_approved") else "smm"

    graph.add_conditional_edges("validator", _route, {"smm": "smm", END: END})
    return graph.compile()
