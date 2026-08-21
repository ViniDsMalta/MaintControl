package services

import (
	"context"
	"strings"

	"MaintControl/internal/models"
)

type ProductionLineStore interface {
	Create(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error)
	ListByUser(ctx context.Context, userID string) ([]models.ProductionLine, error)
	GetByIDForUser(ctx context.Context, id, userID string) (models.ProductionLine, error)
	UpdateForUser(ctx context.Context, line models.ProductionLine) (models.ProductionLine, error)
	DeleteForUser(ctx context.Context, id, userID string) error
	AddMachine(ctx context.Context, linkID, productionLineID, machineID string, position int) (models.ProductionLineMachine, error)
	RemoveMachine(ctx context.Context, productionLineID, machineID string) error
	ListMachines(ctx context.Context, productionLineID string) ([]models.MachineInProductionLine, error)
}

type ProductionLineService struct {
	lines    ProductionLineStore
	machines MachineStore
}

func NewProductionLineService(lines ProductionLineStore, machines MachineStore) *ProductionLineService {
	return &ProductionLineService{lines: lines, machines: machines}
}

func (s *ProductionLineService) Create(ctx context.Context, userID, name string) (models.ProductionLine, error) {
	if err := validateUUID(userID); err != nil {
		return models.ProductionLine{}, ErrUnauthorized
	}

	name = strings.TrimSpace(name)
	if !validName(name) {
		return models.ProductionLine{}, ErrInvalidInput
	}

	id, err := newUUID()
	if err != nil {
		return models.ProductionLine{}, err
	}

	return s.lines.Create(ctx, models.ProductionLine{ID: id, UserID: userID, Name: name})
}

func (s *ProductionLineService) List(ctx context.Context, userID string) ([]models.ProductionLine, error) {
	if err := validateUUID(userID); err != nil {
		return nil, ErrUnauthorized
	}
	return s.lines.ListByUser(ctx, userID)
}

func (s *ProductionLineService) Get(ctx context.Context, userID, id string) (models.ProductionLine, []models.MachineInProductionLine, error) {
	if err := validateUUIDs(userID, id); err != nil {
		return models.ProductionLine{}, nil, ErrInvalidInput
	}

	line, err := s.lines.GetByIDForUser(ctx, id, userID)
	if err != nil {
		return models.ProductionLine{}, nil, err
	}

	machines, err := s.lines.ListMachines(ctx, id)
	if err != nil {
		return models.ProductionLine{}, nil, err
	}

	return line, machines, nil
}

func (s *ProductionLineService) Update(ctx context.Context, userID, id, name string) (models.ProductionLine, error) {
	if err := validateUUIDs(userID, id); err != nil {
		return models.ProductionLine{}, ErrInvalidInput
	}

	name = strings.TrimSpace(name)
	if !validName(name) {
		return models.ProductionLine{}, ErrInvalidInput
	}

	return s.lines.UpdateForUser(ctx, models.ProductionLine{ID: id, UserID: userID, Name: name})
}

func (s *ProductionLineService) Delete(ctx context.Context, userID, id string) error {
	if err := validateUUIDs(userID, id); err != nil {
		return ErrInvalidInput
	}
	return s.lines.DeleteForUser(ctx, id, userID)
}

func (s *ProductionLineService) AddMachine(ctx context.Context, userID, productionLineID, machineID string, position int) (models.ProductionLineMachine, error) {
	if err := validateUUIDs(userID, productionLineID, machineID); err != nil {
		return models.ProductionLineMachine{}, ErrInvalidInput
	}
	if position < 1 {
		return models.ProductionLineMachine{}, ErrInvalidInput
	}

	if _, err := s.lines.GetByIDForUser(ctx, productionLineID, userID); err != nil {
		return models.ProductionLineMachine{}, err
	}
	if _, err := s.machines.GetByIDForUser(ctx, machineID, userID); err != nil {
		return models.ProductionLineMachine{}, err
	}

	linkID, err := newUUID()
	if err != nil {
		return models.ProductionLineMachine{}, err
	}

	return s.lines.AddMachine(ctx, linkID, productionLineID, machineID, position)
}

func (s *ProductionLineService) RemoveMachine(ctx context.Context, userID, productionLineID, machineID string) error {
	if err := validateUUIDs(userID, productionLineID, machineID); err != nil {
		return ErrInvalidInput
	}

	if _, err := s.lines.GetByIDForUser(ctx, productionLineID, userID); err != nil {
		return err
	}
	if _, err := s.machines.GetByIDForUser(ctx, machineID, userID); err != nil {
		return err
	}

	return s.lines.RemoveMachine(ctx, productionLineID, machineID)
}

func (s *ProductionLineService) ListMachines(ctx context.Context, userID, productionLineID string) ([]models.MachineInProductionLine, error) {
	if err := validateUUIDs(userID, productionLineID); err != nil {
		return nil, ErrInvalidInput
	}

	if _, err := s.lines.GetByIDForUser(ctx, productionLineID, userID); err != nil {
		return nil, err
	}

	return s.lines.ListMachines(ctx, productionLineID)
}
