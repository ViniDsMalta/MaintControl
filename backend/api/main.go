package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"MaintControl/internal/clients"
	"MaintControl/internal/config"
	"MaintControl/internal/database"
	"MaintControl/internal/handlers"
	"MaintControl/internal/middleware"
	"MaintControl/internal/repositories"
	"MaintControl/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

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
	telemetryRepo := repositories.NewTelemetryRepository(database.DB)
	aiClient := clients.NewAIClient(cfg.AIServiceURL, cfg.AITimeout)

	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	machineService := services.NewMachineService(machineRepo)
	productionLineService := services.NewProductionLineService(productionLineRepo, machineRepo)
	telemetryService := services.NewTelemetryService(machineRepo, telemetryRepo, aiClient)
	simulatorService := services.NewSimulatorService(machineRepo, cfg.SimulatorToken)

	authHandler := handlers.NewAuthHandler(authService)
	machineHandler := handlers.NewMachineHandler(machineService)
	productionLineHandler := handlers.NewProductionLineHandler(productionLineService)
	telemetryHandler := handlers.NewTelemetryHandler(telemetryService)
	simulatorHandler := handlers.NewSimulatorHandler(simulatorService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/auth/register", authHandler.Register)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.Handle("/auth/me", middleware.Auth(authService, http.HandlerFunc(authHandler.Me)))
	mux.Handle("/machines", middleware.Auth(authService, machineHandler))
	mux.Handle("/machines/", middleware.Auth(authService, machineHandler))
	mux.Handle("/production-lines", middleware.Auth(authService, productionLineHandler))
	mux.Handle("/production-lines/", middleware.Auth(authService, productionLineHandler))
	mux.HandleFunc("/telemetry", telemetryHandler.Receive)
	mux.HandleFunc("/simulator/machines", simulatorHandler.ListMachines)

	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)

	if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
		log.Fatal(err)
	}
}

func runHealthcheck() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	client := http.Client{Timeout: 3 * time.Second}
	response, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil || response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = response.Body.Close()
}
