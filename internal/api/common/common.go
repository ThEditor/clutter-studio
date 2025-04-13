package common

import (
	"context"

	"github.com/ThEditor/clutter-studio/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	Ctx  context.Context
	Repo *repository.Queries
}

func HashPassword(pass string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
