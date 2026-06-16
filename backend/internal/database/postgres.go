package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

var DB *pgx.Conn

func Connect(connString string) error {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return err
	}

	DB = conn

	return nil
}