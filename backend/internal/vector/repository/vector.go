package repository

import (
	"context"

	qdrantlib "github.com/qdrant/go-client/qdrant"
	"misis_kolhoz/pkg/qdrant"
)

type VectorRepository struct {
	client         *qdrantlib.Client
	collectionName string
}

func NewVectorRepository(client *qdrantlib.Client, collectionName string) *VectorRepository {
	return &VectorRepository{
		client:         client,
		collectionName: collectionName,
	}
}

func (r *VectorRepository) Create(ctx context.Context, id uint64, vector []float32) error {
	return qdrant.UpsertVector(ctx, r.client, r.collectionName, id, vector)
}

func (r *VectorRepository) Get(ctx context.Context, id uint64) ([]float32, error) {
	points, err := r.client.Get(ctx, &qdrantlib.GetPoints{
		CollectionName: r.collectionName,
		Ids:            []*qdrantlib.PointId{qdrantlib.NewIDNum(id)},
	})
	if err != nil {
		return nil, err
	}
	if len(points) == 0 {
		return nil, nil
	}
	return points[0].GetVectors().GetVector().GetData(), nil
}

func (r *VectorRepository) Update(ctx context.Context, id uint64, vector []float32) error {
	return qdrant.UpsertVector(ctx, r.client, r.collectionName, id, vector)
}

func (r *VectorRepository) Delete(ctx context.Context, id uint64) error {
	_, err := r.client.Delete(ctx, &qdrantlib.DeletePoints{
		CollectionName: r.collectionName,
		Points:         qdrantlib.NewPointsSelectorIDs([]*qdrantlib.PointId{qdrantlib.NewIDNum(id)}),
	})
	return err
}

func (r *VectorRepository) Search(ctx context.Context, vector []float32, limit uint64) ([]uint64, error) {
	return qdrant.Search(ctx, r.client, r.collectionName, vector, limit)
}