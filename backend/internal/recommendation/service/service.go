package recommendationservice

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
	recommendationmodel "misis_kolhoz/internal/recommendation/model"
	"misis_kolhoz/pkg/logger"

	"go.uber.org/zap"
)

const (
	approxMonthMinDays = 28
	approxMonthMaxDays = 45
	defaultRefreshTick = time.Minute
)

var (
	ErrClientNotFound          = errors.New("client not found")
	ErrRecommendationsNotReady = errors.New("recommendations not ready")
	ErrCatalogNotLoaded        = errors.New("catalog not loaded")
)

type Repository interface {
	GetProductsByIDs(ctx context.Context, ids []int) ([]recommendationmodel.CatalogProduct, error)
	GetProductsByCategory(ctx context.Context, category string) ([]recommendationmodel.CatalogProduct, error)
}

type Service struct {
	repo Repository
	now  func() time.Time

	mu              sync.RWMutex
	cachedResponses map[int]recommendationmodel.Response
	ready           bool
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:            repo,
		now:             time.Now,
		cachedResponses: make(map[int]recommendationmodel.Response),
	}
}

func NewServiceWithNow(repo Repository, now func() time.Time) *Service {
	service := NewService(repo)
	if now != nil {
		service.now = now
	}

	return service
}

func (s *Service) StartPeriodicRefresh(ctx context.Context, filePath string, interval time.Duration) error {
	if interval <= 0 {
		interval = defaultRefreshTick
	}

	initialErr := s.RefreshRecommendations(ctx, filePath)
	if initialErr != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Initial recommendation refresh failed", zap.Error(initialErr))
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.RefreshRecommendations(ctx, filePath); err != nil {
					logger.GetLoggerFromCtx(ctx).Info(ctx, "Recommendation refresh failed", zap.Error(err))
				}
			}
		}
	}()

	return initialErr
}

func (s *Service) RefreshRecommendations(ctx context.Context, filePath string) error {
	responses, err := s.buildRecommendations(ctx, filePath)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.cachedResponses = responses
	s.ready = true
	s.mu.Unlock()

	return nil
}

func (s *Service) GetCachedRecommendations(ctx context.Context, clientID int) (recommendationmodel.Response, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.ready {
		return recommendationmodel.Response{}, ErrRecommendationsNotReady
	}

	response, ok := s.cachedResponses[clientID]
	if !ok {
		return recommendationmodel.Response{}, ErrClientNotFound
	}

	return response, nil
}

func (s *Service) GetRecommendations(ctx context.Context, clientID int, filePath string) (recommendationmodel.Response, error) {
	responses, err := s.buildRecommendations(ctx, filePath)
	if err != nil {
		return recommendationmodel.Response{}, err
	}

	response, ok := responses[clientID]
	if !ok {
		return recommendationmodel.Response{}, ErrClientNotFound
	}

	return response, nil
}

func (s *Service) buildRecommendations(ctx context.Context, filePath string) (map[int]recommendationmodel.Response, error) {
	orders, err := loadOrders(filePath)
	if err != nil {
		return nil, fmt.Errorf("recommendationservice.buildRecommendations load orders: %w", err)
	}

	responses := make(map[int]recommendationmodel.Response)
	if len(orders) == 0 {
		return responses, nil
	}

	ordersByClient := groupOrdersByClient(orders)
	purchasedProductIDs := uniqueProductIDs(orders)
	catalogProducts, err := s.repo.GetProductsByIDs(ctx, purchasedProductIDs)
	if err != nil {
		return nil, fmt.Errorf("recommendationservice.buildRecommendations get purchased products: %w", err)
	}
	if len(purchasedProductIDs) > 0 && len(catalogProducts) == 0 {
		return nil, ErrCatalogNotLoaded
	}

	productByID := make(map[int]recommendationmodel.CatalogProduct, len(catalogProducts))
	for _, product := range catalogProducts {
		productByID[product.ID] = product
	}

	categoryProductsCache := make(map[string][]recommendationmodel.CatalogProduct)
	for clientID, clientOrders := range ordersByClient {
		recommendations, err := s.buildClientRecommendations(ctx, clientOrders, productByID, categoryProductsCache)
		if err != nil {
			return nil, err
		}

		responses[clientID] = recommendationmodel.Response{
			ClientID:        clientID,
			Recommendations: recommendations,
		}
	}

	return responses, nil
}

func (s *Service) buildClientRecommendations(ctx context.Context, clientOrders []recommendationmodel.Order, productByID map[int]recommendationmodel.CatalogProduct, categoryProductsCache map[string][]recommendationmodel.CatalogProduct) ([]recommendationmodel.Recommendation, error) {
	seenProductIDs := make(map[int]struct{}, len(clientOrders))
	latestByCategory := make(map[string]time.Time)
	for _, order := range clientOrders {
		seenProductIDs[order.ProductID] = struct{}{}

		product, ok := productByID[order.ProductID]
		if !ok || product.Category == "" {
			continue
		}

		if order.Date.After(latestByCategory[product.Category]) {
			latestByCategory[product.Category] = order.Date
		}
	}

	categories := qualifyingCategories(latestByCategory, s.now())
	recommendations := make([]recommendationmodel.Recommendation, 0, len(categories))
	for _, category := range categories {
		products, ok := categoryProductsCache[category]
		if !ok {
			var err error
			products, err = s.repo.GetProductsByCategory(ctx, category)
			if err != nil {
				return nil, fmt.Errorf("recommendationservice.buildClientRecommendations get category products: %w", err)
			}
			categoryProductsCache[category] = products
		}

		for _, product := range products {
			if _, seen := seenProductIDs[product.ID]; seen {
				continue
			}

			recommendations = append(recommendations, recommendationmodel.Recommendation{
				ProductID:   product.ID,
				FarmerID:    product.FarmerID,
				ProductName: product.ProductName,
				Category:    product.Category,
				Unit:        product.Unit,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
			break
		}
	}

	return recommendations, nil
}

func loadOrders(filePath string) ([]recommendationmodel.Order, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("recommendationservice.loadOrders open file: %w", err)
	}
	defer file.Close()

	sheetName := firstSheetName(file)
	if sheetName == "" {
		return nil, nil
	}

	rows, err := file.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		return nil, fmt.Errorf("recommendationservice.loadOrders get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, nil
	}

	headerMap := make(map[string]int, len(rows[0]))
	for i, header := range rows[0] {
		headerMap[header] = i
	}

	orderIDIdx, ok := headerMap["Order_ID"]
	if !ok {
		return nil, fmt.Errorf("recommendationservice.loadOrders: Order_ID column not found")
	}
	dateIdx, ok := headerMap["Дата"]
	if !ok {
		return nil, fmt.Errorf("recommendationservice.loadOrders: Дата column not found")
	}
	userIDIdx, ok := headerMap["User_ID"]
	if !ok {
		return nil, fmt.Errorf("recommendationservice.loadOrders: User_ID column not found")
	}
	productIDIdx, ok := headerMap["Product_ID"]
	if !ok {
		return nil, fmt.Errorf("recommendationservice.loadOrders: Product_ID column not found")
	}

	orders := make([]recommendationmodel.Order, 0, len(rows)-1)
	for _, row := range rows[1:] {
		orderID, ok := cellValue(row, orderIDIdx)
		if !ok || orderID == "" {
			continue
		}

		dateValue, ok := cellValue(row, dateIdx)
		if !ok || dateValue == "" {
			continue
		}

		userIDValue, ok := cellValue(row, userIDIdx)
		if !ok || userIDValue == "" {
			continue
		}

		productIDValue, ok := cellValue(row, productIDIdx)
		if !ok || productIDValue == "" {
			continue
		}

		userID, err := parseIntCell(userIDValue)
		if err != nil {
			continue
		}

		productID, err := parseIntCell(productIDValue)
		if err != nil {
			continue
		}

		orderedAt, err := parseDateCell(dateValue)
		if err != nil {
			continue
		}

		orders = append(orders, recommendationmodel.Order{
			OrderID:   orderID,
			Date:      normalizeDate(orderedAt),
			UserID:    userID,
			ProductID: productID,
		})
	}

	return orders, nil
}

func firstSheetName(file *excelize.File) string {
	if file == nil {
		return ""
	}

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return ""
	}

	return sheets[0]
}

func groupOrdersByClient(orders []recommendationmodel.Order) map[int][]recommendationmodel.Order {
	grouped := make(map[int][]recommendationmodel.Order)
	for _, order := range orders {
		grouped[order.UserID] = append(grouped[order.UserID], order)
	}

	return grouped
}

func uniqueProductIDs(orders []recommendationmodel.Order) []int {
	seen := make(map[int]struct{}, len(orders))
	productIDs := make([]int, 0, len(orders))
	for _, order := range orders {
		if _, ok := seen[order.ProductID]; ok {
			continue
		}
		seen[order.ProductID] = struct{}{}
		productIDs = append(productIDs, order.ProductID)
	}

	sort.Ints(productIDs)
	return productIDs
}

func qualifyingCategories(latestByCategory map[string]time.Time, now time.Time) []string {
	categories := make([]string, 0, len(latestByCategory))
	for category, latestDate := range latestByCategory {
		if isApproximatelyMonthAgo(latestDate, now) {
			categories = append(categories, category)
		}
	}

	sort.Strings(categories)
	return categories
}

func isApproximatelyMonthAgo(purchaseDate, now time.Time) bool {
	days := int(normalizeDate(now).Sub(normalizeDate(purchaseDate)).Hours() / 24)
	return days >= approxMonthMinDays && days <= approxMonthMaxDays
}

func normalizeDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func cellValue(row []string, idx int) (string, bool) {
	if idx >= len(row) {
		return "", false
	}

	return row[idx], true
}

func parseIntCell(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err == nil {
		return parsed, nil
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return int(math.Round(floatValue)), nil
}

func parseDateCell(value string) (time.Time, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err == nil {
		wholeDays, fraction := math.Modf(floatValue)
		baseDate := time.Date(1899, time.December, 30, 0, 0, 0, 0, time.UTC)
		date := baseDate.AddDate(0, 0, int(wholeDays))
		if fraction != 0 {
			date = date.Add(time.Duration(fraction * float64(24*time.Hour)))
		}
		return date, nil
	}

	for _, layout := range []string{"2006-01-02", "02.01.2006", time.RFC3339} {
		parsed, parseErr := time.Parse(layout, value)
		if parseErr == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported date value %q", value)
}
