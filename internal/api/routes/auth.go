package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ThEditor/clutter-studio/internal/api/common"
	"github.com/ThEditor/clutter-studio/internal/api/middlewares"
	"github.com/ThEditor/clutter-studio/internal/repository"
	"github.com/go-chi/chi/v5"
)

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type VerifyRequest struct {
	Code string `json:"code" validate:"required,min=6,max=6"`
}

func AuthRouter(s *common.Server) http.Handler {
	r := chi.NewRouter()

	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := common.HashPassword(req.Password)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		user, err := s.Repo.CreateUser(s.Ctx, repository.CreateUserParams{
			Username: req.Username,
			Email:    req.Email,
			Passhash: hashedPassword,
		})

		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		verifyCode, err := s.Repo.CreateVerificationCode(s.Ctx, repository.CreateVerificationCodeParams{
			UserID:    user.ID,
			Code:      common.GenerateRandomCode(6),
			ExpiresAt: time.Now().Add(time.Hour),
		})

		if err == nil {
			common.SendVerificationMail(*s.Mailer, user.Email, verifyCode.Code)
		}

		jwt, err := common.CreateJWT(user.ID, user.Email, user.EmailVerified)

		if err != nil {
			http.Error(w, "Failed creating JWT", http.StatusInternalServerError)
			return
		}

		common.AttachJWTCookie(w, jwt)

		json.NewEncoder(w).Encode(map[string]string{
			"message":      "Successfully created!",
			"access_token": jwt,
		})
	})

	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := common.Validate.Struct(req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		user, err := s.Repo.FindUserByEmail(s.Ctx, req.Email)

		if err != nil || !common.CheckPasswordHash(user.Passhash, req.Password) {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		jwt, err := common.CreateJWT(user.ID, user.Email, user.EmailVerified)

		if err != nil {
			http.Error(w, "Failed creating JWT", http.StatusInternalServerError)
			return
		}

		common.AttachJWTCookie(w, jwt)

		json.NewEncoder(w).Encode(map[string]string{
			"message":      "Successfully logged in!",
			"access_token": jwt,
		})
	})

	r.With(middlewares.AuthWithoutEmailVerifiedMiddleware).
		Post("/generate-code", func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := s.Repo.FindUserByID(s.Ctx, claims.UserID)
			if err != nil {
				http.Error(w, "Cannot find user", http.StatusInternalServerError)
				return
			}

			if user.EmailVerified {
				http.Error(w, "User already has email verified", http.StatusBadRequest)
				return
			}

			verifyCode, err := s.Repo.CreateVerificationCode(s.Ctx, repository.CreateVerificationCodeParams{
				UserID:    user.ID,
				Code:      common.GenerateRandomCode(6),
				ExpiresAt: time.Now().Add(time.Hour),
			})

			if err != nil {
				http.Error(w, "Failed to create verification code", http.StatusInternalServerError)
				return
			}

			err = common.SendVerificationMail(*s.Mailer, user.Email, verifyCode.Code)

			if err != nil {
				http.Error(w, "Failed to create verification code", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{
				"message": "Successfully generated code!",
			})
		})

	r.With(middlewares.AuthWithoutEmailVerifiedMiddleware).
		Post("/verify", func(w http.ResponseWriter, r *http.Request) {
			var req VerifyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := common.Validate.Struct(req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			claims, ok := r.Context().Value(middlewares.ClaimsKey).(*common.Claims)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			user, err := s.Repo.FindUserByID(s.Ctx, claims.UserID)
			if err != nil {
				http.Error(w, "Cannot find user", http.StatusInternalServerError)
				return
			}

			if user.EmailVerified {
				http.Error(w, "User already has email verified", http.StatusBadRequest)
				return
			}

			valid, err := s.Repo.IsVerificationCodeValid(s.Ctx, repository.IsVerificationCodeValidParams{
				UserID: user.ID,
				Code:   req.Code,
			})
			if err != nil || !valid {
				http.Error(w, "Invalid code", http.StatusBadRequest)
				return
			}

			err = s.Repo.UpdateEmailVerificationStatus(s.Ctx, repository.UpdateEmailVerificationStatusParams{
				ID:            user.ID,
				EmailVerified: true,
			})
			if err != nil {
				http.Error(w, "Failed to update email verification status", http.StatusInternalServerError)
				return
			}

			s.Repo.DeleteVerificationCodes(s.Ctx, user.ID)

			jwt, err := common.CreateJWT(user.ID, user.Email, true)

			if err != nil {
				http.Error(w, "Failed creating JWT", http.StatusInternalServerError)
				return
			}

			common.AttachJWTCookie(w, jwt)

			json.NewEncoder(w).Encode(map[string]string{
				"message":      "Successfully verified email!",
				"access_token": jwt,
			})
		})

	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		common.DetachJWTCookie(w)

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Successfully logged out!",
		})
	})

	return r
}
