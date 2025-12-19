package server
package main

import (
	"os"

	"gofr.dev/pkg/gofr"

	"github.com/Websoft9/apprun/core/internal/config"
	"github.com/Websoft9/apprun/core/internal/handlers/auth"
	"github.com/Websoft9/apprun/core/internal/handlers/datamodel"
	"github.com/Websoft9/apprun/core/internal/handlers/workflow"
	"github.com/Websoft9/apprun/core/internal/middleware"
	"github.com/Websoft9/apprun/core/internal/models"
	"github.com/Websoft9/apprun/core/internal/services"
)

func main() {
	// 1. 创建 GoFr 应用
	app := gofr.New()

	// 2. 加载配置
	cfg := config.Load()

	// 3. 数据库自动迁移
	app.Migrate(func(gApp *gofr.Gofr) error {
		return models.AutoMigrate(gApp.GORM())
	})

	// 4. 初始化服务层
	authService := services.NewAuthService()
	eventService := services.NewEventService(cfg.TemporalHost)
	workflowService := services.NewWorkflowService(cfg.TemporalHost)

	// 5. 应用全局中间件（可选，用于受保护的路由）
	// app.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	// app.Use(middleware.TenantMiddleware())

	// 6. 注册公共路由（无需认证）
	app.POST("/auth/register", auth.Register(authService))
	app.POST("/auth/login", auth.Login(authService))
	app.POST("/auth/refresh", auth.RefreshToken(authService))

	// 7. 健康检查路由（GoFr 自动提供 /.well-known/health-check）
	app.GET("/health", func(ctx *gofr.Context) (interface{}, error) {
		return map[string]string{
			"status":  "healthy",
			"version": "1.0.0",
		}, nil
	})

	// 8. 创建受保护的路由组
	protectedRoutes := func(a *gofr.App) {
		// 应用认证和租户中间件
		a.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		a.Use(middleware.TenantMiddleware())

		// 用户管理 API
		a.GET("/api/v1/users", datamodel.GetUsers)
		a.POST("/api/v1/users", datamodel.CreateUser)
		a.GET("/api/v1/users/:id", datamodel.GetUser)
		a.PUT("/api/v1/users/:id", datamodel.UpdateUser)
		a.DELETE("/api/v1/users/:id", datamodel.DeleteUser)

		// 工作流 API
		a.POST("/api/v1/workflows", workflow.StartWorkflow(workflowService))
		a.GET("/api/v1/workflows/:id", workflow.GetWorkflow(workflowService))
		a.POST("/api/v1/workflows/:id/signal", workflow.SignalWorkflow(workflowService))
		a.DELETE("/api/v1/workflows/:id", workflow.CancelWorkflow(workflowService))
	}

	// 应用受保护的路由组
	protectedRoutes(app)

	// 9. 订阅事件（NATS）
	eventService.SubscribeToEvents(app)

	// 10. 启动服务
	// 自动暴露：
	// - HTTP API: :8080
	// - 健康检查: /.well-known/health-check
	// - Prometheus 指标: /metrics
	app.Logger.Infof("Starting apprun-core on port %s", os.Getenv("APP_PORT"))
	app.Run()
}
