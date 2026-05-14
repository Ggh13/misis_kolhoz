package recommendationmodel

import "time"

type Order struct {
	OrderID   string
	Date      time.Time
	UserID    int
	ProductID int
}

type CatalogProduct struct {
	ID          int     `json:"id"`
	FarmerID    int     `json:"farmer_id"`
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type Recommendation struct {
	ProductID   int     `json:"product_id"`
	FarmerID    int     `json:"farmer_id"`
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type Response struct {
	ClientID        int              `json:"client_id"`
	Recommendations []Recommendation `json:"recommendations"`
}
