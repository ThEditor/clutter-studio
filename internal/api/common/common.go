package common

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ThEditor/clutter-studio/internal/config"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/ThEditor/clutter-studio/internal/storage"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	Ctx        context.Context
	Repo       *repository.Queries
	ClickHouse *storage.ClickHouseStorage
}

func HashPassword(pass string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(passHash string, reqPass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passHash), []byte(reqPass)) == nil
}

var Validate = validator.New()

const expirationDuration = 24 * time.Hour

// JWT

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func CreateJWT(userID uuid.UUID, email string) (string, error) {
	cfg := config.Get()
	expirationTime := time.Now().Add(expirationDuration)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.APP_NAME,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(cfg.JWT_SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*Claims, error) {
	cfg := config.Get()
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWT_SECRET), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func AttachJWTCookie(w http.ResponseWriter, jwt string) {
	cfg := config.Get()

	cookie := http.Cookie{
		Name:     "accessToken",
		Value:    jwt,
		Path:     "/",
		MaxAge:   int(expirationDuration.Seconds()),
		HttpOnly: true,
		Secure:   !cfg.DEV_MODE,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
}
