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
			projectRepo := authRepository.NewProjectRepository(dbClient)
			memberRepo := authRepository.NewProjectMemberRepository(dbClient)
			projectSvc := authService.NewProjectService(projectRepo, memberRepo)
			authSvc := authService.NewAuthService(userRepo, projectSvc)
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

		// RBAC routes (Story 5.5.2)
		RegisterRBACRoutes(r, dbClient)

		// Project management routes (Story 5.5 - Project CRUD)
		RegisterProjectRoutes(r, dbClient)

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

// RegisterRBACRoutes registers all RBAC-related routes (Story 5.5.2)
func RegisterRBACRoutes(r chi.Router, dbClient *ent.Client) {
	// Initialize RBAC dependencies
	projectRepo := authRepository.NewProjectRepository(dbClient)
	memberRepo := authRepository.NewProjectMemberRepository(dbClient)
	memberSvc := authService.NewProjectMemberService(memberRepo, projectRepo)
	permissionSvc := authService.NewPermissionService()

	memberHandler := authHandler.NewProjectMemberHandler(memberSvc)
	permissionHandler := authHandler.NewPermissionHandler(permissionSvc)

	// JWT middleware (required for all RBAC routes)
	jwtMiddleware := internalMiddleware.NewJWTMiddleware()

	// Project member management routes
	r.Route("/projects/{project_id}/members", func(r chi.Router) {
		// Apply JWT authentication
		r.Use(jwtMiddleware.JWTAuth)
		// Apply project context middleware (loads project from URL param)
		r.Use(internalMiddleware.ProjectContextMiddleware(dbClient))

		// List members - requires project:read permission
		r.With(internalMiddleware.RequirePermission("project", "read")).Get("/", memberHandler.ListMembers)

		// Add member - requires member:manage permission
		r.With(internalMiddleware.RequirePermission("member", "manage")).Post("/", memberHandler.AddMember)

		// Update and delete member - requires member:manage permission
		r.Route("/{member_id}", func(r chi.Router) {
			r.With(internalMiddleware.RequirePermission("member", "manage")).Put("/", memberHandler.UpdateRole)
			r.With(internalMiddleware.RequirePermission("member", "manage")).Delete("/", memberHandler.RemoveMember)
		})
	})

	// Permission query routes
	r.Route("/projects/{project_id}/permissions", func(r chi.Router) {
		// Apply JWT authentication
		r.Use(jwtMiddleware.JWTAuth)
		// Apply project context middleware
		r.Use(internalMiddleware.ProjectContextMiddleware(dbClient))

		// Get my permissions - requires project:read permission (user checks their own perms)
		r.With(internalMiddleware.RequirePermission("project", "read")).Get("/my", permissionHandler.GetMyPermissions)

		// Check specific permission - requires project:read permission
		r.With(internalMiddleware.RequirePermission("project", "read")).Post("/check", permissionHandler.CheckPermission)
	})
}

// RegisterProjectRoutes registers project CRUD routes
func RegisterProjectRoutes(r chi.Router, dbClient *ent.Client) {
	// Initialize project dependencies
	projectRepo := authRepository.NewProjectRepository(dbClient)
	memberRepo := authRepository.NewProjectMemberRepository(dbClient)
	projectSvc := authService.NewProjectService(projectRepo, memberRepo)
	memberSvc := authService.NewProjectMemberService(memberRepo, projectRepo)
	projectHandler := authHandler.NewProjectHandler(projectSvc, memberSvc)

	// JWT middleware (required for all project routes)
	jwtMiddleware := internalMiddleware.NewJWTMiddleware()

	// Project CRUD routes
	r.Route("/projects", func(r chi.Router) {
		// All project routes require authentication
		r.Use(jwtMiddleware.JWTAuth)

		// List user's projects
		r.Get("/", projectHandler.ListProjects)

		// Create new project
		r.Post("/", projectHandler.CreateProject)

		// Project-specific routes (by UUID)
		r.Route("/{id}", func(r chi.Router) {
			// Get project details
			r.Get("/", projectHandler.GetProject)

			// Update project (requires owner or admin role - checked in handler)
			r.Put("/", projectHandler.UpdateProject)

			// Delete project (requires owner role - checked in handler)
			r.Delete("/", projectHandler.DeleteProject)
		})
	})
}
