package farmermodel

type Farmer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	Address  string `json:"address"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type FarmerProduct struct {
	ID          int     `json:"id"`
	FarmerID    int     `json:"farmer_id"`
	ProductName string  `json:"product_name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

type FarmerWithProducts struct {
	Farmer   Farmer          `json:"farmer"`
	Products []FarmerProduct `json:"products"`
}