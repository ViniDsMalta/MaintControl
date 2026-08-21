package repositories

import (
	"context"

	"MaintControl/internal/models"
	"MaintControl/internal/services"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MachineRepository struct {
	db *pgxpool.Pool
}

func NewMachineRepository(db *pgxpool.Pool) *MachineRepository {
	return &MachineRepository{db: db}
}

func (r *MachineRepository) Create(ctx context.Context, machine models.Machine) (models.Machine, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO machines (id, user_id, name, type, api_key)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at`,
		machine.ID, machine.UserID, machine.Name, machine.Type, machine.APIKey,
	).Scan(&machine.CreatedAt)
	if err != nil {
		return models.Machine{}, err
	}

	_, _ = r.db.Exec(ctx,
		`INSERT INTO machine_status (machine_id, health_score, risk_score, status)
		 VALUES ($1, 100, 0, 'Normal')
		 ON CONFLICT (machine_id) DO NOTHING`,
		machine.ID,
	)

	return machine, nil
}

func (r *MachineRepository) ListByUser(ctx context.Context, userID string) ([]models.Machine, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, type, api_key, created_at
		 FROM machines
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []models.Machine{}
	for rows.Next() {
		var machine models.Machine
		if err := rows.Scan(&machine.ID, &machine.UserID, &machine.Name, &machine.Type, &machine.APIKey, &machine.CreatedAt); err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	return machines, rows.Err()
}

func (r *MachineRepository) GetByIDForUser(ctx context.Context, id, userID string) (models.Machine, error) {
	var machine models.Machine
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, type, api_key, created_at
		 FROM machines
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&machine.ID, &machine.UserID, &machine.Name, &machine.Type, &machine.APIKey, &machine.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.Machine{}, services.ErrNotFound
	}
	if err != nil {
		return models.Machine{}, err
	}

	return machine, nil
}

func (r *MachineRepository) UpdateForUser(ctx context.Context, machine models.Machine) (models.Machine, error) {
	err := r.db.QueryRow(ctx,
		`UPDATE machines
		 SET name = $1, type = $2
		 WHERE id = $3 AND user_id = $4
		 RETURNING id, user_id, name, type, api_key, created_at`,
		machine.Name, machine.Type, machine.ID, machine.UserID,
	).Scan(&machine.ID, &machine.UserID, &machine.Name, &machine.Type, &machine.APIKey, &machine.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.Machine{}, services.ErrNotFound
	}
	if err != nil {
		return models.Machine{}, err
	}

	return machine, nil
}

func (r *MachineRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM machines WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return nil
}
