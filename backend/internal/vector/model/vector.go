package model

type ProductEmbedding struct {
	ID                 int       `json:"id"`
	ProductID          int       `json:"product_id"`
	FarmerID           int       `json:"farmer_id"`
	ProductName        string    `json:"product_name"`
	Category           string    `json:"category"`
	Unit               string    `json:"unit"`
	Price              float64   `json:"price"`
	Quantity           int       `json:"quantity"`
	FarmerDescription  string    `json:"farmer_description"`
	ProductDescription string    `json:"product_description"`
	Embedding          []float32 `json:"-"`
}

type ProductMatch struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	FarmerID  int     `json:"farmer_id"`
	ProductName string `json:"product_name"`
	Category  string  `json:"category"`
	Unit      string  `json:"unit"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Distance  float64 `json:"distance"`
}

type UpsertVectorRequest struct {
	ProductID int       `json:"product_id" binding:"required"`
	Embedding []float32 `json:"embedding" binding:"required"`
}

type SearchRequest struct {
	Vector   []float32 `json:"vector" binding:"required"`
	Limit    int       `json:"limit" binding:"required,min=1"`
	FarmerID int       `json:"farmer_id,omitempty"`
}

type SearchResponse struct {
	Products []ProductEmbedding `json:"products"`
}

type EventEmbedding struct {
	ID         int     `json:"id"`
	EventDate  string  `json:"event_date"`
	HolidayInfo string `json:"holiday_info"`
	Category   string  `json:"category"`
	About      string  `json:"about"`
	FoodCustoms string `json:"food_customs,omitempty"`
	Distance   float64 `json:"distance"`
}

type EventsToProductsRequest struct {
	Limit      int  `json:"limit" binding:"required,min=1"`
	FarmerID   int  `json:"farmer_id,omitempty"`
	FutureOnly bool `json:"future_only"`
}

type ProductsToEventsRequest struct {
	Limit      int  `json:"limit" binding:"required,min=1"`
	FarmerID   int  `json:"farmer_id,omitempty"`
	FutureOnly bool `json:"future_only"`
}

type EventProductsMatch struct {
	Event    EventEmbedding `json:"event"`
	Products []ProductMatch `json:"products"`
}

type ProductEventsMatch struct {
	Product ProductEmbedding `json:"product"`
	Events  []EventEmbedding `json:"events"`
}

type EventsToProductsResponse struct {
	Events []EventProductsMatch `json:"events"`
}

type ProductsToEventsResponse struct {
	Products []ProductEventsMatch `json:"products"`
}

type EventSearchRequest struct {
	Vector []float32 `json:"vector" binding:"required"`
	Limit  int       `json:"limit" binding:"required,min=1"`
}

type EventForProductRequest struct {
	ProductID int `json:"product_id" binding:"required"`
	Limit     int `json:"limit" binding:"required,min=1"`
}

type EventSearchResponse struct {
	Events []EventEmbedding `json:"events"`
}

type DistanceRequest struct {
	VectorA []float32 `json:"vector_a" binding:"required"`
	VectorB []float32 `json:"vector_b" binding:"required"`
}

type DistanceResponse struct {
	Cosine    float64 `json:"cosine"`
	Euclidean float64 `json:"euclidean"`
}
