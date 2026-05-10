package service

import (
	"context"
	"math"

	"misis_kolhoz/internal/vector/repository"
)

type VectorService struct {
	repo *repository.VectorRepository
}

func NewVectorService(repo *repository.VectorRepository) *VectorService {
	return &VectorService{repo: repo}
}

func (s *VectorService) Create(ctx context.Context, id uint64, vector []float32) error {
	return s.repo.Create(ctx, id, vector)
}

func (s *VectorService) Get(ctx context.Context, id uint64) ([]float32, error) {
	return s.repo.Get(ctx, id)
}

func (s *VectorService) Update(ctx context.Context, id uint64, vector []float32) error {
	return s.repo.Update(ctx, id, vector)
}

func (s *VectorService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *VectorService) Search(ctx context.Context, vector []float32, limit uint64) ([]uint64, error) {
	return s.repo.Search(ctx, vector, limit)
}

func (s *VectorService) CalculateDistance(vecA, vecB []float32) (cosine, euclidean float64) {
	if len(vecA) != len(vecB) {
		return 0, 0
	}

	var dotProduct float64
	var normA, normB float64

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