package main

import (
	"context"

	"github.com/ThEditor/clutter-studio/internal/api"
	"github.com/ThEditor/clutter-studio/internal/config"
	"github.com/ThEditor/clutter-studio/internal/mailer"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/ThEditor/clutter-studio/internal/storage"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pgstore, err := storage.NewPostgresStorage(ctx, cfg.DATABASE_URL)
	if err != nil {
		panic(err)
	}
	defer pgstore.Close()

	repo := repository.New(pgstore.Db)

	chstore, err := storage.NewClickHouseStorage(cfg.CLICKHOUSE_URL)
	if err != nil {
		panic(err)
	}
	defer chstore.Close()

	mailer, err := mailer.NewMailer(mailer.MailerConfig{
		Host:     cfg.SMTP_HOST,
		Port:     cfg.SMTP_PORT,
		From:     cfg.SMTP_FROM,
		Username: cfg.SMTP_USERNAME,
		Password: cfg.SMTP_PASSWORD,
	})
	if err != nil {
		panic(err)
	}
	defer mailer.Close()

	api.Start(ctx, cfg.BIND_ADDRESS, cfg.PORT, repo, chstore, mailer)
}
