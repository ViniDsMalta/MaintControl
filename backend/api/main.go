package main

import (
	"log"
	"net/http"

	"MaintControl/internal/config"
	"MaintControl/internal/database"
	"MaintControl/internal/handlers"
	"MaintControl/internal/middleware"
	"MaintControl/internal/repositories"
	"MaintControl/internal/services"

	"github.com/joho/godotenv"
)

func main() {

	_ = godotenv.Load()

	cfg := config.Load()

	if err := database.Connect(cfg.ConnectionString()); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	log.Println("Postgres conectado")

	userRepo := repositories.NewUserRepository(database.DB)
	machineRepo := repositories.NewMachineRepository(database.DB)
	productionLineRepo := repositories.NewProductionLineRepository(database.DB)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	machineService := services.NewMachineService(machineRepo)
	productionLineService := services.NewProductionLineService(productionLineRepo, machineRepo)

	authHandler := handlers.NewAuthHandler(authService)
	machineHandler := handlers.NewMachineHandler(machineService)
	productionLineHandler := handlers.NewProductionLineHandler(productionLineService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.Handle("/auth/me", middleware.Auth(authService, http.HandlerFunc(authHandler.Me)))
	mux.Handle("/machines", middleware.Auth(authService, machineHandler))
	mux.Handle("/machines/", middleware.Auth(authService, machineHandler))
	mux.Handle("/production-lines", middleware.Auth(authService, productionLineHandler))
	mux.Handle("/production-lines/", middleware.Auth(authService, productionLineHandler))

	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)

	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatal(err)
	}
}
