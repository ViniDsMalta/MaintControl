package services

import "errors"

var (
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrNotFound             = errors.New("not found")
	ErrConflict             = errors.New("conflict")
	ErrDuplicateEmail       = errors.New("email already exists")
	ErrDuplicateAssociation = errors.New("machine already belongs to this production line")
	ErrDuplicatePosition    = errors.New("position already in use")
)
