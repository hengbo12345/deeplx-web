package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// DeepLX Configuration
	DeepLXURL   string
	DeepLXToken string

	// Server Configuration
	ServerPort string
	LogLevel   string

	// Upload Configuration
	UploadPath    string
	MaxFileSize   int64

	// Cleanup Configuration
	CleanupInterval time.Duration
	FileMaxAge      time.Duration
	UploadMaxSize   int64

	// Log Configuration
	LogPath        string
	LogMaxSize     int
	LogMaxBackups  int
	LogMaxAge      int

	// Task Configuration
	WorkerCount      int
	TaskMaxAge       time.Duration
	TaskCleanupInterval time.Duration

	// Auth Configuration
	AuthToken string
}

func Load() *Config {
	return &Config{
		DeepLXURL:       getEnv("DEEPLX_URL", "http://localhost:1188"),
		DeepLXToken:     getEnv("DEEPLX_TOKEN", ""),
		ServerPort:      getEnv("SERVER_PORT", "9448"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		UploadPath:      getEnv("UPLOAD_PATH", "./uploads"),
		MaxFileSize:     getInt64Env("MAX_FILE_SIZE", 10485760), // 10MB
		CleanupInterval: getDurationEnv("CLEANUP_INTERVAL", 1*time.Hour),
		FileMaxAge:      getDurationEnv("FILE_MAX_AGE", 24*time.Hour),
		UploadMaxSize:   getInt64Env("UPLOAD_MAX_SIZE", 1073741824), // 1GB
		LogPath:         getEnv("LOG_PATH", "./logs"),
		LogMaxSize:      getIntEnv("LOG_MAX_SIZE", 100),
		LogMaxBackups:   getIntEnv("LOG_MAX_BACKUPS", 3),
		LogMaxAge:       getIntEnv("LOG_MAX_AGE", 7),
		WorkerCount:     getIntEnv("WORKER_COUNT", 5),
		TaskMaxAge:      getDurationEnv("TASK_MAX_AGE", 1*time.Hour),
		TaskCleanupInterval: getDurationEnv("TASK_CLEANUP_INTERVAL", 5*time.Minute),
		AuthToken:           getEnv("AUTH_TOKEN", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}