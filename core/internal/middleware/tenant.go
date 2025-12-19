package middleware

import (
	"gofr.dev/pkg/gofr"
)

// TenantMiddleware 多租户隔离中间件
// 确保所有数据库查询自动过滤租户 ID
func TenantMiddleware() func(*gofr.Handler) gofr.Handler {
	return func(next gofr.Handler) gofr.Handler {
		return func(ctx *gofr.Context) (interface{}, error) {
			// 从上下文中获取租户 ID（由 AuthMiddleware 注入）
			tenantID, ok := ctx.Get("tenant_id").(uint)
			if !ok {
				return nil, &gofr.HTTP{
					StatusCode: 400,
					Message:    "tenant_id not found in context",
				}
			}

			// 验证租户 ID 的有效性
			if tenantID == 0 {
				return nil, &gofr.HTTP{
					StatusCode: 400,
					Message:    "invalid tenant_id",
				}
			}

			// 在 GORM 中设置全局 WHERE 条件（自动租户隔离）
			ctx.GORM().Where("tenant_id = ?", tenantID)

			// 继续执行下一个处理器
			return next(ctx)
		}
	}
}
