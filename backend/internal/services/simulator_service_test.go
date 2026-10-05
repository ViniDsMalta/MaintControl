package services

import (
	"context"
	"testing"

	"MaintControl/internal/models"
)

func TestSimulatorMachineDiscoveryRequiresToken(t *testing.T) {
	store := &fakeSimulatorMachineStore{machines: []models.Machine{
		{ID: testMachineA, Name: "Motor", Type: "motor", APIKey: "machine-key", UserID: testUserA},
	}}
	service := NewSimulatorService(store, "simulator-secret")

	if _, err := service.ListMachines(context.Background(), "wrong"); err != ErrUnauthorized {
		t.Fatalf("expected unauthorized, got %v", err)
	}
	machines, err := service.ListMachines(context.Background(), "simulator-secret")
	if err != nil || len(machines) != 1 {
		t.Fatalf("expected one machine, got %+v and %v", machines, err)
	}
	if machines[0].APIKey != "machine-key" || machines[0].Type != "motor" {
		t.Fatalf("unexpected simulator machine: %+v", machines[0])
	}
}

type fakeSimulatorMachineStore struct {
	machines []models.Machine
}

func (s *fakeSimulatorMachineStore) ListAll(_ context.Context) ([]models.Machine, error) {
	return s.machines, nil
}
