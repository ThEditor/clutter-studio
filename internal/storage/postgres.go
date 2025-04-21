package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	Db  *pgxpool.Pool
	ctx context.Context
}

func NewPostgresStorage(ctx context.Context, dsn string) (*PostgresStorage, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to create postgres pool: %w", err)
	}

	storage := &PostgresStorage{Db: dbpool, ctx: ctx}

	return storage, nil
}

func (s *PostgresStorage) Close() error {
	s.Db.Close()
	return nil
}
