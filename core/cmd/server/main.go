package main

import (
	"context"
	"log"
	"net/http"
	"os"

	_ "apprun/docs" // Swagger docs (自动生成)
	internalConfig "apprun/internal/config"
	"apprun/modules/config"
	"apprun/routes"

	_ "github.com/lib/pq"
)

// @title           AppRun API
// @version         1.0
// @description     AppRun 平台 REST API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.email   support@websoft9.com

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /api

// @schemes         http https
func main() {
	ctx := context.Background()

	// 创建配置引导器
	bootstrap := config.NewBootstrap(getEnv("CONFIG_DIR", "./config"))

	// 1. 加载初始配置
	cfg, err := bootstrap.LoadInitialConfig(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to load initial config: %v", err)
	}
	log.Printf("✅ Config loaded: %s v%s", cfg.App.Name, cfg.App.Version)

	// 2. 初始化数据库
	dbClient, err := bootstrap.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to initialize database: %v", err)
	}
	defer dbClient.Close()
	log.Println("✅ Database connected")

	// 3. 创建配置服务
	configService, err := bootstrap.CreateService(ctx, dbClient)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create config service: %v", err)
		log.Println("⚠️  Config API routes will not be registered")
	} else {
		log.Println("✅ Config service initialized with DB support")
	}

	// 4. 设置路由
	router := routes.SetupRoutes(configService)

	// 5. 启动服务器
	startServer(router, cfg)
}

// startServer 启动 HTTP/HTTPS 服务器
func startServer(router http.Handler, cfg *internalConfig.Config) {
	// 获取 TLS 配置
	sslCertFile := os.Getenv("SSL_CERT_FILE")
	sslKeyFile := os.Getenv("SSL_KEY_FILE")
	httpPort := getEnv("SERVER_PORT", "8080")
	httpsPort := getEnv("HTTPS_PORT", "8443")

	// 检查是否启用 TLS
	if sslCertFile != "" && sslKeyFile != "" {
		// 启动 HTTPS 服务器
		log.Printf("🔒 Starting HTTPS server on :%s", httpsPort)
		log.Printf("📄 Using certificate: %s", sslCertFile)

		// 同时启动 HTTP 服务器（用于健康检查和可能的重定向）
		go func() {
			httpAddr := ":" + httpPort
			log.Printf("🌐 Starting HTTP server on %s (for health checks)", httpAddr)
			if err := http.ListenAndServe(httpAddr, router); err != nil {
				log.Fatalf("HTTP server failed: %v", err)
			}
		}()

		// 启动 HTTPS 服务器
		httpsAddr := ":" + httpsPort
		if err := http.ListenAndServeTLS(httpsAddr, sslCertFile, sslKeyFile, router); err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	} else {
		// 仅启动 HTTP 服务器
		addr := ":" + httpPort
		log.Printf("🌐 Starting HTTP server on %s", addr)
		log.Printf("💡 Tip: Set SSL_CERT_FILE and SSL_KEY_FILE to enable HTTPS")
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
