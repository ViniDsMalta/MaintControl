package repositories

import (
	"context"

	"MaintControl/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TelemetryRepository struct {
	db *pgxpool.Pool
}

func NewTelemetryRepository(db *pgxpool.Pool) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

func (r *TelemetryRepository) Create(ctx context.Context, telemetry models.Telemetry) (models.Telemetry, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO telemetry (machine_id, temperature, vibration, rpm, pressure, flow_rate)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		telemetry.MachineID,
		telemetry.Temperature,
		telemetry.Vibration,
		telemetry.RPM,
		telemetry.Pressure,
		telemetry.FlowRate,
	).Scan(&telemetry.ID, &telemetry.CreatedAt)
	if err != nil {
		return models.Telemetry{}, err
	}
	return telemetry, nil
}

func (r *TelemetryRepository) LatestByMachine(ctx context.Context, machineID string, limit int) ([]models.Telemetry, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, machine_id, temperature, vibration, rpm, pressure, flow_rate, created_at
		 FROM telemetry
		 WHERE machine_id = $1
		 ORDER BY created_at DESC, id DESC
		 LIMIT $2`,
		machineID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []models.Telemetry{}
	for rows.Next() {
		var reading models.Telemetry
		if err := rows.Scan(
			&reading.ID,
			&reading.MachineID,
			&reading.Temperature,
			&reading.Vibration,
			&reading.RPM,
			&reading.Pressure,
			&reading.FlowRate,
			&reading.CreatedAt,
		); err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for left, right := 0, len(readings)-1; left < right; left, right = left+1, right-1 {
		readings[left], readings[right] = readings[right], readings[left]
	}
	return readings, nil
}

func (r *TelemetryRepository) UpdateStatus(ctx context.Context, status models.MachineStatus) (models.MachineStatus, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO machine_status (machine_id, health_score, risk_score, status, updated_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (machine_id) DO UPDATE SET
		   health_score = EXCLUDED.health_score,
		   risk_score = EXCLUDED.risk_score,
		   status = EXCLUDED.status,
		   updated_at = NOW()
		 RETURNING updated_at`,
		status.MachineID,
		status.HealthScore,
		status.RiskScore,
		status.Status,
	).Scan(&status.UpdatedAt)
	if err != nil {
		return models.MachineStatus{}, err
	}
	return status, nil
}
