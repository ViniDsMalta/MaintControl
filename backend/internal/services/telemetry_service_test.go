package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"MaintControl/internal/models"
)

func TestTelemetryRejectsInvalidAPIKeyAndMetrics(t *testing.T) {
	machines := &fakeTelemetryMachineStore{byKey: map[string]models.Machine{
		"motor-key": {ID: testMachineA, Type: "motor"},
	}}
	store := &fakeTelemetryStore{}
	predictor := &fakePredictor{}
	service := NewTelemetryService(machines, store, predictor)

	if _, err := service.Receive(context.Background(), "wrong-key", motorReading()); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized for unknown API key, got %v", err)
	}

	invalid := motorReading()
	invalid.RPM = nil
	invalid.Pressure = floatPointer(4.5)
	if _, err := service.Receive(context.Background(), "motor-key", invalid); err != ErrInvalidInput {
		t.Fatalf("expected invalid input for incompatible metrics, got %v", err)
	}
	if len(store.readings) != 0 {
		t.Fatal("invalid telemetry must not be stored")
	}
}

func TestTelemetryWaitsUntilThirtyReadings(t *testing.T) {
	machines := &fakeTelemetryMachineStore{byKey: map[string]models.Machine{
		"motor-key": {ID: testMachineA, Type: "motor"},
	}}
	store := &fakeTelemetryStore{}
	predictor := &fakePredictor{}
	service := NewTelemetryService(machines, store, predictor)

	result, err := service.Receive(context.Background(), "motor-key", motorReading())
	if err != nil {
		t.Fatalf("expected telemetry acceptance, got %v", err)
	}
	if result.PredictionPerformed || predictor.calls != 0 {
		t.Fatal("AI must not be called with fewer than 30 readings")
	}
	if result.ReadingsCollected != 1 || result.MachineStatus.Status != models.StatusWaitingData {
		t.Fatalf("unexpected waiting result: %+v", result)
	}
}

func TestTelemetryPredictsAndUpdatesStatusAtThirtyReadings(t *testing.T) {
	machines := &fakeTelemetryMachineStore{byKey: map[string]models.Machine{
		"motor-key": {ID: testMachineA, Type: "motor"},
	}}
	store := &fakeTelemetryStore{readings: makeMotorTelemetry(29)}
	predictor := &fakePredictor{prediction: models.AIPrediction{
		Prediction: models.StatusFailure,
		Probabilities: map[string]float64{
			models.StatusNormal: 0.2, models.StatusDegradation: 0.3, models.StatusFailure: 0.5,
		},
	}}
	service := NewTelemetryService(machines, store, predictor)

	result, err := service.Receive(context.Background(), "motor-key", motorReading())
	if err != nil {
		t.Fatalf("expected prediction, got %v", err)
	}
	if !result.PredictionPerformed || predictor.calls != 1 || len(predictor.received) != 30 {
		t.Fatalf("expected one prediction with 30 readings: %+v", result)
	}
	if *result.MachineStatus.RiskScore != 65 || *result.MachineStatus.HealthScore != 35 {
		t.Fatalf("unexpected derived scores: %+v", result.MachineStatus)
	}
	if result.MachineStatus.Status != models.StatusFailure {
		t.Fatalf("expected failure status, got %s", result.MachineStatus.Status)
	}
}

func TestTelemetryKeepsStoredReadingWhenAIIsUnavailable(t *testing.T) {
	machines := &fakeTelemetryMachineStore{byKey: map[string]models.Machine{
		"motor-key": {ID: testMachineA, Type: "motor"},
	}}
	store := &fakeTelemetryStore{readings: makeMotorTelemetry(29)}
	predictor := &fakePredictor{err: errors.New("offline")}
	service := NewTelemetryService(machines, store, predictor)

	result, err := service.Receive(context.Background(), "motor-key", motorReading())
	if err != nil {
		t.Fatalf("AI outage must not reject stored telemetry, got %v", err)
	}
	if len(store.readings) != 30 || result.PredictionError == "" || result.PredictionPerformed {
		t.Fatalf("unexpected degraded result: %+v", result)
	}
}

type fakeTelemetryMachineStore struct {
	byKey map[string]models.Machine
}

func (s *fakeTelemetryMachineStore) GetByAPIKey(_ context.Context, apiKey string) (models.Machine, error) {
	machine, ok := s.byKey[apiKey]
	if !ok {
		return models.Machine{}, ErrNotFound
	}
	return machine, nil
}

type fakeTelemetryStore struct {
	readings []models.Telemetry
	status   models.MachineStatus
}

func (s *fakeTelemetryStore) Create(_ context.Context, telemetry models.Telemetry) (models.Telemetry, error) {
	telemetry.ID = int64(len(s.readings) + 1)
	telemetry.CreatedAt = time.Unix(telemetry.ID, 0)
	s.readings = append(s.readings, telemetry)
	return telemetry, nil
}

func (s *fakeTelemetryStore) LatestByMachine(_ context.Context, _ string, limit int) ([]models.Telemetry, error) {
	start := len(s.readings) - limit
	if start < 0 {
		start = 0
	}
	return append([]models.Telemetry(nil), s.readings[start:]...), nil
}

func (s *fakeTelemetryStore) UpdateStatus(_ context.Context, status models.MachineStatus) (models.MachineStatus, error) {
	status.UpdatedAt = time.Now()
	s.status = status
	return status, nil
}

type fakePredictor struct {
	prediction models.AIPrediction
	err        error
	calls      int
	received   []models.TelemetryReading
}

func (p *fakePredictor) Predict(_ context.Context, _ string, readings []models.TelemetryReading) (models.AIPrediction, error) {
	p.calls++
	p.received = append([]models.TelemetryReading(nil), readings...)
	return p.prediction, p.err
}

func motorReading() models.TelemetryReading {
	return models.TelemetryReading{
		Temperature: floatPointer(55), Vibration: floatPointer(1.5), RPM: floatPointer(1800),
	}
}

func makeMotorTelemetry(count int) []models.Telemetry {
	readings := make([]models.Telemetry, count)
	for index := range readings {
		readings[index] = models.Telemetry{
			ID: int64(index + 1), MachineID: testMachineA, TelemetryReading: motorReading(),
		}
	}
	return readings
}

func floatPointer(value float64) *float64 {
	return &value
}
