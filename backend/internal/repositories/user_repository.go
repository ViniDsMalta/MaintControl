package repositories

import (
	"context"
	"errors"

	"MaintControl/internal/models"
	"MaintControl/internal/services"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user models.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
		user.ID, user.Name, user.Email, user.PasswordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return services.ErrDuplicateEmail
		}
		return err
	}

	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	return r.getOne(ctx, `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`, email)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (models.User, error) {
	return r.getOne(ctx, `SELECT id, name, email, password_hash, created_at FROM users WHERE id = $1`, id)
}

func (r *UserRepository) getOne(ctx context.Context, query string, args ...any) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return models.User{}, services.ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
