package recommendationservice

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
	recommendationmodel "misis_kolhoz/internal/recommendation/model"
)

type mockRepository struct {
	productsByID       map[int]recommendationmodel.CatalogProduct
	productsByCategory map[string][]recommendationmodel.CatalogProduct
}

func (m *mockRepository) GetProductsByIDs(ctx context.Context, ids []int) ([]recommendationmodel.CatalogProduct, error) {
	products := make([]recommendationmodel.CatalogProduct, 0, len(ids))
	for _, id := range ids {
		if product, ok := m.productsByID[id]; ok {
			products = append(products, product)
		}
	}
	return products, nil
}

func (m *mockRepository) GetProductsByCategory(ctx context.Context, category string) ([]recommendationmodel.CatalogProduct, error) {
	return m.productsByCategory[category], nil
}

func TestGetRecommendationsReturnsDeterministicCategoryMatches(t *testing.T) {
	now := time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC)
	filePath := writeOrdersWorkbook(t, []orderRow{
		{orderID: "ORD-1", date: now.AddDate(0, 0, -30), userID: 7, productID: 101, productName: "Forest Tea"},
		{orderID: "ORD-2", date: now.AddDate(0, 0, -29), userID: 7, productID: 201, productName: "Berry Jam"},
		{orderID: "ORD-3", date: now.AddDate(0, 0, -10), userID: 7, productID: 301, productName: "Milk"},
		{orderID: "ORD-4", date: now.AddDate(0, 0, -30), userID: 99, productID: 102, productName: "Other User Tea"},
	})

	repo := &mockRepository{
		productsByID: map[int]recommendationmodel.CatalogProduct{
			101: {ID: 101, FarmerID: 1, ProductName: "Forest Tea", Category: "Drinks", Unit: "pcs", Price: 100},
			201: {ID: 201, FarmerID: 2, ProductName: "Berry Jam", Category: "Preserves", Unit: "pcs", Price: 150},
			301: {ID: 301, FarmerID: 3, ProductName: "Milk", Category: "Dairy", Unit: "l", Price: 120},
		},
		productsByCategory: map[string][]recommendationmodel.CatalogProduct{
			"Dairy": {
				{ID: 301, FarmerID: 3, ProductName: "Milk", Category: "Dairy", Unit: "l", Price: 120},
				{ID: 302, FarmerID: 3, ProductName: "Cheese", Category: "Dairy", Unit: "pcs", Price: 240},
			},
			"Drinks": {
				{ID: 101, FarmerID: 1, ProductName: "Forest Tea", Category: "Drinks", Unit: "pcs", Price: 100},
				{ID: 102, FarmerID: 1, ProductName: "Mint Tea", Category: "Drinks", Unit: "pcs", Price: 110},
			},
			"Preserves": {
				{ID: 201, FarmerID: 2, ProductName: "Berry Jam", Category: "Preserves", Unit: "pcs", Price: 150},
				{ID: 202, FarmerID: 2, ProductName: "Pine Jam", Category: "Preserves", Unit: "pcs", Price: 180},
			},
		},
	}

	service := NewServiceWithNow(repo, func() time.Time { return now })

	response, err := service.GetRecommendations(context.Background(), 7, filePath)
	if err != nil {
		t.Fatalf("GetRecommendations returned error: %v", err)
	}

	if response.ClientID != 7 {
		t.Fatalf("expected client id 7, got %d", response.ClientID)
	}

	if len(response.Recommendations) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(response.Recommendations))
	}

	if response.Recommendations[0].Category != "Drinks" || response.Recommendations[0].ProductID != 102 {
		t.Fatalf("unexpected first recommendation: %+v", response.Recommendations[0])
	}

	if response.Recommendations[1].Category != "Preserves" || response.Recommendations[1].ProductID != 202 {
		t.Fatalf("unexpected second recommendation: %+v", response.Recommendations[1])
	}
}

func TestGetRecommendationsSkipsSeenOutOfWindowAndExhaustedCategories(t *testing.T) {
	now := time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC)
	filePath := writeOrdersWorkbook(t, []orderRow{
		{orderID: "ORD-1", date: now.AddDate(0, 0, -30), userID: 5, productID: 401, productName: "Old Honey"},
		{orderID: "ORD-2", date: now.AddDate(0, 0, -20), userID: 5, productID: 501, productName: "Fresh Tea"},
	})

	repo := &mockRepository{
		productsByID: map[int]recommendationmodel.CatalogProduct{
			401: {ID: 401, FarmerID: 4, ProductName: "Old Honey", Category: "Honey", Unit: "pcs", Price: 210},
			501: {ID: 501, FarmerID: 5, ProductName: "Fresh Tea", Category: "Tea", Unit: "pcs", Price: 130},
		},
		productsByCategory: map[string][]recommendationmodel.CatalogProduct{
			"Honey": {
				{ID: 401, FarmerID: 4, ProductName: "Old Honey", Category: "Honey", Unit: "pcs", Price: 210},
			},
			"Tea": {
				{ID: 501, FarmerID: 5, ProductName: "Fresh Tea", Category: "Tea", Unit: "pcs", Price: 130},
				{ID: 502, FarmerID: 5, ProductName: "Herbal Tea", Category: "Tea", Unit: "pcs", Price: 135},
			},
		},
	}

	service := NewServiceWithNow(repo, func() time.Time { return now })

	response, err := service.GetRecommendations(context.Background(), 5, filePath)
	if err != nil {
		t.Fatalf("GetRecommendations returned error: %v", err)
	}

	if len(response.Recommendations) != 0 {
		t.Fatalf("expected no recommendations, got %+v", response.Recommendations)
	}
}

func TestRefreshRecommendationsCachesResponses(t *testing.T) {
	now := time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC)
	filePath := writeOrdersWorkbook(t, []orderRow{
		{orderID: "ORD-1", date: now.AddDate(0, 0, -30), userID: 3, productID: 601, productName: "Parmesan"},
	})

	repo := &mockRepository{
		productsByID: map[int]recommendationmodel.CatalogProduct{
			601: {ID: 601, FarmerID: 6, ProductName: "Parmesan", Category: "Cheese", Unit: "pcs", Price: 500},
		},
		productsByCategory: map[string][]recommendationmodel.CatalogProduct{
			"Cheese": {
				{ID: 601, FarmerID: 6, ProductName: "Parmesan", Category: "Cheese", Unit: "pcs", Price: 500},
				{ID: 602, FarmerID: 6, ProductName: "Montasio", Category: "Cheese", Unit: "pcs", Price: 450},
			},
		},
	}

	service := NewServiceWithNow(repo, func() time.Time { return now })
	if err := service.RefreshRecommendations(context.Background(), filePath); err != nil {
		t.Fatalf("RefreshRecommendations returned error: %v", err)
	}

	response, err := service.GetCachedRecommendations(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetCachedRecommendations returned error: %v", err)
	}

	if len(response.Recommendations) != 1 || response.Recommendations[0].ProductID != 602 {
		t.Fatalf("unexpected cached recommendations: %+v", response.Recommendations)
	}
}

func TestGetCachedRecommendationsReturnsNotReady(t *testing.T) {
	service := NewService(&mockRepository{})

	_, err := service.GetCachedRecommendations(context.Background(), 1)
	if !errors.Is(err, ErrRecommendationsNotReady) {
		t.Fatalf("expected ErrRecommendationsNotReady, got %v", err)
	}
}

func TestGetRecommendationsReturnsClientNotFound(t *testing.T) {
	now := time.Date(2026, time.May, 12, 0, 0, 0, 0, time.UTC)
	filePath := writeOrdersWorkbook(t, []orderRow{
		{orderID: "ORD-1", date: now.AddDate(0, 0, -30), userID: 8, productID: 701, productName: "Tea"},
	})

	repo := &mockRepository{
		productsByID: map[int]recommendationmodel.CatalogProduct{
			701: {ID: 701, FarmerID: 7, ProductName: "Tea", Category: "Tea", Unit: "pcs", Price: 100},
		},
		productsByCategory: map[string][]recommendationmodel.CatalogProduct{
			"Tea": {
				{ID: 701, FarmerID: 7, ProductName: "Tea", Category: "Tea", Unit: "pcs", Price: 100},
				{ID: 702, FarmerID: 7, ProductName: "Mint Tea", Category: "Tea", Unit: "pcs", Price: 120},
			},
		},
	}

	service := NewServiceWithNow(repo, func() time.Time { return now })
	_, err := service.GetRecommendations(context.Background(), 999, filePath)
	if !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound, got %v", err)
	}
}

type orderRow struct {
	orderID     string
	date        time.Time
	userID      int
	productID   int
	productName string
}

func writeOrdersWorkbook(t *testing.T, rows []orderRow) string {
	t.Helper()

	file := excelize.NewFile()
	defer file.Close()

	sheetName := file.GetSheetName(file.GetActiveSheetIndex())
	data := [][]any{{"Order_ID", "Дата", "User_ID", "Product_ID", "Товар", "Кол-во", "Сумма"}}
	for _, row := range rows {
		data = append(data, []any{
			row.orderID,
			excelDate(row.date),
			row.userID,
			row.productID,
			row.productName,
			1,
			100,
		})
	}

	for rowIdx, row := range data {
		cell, err := excelize.CoordinatesToCellName(1, rowIdx+1)
		if err != nil {
			t.Fatalf("CoordinatesToCellName returned error: %v", err)
		}
		if err := file.SetSheetRow(sheetName, cell, &row); err != nil {
			t.Fatalf("SetSheetRow returned error: %v", err)
		}
	}

	filePath := filepath.Join(t.TempDir(), "orders.xlsx")
	if err := file.SaveAs(filePath); err != nil {
		t.Fatalf("SaveAs returned error: %v", err)
	}

	return filePath
}

func excelDate(value time.Time) float64 {
	base := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
	return value.Sub(base).Hours() / 24
}
