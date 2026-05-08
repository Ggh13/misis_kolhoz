package config

import (
	"context"
	"fmt"
	"os"

	"misis_kolhoz/pkg/logger"
	"misis_kolhoz/pkg/postgres"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	RestPort    string         `yaml:"rest_port" env-default:"8080"`
	RestHost    string         `yaml:"rest_host" env-default:"0.0.0.0"`
	PostgresCFG PostgresConfig `yaml:"postgres"`
	MongoCFG    MongoConfig    `yaml:"mongo"`
	S3CFG       S3Config       `yaml:"s3"`
}

type PostgresConfig = postgres.Config

type MongoConfig struct {
	Host       string `yaml:"host" env-default:"mongo"`
	Port       uint16 `yaml:"port" env-default:"27017"`
	Username   string `yaml:"username" env-default:"root"`
	Password   string `yaml:"password" env-default:"password"`
	Database   string `yaml:"database" env-default:"mydb"`
	AuthSource string `yaml:"auth_source" env-default:"admin"`
}

type S3Config struct {
	Endpoint string `yaml:"endpoint" env-default:""`
	Bucket   string `yaml:"bucket" env-default:""`
	Region   string `yaml:"region" env-default:"us-east-1"`
}

func NewConfig(ctx context.Context) (*Config, error) {
	godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	logger.GetLoggerFromCtx(ctx).Info(ctx, fmt.Sprintf("REST PORT: %s", cfg.RestPort))

	return &cfg, nil
}
