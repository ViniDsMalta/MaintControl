package main

import (
	"log"

	"MaintControl/internal/config"
	"MaintControl/internal/database"

	"github.com/joho/godotenv"

	"net/http"

	"MaintControl/internal/handlers"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load()

	err = database.Connect(cfg.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Postgres conectado")

	http.HandleFunc("/health", handlers.HealthHandler)

log.Println("Servidor iniciado na porta 8080")

err = http.ListenAndServe(":8080", nil)
if err != nil {
	log.Fatal(err)
}
}