package loyaltyrepository

import (
	"context"
	"fmt"

	loyaltymodel "misis_kolhoz/internal/loyalty/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	CreateClientsTable = `CREATE TABLE IF NOT EXISTS public.clients (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		phone VARCHAR(50),
		created_at TIMESTAMP DEFAULT NOW()
	)`

	CreateBonusTransactionsTable = `CREATE TABLE IF NOT EXISTS public.bonus_transactions (
		id SERIAL PRIMARY KEY,
		client_id INTEGER REFERENCES clients(id),
		type VARCHAR(20) NOT NULL,
		amount INTEGER NOT NULL,
		order_id INTEGER,
		created_at TIMESTAMP DEFAULT NOW()
	)`

	InsertClient = `INSERT INTO public.clients (id, name, email, phone, created_at) 
		VALUES($1, $2, $3, $4, $5) 
		ON CONFLICT (id) DO UPDATE SET 
			name = EXCLUDED.name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone`

	InsertBonusTransaction = `INSERT INTO public.bonus_transactions 
		(client_id, type, amount, order_id, created_at) 
		VALUES($1, $2, $3, $4, NOW())
		RETURNING id`

	SelectClientByID = `SELECT id, name, email, phone, created_at FROM clients WHERE id = $1`

	SelectClientByEmail = `SELECT id, name, email, phone, created_at FROM clients WHERE email = $1`

	SelectTransactionsByClientID = `SELECT id, client_id, type, amount, order_id, created_at 
		FROM bonus_transactions WHERE client_id = $1 ORDER BY created_at DESC`

	SelectClientBalance = `SELECT COALESCE(SUM(CASE WHEN type = 'accrual' THEN amount ELSE -amount END), 0)
		FROM bonus_transactions WHERE client_id = $1`

	SelectAllClients = `SELECT id, name, email, phone, created_at FROM clients ORDER BY id`
)

type Repository struct {
	pgDB *pgxpool.Pool
}

func NewRepository(pgDB *pgxpool.Pool) *Repository {
	return &Repository{pgDB: pgDB}
}

func (r *Repository) InitTables(ctx context.Context) error {
	_, err := r.pgDB.Exec(ctx, CreateClientsTable)
	if err != nil {
		return fmt.Errorf("loyaltyrepository.CreateClientsTable: %w", err)
	}

	_, err = r.pgDB.Exec(ctx, CreateBonusTransactionsTable)
	if err != nil {
		return fmt.Errorf("loyaltyrepository.CreateBonusTransactionsTable: %w", err)
	}

	return nil
}

func (r *Repository) UpsertClient(ctx context.Context, client loyaltymodel.Client) error {
	_, err := r.pgDB.Exec(ctx, InsertClient,
		client.ID,
		client.Name,
		client.Email,
		client.Phone,
		client.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("loyaltyrepository.UpsertClient: %w", err)
	}
	return nil
}

func (r *Repository) GetClientByID(ctx context.Context, id int) (loyaltymodel.Client, error) {
	var client loyaltymodel.Client
	err := r.pgDB.QueryRow(ctx, SelectClientByID, id).Scan(
		&client.ID,
		&client.Name,
		&client.Email,
		&client.Phone,
		&client.CreatedAt,
	)
	if err != nil {
		return client, fmt.Errorf("loyaltyrepository.GetClientByID: %w", err)
	}
	return client, nil
}

func (r *Repository) GetClientByEmail(ctx context.Context, email string) (loyaltymodel.Client, error) {
	var client loyaltymodel.Client
	err := r.pgDB.QueryRow(ctx, SelectClientByEmail, email).Scan(
		&client.ID,
		&client.Name,
		&client.Email,
		&client.Phone,
		&client.CreatedAt,
	)
	if err != nil {
		return client, fmt.Errorf("loyaltyrepository.GetClientByEmail: %w", err)
	}
	return client, nil
}

func (r *Repository) GetClientTransactions(ctx context.Context, clientID int) ([]loyaltymodel.BonusTransaction, error) {
	rows, err := r.pgDB.Query(ctx, SelectTransactionsByClientID, clientID)
	if err != nil {
		return nil, fmt.Errorf("loyaltyrepository.GetClientTransactions: %w", err)
	}
	defer rows.Close()

	var transactions []loyaltymodel.BonusTransaction
	for rows.Next() {
		var t loyaltymodel.BonusTransaction
		err := rows.Scan(
			&t.ID,
			&t.ClientID,
			&t.Type,
			&t.Amount,
			&t.OrderID,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("loyaltyrepository.GetClientTransactions scan: %w", err)
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func (r *Repository) GetClientBalance(ctx context.Context, clientID int) (int, error) {
	var balance int
	err := r.pgDB.QueryRow(ctx, SelectClientBalance, clientID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("loyaltyrepository.GetClientBalance: %w", err)
	}
	return balance, nil
}

func (r *Repository) AccrueBonus(ctx context.Context, clientID int, amount int, orderID int) error {
	_, err := r.pgDB.Exec(ctx, InsertBonusTransaction,
		clientID,
		"accrual",
		amount,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("loyaltyrepository.AccrueBonus: %w", err)
	}
	return nil
}

func (r *Repository) SpendBonus(ctx context.Context, clientID int, amount int) error {
	_, err := r.pgDB.Exec(ctx, InsertBonusTransaction,
		clientID,
		"spend",
		amount,
		nil,
	)
	if err != nil {
		return fmt.Errorf("loyaltyrepository.SpendBonus: %w", err)
	}
	return nil
}

func (r *Repository) GetAllClients(ctx context.Context) ([]loyaltymodel.Client, error) {
	rows, err := r.pgDB.Query(ctx, SelectAllClients)
	if err != nil {
		return nil, fmt.Errorf("loyaltyrepository.GetAllClients: %w", err)
	}
	defer rows.Close()

	var clients []loyaltymodel.Client
	for rows.Next() {
		var c loyaltymodel.Client
		err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.Email,
			&c.Phone,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("loyaltyrepository.GetAllClients scan: %w", err)
		}
		clients = append(clients, c)
	}

	return clients, nil
}
