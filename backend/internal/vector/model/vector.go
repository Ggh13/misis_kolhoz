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

type UpsertVectorRequest struct {
	ProductID int       `json:"product_id" binding:"required"`
	Embedding []float32 `json:"embedding" binding:"required"`
}

type SearchRequest struct {
	Vector []float32 `json:"vector" binding:"required"`
	Limit  int       `json:"limit" binding:"required,min=1"`
}

type SearchResponse struct {
	Products []ProductEmbedding `json:"products"`
}

type DistanceRequest struct {
	VectorA []float32 `json:"vector_a" binding:"required"`
	VectorB []float32 `json:"vector_b" binding:"required"`
}

type DistanceResponse struct {
	Cosine    float64 `json:"cosine"`
	Euclidean float64 `json:"euclidean"`
}
