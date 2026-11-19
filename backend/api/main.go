package main

import (
	"context"
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

	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../.env")
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}

	if err := database.Connect(cfg.ConnectionString()); err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	log.Println("Postgres conectado")

	userRepo := repositories.NewUserRepository(database.DB)
	machineRepo := repositories.NewMachineRepository(database.DB)
	productionLineRepo := repositories.NewProductionLineRepository(database.DB)
	telemetryRepo := repositories.NewTelemetryRepository(database.DB)
	startTelemetryRetention(telemetryRepo, cfg.RetentionDays)
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
	simulatorHandler := handlers.NewSimulatorHandler(simulatorService, telemetryService)

	mux := http.NewServeMux()
	loginLimiter := middleware.NewIPRateLimiter(10, 5*time.Minute)
	registerLimiter := middleware.NewIPRateLimiter(5, 10*time.Minute)
	mux.HandleFunc("/health", handlers.HealthHandler)
	mux.HandleFunc("/auth/register", registerLimiter.Wrap(authHandler.Register))
	mux.HandleFunc("/auth/login", loginLimiter.Wrap(authHandler.Login))
	mux.Handle("/auth/me", middleware.Auth(authService, http.HandlerFunc(authHandler.Me)))
	mux.Handle("/machines", middleware.Auth(authService, machineHandler))
	mux.Handle("/machines/", middleware.Auth(authService, machineHandler))
	mux.Handle("/production-lines", middleware.Auth(authService, productionLineHandler))
	mux.Handle("/production-lines/", middleware.Auth(authService, productionLineHandler))
	mux.HandleFunc("/telemetry", telemetryHandler.Receive)
	mux.HandleFunc("/simulator/machines", simulatorHandler.ListMachines)
	mux.HandleFunc("/simulator/telemetry", simulatorHandler.ReceiveTelemetry)

	log.Printf("Servidor iniciado na porta %s", cfg.ServerPort)

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func startTelemetryRetention(repository *repositories.TelemetryRepository, retentionDays int) {
	cleanup := func() {
		cutoff := time.Now().AddDate(0, 0, -retentionDays)
		deleted, err := repository.DeleteOlderThan(context.Background(), cutoff)
		if err != nil {
			log.Printf("telemetry retention failed: %v", err)
			return
		}
		if deleted > 0 {
			log.Printf("telemetry retention removed %d old readings", deleted)
		}
	}

	cleanup()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanup()
		}
	}()
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
