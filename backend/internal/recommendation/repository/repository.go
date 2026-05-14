package recommendationrepository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	recommendationmodel "misis_kolhoz/internal/recommendation/model"
)

const (
	SelectProductsByIDs = `SELECT id, farmer_id, product_name, category, unit, price, quantity
		FROM public.farmer_products
		WHERE id = ANY($1)`

	SelectProductsByCategory = `SELECT id, farmer_id, product_name, category, unit, price, quantity
		FROM public.farmer_products
		WHERE category = $1
		ORDER BY id`
)

type Repository struct {
	pgDB *pgxpool.Pool
}

func NewRepository(pgDB *pgxpool.Pool) *Repository {
	return &Repository{pgDB: pgDB}
}

func (r *Repository) GetProductsByIDs(ctx context.Context, ids []int) ([]recommendationmodel.CatalogProduct, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.pgDB.Query(ctx, SelectProductsByIDs, ids)
	if err != nil {
		return nil, fmt.Errorf("recommendationrepository.GetProductsByIDs: %w", err)
	}
	defer rows.Close()

	var products []recommendationmodel.CatalogProduct
	for rows.Next() {
		var product recommendationmodel.CatalogProduct
		err := rows.Scan(
			&product.ID,
			&product.FarmerID,
			&product.ProductName,
			&product.Category,
			&product.Unit,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("recommendationrepository.GetProductsByIDs scan: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *Repository) GetProductsByCategory(ctx context.Context, category string) ([]recommendationmodel.CatalogProduct, error) {
	rows, err := r.pgDB.Query(ctx, SelectProductsByCategory, category)
	if err != nil {
		return nil, fmt.Errorf("recommendationrepository.GetProductsByCategory: %w", err)
	}
	defer rows.Close()

	var products []recommendationmodel.CatalogProduct
	for rows.Next() {
		var product recommendationmodel.CatalogProduct
		err := rows.Scan(
			&product.ID,
			&product.FarmerID,
			&product.ProductName,
			&product.Category,
			&product.Unit,
			&product.Price,
			&product.Quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("recommendationrepository.GetProductsByCategory scan: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}
