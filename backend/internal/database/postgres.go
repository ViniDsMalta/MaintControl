package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(connString string) error {
	conn, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}

	if err := conn.Ping(context.Background()); err != nil {
		conn.Close()
		return err
	}

	DB = conn

	return nil
}
