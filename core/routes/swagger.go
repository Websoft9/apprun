package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterSwagger 注册 Swagger UI 路由
// 访问路径: /api/docs 或 /api/docs/
// OpenAPI spec: /api/docs/doc.json
//
// 注意: 需要先运行 `make swagger` 生成文档
// 如果文档未生成，编译时 import _ "apprun/docs" 会失败
func RegisterSwagger(r chi.Router) {
	// 处理 /api/docs 重定向到 /api/docs/
	r.Get("/api/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/docs/", http.StatusMovedPermanently)
	})

	// Swagger UI 路由（使用相对路径，自动适配部署环境）
	r.Get("/api/docs/*", httpSwagger.Handler(
		httpSwagger.URL("doc.json"), // 使用相对路径，不绑定 host
	))
}
