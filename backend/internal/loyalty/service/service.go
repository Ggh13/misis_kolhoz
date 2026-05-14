package loyaltyservice

import (
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
	loyaltymodel "misis_kolhoz/internal/loyalty/model"
)

const (
	BonusRate      = 0.05
	BonusThreshold = 1000.0
)

type Repository interface {
	InitTables(ctx context.Context) error
	UpsertClient(ctx context.Context, client loyaltymodel.Client) error
	GetClientByID(ctx context.Context, id int) (loyaltymodel.Client, error)
	GetClientBalance(ctx context.Context, clientID int) (int, error)
	GetClientTransactions(ctx context.Context, clientID int) ([]loyaltymodel.BonusTransaction, error)
	AccrueBonus(ctx context.Context, clientID int, amount int, orderID int) error
	SpendBonus(ctx context.Context, clientID int, amount int) error
	GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LoadOrdersFromExcel(ctx context.Context, filePath string) error {
	err := s.repo.InitTables(ctx)
	if err != nil {
		return fmt.Errorf("loyaltyservice.LoadOrdersFromExcel init tables: %w", err)
	}

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("loyaltyservice.LoadOrdersFromExcel open file: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetMap()
	sheetName := ""
	for _, name := range sheets {
		if len(name) > 0 {
			sheetName = name
			break
		}
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("loyaltyservice.LoadOrdersFromExcel get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil
	}

	headers := rows[0]
	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[h] = i
	}

	orderIDIdx := 0
	userIDIdx := 2
	priceIdx := 6

	processedOrders := make(map[string]bool)

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]

		if len(row) <= userIDIdx || len(row) <= priceIdx {
			continue
		}

		var clientID int
		var orderID string

		if userIDIdx < len(row) {
			fmt.Sscanf(row[userIDIdx], "%d", &clientID)
		}

		if orderIDIdx < len(row) {
			orderID = row[orderIDIdx]
		}

		if clientID <= 0 || orderID == "" {
			continue
		}

		if processedOrders[orderID] {
			continue
		}
		processedOrders[orderID] = true

		client := loyaltymodel.Client{
			ID:        clientID,
			CreatedAt: time.Now(),
		}

		err = s.repo.UpsertClient(ctx, client)
		if err != nil {
			return fmt.Errorf("loyaltyservice.LoadOrdersFromExcel upsert client: %w", err)
		}

		var orderTotal float64
		for r := 1; r < len(rows); r++ {
			checkRow := rows[r]
			if len(checkRow) <= orderIDIdx || len(checkRow) <= priceIdx {
				continue
			}
			checkOrderID := checkRow[orderIDIdx]
			if checkOrderID == orderID {
				var p float64
				fmt.Sscanf(checkRow[priceIdx], "%f", &p)
				orderTotal += p
			}
		}

		if orderTotal >= BonusThreshold {
			bonusAmount := int(orderTotal * BonusRate)
			if bonusAmount > 0 {
				err := s.repo.AccrueBonus(ctx, clientID, bonusAmount, 0)
				if err != nil {
					return fmt.Errorf("loyaltyservice.LoadOrdersFromExcel accrue bonus: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *Service) GetClientBonusInfo(ctx context.Context, clientID int) (loyaltymodel.ClientWithBonus, error) {
	client, err := s.repo.GetClientByID(ctx, clientID)
	if err != nil {
		return loyaltymodel.ClientWithBonus{}, fmt.Errorf("loyaltyservice.GetClientBonusInfo get client: %w", err)
	}

	balance, err := s.repo.GetClientBalance(ctx, client.ID)
	if err != nil {
		return loyaltymodel.ClientWithBonus{}, fmt.Errorf("loyaltyservice.GetClientBonusInfo get balance: %w", err)
	}

	transactions, err := s.repo.GetClientTransactions(ctx, client.ID)
	if err != nil {
		return loyaltymodel.ClientWithBonus{}, fmt.Errorf("loyaltyservice.GetClientBonusInfo get transactions: %w", err)
	}

	return loyaltymodel.ClientWithBonus{
		Client:       client,
		Balance:      balance,
		Transactions: transactions,
	}, nil
}

func (s *Service) SpendClientBonus(ctx context.Context, clientID int, amount int) error {
	balance, err := s.repo.GetClientBalance(ctx, clientID)
	if err != nil {
		return fmt.Errorf("loyaltyservice.SpendClientBonus get balance: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("loyaltyservice.SpendClientBonus: insufficient balance")
	}

	err = s.repo.SpendBonus(ctx, clientID, amount)
	if err != nil {
		return fmt.Errorf("loyaltyservice.SpendClientBonus: %w", err)
	}

	return nil
}

func (s *Service) GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error) {
	return s.repo.GetAllClients(ctx)
}
