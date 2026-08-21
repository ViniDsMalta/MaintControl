package services

import (
	"context"
	"testing"

	"MaintControl/internal/models"
)

const (
	testLineA = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
)

func TestProductionLineCRUDAndMachineMembership(t *testing.T) {
	lineStore := newFakeProductionLineStore()
	machineStore := newFakeMachineStore()
	machineStore.items[testMachineA] = models.Machine{ID: testMachineA, UserID: testUserA, Name: "Motor", Type: "motor"}
	service := NewProductionLineService(lineStore, machineStore)

	line, err := service.Create(context.Background(), testUserA, "Linha A")
	if err != nil {
		t.Fatalf("expected create line, got %v", err)
	}

	if _, err := service.Update(context.Background(), testUserA, line.ID, "Linha Atualizada"); err != nil {
		t.Fatalf("expected update line, got %v", err)
	}

	link, err := service.AddMachine(context.Background(), testUserA, line.ID, testMachineA, 1)
	if err != nil {
		t.Fatalf("expected add machine, got %v", err)
	}
	if link.Position != 1 {
		t.Fatal("expected position 1")
	}

	machines, err := service.ListMachines(context.Background(), testUserA, line.ID)
	if err != nil || len(machines) != 1 {
		t.Fatalf("expected one line machine, got %d and err %v", len(machines), err)
	}

	if err := service.RemoveMachine(context.Background(), testUserA, line.ID, testMachineA); err != nil {
		t.Fatalf("expected remove machine, got %v", err)
	}

	if err := service.Delete(context.Background(), testUserA, line.ID); err != nil {
		t.Fatalf("expected delete line, got %v", err)
	}
}

func TestProductionLineIsolationAndForeignMachineProtection(t *testing.T) {
	lineStore := newFakeProductionLineStore()
	lineStore.items[testLineA] = models.ProductionLine{ID: testLineA, UserID: testUserA, Name: "Linha A"}
	machineStore := newFakeMachineStore()
	machineStore.items[testMachineA] = models.Machine{ID: testMachineA, UserID: testUserB, Name: "Motor B", Type: "motor"}

	service := NewProductionLineService(lineStore, machineStore)

	if _, _, err := service.Get(context.Background(), testUserB, testLineA); err != ErrNotFound {
		t.Fatalf("expected not found for another user's line, got %v", err)
	}

	if _, err := service.AddMachine(context.Background(), testUserA, testLineA, testMachineA, 1); err != ErrNotFound {
		t.Fatalf("expected not found when adding another user's machine, got %v", err)
	}
}

type fakeProductionLineStore struct {
	items    map[string]models.ProductionLine
	machines map[string]models.ProductionLineMachine
}

func newFakeProductionLineStore() *fakeProductionLineStore {
	return &fakeProductionLineStore{
		items:    map[string]models.ProductionLine{},
		machines: map[string]models.ProductionLineMachine{},
	}
}

func (s *fakeProductionLineStore) Create(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error) {
	s.items[line.ID] = line
	return line, nil
}

func (s *fakeProductionLineStore) ListByUser(ctx context.Context, userID string) ([]models.ProductionLine, error) {
	result := []models.ProductionLine{}
	for _, line := range s.items {
		if line.UserID == userID {
			result = append(result, line)
		}
	}
	return result, nil
}

func (s *fakeProductionLineStore) GetByIDForUser(ctx context.Context, id, userID string) (models.ProductionLine, error) {
	line, ok := s.items[id]
	if !ok || line.UserID != userID {
		return models.ProductionLine{}, ErrNotFound
	}
	return line, nil
}

func (s *fakeProductionLineStore) UpdateForUser(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error) {
	existing, ok := s.items[line.ID]
	if !ok || existing.UserID != line.UserID {
		return models.ProductionLine{}, ErrNotFound
	}
	existing.Name = line.Name
	s.items[line.ID] = existing
	return existing, nil
}

func (s *fakeProductionLineStore) DeleteForUser(ctx context.Context, id, userID string) error {
	existing, ok := s.items[id]
	if !ok || existing.UserID != userID {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}

func (s *fakeProductionLineStore) AddMachine(ctx context.Context, linkID, productionLineID, machineID string, position int) (models.ProductionLineMachine, error) {
	for _, link := range s.machines {
		if link.ProductionLineID == productionLineID && link.MachineID == machineID {
			return models.ProductionLineMachine{}, ErrDuplicateAssociation
		}
		if link.ProductionLineID == productionLineID && link.Position == position {
			return models.ProductionLineMachine{}, ErrDuplicatePosition
		}
	}
	link := models.ProductionLineMachine{ID: linkID, ProductionLineID: productionLineID, MachineID: machineID, Position: position}
	s.machines[linkID] = link
	return link, nil
}

func (s *fakeProductionLineStore) RemoveMachine(ctx context.Context, productionLineID, machineID string) error {
	for id, link := range s.machines {
		if link.ProductionLineID == productionLineID && link.MachineID == machineID {
			delete(s.machines, id)
			return nil
		}
	}
	return ErrNotFound
}

func (s *fakeProductionLineStore) ListMachines(ctx context.Context, productionLineID string) ([]models.MachineInProductionLine, error) {
	result := []models.MachineInProductionLine{}
	for _, link := range s.machines {
		if link.ProductionLineID == productionLineID {
			result = append(result, models.MachineInProductionLine{
				Machine:  models.Machine{ID: link.MachineID},
				Position: link.Position,
			})
		}
	}
	return result, nil
}
