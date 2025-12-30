package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"apprun/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 1. 初始化 logger
	cfg := logger.Config{
		Level: logger.LevelDebug,
		Output: logger.OutputConfig{
			Targets: []string{"stdout"},
		},
	}

	log, err := logger.NewZapLogger(cfg)
	if err != nil {
		// Use fmt for initial logging before logger is ready
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close() // Ensure resources are cleaned up
	logger.SetLogger(log)

	// 2. 基本日志示例
	logger.Info("Application starting", logger.Field{"version", "1.0.0"})
	logger.Debug("Debug mode enabled")

	// 3. 结构化字段
	logger.Info("Server configuration",
		logger.Field{"port", 8080},
		logger.Field{"env", "development"},
		logger.Field{"max_connections", 100})

	// 4. 带固定字段的 logger
	dbLogger := logger.L().With(logger.Field{"component", "database"})
	dbLogger.Info("Database connected", logger.Field{"host", "localhost"})
	dbLogger.Info("Pool size set", logger.Field{"size", 10})

	// 5. HTTP 服务示例（自动 request_id）
	r := chi.NewRouter()
	r.Use(middleware.RequestID) // 重要：必须先注入 RequestID middleware
	r.Use(LoggingMiddleware)    // 自定义日志 middleware
	r.Get("/", HandleHome)
	r.Get("/user/{id}", HandleUser)

	logger.Info("Server starting", logger.Field{"address", ":8080"})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Fatal("Server failed", logger.Field{"error", err})
	}
}

// LoggingMiddleware demonstrates request logging with auto request_id
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 使用 WithContext 自动注入 request_id
		log := logger.L().WithContext(r.Context())

		log.Info("Request received",
			logger.Field{"method", r.Method},
			logger.Field{"path", r.URL.Path},
			logger.Field{"remote_addr", r.RemoteAddr})

		next.ServeHTTP(w, r)

		log.Info("Request completed",
			logger.Field{"duration_ms", time.Since(start).Milliseconds()})
	})
}

// HandleHome demonstrates basic handler logging
func HandleHome(w http.ResponseWriter, r *http.Request) {
	log := logger.L().WithContext(r.Context())
	log.Info("Home page accessed")

	w.Write([]byte("Welcome! Check logs for request_id tracking.\n"))
}

// HandleUser demonstrates service-scoped logging
func HandleUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// 创建带服务标识的 logger
	log := logger.L().WithContext(r.Context()).With(logger.Field{"service", "user"})

	log.Info("Fetching user", logger.Field{"user_id", userID})

	// 模拟业务逻辑
	if userID == "123" {
		log.Info("User found", logger.Field{"username", "john"})
		w.Write([]byte("User: john\n"))
	} else {
		log.Warn("User not found", logger.Field{"user_id", userID})
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found\n"))
	}
}

// Example of contextual logging in business logic
func BusinessLogicExample(ctx context.Context, userID int) error {
	// 业务逻辑中使用 WithContext 获取带 request_id 的 logger
	log := logger.L().WithContext(ctx)

	log.Debug("Starting business operation", logger.Field{"user_id", userID})

	// 模拟操作
	time.Sleep(10 * time.Millisecond)

	log.Info("Operation completed successfully", logger.Field{"user_id", userID})
	return nil
}
