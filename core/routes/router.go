package routes

import (
	"apprun/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func SetupRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Use go-chi middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 创建处理器
	configHandler := handlers.NewConfigHandler()

	// 基础路由
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, apprun! This is a demo using go-chi framework."))
	})

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"apprun"}`))
	})

	// API路由组
	r.Route("/api", func(r chi.Router) {
		r.Route("/config", func(r chi.Router) {
			r.Get("/", configHandler.GetConfig)          // GET /api/config
			r.Put("/", configHandler.UpdateConfig)       // PUT /api/config
			r.Get("/{key}", configHandler.GetConfigItem) // GET /api/config/{key}
		})
	})

	// 兼容旧路由
	r.Route("/config", func(r chi.Router) {
		r.Get("/", configHandler.GetConfig)    // GET /config
		r.Put("/", configHandler.UpdateConfig) // PUT /config
	})

	return r
}
