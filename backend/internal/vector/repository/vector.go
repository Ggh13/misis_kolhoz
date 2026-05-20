package repository

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"misis_kolhoz/internal/vector/model"
)

type VectorRepository struct {
	pgDB *pgxpool.Pool
}

const (
	CreateExtensionVector = `CREATE EXTENSION IF NOT EXISTS vector`

	CreateProductEmbeddingsTable = `CREATE TABLE IF NOT EXISTS public.product_embeddings (
		id SERIAL PRIMARY KEY,
		product_id INTEGER UNIQUE REFERENCES farmer_products(id) ON DELETE CASCADE,
		farmer_id INTEGER,
		product_name VARCHAR(255),
		category VARCHAR(100),
		unit VARCHAR(50),
		price DECIMAL(10,2),
		quantity INTEGER,
		farmer_description TEXT,
		product_description TEXT,
		embedding vector(384)
	)`

	CreateEmbeddingIndex = `CREATE INDEX IF NOT EXISTS idx_product_embeddings_embedding
		ON product_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`

	SelectAllProductsForEmbedding = `SELECT fp.id, fp.farmer_id, fp.product_name, fp.category, fp.unit, fp.price, fp.quantity,
		COALESCE(f.farmer_description, ''), COALESCE(fp.product_description, '')
		FROM farmer_products fp
		LEFT JOIN farmers f ON f.id = fp.farmer_id`

	InsertProductEmbedding = `INSERT INTO product_embeddings
		(product_id, farmer_id, product_name, category, unit, price, quantity, farmer_description, product_description, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::vector)
		ON CONFLICT (product_id) DO UPDATE SET
			farmer_id = EXCLUDED.farmer_id,
			product_name = EXCLUDED.product_name,
			category = EXCLUDED.category,
			unit = EXCLUDED.unit,
			price = EXCLUDED.price,
			quantity = EXCLUDED.quantity,
			farmer_description = EXCLUDED.farmer_description,
			product_description = EXCLUDED.product_description,
			embedding = EXCLUDED.embedding`

	GetVectorByProductID = `SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity,
		farmer_description, product_description, embedding::text
		FROM product_embeddings WHERE product_id = $1`

	DeleteVectorByProductID = `DELETE FROM product_embeddings WHERE product_id = $1`

	SearchVectors = `SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity,
		farmer_description, product_description
		FROM product_embeddings ORDER BY embedding <=> $1::vector LIMIT $2`

	SearchVectorsByFarmer = `SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity,
		farmer_description, product_description
		FROM product_embeddings WHERE farmer_id = $3 ORDER BY embedding <=> $1::vector LIMIT $2`

	SearchEventVectors = `SELECT id, event_date, holiday_info, category, about, food_customs,
		embedding <=> $1::vector AS distance
		FROM event_embeddings
		WHERE event_date::date >= CURRENT_DATE
		ORDER BY distance ASC LIMIT $2`

	AlterProductEmbeddingsAddFarmerDescription = `ALTER TABLE public.product_embeddings
		ADD COLUMN IF NOT EXISTS farmer_description TEXT`

	AlterProductEmbeddingsAddProductDescription = `ALTER TABLE public.product_embeddings
		ADD COLUMN IF NOT EXISTS product_description TEXT`
)

func NewVectorRepository(pgDB *pgxpool.Pool) *VectorRepository {
	return &VectorRepository{pgDB: pgDB}
}

func (r *VectorRepository) Init(ctx context.Context) error {
	_, err := r.pgDB.Exec(ctx, CreateExtensionVector)
	if err != nil {
		return fmt.Errorf("vector create extension: %w", err)
	}
	_, err = r.pgDB.Exec(ctx, CreateProductEmbeddingsTable)
	if err != nil {
		return fmt.Errorf("vector create table: %w", err)
	}
	_, err = r.pgDB.Exec(ctx, AlterProductEmbeddingsAddFarmerDescription)
	if err != nil {
		return fmt.Errorf("vector alter table add farmer_description: %w", err)
	}
	_, err = r.pgDB.Exec(ctx, AlterProductEmbeddingsAddProductDescription)
	if err != nil {
		return fmt.Errorf("vector alter table add product_description: %w", err)
	}
	_, err = r.pgDB.Exec(ctx, CreateEmbeddingIndex)
	if err != nil {
		return fmt.Errorf("vector create index: %w", err)
	}
	return nil
}

func (r *VectorRepository) BulkInsertFromProducts(ctx context.Context) error {
	rows, err := r.pgDB.Query(ctx, SelectAllProductsForEmbedding)
	if err != nil {
		return fmt.Errorf("vector select products: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var productID, farmerID int
		var productName, category, unit, farmerDescription, productDescription string
		var price float64
		var quantity int
		if err := rows.Scan(&productID, &farmerID, &productName, &category, &unit, &price, &quantity, &farmerDescription, &productDescription); err != nil {
			return fmt.Errorf("vector scan product: %w", err)
		}

		zeroVec := make([]float32, 384)
		vecStr := formatVectorForSQL(zeroVec)

		_, err := r.pgDB.Exec(ctx, InsertProductEmbedding,
			productID, farmerID, productName, category, unit, price, quantity, farmerDescription, productDescription, vecStr)
		if err != nil {
			return fmt.Errorf("vector insert embedding for product %d: %w", productID, err)
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("vector rows error: %w", err)
	}

	if count > 0 {
		fmt.Printf("Created %d embedding rows\n", count)
	}
	return nil
}

func (r *VectorRepository) Upsert(ctx context.Context, productID int, embedding []float32) error {
	if len(embedding) != 384 {
		return fmt.Errorf("invalid embedding size: expected 384, got %d", len(embedding))
	}

	var farmerID int
	var productName, category, unit, farmerDescription, productDescription string
	var price float64
	var quantity int
	err := r.pgDB.QueryRow(ctx, `
		SELECT fp.farmer_id, fp.product_name, fp.category, fp.unit, fp.price, fp.quantity,
			COALESCE(f.farmer_description, ''), COALESCE(fp.product_description, '')
		FROM farmer_products fp
		LEFT JOIN farmers f ON f.id = fp.farmer_id
		WHERE fp.id = $1`, productID).
		Scan(&farmerID, &productName, &category, &unit, &price, &quantity, &farmerDescription, &productDescription)
	if err != nil {
		return fmt.Errorf("vector get product: %w", err)
	}

	vecStr := formatVectorForSQL(embedding)
	_, err = r.pgDB.Exec(ctx, InsertProductEmbedding,
		productID, farmerID, productName, category, unit, price, quantity, farmerDescription, productDescription, vecStr)
	if err != nil {
		return fmt.Errorf("vector upsert: %w", err)
	}
	return nil
}

func (r *VectorRepository) Get(ctx context.Context, productID int) (*model.ProductEmbedding, error) {
	row := r.pgDB.QueryRow(ctx, GetVectorByProductID, productID)

	var e model.ProductEmbedding
	var embeddingStr string
	err := row.Scan(&e.ID, &e.ProductID, &e.FarmerID, &e.ProductName, &e.Category,
		&e.Unit, &e.Price, &e.Quantity, &e.FarmerDescription, &e.ProductDescription, &embeddingStr)
	if err != nil {
		return nil, fmt.Errorf("vector get: %w", err)
	}
	e.Embedding = parseVectorString(embeddingStr)
	return &e, nil
}

func (r *VectorRepository) Update(ctx context.Context, productID int, embedding []float32) error {
	return r.Upsert(ctx, productID, embedding)
}

func (r *VectorRepository) Delete(ctx context.Context, productID int) error {
	_, err := r.pgDB.Exec(ctx, DeleteVectorByProductID, productID)
	if err != nil {
		return fmt.Errorf("vector delete: %w", err)
	}
	return nil
}

func (r *VectorRepository) Search(ctx context.Context, embedding []float32, limit int, farmerID int) ([]model.ProductEmbedding, error) {
	if len(embedding) != 384 {
		return nil, fmt.Errorf("invalid embedding size: expected 384, got %d", len(embedding))
	}

	vecStr := formatVectorForSQL(embedding)

	var rows pgx.Rows
	var err error
	if farmerID > 0 {
		rows, err = r.pgDB.Query(ctx, SearchVectorsByFarmer, vecStr, limit, farmerID)
	} else {
		rows, err = r.pgDB.Query(ctx, SearchVectors, vecStr, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()

	var results []model.ProductEmbedding
	for rows.Next() {
		var e model.ProductEmbedding
		if err := rows.Scan(&e.ID, &e.ProductID, &e.FarmerID, &e.ProductName, &e.Category,
			&e.Unit, &e.Price, &e.Quantity, &e.FarmerDescription, &e.ProductDescription); err != nil {
			return nil, fmt.Errorf("vector search scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *VectorRepository) SearchEvents(ctx context.Context, embedding []float32, limit int) ([]model.EventEmbedding, error) {
	if len(embedding) != 384 {
		return nil, fmt.Errorf("invalid embedding size: expected 384, got %d", len(embedding))
	}

	vecStr := formatVectorForSQL(embedding)

	rows, err := r.pgDB.Query(ctx, SearchEventVectors, vecStr, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search events: %w", err)
	}
	defer rows.Close()

	var results []model.EventEmbedding
	for rows.Next() {
		var e model.EventEmbedding
		if err := rows.Scan(&e.ID, &e.EventDate, &e.HolidayInfo, &e.Category,
			&e.About, &e.FoodCustoms, &e.Distance); err != nil {
			return nil, fmt.Errorf("vector search events scan: %w", err)
		}
		if math.IsNaN(e.Distance) || math.IsInf(e.Distance, 0) {
			e.Distance = 0
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *VectorRepository) MatchEventsToProducts(ctx context.Context, limit int, farmerID int, futureOnly bool) ([]model.EventProductsMatch, error) {
	tx, err := r.pgDB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("vector begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SET LOCAL ivfflat.probes = 100")
	if err != nil {
		return nil, fmt.Errorf("vector set probes: %w", err)
	}

	query := `
		SELECT e.id, e.event_date, e.holiday_info, e.category, e.about, e.food_customs,
			p.id, p.product_id, p.farmer_id, p.product_name, p.category, p.unit, p.price, p.quantity,
			(p.embedding <=> e.embedding) AS distance
		FROM event_embeddings e
		CROSS JOIN LATERAL (
			SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity, embedding
			FROM product_embeddings
			WHERE ($2 = 0 OR farmer_id = $2)
			ORDER BY embedding <=> e.embedding
			LIMIT $1
		) p
		WHERE ($3 = false OR e.event_date::date >= CURRENT_DATE)
		ORDER BY e.id, distance`

	rows, err := tx.Query(ctx, query, limit, farmerID, futureOnly)
	if err != nil {
		return nil, fmt.Errorf("vector match events to products: %w", err)
	}
	defer rows.Close()

	resultMap := make(map[int]*model.EventProductsMatch)
	order := make([]int, 0)
	for rows.Next() {
		var event model.EventEmbedding
		var product model.ProductMatch
		if err := rows.Scan(
			&event.ID, &event.EventDate, &event.HolidayInfo, &event.Category, &event.About, &event.FoodCustoms,
			&product.ID, &product.ProductID, &product.FarmerID, &product.ProductName, &product.Category,
			&product.Unit, &product.Price, &product.Quantity, &product.Distance,
		); err != nil {
			return nil, fmt.Errorf("vector match events scan: %w", err)
		}

		bucket, ok := resultMap[event.ID]
		if !ok {
			resultMap[event.ID] = &model.EventProductsMatch{Event: event, Products: []model.ProductMatch{product}}
			order = append(order, event.ID)
			continue
		}
		bucket.Products = append(bucket.Products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vector match events rows: %w", err)
	}

	result := make([]model.EventProductsMatch, 0, len(order))
	for _, id := range order {
		result = append(result, *resultMap[id])
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("vector match events commit: %w", err)
	}

	return result, nil
}

func (r *VectorRepository) MatchProductsToEvents(ctx context.Context, limit int, farmerID int, futureOnly bool) ([]model.ProductEventsMatch, error) {
	tx, err := r.pgDB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("vector begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "SET LOCAL ivfflat.probes = 100")
	if err != nil {
		return nil, fmt.Errorf("vector set probes: %w", err)
	}

	query := `
		SELECT p.id, p.product_id, p.farmer_id, p.product_name, p.category, p.unit, p.price, p.quantity,
			e.id, e.event_date, e.holiday_info, e.category, e.about, e.food_customs,
			(e.embedding <=> p.embedding) AS distance
		FROM product_embeddings p
		CROSS JOIN LATERAL (
			SELECT id, event_date, holiday_info, category, about, food_customs, embedding
			FROM event_embeddings
			WHERE ($3 = false OR event_date::date >= CURRENT_DATE)
			ORDER BY embedding <=> p.embedding
			LIMIT $1
		) e
		WHERE ($2 = 0 OR p.farmer_id = $2)
		ORDER BY p.product_id, distance`

	rows, err := tx.Query(ctx, query, limit, farmerID, futureOnly)
	if err != nil {
		return nil, fmt.Errorf("vector match products to events: %w", err)
	}
	defer rows.Close()

	resultMap := make(map[int]*model.ProductEventsMatch)
	order := make([]int, 0)
	for rows.Next() {
		var product model.ProductEmbedding
		var event model.EventEmbedding
		if err := rows.Scan(
			&product.ID, &product.ProductID, &product.FarmerID, &product.ProductName, &product.Category,
			&product.Unit, &product.Price, &product.Quantity,
			&event.ID, &event.EventDate, &event.HolidayInfo, &event.Category, &event.About, &event.FoodCustoms,
			&event.Distance,
		); err != nil {
			return nil, fmt.Errorf("vector match products scan: %w", err)
		}

		bucket, ok := resultMap[product.ProductID]
		if !ok {
			resultMap[product.ProductID] = &model.ProductEventsMatch{Product: product, Events: []model.EventEmbedding{event}}
			order = append(order, product.ProductID)
			continue
		}
		bucket.Events = append(bucket.Events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vector match products rows: %w", err)
	}

	result := make([]model.ProductEventsMatch, 0, len(order))
	for _, id := range order {
		result = append(result, *resultMap[id])
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("vector match products commit: %w", err)
	}

	return result, nil
}

func formatVectorForSQL(emb []float32) string {
	parts := make([]string, len(emb))
	for i, v := range emb {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func parseVectorString(s string) []float32 {
	s = strings.Trim(s, "{}[]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	vec := make([]float32, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			continue
		}
		vec = append(vec, float32(v))
	}
	return vec
}
