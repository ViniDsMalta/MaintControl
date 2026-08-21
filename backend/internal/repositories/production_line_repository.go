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

type ProductionLineRepository struct {
	db *pgxpool.Pool
}

func NewProductionLineRepository(db *pgxpool.Pool) *ProductionLineRepository {
	return &ProductionLineRepository{db: db}
}

func (r *ProductionLineRepository) Create(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO production_lines (id, user_id, name)
		 VALUES ($1, $2, $3)
		 RETURNING created_at`,
		line.ID, line.UserID, line.Name,
	).Scan(&line.CreatedAt)
	if err != nil {
		return models.ProductionLine{}, err
	}

	return line, nil
}

func (r *ProductionLineRepository) ListByUser(ctx context.Context, userID string) ([]models.ProductionLine, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, created_at
		 FROM production_lines
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := []models.ProductionLine{}
	for rows.Next() {
		var line models.ProductionLine
		if err := rows.Scan(&line.ID, &line.UserID, &line.Name, &line.CreatedAt); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}

	return lines, rows.Err()
}

func (r *ProductionLineRepository) GetByIDForUser(ctx context.Context, id, userID string) (models.ProductionLine, error) {
	var line models.ProductionLine
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, created_at
		 FROM production_lines
		 WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&line.ID, &line.UserID, &line.Name, &line.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.ProductionLine{}, services.ErrNotFound
	}
	if err != nil {
		return models.ProductionLine{}, err
	}

	return line, nil
}

func (r *ProductionLineRepository) UpdateForUser(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error) {
	err := r.db.QueryRow(ctx,
		`UPDATE production_lines
		 SET name = $1
		 WHERE id = $2 AND user_id = $3
		 RETURNING id, user_id, name, created_at`,
		line.Name, line.ID, line.UserID,
	).Scan(&line.ID, &line.UserID, &line.Name, &line.CreatedAt)
	if err == pgx.ErrNoRows {
		return models.ProductionLine{}, services.ErrNotFound
	}
	if err != nil {
		return models.ProductionLine{}, err
	}

	return line, nil
}

func (r *ProductionLineRepository) DeleteForUser(ctx context.Context, id, userID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM production_lines WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return nil
}

func (r *ProductionLineRepository) AddMachine(ctx context.Context, linkID, productionLineID, machineID string, position int) (models.ProductionLineMachine, error) {
	link := models.ProductionLineMachine{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO production_line_machines (id, production_line_id, machine_id, position)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, production_line_id, machine_id, position`,
		linkID, productionLineID, machineID, position,
	).Scan(&link.ID, &link.ProductionLineID, &link.MachineID, &link.Position)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "production_line_machine_unique" {
				return models.ProductionLineMachine{}, services.ErrDuplicateAssociation
			}
			return models.ProductionLineMachine{}, services.ErrDuplicatePosition
		}
		return models.ProductionLineMachine{}, err
	}

	return link, nil
}

func (r *ProductionLineRepository) RemoveMachine(ctx context.Context, productionLineID, machineID string) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM production_line_machines
		 WHERE production_line_id = $1 AND machine_id = $2`,
		productionLineID, machineID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return services.ErrNotFound
	}

	return nil
}

func (r *ProductionLineRepository) ListMachines(ctx context.Context, productionLineID string) ([]models.MachineInProductionLine, error) {
	rows, err := r.db.Query(ctx,
		`SELECT m.id, m.user_id, m.name, m.type, m.api_key, m.created_at, plm.position
		 FROM production_line_machines plm
		 INNER JOIN machines m ON m.id = plm.machine_id
		 WHERE plm.production_line_id = $1
		 ORDER BY plm.position ASC`,
		productionLineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []models.MachineInProductionLine{}
	for rows.Next() {
		var machine models.MachineInProductionLine
		if err := rows.Scan(
			&machine.ID,
			&machine.UserID,
			&machine.Name,
			&machine.Type,
			&machine.APIKey,
			&machine.CreatedAt,
			&machine.Position,
		); err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	return machines, rows.Err()
}
