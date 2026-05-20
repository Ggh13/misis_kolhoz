import argparse
import os
import sys
import time
from typing import Dict, Iterable, List, Optional, Set

import psycopg
from pgvector.psycopg import register_vector

PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

from embedder.emb import build_product_payload, build_event_payload, e5_embs
from extractor import FeatureExtractor, build_default_extractor


DEFAULT_DB_URL = "postgresql://postgres:password@localhost:5432/mydb"
BATCH_SIZE = 25


def get_db_url() -> str:
    return os.getenv("DATABASE_URL", DEFAULT_DB_URL)


def build_event_text(row: Dict[str, str]) -> str:
    holiday_info = (row.get("holiday_info") or "").strip()
    category = (row.get("category") or "").strip()
    about = (row.get("about") or "").strip()
    food_customs = (row.get("food_customs") or "").strip()
    event_date = (row.get("event_date") or "").strip()

    event_name = holiday_info or category or "event"
    desc_parts = [about, food_customs, category, event_date]
    description = ". ".join([p for p in desc_parts if p])

    if description:
        return f"{event_name}. {description}"
    return event_name


def build_raw_product_text(row: Dict[str, object]) -> str:
    farmer_text = str(row.get("farmer_description") or "").strip()
    product_text = str(row.get("product_description") or "").strip()
    combined = f"FARMER: {farmer_text}\nPRODUCT: {product_text}".strip()
    return combined if combined else "product"


def fetch_all_products(conn: psycopg.Connection) -> List[Dict[str, object]]:
    query = """
        SELECT fp.id AS product_id, fp.farmer_id, fp.product_name, fp.category, fp.unit,
               fp.price, fp.quantity,
               COALESCE(f.farmer_description, '') AS farmer_description,
               COALESCE(fp.product_description, '') AS product_description
        FROM farmer_products fp
        LEFT JOIN farmers f ON f.id = fp.farmer_id
        ORDER BY fp.id
    """
    with conn.cursor() as cur:
        cur.execute(query)
        return cur.fetchall()


def fetch_all_events(conn: psycopg.Connection) -> List[Dict[str, object]]:
    query = """
        SELECT id, event_date, holiday_info, category, about, food_customs
        FROM event_embeddings
        ORDER BY id
    """
    with conn.cursor() as cur:
        cur.execute(query)
        return cur.fetchall()


def upsert_product_embedding(
    conn: psycopg.Connection, row: Dict[str, object], embedding: List[float]
) -> None:
    query = """
        INSERT INTO product_embeddings
            (product_id, farmer_id, product_name, category, unit, price, quantity,
             farmer_description, product_description, embedding)
        VALUES
            (%(product_id)s, %(farmer_id)s, %(product_name)s, %(category)s, %(unit)s,
             %(price)s, %(quantity)s, %(farmer_description)s, %(product_description)s,
             %(embedding)s)
        ON CONFLICT (product_id) DO UPDATE SET
            farmer_id = EXCLUDED.farmer_id,
            product_name = EXCLUDED.product_name,
            category = EXCLUDED.category,
            unit = EXCLUDED.unit,
            price = EXCLUDED.price,
            quantity = EXCLUDED.quantity,
            farmer_description = EXCLUDED.farmer_description,
            product_description = EXCLUDED.product_description,
            embedding = EXCLUDED.embedding
    """
    payload = dict(row)
    payload["embedding"] = embedding
    with conn.cursor() as cur:
        cur.execute(query, payload)


def update_event_embedding(
    conn: psycopg.Connection, event_id: int, embedding: List[float]
) -> None:
    query = """
        UPDATE event_embeddings
        SET embedding = %(embedding)s
        WHERE id = %(id)s
    """
    with conn.cursor() as cur:
        cur.execute(query, {"id": event_id, "embedding": embedding})


def ensure_vector_extension(conn: psycopg.Connection) -> None:
    with conn.cursor() as cur:
        cur.execute("CREATE EXTENSION IF NOT EXISTS vector")


def reseed_product_embeddings(
    conn: psycopg.Connection,
    extractor: FeatureExtractor,
    extract_farmer_ids: Optional[Set[int]] = None,
) -> None:
    rows = fetch_all_products(conn)
    total = len(rows)
    if total == 0:
        print("No products found. Skipping product embeddings.")
        return

    print(f"Found {total} products. Generating embeddings...")
    updated = 0
    failed = 0
    used_extract = 0
    used_raw = 0

    for idx, row in enumerate(rows, start=1):
        try:
            farmer_id = int(row.get("farmer_id") or 0)
            use_extract = extract_farmer_ids is None or farmer_id in extract_farmer_ids

            if use_extract:
                farmer_text = str(row.get("farmer_description") or "")
                product_text = str(row.get("product_description") or "")
                farmer_features = (
                    extractor.extract_farm_features(farmer_text) if farmer_text.strip() else {}
                )
                product_features = (
                    extractor.extract_product_features(product_text) if product_text.strip() else {}
                )
                payload = build_product_payload(product_features, farmer_features)
                embedding = payload["embedding"]
                used_extract += 1
            else:
                raw_text = build_raw_product_text(row)
                embedding = e5_embs(raw_text, "passage: ")
                used_raw += 1

            if len(embedding) != 384:
                raise ValueError(f"product {row.get('product_id')} embedding size {len(embedding)}")

            upsert_product_embedding(conn, row, embedding)
            updated += 1
        except Exception as exc:
            failed += 1
            print(f"Product {row.get('product_id')} failed: {exc}")

        if idx % BATCH_SIZE == 0:
            conn.commit()
            print(f"Progress: {idx}/{total} products")
            time.sleep(0.2)

    conn.commit()
    print(
        f"Product embeddings updated: {updated}, failed: {failed}, "
        f"extract: {used_extract}, raw: {used_raw}"
    )


def reseed_event_embeddings(conn: psycopg.Connection) -> None:
    rows = fetch_all_events(conn)
    total = len(rows)
    if total == 0:
        print("No events found. Skipping event embeddings.")
        return

    print(f"Found {total} events. Generating embeddings...")
    updated = 0
    failed = 0

    for idx, row in enumerate(rows, start=1):
        try:
            event_name = (row.get("holiday_info") or row.get("category") or "event").strip()
            about = (row.get("about") or "").strip()
            food_customs = (row.get("food_customs") or "").strip()
            desc = f"{about} {food_customs}".strip()

            payload = build_event_payload(event_name, desc)
            embedding = payload["embedding"]

            if len(embedding) != 384:
                raise ValueError(f"event {row.get('id')} embedding size {len(embedding)}")

            update_event_embedding(conn, int(row["id"]), embedding)
            updated += 1
        except Exception as exc:
            failed += 1
            print(f"Event {row.get('id')} failed: {exc}")

        if idx % BATCH_SIZE == 0:
            conn.commit()
            print(f"Progress: {idx}/{total} events")
            time.sleep(0.1)

    conn.commit()
    print(f"Event embeddings updated: {updated}, failed: {failed}")


def parse_farmer_ids(value: str) -> Set[int]:
    parts = [p.strip() for p in value.split(",") if p.strip()]
    ids: Set[int] = set()
    for part in parts:
        ids.add(int(part))
    return ids


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Reseed product and event embeddings in Postgres."
    )
    parser.add_argument(
        "--farmer-ids",
        help="Comma-separated farmer IDs to use entity extraction for. Others use raw text.",
    )
    args = parser.parse_args()

    extract_farmer_ids = None
    if args.farmer_ids:
        extract_farmer_ids = parse_farmer_ids(args.farmer_ids)

    db_url = get_db_url()
    print("Connecting to database...")

    try:
        conn = psycopg.connect(db_url, row_factory=psycopg.rows.dict_row)
    except Exception as exc:
        print(f"Failed to connect to DB: {exc}")
        return 1

    with conn:
        register_vector(conn)
        ensure_vector_extension(conn)
        extractor = build_default_extractor()

        if not os.getenv("GROQ_API_KEY"):
            print("GROQ_API_KEY is missing. Set it before running.")
            return 1

        reseed_product_embeddings(conn, extractor, extract_farmer_ids)
        reseed_event_embeddings(conn)

    print("Done.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
