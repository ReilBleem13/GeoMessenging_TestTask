package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PostgresClient struct {
	db *sqlx.DB
}

func NewPostgresClient(ctx context.Context, dbURL string) (*PostgresClient, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := sqlx.ConnectContext(ctx, "postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect db: %w", err)
	}
	return &PostgresClient{db: db}, nil
}

func (p *PostgresClient) Close() error {
	if p != nil {
		return p.db.Close()
	}
	return nil
}

func (p *PostgresClient) Client() *sqlx.DB {
	return p.db
}
