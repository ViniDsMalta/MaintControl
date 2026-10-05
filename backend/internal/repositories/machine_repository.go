package repositories

import (
	"context"
	"time"

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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Machine{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx,
		`INSERT INTO machines (id, user_id, name, type, api_key)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING created_at`,
		machine.ID, machine.UserID, machine.Name, machine.Type, machine.APIKey,
	).Scan(&machine.CreatedAt)
	if err != nil {
		return models.Machine{}, err
	}

	status := models.MachineStatus{MachineID: machine.ID, Status: models.StatusWaitingData}
	if err := tx.QueryRow(ctx,
		`INSERT INTO machine_status (machine_id, health_score, risk_score, status)
		 VALUES ($1, NULL, NULL, $2)
		 RETURNING updated_at`,
		machine.ID,
		status.Status,
	).Scan(&status.UpdatedAt); err != nil {
		return models.Machine{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Machine{}, err
	}
	machine.Status = &status

	return machine, nil
}

func (r *MachineRepository) ListByUser(ctx context.Context, userID string) ([]models.Machine, error) {
	rows, err := r.db.Query(ctx,
		machineWithStatusQuery+`
		 WHERE m.user_id = $1
		 ORDER BY m.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []models.Machine{}
	for rows.Next() {
		var machine models.Machine
		if err := scanMachine(rows, &machine); err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	return machines, rows.Err()
}

func (r *MachineRepository) GetByIDForUser(ctx context.Context, id, userID string) (models.Machine, error) {
	var machine models.Machine
	err := scanMachine(r.db.QueryRow(ctx, machineWithStatusQuery+` WHERE m.id = $1 AND m.user_id = $2`, id, userID), &machine)
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

	return r.GetByIDForUser(ctx, machine.ID, machine.UserID)
}

func (r *MachineRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM machines WHERE id = $1 AND user_id = $2)`, id, userID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return services.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `DELETE FROM telemetry WHERE machine_id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM machine_status WHERE machine_id = $1`, id); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `DELETE FROM machines WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *MachineRepository) GetByAPIKey(ctx context.Context, apiKey string) (models.Machine, error) {
	var machine models.Machine
	err := scanMachine(r.db.QueryRow(ctx, machineWithStatusQuery+` WHERE m.api_key = $1`, apiKey), &machine)
	if err == pgx.ErrNoRows {
		return models.Machine{}, services.ErrNotFound
	}
	if err != nil {
		return models.Machine{}, err
	}
	return machine, nil
}

func (r *MachineRepository) ListAll(ctx context.Context) ([]models.Machine, error) {
	rows, err := r.db.Query(ctx, machineWithStatusQuery+` ORDER BY m.created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []models.Machine{}
	for rows.Next() {
		var machine models.Machine
		if err := scanMachine(rows, &machine); err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}
	return machines, rows.Err()
}

const machineWithStatusQuery = `SELECT
		m.id, m.user_id, m.name, m.type, m.api_key, m.created_at,
		ms.health_score, ms.risk_score, ms.status, ms.updated_at
	 FROM machines m
	 LEFT JOIN machine_status ms ON ms.machine_id = m.id`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanMachine(row rowScanner, machine *models.Machine) error {
	var healthScore, riskScore *float64
	var status *string
	var updatedAt *time.Time
	err := row.Scan(
		&machine.ID,
		&machine.UserID,
		&machine.Name,
		&machine.Type,
		&machine.APIKey,
		&machine.CreatedAt,
		&healthScore,
		&riskScore,
		&status,
		&updatedAt,
	)
	if err != nil {
		return err
	}
	if status != nil && updatedAt != nil {
		machine.Status = &models.MachineStatus{
			MachineID:   machine.ID,
			HealthScore: healthScore,
			RiskScore:   riskScore,
			Status:      *status,
			UpdatedAt:   *updatedAt,
		}
	}
	return nil
}
