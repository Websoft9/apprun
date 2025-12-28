package main

import (
	"apprun/routes"
	"log"
	"net/http"
	"os"
)

func main() {
	// 设置路由
	router := routes.SetupRoutes()

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
