package routes

import (
	"net/http"

	"apprun/handlers"
	internalMiddleware "apprun/internal/middleware"
	configModule "apprun/modules/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRoutes 设置所有路由
// configService 参数可选，如果提供则注册配置 API 路由
func SetupRoutes(configService *configModule.Service) *chi.Mux {
	r := chi.NewRouter()

	// Use go-chi middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Use i18n language detector middleware
	r.Use(internalMiddleware.LanguageDetector())

	// Health check at root
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"apprun"}`))
	})

	// API routes group
	r.Route("/api", func(r chi.Router) {
		// root route
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Hello, apprun API"))
		})

		// demo routes
		r.Route("/demo", func(r chi.Router) {
			r.Get("/i18n", handlers.I18nDemoHandler)
		})

		// feature/config routes (如果提供了配置服务)
		if configService != nil {
			configHandler := configModule.NewHandler(configService)
			configHandler.RegisterRoutes(r)
		}
	})

	// Swagger 文档路由（挂载到 /api/docs/）
	RegisterSwagger(r)

	return r
}
