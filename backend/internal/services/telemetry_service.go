package services

import (
	"context"
	"errors"
	"log"
	"math"
	"strings"

	"MaintControl/internal/models"
)

const telemetryWindowSize = 30

type TelemetryMachineStore interface {
	GetByAPIKey(ctx context.Context, apiKey string) (models.Machine, error)
}

type TelemetryStore interface {
	Create(ctx context.Context, telemetry models.Telemetry) (models.Telemetry, error)
	LatestByMachine(ctx context.Context, machineID string, limit int) ([]models.Telemetry, error)
	UpdateStatus(ctx context.Context, status models.MachineStatus) (models.MachineStatus, error)
}

type Predictor interface {
	Predict(ctx context.Context, machineType string, readings []models.TelemetryReading) (models.AIPrediction, error)
}

type TelemetryResult struct {
	Telemetry           models.Telemetry      `json:"telemetry"`
	ReadingsCollected   int                   `json:"readings_collected"`
	RequiredReadings    int                   `json:"required_readings"`
	PredictionPerformed bool                  `json:"prediction_performed"`
	PredictionError     string                `json:"prediction_error,omitempty"`
	MachineStatus       *models.MachineStatus `json:"machine_status,omitempty"`
}

type TelemetryService struct {
	machines  TelemetryMachineStore
	telemetry TelemetryStore
	predictor Predictor
}

func NewTelemetryService(machines TelemetryMachineStore, telemetry TelemetryStore, predictor Predictor) *TelemetryService {
	return &TelemetryService{machines: machines, telemetry: telemetry, predictor: predictor}
}

func (s *TelemetryService) Receive(ctx context.Context, apiKey string, reading models.TelemetryReading) (TelemetryResult, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return TelemetryResult{}, ErrUnauthorized
	}

	machine, err := s.machines.GetByAPIKey(ctx, apiKey)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return TelemetryResult{}, ErrUnauthorized
		}
		return TelemetryResult{}, err
	}
	if !validTelemetry(machine.Type, reading) {
		return TelemetryResult{}, ErrInvalidInput
	}

	stored, err := s.telemetry.Create(ctx, models.Telemetry{
		MachineID:        machine.ID,
		TelemetryReading: reading,
	})
	if err != nil {
		return TelemetryResult{}, err
	}

	latest, err := s.telemetry.LatestByMachine(ctx, machine.ID, telemetryWindowSize)
	if err != nil {
		return TelemetryResult{}, err
	}
	result := TelemetryResult{
		Telemetry:         stored,
		ReadingsCollected: len(latest),
		RequiredReadings:  telemetryWindowSize,
	}

	if len(latest) < telemetryWindowSize {
		status, updateErr := s.telemetry.UpdateStatus(ctx, models.MachineStatus{
			MachineID: machine.ID,
			Status:    models.StatusWaitingData,
		})
		if updateErr != nil {
			return TelemetryResult{}, updateErr
		}
		result.MachineStatus = &status
		return result, nil
	}

	window := make([]models.TelemetryReading, len(latest))
	for index, telemetry := range latest {
		window[index] = telemetry.TelemetryReading
	}
	prediction, err := s.predictor.Predict(ctx, machine.Type, window)
	if err != nil {
		log.Printf("AI prediction unavailable for machine %s: %v", machine.ID, err)
		result.PredictionError = "prediction temporarily unavailable"
		result.MachineStatus = machine.Status
		return result, nil
	}

	status, err := statusFromPrediction(machine.ID, prediction)
	if err != nil {
		log.Printf("AI prediction invalid for machine %s: %v", machine.ID, err)
		result.PredictionError = "invalid prediction response"
		result.MachineStatus = machine.Status
		return result, nil
	}
	updated, err := s.telemetry.UpdateStatus(ctx, status)
	if err != nil {
		return TelemetryResult{}, err
	}
	result.PredictionPerformed = true
	result.MachineStatus = &updated
	return result, nil
}

func validTelemetry(machineType string, reading models.TelemetryReading) bool {
	if !finitePointers(reading.Temperature, reading.Vibration, reading.RPM, reading.Pressure, reading.FlowRate) {
		return false
	}
	switch machineType {
	case "motor":
		return reading.Temperature != nil && reading.Vibration != nil && reading.RPM != nil &&
			reading.Pressure == nil && reading.FlowRate == nil
	case "bomba":
		return reading.Pressure != nil && reading.FlowRate != nil && reading.Vibration != nil &&
			reading.Temperature == nil && reading.RPM == nil
	case "compressor":
		return reading.Temperature != nil && reading.Pressure != nil && reading.Vibration != nil &&
			reading.RPM == nil && reading.FlowRate == nil
	default:
		return false
	}
}

func finitePointers(values ...*float64) bool {
	for _, value := range values {
		if value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0)) {
			return false
		}
	}
	return true
}

func statusFromPrediction(machineID string, prediction models.AIPrediction) (models.MachineStatus, error) {
	if prediction.Prediction != models.StatusNormal &&
		prediction.Prediction != models.StatusDegradation &&
		prediction.Prediction != models.StatusFailure {
		return models.MachineStatus{}, ErrInvalidInput
	}

	normal, normalOK := prediction.Probabilities[models.StatusNormal]
	degradation, degradationOK := prediction.Probabilities[models.StatusDegradation]
	failure, failureOK := prediction.Probabilities[models.StatusFailure]
	if !normalOK || !degradationOK || !failureOK ||
		!validProbability(normal) || !validProbability(degradation) || !validProbability(failure) {
		return models.MachineStatus{}, ErrInvalidInput
	}
	if math.Abs(normal+degradation+failure-1) > 0.01 {
		return models.MachineStatus{}, ErrInvalidInput
	}

	// Degradation contributes half risk and failure contributes full risk.
	risk := math.Round((degradation*50+failure*100)*100) / 100
	health := math.Round((100-risk)*100) / 100
	return models.MachineStatus{
		MachineID:   machineID,
		HealthScore: &health,
		RiskScore:   &risk,
		Status:      prediction.Prediction,
	}, nil
}

func validProbability(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
