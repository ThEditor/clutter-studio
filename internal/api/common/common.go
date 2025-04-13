package common

import (
	"context"

	"github.com/ThEditor/clutter-studio/internal/repository"
)

type Server struct {
	Ctx  context.Context
	Repo *repository.Queries
}
