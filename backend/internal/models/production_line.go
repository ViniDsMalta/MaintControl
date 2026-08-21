package models

import "time"

type ProductionLine struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductionLineMachine struct {
	ID               string `json:"id"`
	ProductionLineID string `json:"production_line_id"`
	MachineID        string `json:"machine_id"`
	Position         int    `json:"position"`
}

type MachineInProductionLine struct {
	Machine
	Position int `json:"position"`
}
