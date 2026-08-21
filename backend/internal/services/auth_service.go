package services

import (
	"context"
	"strings"
	"time"

	"MaintControl/internal/models"
)

type UserStore interface {
	Create(ctx context.Context, user models.User) error
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByID(ctx context.Context, id string) (models.User, error)
}

type AuthService struct {
	users     UserStore
	jwtSecret string
}

func NewAuthService(users UserStore, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (models.User, string, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)

	if name == "" || !validEmail(email) || len(password) < 6 {
		return models.User{}, "", ErrInvalidInput
	}

	id, err := newUUID()
	if err != nil {
		return models.User{}, "", err
	}
	passwordHash, err := hashPassword(password)
	if err != nil {
		return models.User{}, "", err
	}

	user := models.User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return models.User{}, "", err
	}

	token, err := createJWT(user.ID, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return models.User{}, "", err
	}

	user.PasswordHash = ""
	return user, token, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	email = normalizeEmail(email)
	if !validEmail(email) || password == "" {
		return models.User{}, "", ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return models.User{}, "", ErrInvalidCredentials
	}
	if !verifyPassword(password, user.PasswordHash) {
		return models.User{}, "", ErrInvalidCredentials
	}

	token, err := createJWT(user.ID, s.jwtSecret, 24*time.Hour)
	if err != nil {
		return models.User{}, "", err
	}

	user.PasswordHash = ""
	return user, token, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (models.User, error) {
	if err := validateUUID(userID); err != nil {
		return models.User{}, ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) AuthenticateToken(token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrUnauthorized
	}

	return parseJWT(token, s.jwtSecret)
}
