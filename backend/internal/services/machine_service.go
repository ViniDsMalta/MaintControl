package services

import (
	"context"
	"strings"

	"MaintControl/internal/models"
)

type MachineStore interface {
	Create(ctx context.Context, machine models.Machine) (models.Machine, error)
	ListByUser(ctx context.Context, userID string) ([]models.Machine, error)
	GetByIDForUser(ctx context.Context, id, userID string) (models.Machine, error)
	UpdateForUser(ctx context.Context, machine models.Machine) (models.Machine, error)
	DeleteForUser(ctx context.Context, id, userID string) error
}

type MachineService struct {
	machines MachineStore
}

func NewMachineService(machines MachineStore) *MachineService {
	return &MachineService{machines: machines}
}

func (s *MachineService) Create(ctx context.Context, userID, name, machineType string) (models.Machine, error) {
	if err := validateUUID(userID); err != nil {
		return models.Machine{}, ErrUnauthorized
	}

	name = strings.TrimSpace(name)
	machineType = strings.TrimSpace(machineType)
	if !validName(name) || !validType(machineType) {
		return models.Machine{}, ErrInvalidInput
	}

	id, err := newUUID()
	if err != nil {
		return models.Machine{}, err
	}
	apiKey, err := randomToken(32)
	if err != nil {
		return models.Machine{}, err
	}

	return s.machines.Create(ctx, models.Machine{
		ID:     id,
		UserID: userID,
		Name:   name,
		Type:   machineType,
		APIKey: apiKey,
	})
}

func (s *MachineService) List(ctx context.Context, userID string) ([]models.Machine, error) {
	if err := validateUUID(userID); err != nil {
		return nil, ErrUnauthorized
	}
	return s.machines.ListByUser(ctx, userID)
}

func (s *MachineService) Get(ctx context.Context, userID, id string) (models.Machine, error) {
	if err := validateUUIDs(userID, id); err != nil {
		return models.Machine{}, ErrInvalidInput
	}
	return s.machines.GetByIDForUser(ctx, id, userID)
}

func (s *MachineService) Update(ctx context.Context, userID, id, name, machineType string) (models.Machine, error) {
	if err := validateUUIDs(userID, id); err != nil {
		return models.Machine{}, ErrInvalidInput
	}

	name = strings.TrimSpace(name)
	machineType = strings.TrimSpace(machineType)
	if !validName(name) || !validType(machineType) {
		return models.Machine{}, ErrInvalidInput
	}

	return s.machines.UpdateForUser(ctx, models.Machine{
		ID:     id,
		UserID: userID,
		Name:   name,
		Type:   machineType,
	})
}

func (s *MachineService) Delete(ctx context.Context, userID, id string) error {
	if err := validateUUIDs(userID, id); err != nil {
		return ErrInvalidInput
	}
	return s.machines.DeleteForUser(ctx, id, userID)
}

func validateUUIDs(ids ...string) error {
	for _, id := range ids {
		if err := validateUUID(id); err != nil {
			return err
		}
	}
	return nil
}
