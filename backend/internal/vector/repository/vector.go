package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
		embedding vector(384)
	)`

	CreateEmbeddingIndex = `CREATE INDEX IF NOT EXISTS idx_product_embeddings_embedding
		ON product_embeddings USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)`

	SelectAllProductsForEmbedding = `SELECT id, farmer_id, product_name, category, unit, price, quantity FROM farmer_products`

	InsertProductEmbedding = `INSERT INTO product_embeddings (product_id, farmer_id, product_name, category, unit, price, quantity, embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector)
		ON CONFLICT (product_id) DO UPDATE SET
			farmer_id = EXCLUDED.farmer_id,
			product_name = EXCLUDED.product_name,
			category = EXCLUDED.category,
			unit = EXCLUDED.unit,
			price = EXCLUDED.price,
			quantity = EXCLUDED.quantity,
			embedding = EXCLUDED.embedding`

	GetVectorByProductID = `SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity, embedding::text
		FROM product_embeddings WHERE product_id = $1`

	DeleteVectorByProductID = `DELETE FROM product_embeddings WHERE product_id = $1`

	SearchVectors = `SELECT id, product_id, farmer_id, product_name, category, unit, price, quantity
		FROM product_embeddings ORDER BY embedding <=> $1::vector LIMIT $2`
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
		var productName, category, unit string
		var price float64
		var quantity int
		if err := rows.Scan(&productID, &farmerID, &productName, &category, &unit, &price, &quantity); err != nil {
			return fmt.Errorf("vector scan product: %w", err)
		}

		zeroVec := make([]float32, 384)
		vecStr := formatVectorForSQL(zeroVec)

		_, err := r.pgDB.Exec(ctx, InsertProductEmbedding,
			productID, farmerID, productName, category, unit, price, quantity, vecStr)
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
	var productName, category, unit string
	var price float64
	var quantity int
	err := r.pgDB.QueryRow(ctx, `
		SELECT farmer_id, product_name, category, unit, price, quantity
		FROM farmer_products WHERE id = $1`, productID).
		Scan(&farmerID, &productName, &category, &unit, &price, &quantity)
	if err != nil {
		return fmt.Errorf("vector get product: %w", err)
	}

	vecStr := formatVectorForSQL(embedding)
	_, err = r.pgDB.Exec(ctx, InsertProductEmbedding,
		productID, farmerID, productName, category, unit, price, quantity, vecStr)
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
		&e.Unit, &e.Price, &e.Quantity, &embeddingStr)
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

func (r *VectorRepository) Search(ctx context.Context, embedding []float32, limit int) ([]model.ProductEmbedding, error) {
	if len(embedding) != 384 {
		return nil, fmt.Errorf("invalid embedding size: expected 384, got %d", len(embedding))
	}

	vecStr := formatVectorForSQL(embedding)

	rows, err := r.pgDB.Query(ctx, SearchVectors, vecStr, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search: %w", err)
	}
	defer rows.Close()

	var results []model.ProductEmbedding
	for rows.Next() {
		var e model.ProductEmbedding
		if err := rows.Scan(&e.ID, &e.ProductID, &e.FarmerID, &e.ProductName, &e.Category,
			&e.Unit, &e.Price, &e.Quantity); err != nil {
			return nil, fmt.Errorf("vector search scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func formatVectorForSQL(emb []float32) string {
	parts := make([]string, len(emb))
	for i, v := range emb {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func parseVectorString(s string) []float32 {
	s = strings.Trim(s, "{}")
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
