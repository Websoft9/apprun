// Package routes provides HTTP routing configuration and middleware setup.
package routes

import (
	"log"
	"net/http"

	"apprun/ent"
	"apprun/handlers"
	internalMiddleware "apprun/internal/middleware"
	authHandler "apprun/modules/auth/handler"
	authRepository "apprun/modules/auth/repository"
	authService "apprun/modules/auth/service"
	configModule "apprun/modules/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRoutes 设置所有路由
// dbClient: 数据库客户端（必需，用于认证等模块）
// configService: 配置服务（可选）
func SetupRoutes(dbClient *ent.Client, configService *configModule.Service) *chi.Mux {
	r := chi.NewRouter()

	// Use go-chi middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Use i18n language detector middleware
	r.Use(internalMiddleware.LanguageDetector())

	// Health check at root (documented in Swagger)
	r.Get("/health", handlers.HealthHandler)

	// API routes group
	r.Route("/api", func(r chi.Router) {
		// root route
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("Hello, apprun API")); err != nil {
				log.Printf("Failed to write API root response: %v", err)
			}
		})

		// Authentication routes (public, no version prefix)
		r.Route("/auth", func(r chi.Router) {
			// Initialize auth dependencies
			userRepo := authRepository.NewUserRepository(dbClient)
			authSvc := authService.NewAuthService(userRepo)
			authHdl := authHandler.NewAuthHandler(authSvc)

			// Public endpoints (no JWT middleware)
			r.Post("/register", authHdl.Register)
			r.Post("/login", authHdl.Login)

			// Token refresh endpoint with rate limiting (10 req/hour per IP)
			r.With(middleware.Throttle(10)).Post("/refresh", authHdl.Refresh)

			// Protected endpoints (require JWT middleware)
			r.Group(func(r chi.Router) {
				jwtMiddleware := internalMiddleware.NewJWTMiddleware()
				r.Use(jwtMiddleware.JWTAuth)

				r.Get("/me", authHdl.Me)
			})
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
