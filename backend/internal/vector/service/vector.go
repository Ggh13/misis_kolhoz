package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"misis_kolhoz/internal/vector/model"
	vectorepository "misis_kolhoz/internal/vector/repository"
)

type VectorService struct {
	repo *vectorepository.VectorRepository
}

func NewVectorService(repo *vectorepository.VectorRepository) *VectorService {
	return &VectorService{repo: repo}
}

func (s *VectorService) Init(ctx context.Context) error {
	return s.repo.Init(ctx)
}

func (s *VectorService) BulkInsertFromProducts(ctx context.Context) error {
	return s.repo.BulkInsertFromProducts(ctx)
}

func (s *VectorService) Upsert(ctx context.Context, productID int, embedding []float32) error {
	return s.repo.Upsert(ctx, productID, embedding)
}

func (s *VectorService) Get(ctx context.Context, productID int) (*model.ProductEmbedding, error) {
	return s.repo.Get(ctx, productID)
}

func (s *VectorService) Update(ctx context.Context, productID int, embedding []float32) error {
	return s.repo.Update(ctx, productID, embedding)
}

func (s *VectorService) Delete(ctx context.Context, productID int) error {
	return s.repo.Delete(ctx, productID)
}

func (s *VectorService) Search(ctx context.Context, embedding []float32, limit int, farmerID int) ([]model.ProductEmbedding, error) {
	if len(embedding) == 0 {
		return nil, errors.New("embedding is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	return s.repo.Search(ctx, embedding, limit, farmerID)
}

func (s *VectorService) SearchEventsByProduct(ctx context.Context, productID int, limit int) ([]model.EventEmbedding, error) {
	product, err := s.repo.Get(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get product embedding: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	return s.repo.SearchEvents(ctx, product.Embedding, limit)
}

func (s *VectorService) SearchEvents(ctx context.Context, embedding []float32, limit int) ([]model.EventEmbedding, error) {
	if len(embedding) == 0 {
		return nil, errors.New("embedding is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	return s.repo.SearchEvents(ctx, embedding, limit)
}

func (s *VectorService) MatchEventsToProducts(ctx context.Context, limit int, farmerID int, futureOnly bool) ([]model.EventProductsMatch, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	return s.repo.MatchEventsToProducts(ctx, limit, farmerID, futureOnly)
}

func (s *VectorService) MatchProductsToEvents(ctx context.Context, limit int, farmerID int, futureOnly bool) ([]model.ProductEventsMatch, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	return s.repo.MatchProductsToEvents(ctx, limit, farmerID, futureOnly)
}

func (s *VectorService) CalculateDistance(vecA, vecB []float32) (cosine, euclidean float64) {
	if len(vecA) != len(vecB) || len(vecA) == 0 {
		return 0, 0
	}

	var dotProduct, normA, normB float64
	for i := range vecA {
		dotProduct += float64(vecA[i]) * float64(vecB[i])
		normA += float64(vecA[i]) * float64(vecA[i])
		normB += float64(vecB[i]) * float64(vecB[i])
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA > 0 && normB > 0 {
		cosine = dotProduct / (normA * normB)
	}

	var sumSquares float64
	for i := range vecA {
		diff := float64(vecA[i]) - float64(vecB[i])
		sumSquares += diff * diff
	}
	euclidean = math.Sqrt(sumSquares)

	return cosine, euclidean
}
