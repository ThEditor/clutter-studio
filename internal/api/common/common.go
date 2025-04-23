package common

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/ThEditor/clutter-studio/internal/config"
	"github.com/ThEditor/clutter-studio/internal/mailer"
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
	Mailer     *mailer.Mailer
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

// Email verification
func GenerateRandomCode(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, n)
	rand.Read(b)

	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}

	return string(b)
}

func SendVerificationMail(mailer mailer.Mailer, to string, code string) error {
	return mailer.Send([]string{to}, "Clutter Verification Code", "Your verification code for Clutter Analytics is: "+code)
}

// Validator
func IsYYYYMMDDDate(fl validator.FieldLevel) bool {
	YYYYMMDDDateRegexString := "^(\\d{4})-(0[1-9]|1[0-2])-(0[1-9]|[12]\\d|3[01])$"
	YYYYMMDDDateRegex := regexp.MustCompile(YYYYMMDDDateRegexString)
	return YYYYMMDDDateRegex.MatchString(fl.Field().String())
}

var Validate = validator.New()
var _ = Validate.RegisterValidation("YYYYMMDDdate", IsYYYYMMDDDate)

const expirationDuration = 24 * time.Hour

// JWT

type Claims struct {
	UserID        uuid.UUID `json:"user_id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	jwt.RegisteredClaims
}

func CreateJWT(userID uuid.UUID, email string, verified bool) (string, error) {
	cfg := config.Get()
	expirationTime := time.Now().Add(expirationDuration)

	claims := &Claims{
		UserID:        userID,
		Email:         email,
		EmailVerified: verified,
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
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}

func DetachJWTCookie(w http.ResponseWriter) {
	cfg := config.Get()

	cookie := http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !cfg.DEV_MODE,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}
