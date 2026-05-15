package farmerrepository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	farmermodel "misis_kolhoz/internal/farmer/model"
)

const (
	CreateFarmersTable = `CREATE TABLE IF NOT EXISTS public.farmers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		region VARCHAR(255),
		address TEXT,
		phone VARCHAR(50),
		email VARCHAR(255),
		farmer_description TEXT
	)`

	CreateFarmerProductsTable = `CREATE TABLE IF NOT EXISTS public.farmer_products (
		id INTEGER PRIMARY KEY,
		farmer_id INTEGER REFERENCES farmers(id),
		product_name VARCHAR(255) NOT NULL,
		category VARCHAR(100),
		unit VARCHAR(50),
		price DECIMAL(10,2),
		quantity INTEGER,
		product_description TEXT
	)`

	InsertFarmer = `INSERT INTO public.farmers (id, name, region, address, phone, email, farmer_description)
		VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			region = EXCLUDED.region,
			address = EXCLUDED.address,
			phone = EXCLUDED.phone,
			email = EXCLUDED.email,
			farmer_description = EXCLUDED.farmer_description`

	InsertFarmerProduct = `INSERT INTO public.farmer_products
		(id, farmer_id, product_name, category, unit, price, quantity, product_description)
		VALUES($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			farmer_id = EXCLUDED.farmer_id,
			product_name = EXCLUDED.product_name,
			category = EXCLUDED.category,
			unit = EXCLUDED.unit,
			price = EXCLUDED.price,
			quantity = EXCLUDED.quantity,
			product_description = EXCLUDED.product_description`

	SelectFarmerByID = `SELECT id, name, region, address, phone, email, farmer_description FROM farmers WHERE id = $1`

	SelectFarmerProductsByFarmerID = `SELECT id, farmer_id, product_name, category, unit, price, quantity, product_description
		FROM farmer_products WHERE farmer_id = $1`

	AlterFarmersAddDescription = `ALTER TABLE public.farmers
		ADD COLUMN IF NOT EXISTS farmer_description TEXT`

	AlterFarmerProductsAddDescription = `ALTER TABLE public.farmer_products
		ADD COLUMN IF NOT EXISTS product_description TEXT`
)

type Repository struct {
	pgDB *pgxpool.Pool
}

func NewRepository(pgDB *pgxpool.Pool) *Repository {
	return &Repository{pgDB: pgDB}
}

func (r *Repository) InitTables(ctx context.Context) error {
	_, err := r.pgDB.Exec(ctx, CreateFarmersTable)
	if err != nil {
		return fmt.Errorf("farmerrepository.CreateFarmersTable: %w", err)
	}

	_, err = r.pgDB.Exec(ctx, CreateFarmerProductsTable)
	if err != nil {
		return fmt.Errorf("farmerrepository.CreateFarmerProductsTable: %w", err)
	}

	_, err = r.pgDB.Exec(ctx, AlterFarmersAddDescription)
	if err != nil {
		return fmt.Errorf("farmerrepository.AlterFarmersAddDescription: %w", err)
	}

	_, err = r.pgDB.Exec(ctx, AlterFarmerProductsAddDescription)
	if err != nil {
		return fmt.Errorf("farmerrepository.AlterFarmerProductsAddDescription: %w", err)
	}

	return nil
}

func (r *Repository) AddFarmer(ctx context.Context, farmer farmermodel.Farmer) error {
	_, err := r.pgDB.Exec(ctx, InsertFarmer,
		farmer.ID,
		farmer.Name,
		farmer.Region,
		farmer.Address,
		farmer.Phone,
		farmer.Email,
		farmer.FarmerDescription,
	)
	if err != nil {
		return fmt.Errorf("farmerrepository.AddFarmer: %w", err)
	}
	return nil
}

func (r *Repository) AddFarmerProduct(ctx context.Context, product farmermodel.FarmerProduct) error {
	_, err := r.pgDB.Exec(ctx, InsertFarmerProduct,
		product.ID,
		product.FarmerID,
		product.ProductName,
		product.Category,
		product.Unit,
		product.Price,
		product.Quantity,
		product.ProductDescription,
	)
	if err != nil {
		return fmt.Errorf("farmerrepository.AddFarmerProduct: %w", err)
	}
	return nil
}

func (r *Repository) GetFarmerByID(ctx context.Context, id int) (farmermodel.Farmer, error) {
	var farmer farmermodel.Farmer
	err := r.pgDB.QueryRow(ctx, SelectFarmerByID, id).Scan(
		&farmer.ID,
		&farmer.Name,
		&farmer.Region,
		&farmer.Address,
		&farmer.Phone,
		&farmer.Email,
		&farmer.FarmerDescription,
	)
	if err != nil {
		return farmer, fmt.Errorf("farmerrepository.GetFarmerByID: %w", err)
	}
	return farmer, nil
}

func (r *Repository) GetFarmerProducts(ctx context.Context, farmerID int) ([]farmermodel.FarmerProduct, error) {
	rows, err := r.pgDB.Query(ctx, SelectFarmerProductsByFarmerID, farmerID)
	if err != nil {
		return nil, fmt.Errorf("farmerrepository.GetFarmerProducts: %w", err)
	}
	defer rows.Close()

	var products []farmermodel.FarmerProduct
	for rows.Next() {
		var p farmermodel.FarmerProduct
		err := rows.Scan(
			&p.ID,
			&p.FarmerID,
			&p.ProductName,
			&p.Category,
			&p.Unit,
			&p.Price,
			&p.Quantity,
			&p.ProductDescription,
		)
		if err != nil {
			return nil, fmt.Errorf("farmerrepository.GetFarmerProducts scan: %w", err)
		}
		products = append(products, p)
	}

	return products, nil
}
