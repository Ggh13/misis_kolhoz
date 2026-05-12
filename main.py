import json

from extractor import build_default_extractor, extract_for_product_id


def _print_json(title: str, data: dict) -> None:
	print(f"\n{title}")
	print(json.dumps(data, ensure_ascii=False, indent=2))

id = 180713
extractor = build_default_extractor()
result = extract_for_product_id("table.xlsx", id, extractor)

_print_json("farmer_features", result.get("farmer_features", {}))
_print_json("product_features", result.get("product_features", {}))