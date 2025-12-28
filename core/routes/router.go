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

	// Health check at root
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"apprun"}`))
	})

	// Create handlers
	configHandler := handlers.NewConfigHandler()

	// API routes group
	r.Route("/api", func(r chi.Router) {
		// root route
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, apprun API"))
		})

		// feature/config routes
		r.Route("/config", func(r chi.Router) {
			r.Get("/", configHandler.GetConfig)          // GET /api/config
			r.Put("/", configHandler.UpdateConfig)       // PUT /api/config
			r.Get("/{key}", configHandler.GetConfigItem) // GET /api/config/{key}
		})
	})

	return r
}
