package models

import "time"

const (
	StatusWaitingData = "AGUARDANDO_DADOS"
	StatusNormal      = "NORMAL"
	StatusDegradation = "DEGRADATION"
	StatusFailure     = "FAILURE"
)

type TelemetryReading struct {
	Temperature *float64 `json:"temperature,omitempty"`
	Vibration   *float64 `json:"vibration,omitempty"`
	RPM         *float64 `json:"rpm,omitempty"`
	Pressure    *float64 `json:"pressure,omitempty"`
	FlowRate    *float64 `json:"flow_rate,omitempty"`
}

type Telemetry struct {
	ID        int64  `json:"id"`
	MachineID string `json:"machine_id"`
	TelemetryReading
	CreatedAt time.Time `json:"created_at"`
}

type MachineStatus struct {
	MachineID   string    `json:"machine_id"`
	HealthScore *float64  `json:"health_score"`
	RiskScore   *float64  `json:"risk_score"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AIPrediction struct {
	Prediction    string             `json:"prediction"`
	Probabilities map[string]float64 `json:"probabilities"`
}
