package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"deeplx-web/internal/config"
	"deeplx-web/internal/handler"
	"deeplx-web/internal/middleware"
	"deeplx-web/internal/service"
	"deeplx-web/internal/worker"
	"deeplx-web/pkg/utils"
)

// checkPythonDependency verifies that python3 and python-docx are available
func checkPythonDependency() {
	cmd := exec.Command("python3", "-c", "import docx")
	if err := cmd.Run(); err != nil {
		utils.Logger.Warn("Python python-docx not available — DOCX translation will fail. Install: pip3 install python-docx",
			zap.Error(err),
		)
		return
	}
	utils.Logger.Info("Python python-docx dependency verified")
}

func main() {
	cfg := config.Load()

	// Initialize logger with file rotation
	utils.InitLogger(cfg.LogPath, cfg.LogLevel, cfg.LogMaxSize, cfg.LogMaxBackups, cfg.LogMaxAge)
	defer utils.Sync()

	// Initialize middleware logger
	middleware.InitLogger(utils.Logger)

	// Auto-generate auth token if not set
	if cfg.AuthToken == "" {
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			utils.Logger.Fatal("Failed to generate auth token", zap.Error(err))
		}
		cfg.AuthToken = hex.EncodeToString(bytes)
		utils.Logger.Info("Generated auth token (set AUTH_TOKEN env var to override)", zap.String("token", cfg.AuthToken))
	}

	// Check Python docx_helper dependency
	checkPythonDependency()

	deeplxService := service.NewDeepLXService(cfg.DeepLXURL, cfg.DeepLXToken)
	docxService := service.NewDocxService(deeplxService)

	// Initialize task manager
	taskManager, err := service.NewTaskManager()
	if err != nil {
		utils.Logger.Fatal("Failed to initialize task manager", zap.Error(err))
	}

	// Initialize document worker
	documentWorker := worker.NewDocumentWorker(taskManager, docxService, cfg.UploadPath, cfg.WorkerCount)
	documentWorker.Start()
	defer documentWorker.Stop()

	translateHandler := handler.NewTranslateHandler(deeplxService)
	healthHandler := handler.NewHealthHandler()
	taskHandler := handler.NewTaskHandler(taskManager, documentWorker, cfg.UploadPath, cfg.MaxFileSize)
	authHandler := handler.NewAuthHandler(cfg.AuthToken)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Auth middleware - whitelist health and verify endpoints
	r.Use(middleware.TokenAuth(cfg.AuthToken, map[string]bool{
		"/health":          true,
		"/api/auth/verify": true,
	}))

	r.Get("/health", healthHandler.ServeHTTP)
	r.Post("/api/auth/verify", authHandler.Verify)
	r.Post("/api/translate", translateHandler.ServeHTTP)
	r.Post("/api/translate/document", taskHandler.CreateDocumentTask)
	r.Get("/api/tasks/{id}/status", taskHandler.GetTaskStatus)
	r.Get("/api/tasks/{id}/download", taskHandler.DownloadTaskResult)

	utils.StartCleanupService(cfg.UploadPath, cfg.FileMaxAge, cfg.UploadMaxSize, cfg.CleanupInterval)

	// Setup graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		utils.Logger.Info("Shutting down server...")
		documentWorker.Stop()
		utils.Logger.Info("Server stopped")
		os.Exit(0)
	}()

	addr := "127.0.0.1:" + cfg.ServerPort
	utils.Logger.Info("Starting server",
		zap.String("address", addr),
		zap.String("deepLX_url", cfg.DeepLXURL),
	)
	if err := http.ListenAndServe(addr, r); err != nil {
		utils.Logger.Fatal("Server failed", zap.String("error", err.Error()))
	}
}
