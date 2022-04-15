package main

import (
	"net/http"
	"time"

	"github.com/senicko/run-api/pkg/api"
	"github.com/senicko/run-api/pkg/controllers"
	"github.com/senicko/run-api/pkg/services"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sandboxService, err := services.NewSandboxService(logger, services.SandboxServiceConfig{
		PoolWorkers: 8,
	})
	if err != nil {
		logger.Error("failed to create sandbox service")
	}

	sandboxController := controllers.NewSandboxController(logger, sandboxService)

	mux := http.NewServeMux()
	mux.HandleFunc("/run", api.Method(http.MethodPost, sandboxController.Run))

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}

	logger.Info("Starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("failed to start")
	}
}
