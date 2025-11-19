package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/namex/goflare/internal/config"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	Postgres   *pgxpool.Pool
	Redis      *redis.Client
	ClickHouse clickhouse.Conn
}

func New(cfg *config.Config) (*Container, error) {
	// Postgres
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	pgPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// ClickHouse
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.ClickHouse.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.ClickHouse.DB,
			Username: "default", // Default user
			Password: "",        // Default password
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		// Don't fail hard on ClickHouse for now as it might take time to start
		fmt.Printf("Warning: failed to ping clickhouse: %v\n", err)
	}

	return &Container{
		Postgres:   pgPool,
		Redis:      rdb,
		ClickHouse: conn,
	}, nil
}

func (c *Container) Close() {
	if c.Postgres != nil {
		c.Postgres.Close()
	}
	if c.Redis != nil {
		c.Redis.Close()
	}
	if c.ClickHouse != nil {
		_ = c.ClickHouse.Close()
	}
}
