package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Host       string `yaml:"host"`
	Port       uint16 `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Database   string `yaml:"database"`
	AuthSource string `yaml:"auth_source"`
}

func NewMongo(cfg Config) (*mongo.Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.AuthSource,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed connect mongo.NewMongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed ping mongo.NewMongo: %w", err)
	}

	mongoDB := client.Database(cfg.Database)

	return mongoDB, nil
}

func Disconnect(ctx context.Context, client *mongo.Client) error {
	return client.Disconnect(ctx)
}
