package loyaltymodel

import "time"

type Client struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type BonusTransaction struct {
	ID        int       `json:"id"`
	ClientID  int       `json:"client_id"`
	Type      string    `json:"type"`
	Amount    int       `json:"amount"`
	OrderID   int       `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ClientWithBonus struct {
	Client       Client             `json:"client"`
	Balance      int                `json:"balance"`
	Transactions []BonusTransaction `json:"transactions"`
}

type Order struct {
	ID          int     `json:"id"`
	ClientName  string  `json:"client_name"`
	ClientEmail string  `json:"client_email"`
	TotalPrice  float64 `json:"total_price"`
	Status      string  `json:"status"`
}
