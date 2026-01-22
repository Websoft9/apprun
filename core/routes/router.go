// Package routes provides HTTP routing configuration and middleware setup.
package routes

import (
	"log"
	"net/http"

	"apprun/ent"
	"apprun/handlers"
	internalMiddleware "apprun/internal/middleware"
	adminHandler "apprun/modules/admin/handler"
	adminService "apprun/modules/admin/service"
	"apprun/modules/audit"
	auditHandler "apprun/modules/audit/handler"
	auditMiddleware "apprun/modules/audit/middleware"
	auditService "apprun/modules/audit/service"
	auditStorage "apprun/modules/audit/storage"
	authHandler "apprun/modules/auth/handler"
	authRepository "apprun/modules/auth/repository"
	authService "apprun/modules/auth/service"
	configModule "apprun/modules/config"
	"apprun/modules/metrics"
	"apprun/pkg/cache"
	"apprun/pkg/logger"
	"apprun/pkg/metricstore"
	"apprun/pkg/metricstore/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes 设置所有路由
// dbClient: 数据库客户端（必需，用于认证等模块）
// configService: 配置服务（可选）
// cacheClient: 缓存客户端（可选，用于metrics等模块）
func SetupRoutes(dbClient *ent.Client, configService *configModule.Service, cacheClient cache.Client) *chi.Mux {
	r := chi.NewRouter()

	// Use go-chi middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Use i18n language detector middleware
	r.Use(internalMiddleware.LanguageDetector())

	// Initialize audit service and middleware
	var auditSvc *auditService.Service

	// Get audit config from config service, or use defaults
	auditConfig := audit.DefaultConfig()
	// TODO: Get audit config from configService when it's available
	// if configService != nil {
	// 	auditConfig = configService.GetAuditConfig()
	// }

	auditStor := auditStorage.NewDatabaseStorage(dbClient)
	svc, err := auditService.NewService(auditStor, auditConfig.Service, auditConfig.Middleware.SensitiveFields)
	if err != nil {
		log.Printf("Failed to initialize audit service: %v", err)
	} else {
		auditSvc = svc
		if err := auditSvc.Start(); err != nil {
			log.Printf("Failed to start audit service: %v", err)
		} else {
			// Apply audit middleware globally
			auditMw := auditMiddleware.New(auditSvc, auditConfig.Middleware)
			r.Use(auditMw.Handler)
		}
	}

	// Health check at root (documented in Swagger)
	// Initialize comprehensive health handler with all dependencies
	healthHandler := handlers.NewHealthHandler(dbClient, cacheClient, true) //nolint:gocritic // Comment is informative
	r.Get("/health", healthHandler.Check)

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

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
		})

		// demo routes
		r.Route("/demo", func(r chi.Router) {
			r.Get("/i18n", handlers.I18nDemoHandler)
		})

		// RBAC routes (Story 5.5.2)
		RegisterRBACRoutes(r, dbClient)

		// Project management routes (Story 5.5 - Project CRUD)
		RegisterProjectRoutes(r, dbClient)

		// User self-service routes (Story 5.6)
		RegisterUserRoutes(r, dbClient)

		// Admin user management routes (Story 5.7)
		RegisterAdminUserRoutes(r, dbClient)

		// Metrics routes (Story 9.1 - Observability)
		if cacheClient != nil {
			RegisterMetricsRoutes(r, dbClient, cacheClient)
		}

		// Audit log routes (Story 5.9)
		if auditSvc != nil {
			RegisterAuditRoutes(r, dbClient, auditSvc)
		}

		// Config routes with RBAC protection (Story 3-2-1)
		if configService != nil {
			RegisterConfigRoutes(r, configService)
		}

		// Swagger documentation routes
		RegisterSwaggerInAPI(r)
	})

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

	// JWT middleware with DB client for token version validation (Story 5.7)
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

	// Project member management routes
	r.Route("/projects/{project_id}/members", func(r chi.Router) {
		// Apply JWT authentication with token version validation
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
		r.With(internalMiddleware.RequirePermission("project", "read")).Get("/me", permissionHandler.GetMyPermissions)

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

	// JWT middleware with DB client for token version validation (Story 5.7)
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

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

// RegisterUserRoutes registers user self-service routes (Story 5.6)
func RegisterUserRoutes(r chi.Router, dbClient *ent.Client) {
	// Initialize user service dependencies
	userRepo := authRepository.NewUserRepository(dbClient)
	userSvc := authService.NewUserService(userRepo)
	profileHandler := authHandler.NewProfileHandler(userSvc)
	passwordHandler := authHandler.NewPasswordHandler(userSvc)

	// JWT middleware with DB client for token version validation (Story 5.7)
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

	// User self-service routes
	r.Route("/profile", func(r chi.Router) {
		// All routes require authentication
		r.Use(jwtMiddleware.JWTAuth)

		// Get current user profile
		r.Get("/", profileHandler.GetProfile)

		// Update current user profile
		r.Put("/", profileHandler.UpdateProfile)

		// Change password
		r.Put("/password", passwordHandler.ChangePassword)
	})
}

// RegisterAuditRoutes registers audit log query routes (Story 5.9)
func RegisterAuditRoutes(r chi.Router, dbClient *ent.Client, auditSvc *auditService.Service) {
	// Initialize audit handler
	auditHdl := auditHandler.New(auditSvc, dbClient)

	// JWT middleware with DB client for token version validation
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

	// Admin audit routes (platform_admin only)
	r.Route("/admin/audit-logs", func(r chi.Router) {
		// Require authentication
		r.Use(jwtMiddleware.JWTAuth)
		// Require platform:audit:read permission (Story 5.5.4)
		r.Use(internalMiddleware.RequirePermission("platform:audit", "read"))

		// Query audit logs
		r.Get("/", auditHdl.QueryLogs)
	})
}

// RegisterAdminUserRoutes registers admin user management routes (Story 5.7)
func RegisterAdminUserRoutes(r chi.Router, dbClient *ent.Client) {
	// Initialize user management dependencies
	userMgmtSvc := adminService.NewUserMgmtService(dbClient)
	usersHandler := adminHandler.NewUsersHandler(userMgmtSvc)

	// JWT middleware with DB client for token version validation (Story 5.7)
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

	// Admin user management routes
	r.Route("/admin/users", func(r chi.Router) {
		// All routes require authentication
		r.Use(jwtMiddleware.JWTAuth)
		// All routes require platform:user:manage permission (Story 5.5.4)
		r.Use(internalMiddleware.RequirePermission("platform:user", "manage"))

		// Rate limit admin user management to 60 requests per minute per IP
		r.Use(middleware.Throttle(60))

		// List all users with filtering
		r.Get("/", usersHandler.ListUsers)

		// Create new user
		r.Post("/", usersHandler.CreateUser)

		// User-specific routes
		r.Route("/{id}", func(r chi.Router) {
			// Get user details
			r.Get("/", usersHandler.GetUser)

			// Change user role
			r.Put("/role", usersHandler.ChangeUserRole)

			// Change user status
			r.Put("/status", usersHandler.ChangeUserStatus)

			// Delete user (soft delete)
			r.Delete("/", usersHandler.DeleteUser)
		})
	})
}

// RegisterMetricsRoutes registers metrics routes (Story 9.1 - Observability)
func RegisterMetricsRoutes(r chi.Router, dbClient *ent.Client, cacheClient cache.Client) {
	// Initialize metrics storage repository (Story 9.1 - Storage integration)
	metricsCfg, err := metricstore.LoadConfig()
	if err != nil {
		// Log error but continue - metrics will work without persistence
		logger.L().Warn("Failed to load metrics config, persistence disabled", logger.Field{Key: "error", Value: err})
		metricsCfg = nil
	}

	var metricsRepo *metricstore.Repository
	if metricsCfg != nil {
		storageBackend, err := storage.NewStorage(metricsCfg.ToStorageConfig())
		if err != nil {
			logger.L().Warn("Failed to initialize metrics storage, persistence disabled", logger.Field{Key: "error", Value: err})
		} else {
			metricsRepo = metricstore.NewRepository(storageBackend, metricsCfg)
		}
	}

	// Initialize metrics service and handlers
	metricsService := metrics.NewMetricsService(dbClient, cacheClient, metricsRepo)
	metricsHandler := metrics.NewMetricsHandler(metricsService)

	// JWT middleware with DB client for token version validation
	jwtMiddleware := internalMiddleware.NewJWTMiddlewareWithDB(dbClient)

	// Metrics routes (platform_admin only)
	r.Route("/metrics", func(r chi.Router) {
		// Apply rate limiting: 100 requests per minute
		r.Use(middleware.Throttle(int(metrics.MetricsRateLimitRequests)))

		// Require authentication
		r.Use(jwtMiddleware.JWTAuth)
		// Require platform:metrics:read permission
		r.Use(internalMiddleware.RequirePermission("platform:metrics", "read"))

		// Core data endpoints (Refactored API)
		r.Get("/snapshot", metricsHandler.GetSnapshot)                                                                   // Replaces: /, /users, /system, /performance
		r.Get("/history", metricsHandler.GetHistory)                                                                     // Unchanged
		r.With(internalMiddleware.RequirePermission("platform:metrics", "write")).Post("/ingest", metricsHandler.Ingest) // Moved from /storage/ingest

		// Discovery & metadata endpoints (NEW)
		r.Get("/keys", metricsHandler.GetKeys)     // NEW: List available metric names
		r.Get("/scopes", metricsHandler.GetScopes) // NEW: List available snapshot scopes
	})
}

// RegisterConfigRoutes registers configuration management routes with RBAC protection (Story 3-2-1)
func RegisterConfigRoutes(r chi.Router, configService *configModule.Service) {
	configHandler := configModule.NewHandler(configService)

	// JWT middleware for authentication
	jwtMiddleware := internalMiddleware.NewJWTMiddleware()

	// Config routes - Platform-level permissions (projectID = 0)
	r.Route("/config", func(r chi.Router) {
		// Apply JWT authentication
		r.Use(jwtMiddleware.JWTAuth)

		// Read operations - require platform:config:read permission
		r.With(internalMiddleware.RequirePermission("platform:config", "read")).Get("/", configHandler.GetConfig)
		r.With(internalMiddleware.RequirePermission("platform:config", "read")).Get("/list", configHandler.ListConfigs)
		r.With(internalMiddleware.RequirePermission("platform:config", "read")).Get("/allowed", configHandler.GetAllowedKeys)

		// Write operations - require platform:config:write permission
		r.With(internalMiddleware.RequirePermission("platform:config", "write")).Put("/", configHandler.UpdateConfig)
		r.With(internalMiddleware.RequirePermission("platform:config", "write")).Delete("/", configHandler.DeleteConfig)
	})
}
