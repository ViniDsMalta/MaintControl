package services

import (
	"context"
	"testing"

	"MaintControl/internal/models"
)

const (
	testUserA    = "11111111-1111-4111-8111-111111111111"
	testUserB    = "22222222-2222-4222-8222-222222222222"
	testMachineA = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
)

func TestMachineCreateListGetUpdateDelete(t *testing.T) {
	store := newFakeMachineStore()
	service := NewMachineService(store)

	created, err := service.Create(context.Background(), testUserA, "Motor Principal", "motor")
	if err != nil {
		t.Fatalf("expected create, got %v", err)
	}
	if created.APIKey == "" {
		t.Fatal("expected generated api key")
	}

	machines, err := service.List(context.Background(), testUserA)
	if err != nil || len(machines) != 1 {
		t.Fatalf("expected one machine, got %d and err %v", len(machines), err)
	}

	got, err := service.Get(context.Background(), testUserA, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("expected get own machine, got %v and err %v", got, err)
	}

	updated, err := service.Update(context.Background(), testUserA, created.ID, "Motor Atualizado", "motor")
	if err != nil || updated.Name != "Motor Atualizado" {
		t.Fatalf("expected update, got %v and err %v", updated, err)
	}

	if err := service.Delete(context.Background(), testUserA, created.ID); err != nil {
		t.Fatalf("expected delete, got %v", err)
	}
}

func TestMachineIsolationBetweenUsers(t *testing.T) {
	store := newFakeMachineStore()
	store.items[testMachineA] = models.Machine{ID: testMachineA, UserID: testUserA, Name: "Motor", Type: "motor"}
	service := NewMachineService(store)

	if _, err := service.Get(context.Background(), testUserB, testMachineA); err != ErrNotFound {
		t.Fatalf("expected not found for another user's machine, got %v", err)
	}
}

type fakeMachineStore struct {
	items map[string]models.Machine
}

func newFakeMachineStore() *fakeMachineStore {
	return &fakeMachineStore{items: map[string]models.Machine{}}
}

func (s *fakeMachineStore) Create(ctx context.Context, machine models.Machine) (models.Machine, error) {
	s.items[machine.ID] = machine
	return machine, nil
}

func (s *fakeMachineStore) ListByUser(ctx context.Context, userID string) ([]models.Machine, error) {
	result := []models.Machine{}
	for _, machine := range s.items {
		if machine.UserID == userID {
			result = append(result, machine)
		}
	}
	return result, nil
}

func (s *fakeMachineStore) GetByIDForUser(ctx context.Context, id, userID string) (models.Machine, error) {
	machine, ok := s.items[id]
	if !ok || machine.UserID != userID {
		return models.Machine{}, ErrNotFound
	}
	return machine, nil
}

func (s *fakeMachineStore) UpdateForUser(ctx context.Context, machine models.Machine) (models.Machine, error) {
	existing, ok := s.items[machine.ID]
	if !ok || existing.UserID != machine.UserID {
		return models.Machine{}, ErrNotFound
	}
	existing.Name = machine.Name
	existing.Type = machine.Type
	s.items[machine.ID] = existing
	return existing, nil
}

func (s *fakeMachineStore) DeleteForUser(ctx context.Context, id, userID string) error {
	existing, ok := s.items[id]
	if !ok || existing.UserID != userID {
		return ErrNotFound
	}
	delete(s.items, id)
	return nil
}
