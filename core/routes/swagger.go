package routes

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// RegisterSwagger 注册 Swagger UI 路由
// 访问路径: /api/docs/
// OpenAPI spec: /api/docs/doc.json
func RegisterSwagger(r chi.Router) {
	r.Get("/api/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/api/docs/doc.json"),
	))
}
