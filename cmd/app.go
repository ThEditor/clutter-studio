package main

import (
	"context"

	"github.com/ThEditor/clutter-studio/internal/api"
	"github.com/ThEditor/clutter-studio/internal/config"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/ThEditor/clutter-studio/internal/storage"
	"github.com/jackc/pgx/v5"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	conn, err := pgx.Connect(context.Background(), cfg.DATABASE_URL)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	repo := repository.New(conn)

	store, err := storage.NewClickHouseStorage(cfg.CLICKHOUSE_URL)
	if err != nil {
		panic(err)
	}
	defer store.Close()

	api.Start(ctx, cfg.BIND_ADDRESS, cfg.PORT, repo, store)
}
