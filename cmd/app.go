package main

import (
	"github.com/ThEditor/clutter-studio/internal/api"
	"github.com/ThEditor/clutter-studio/internal/config"
)

func main() {
	cfg := config.Load()

	api.Start(cfg.BIND_ADDRESS, cfg.PORT)
}
