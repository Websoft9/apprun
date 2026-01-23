package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterSwagger 注册 Swagger UI 路由（在根路由上）
// 访问路径: /api/docs 或 /api/docs/
// OpenAPI spec: /api/docs/doc.json
//
// 注意: 需要先运行 `make swagger` 生成文档到 core/apidocs/
// 如果文档未生成，编译时 import _ "apprun/apidocs" 会失败
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

// RegisterSwaggerInAPI 在 /api 路由组内注册 Swagger
func RegisterSwaggerInAPI(r chi.Router) {
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/api/docs/doc.json"),
	)

	// 处理 /docs 重定向到 /docs/index.html
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/docs/index.html", http.StatusMovedPermanently)
	})

	// 处理 /docs/ (带尾部斜杠) - chi 的 /* 不匹配空字符串
	r.Get("/docs/", swaggerHandler)

	// Swagger UI 路由 (处理 /docs/index.html, /docs/doc.json 等)
	r.Get("/docs/*", swaggerHandler)
}
