package services

import (
	"context"
	"crypto/subtle"

	"MaintControl/internal/models"
)

type SimulatorMachineStore interface {
	ListAll(ctx context.Context) ([]models.Machine, error)
}

type SimulatorMachine struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	APIKey string `json:"api_key"`
}

type SimulatorService struct {
	machines SimulatorMachineStore
	token    string
}

func NewSimulatorService(machines SimulatorMachineStore, token string) *SimulatorService {
	return &SimulatorService{machines: machines, token: token}
}

func (s *SimulatorService) ListMachines(ctx context.Context, token string) ([]SimulatorMachine, error) {
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
		return nil, ErrUnauthorized
	}

	machines, err := s.machines.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]SimulatorMachine, 0, len(machines))
	for _, machine := range machines {
		result = append(result, SimulatorMachine{
			ID: machine.ID, Name: machine.Name, Type: machine.Type, APIKey: machine.APIKey,
		})
	}
	return result, nil
}
