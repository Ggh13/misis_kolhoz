package model

type Vector struct {
	ID     uint64    `json:"id"`
	Vector []float32 `json:"vector"`
}

type CreateVectorRequest struct {
	Vector []float32 `json:"vector" binding:"required"`
}

type SearchRequest struct {
	Vector []float32 `json:"vector" binding:"required"`
	Limit  uint64    `json:"limit" binding:"required,min=1"`
}

type DistanceRequest struct {
	VectorA []float32 `json:"vector_a" binding:"required"`
	VectorB []float32 `json:"vector_b" binding:"required"`
}

type SearchResponse struct {
	IDs []uint64 `json:"ids"`
}

type DistanceResponse struct {
	Cosine   float64 `json:"cosine"`
	Euclidean float64 `json:"euclidean"`
}