package main

import (
	"apprun/ent"
	"apprun/internal/config"
	"apprun/routes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量构建数据库连接字符串
	// 环境变量使用 W9_ 前缀
	dbHost := getEnv("W9_DATABASE_HOST", "localhost")
	dbPort := getEnv("W9_DATABASE_PORT", "5432")
	dbUser := getEnv("W9_DATABASE_USER", "postgres")
	dbPassword := getEnv("W9_DATABASE_PASSWORD", "postgres")
	dbName := getEnv("W9_DATABASE_DBNAME", "apprun")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to database at %s:%s/%s", dbHost, dbPort, dbName)

	// 初始化数据库连接
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer client.Close()

	// 运行数据库迁移
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// 初始化配置系统
	if err := config.InitConfig(client); err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded successfully: App=%s, Version=%s", cfg.App.Name, cfg.App.Version)

	// 设置路由
	router := routes.SetupRoutes()

	// 启动HTTP服务器，监听8080端口
	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
