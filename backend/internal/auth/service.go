package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	AddUser(ctx context.Context, email, hash string) error
	GetUserID(ctx context.Context, email string) (string, error)
	GetPasswordHash(ctx context.Context, email string) (string, error)
}

type ServiceAuth struct {
	repo      Repository
	jwtSecret []byte
}

func NewService(repo Repository, jwtSecret string) *ServiceAuth {
	return &ServiceAuth{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

type UserClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *ServiceAuth) generateToken(userID, email string) (string, error) {
	claims := &UserClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "cryptofund-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}

func (s *ServiceAuth) RegisterService(ctx context.Context, data RegisterRequest) (ResponseAuth, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return ResponseAuth{}, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.AddUser(ctx, data.Email, string(hash)); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return ResponseAuth{}, ErrUserAlreadyExists
		}
		return ResponseAuth{}, fmt.Errorf("failed to add user: %w", err)
	}

	userID, err := s.repo.GetUserID(ctx, data.Email)
	if err != nil {
		return ResponseAuth{}, fmt.Errorf("eror getting user ID: %w", err)
	}

	token, err := s.generateToken(userID, data.Email)
	if err != nil {
		return ResponseAuth{}, fmt.Errorf("generate token: %w", err)
	}

	return ResponseAuth{
		Token:  token,
		UserID: userID,
		Email:  data.Email,
	}, nil
}

func (s *ServiceAuth) LoginService(ctx context.Context, data LoginRequest) (ResponseAuth, error) {
	hash, err := s.repo.GetPasswordHash(ctx, data.Email)
	if err != nil {
		return ResponseAuth{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(data.Password)); err != nil {
		return ResponseAuth{}, ErrInvalidCredentials
	}

	userID, err := s.repo.GetUserID(ctx, data.Email)
	if err != nil {
		return ResponseAuth{}, fmt.Errorf("failed to get user ID: %w", err)
	}

	token, err := s.generateToken(userID, data.Email)
	if err != nil {
		return ResponseAuth{}, fmt.Errorf("generate token: %w", err)
	}

	return ResponseAuth{
		Token:  token,
		UserID: userID,
		Email:  data.Email,
	}, nil
}
