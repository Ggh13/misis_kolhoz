package qdrant

import (
	"context"

	"github.com/qdrant/go-client/qdrant"
)

type Config struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	CollectionName string `yaml:"collection_name"`
	VectorSize     uint64 `yaml:"vector_size"`
}

func NewQdrant(ctx context.Context, cfg *Config) (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:                     cfg.Host,
		Port:                     cfg.Port,
		SkipCompatibilityCheck:   true,
	})
	if err != nil {
		return nil, err
	}

	exists, err := client.CollectionExists(ctx, cfg.CollectionName)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: cfg.CollectionName,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     cfg.VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func UpsertVector(ctx context.Context, client *qdrant.Client, collectionName string, id uint64, vector []float32) error {
	_, err := client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points: []*qdrant.PointStruct{
			{
				Id:      qdrant.NewIDNum(id),
				Vectors: qdrant.NewVectors(vector...),
			},
		},
	})
	return err
}

func Search(ctx context.Context, client *qdrant.Client, collectionName string, vector []float32, limit uint64) ([]uint64, error) {
	res, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
	})
	if err != nil {
		return nil, err
	}

	ids := make([]uint64, 0, len(res))
	for _, p := range res {
		id := p.GetId()
		if id.GetNum() != 0 {
			ids = append(ids, id.GetNum())
		}
	}
	return ids, nil
}