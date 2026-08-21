package services

import (
	"context"
	"testing"

	"MaintControl/internal/models"
)

func TestAuthRegisterLoginAndDuplicateEmail(t *testing.T) {
	store := newFakeUserStore()
	service := NewAuthService(store, "test-secret")

	user, token, err := service.Register(context.Background(), "Joao", "joao@example.com", "senha123")
	if err != nil {
		t.Fatalf("expected valid register, got %v", err)
	}
	if user.ID == "" || token == "" {
		t.Fatal("expected user id and token")
	}
	if store.byEmail["joao@example.com"].PasswordHash == "senha123" {
		t.Fatal("password must not be stored as plain text")
	}

	if _, _, err := service.Register(context.Background(), "Joao", "joao@example.com", "senha123"); err != ErrDuplicateEmail {
		t.Fatalf("expected duplicate email, got %v", err)
	}

	loggedUser, loginToken, err := service.Login(context.Background(), "joao@example.com", "senha123")
	if err != nil {
		t.Fatalf("expected valid login, got %v", err)
	}
	if loggedUser.ID != user.ID || loginToken == "" {
		t.Fatal("expected matching logged user and token")
	}
}

func TestAuthLoginWrongPasswordAndInvalidJWT(t *testing.T) {
	store := newFakeUserStore()
	service := NewAuthService(store, "test-secret")

	if _, _, err := service.Register(context.Background(), "Maria", "maria@example.com", "senha123"); err != nil {
		t.Fatalf("register setup failed: %v", err)
	}

	if _, _, err := service.Login(context.Background(), "maria@example.com", "errada123"); err != ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}

	if _, err := service.AuthenticateToken("invalid-token"); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized token, got %v", err)
	}
}

type fakeUserStore struct {
	byID    map[string]models.User
	byEmail map[string]models.User
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{byID: map[string]models.User{}, byEmail: map[string]models.User{}}
}

func (s *fakeUserStore) Create(ctx context.Context, user models.User) error {
	if _, exists := s.byEmail[user.Email]; exists {
		return ErrDuplicateEmail
	}
	s.byID[user.ID] = user
	s.byEmail[user.Email] = user
	return nil
}

func (s *fakeUserStore) GetByEmail(ctx context.Context, email string) (models.User, error) {
	user, ok := s.byEmail[email]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return user, nil
}

func (s *fakeUserStore) GetByID(ctx context.Context, id string) (models.User, error) {
	user, ok := s.byID[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return user, nil
}
