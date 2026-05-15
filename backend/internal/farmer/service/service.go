package farmerservice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
	farmermodel "misis_kolhoz/internal/farmer/model"
)

type Repository interface {
	InitTables(ctx context.Context) error
	AddFarmer(ctx context.Context, farmer farmermodel.Farmer) error
	AddFarmerProduct(ctx context.Context, product farmermodel.FarmerProduct) error
	GetFarmerByID(ctx context.Context, id int) (farmermodel.Farmer, error)
	GetFarmerProducts(ctx context.Context, farmerID int) ([]farmermodel.FarmerProduct, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LoadDataFromExcel(ctx context.Context, filePath string) error {
	err := s.repo.InitTables(ctx)
	if err != nil {
		return fmt.Errorf("farmerservice.LoadDataFromExcel init tables: %w", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("farmerservice.LoadDataFromExcel open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetMap()
	for _, sheetName := range sheets {
		err := s.processSheet(ctx, f, sheetName)
		if err != nil {
			return fmt.Errorf("farmerservice.LoadDataFromExcel process sheet: %w", err)
		}
	}

	return nil
}

func (s *Service) processSheet(ctx context.Context, f *excelize.File, sheetName string) error {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("farmerservice.processSheet get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil
	}

	headers := rows[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	orgIDIdx := headerMap["organization_id"]
	shopNameIdx := headerMap["shop_name"]
	regionIdx := headerMap["region"]
	farmerDescIdx := headerMap["farmer_description"]
	productNameIdx := headerMap["name_product"]
	categoryIdx := headerMap["category"]
	priceIdx := headerMap["price"]
	productIDIdx := headerMap["product_id"]
	productDescIdx := headerMap["product_description"]

	currentFarmerID := 0

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]

		if len(row) <= orgIDIdx {
			continue
		}

		var farmerID int
		if orgIDIdx < len(row) && row[orgIDIdx] != "" {
			fmt.Sscanf(row[orgIDIdx], "%d", &farmerID)
		}

		if farmerID != currentFarmerID && farmerID > 0 {
			farmer := farmermodel.Farmer{
				ID: farmerID,
			}

			if shopNameIdx < len(row) {
				farmer.Name = row[shopNameIdx]
			}
			if regionIdx < len(row) {
				farmer.Region = row[regionIdx]
			}
			if farmerDescIdx < len(row) {
				farmer.FarmerDescription = row[farmerDescIdx]
				farmer.Address = row[farmerDescIdx]
			}

			err := s.repo.AddFarmer(ctx, farmer)
			if err != nil {
				return fmt.Errorf("farmerservice.processSheet add farmer: %w", err)
			}

			currentFarmerID = farmerID
		}

		if farmerID > 0 && productNameIdx < len(row) && row[productNameIdx] != "" {
			var productID int
			if productIDIdx < len(row) && row[productIDIdx] != "" {
				productID, _ = strconv.Atoi(row[productIDIdx])
			}

			if productID > 0 {
				product := farmermodel.FarmerProduct{
					ID:          productID,
					FarmerID:    farmerID,
					ProductName: row[productNameIdx],
					Unit:        "шт",
				}

				if categoryIdx < len(row) {
					product.Category = row[categoryIdx]
				}
				if productDescIdx < len(row) {
					product.ProductDescription = row[productDescIdx]
				}
				if priceIdx < len(row) && row[priceIdx] != "" {
					fmt.Sscanf(row[priceIdx], "%f", &product.Price)
				}
				product.Quantity = 0

				err := s.repo.AddFarmerProduct(ctx, product)
				if err != nil {
					return fmt.Errorf("farmerservice.processSheet add product: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *Service) GetFarmerByID(ctx context.Context, id int) (farmermodel.FarmerWithProducts, error) {
	farmer, err := s.repo.GetFarmerByID(ctx, id)
	if err != nil {
		return farmermodel.FarmerWithProducts{}, fmt.Errorf("farmerservice.GetFarmerByID: %w", err)
	}

	products, err := s.repo.GetFarmerProducts(ctx, id)
	if err != nil {
		return farmermodel.FarmerWithProducts{}, fmt.Errorf("farmerservice.GetFarmerByID get products: %w", err)
	}

	return farmermodel.FarmerWithProducts{
		Farmer:   farmer,
		Products: products,
	}, nil
}
