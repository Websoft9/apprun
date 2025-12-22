package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"apprun/ent"
	"apprun/internal/config"

	_ "github.com/lib/pq"
)

func main() {
	// 从环境变量构建数据库连接字符串
	dbHost := getEnv("APP_DATABASE_HOST", "localhost")
	dbPort := getEnv("APP_DATABASE_PORT", "5432")
	dbUser := getEnv("APP_DATABASE_USER", "postgres")
	dbPassword := getEnv("APP_DATABASE_PASSWORD", "postgres")
	dbName := getEnv("APP_DATABASE_DBNAME", "apprun")

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

	// 定义handler函数
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, apprun! This is a demo using native Go net/http.")
	})

	// GET /config - 返回所有配置项（JSON数组），并标记哪些项可修改
	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetConfig(w, r)
		case http.MethodPut:
			handlePutConfig(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 启动HTTP服务器，监听8080端口
	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleGetConfig 处理 GET /config 请求
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	items, err := config.GetAllConfigItems()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode config items: %v", err)
	}
}

// handlePutConfig 处理 PUT /config 请求
func handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(updates) == 0 {
		http.Error(w, "No updates provided", http.StatusBadRequest)
		return
	}

	// 更新配置
	if err := config.UpdateConfig(updates); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "config key not allowed to be modified" {
			status = http.StatusForbidden
		} else if err.Error() == "config validation failed after update" {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	// 返回更新后的配置
	items, err := config.GetAllConfigItems()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(items); err != nil {
		log.Printf("Failed to encode updated config: %v", err)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
