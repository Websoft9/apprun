package middleware
package middleware

import (
	"strings"

	"gofr.dev/pkg/gofr"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(jwtSecret string) func(*gofr.Handler) gofr.Handler {
	return func(next gofr.Handler) gofr.Handler {
		return func(ctx *gofr.Context) (interface{}, error) {
			// 获取 Authorization header
			authHeader := ctx.Request.Header.Get("Authorization")
			if authHeader == "" {
				return nil, &gofr.HTTP{
					StatusCode: 401,
					Message:    "missing authorization header",
				}
			}

			// 检查 Bearer 前缀
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return nil, &gofr.HTTP{
					StatusCode: 401,
					Message:    "invalid authorization format",
				}
			}

			// 提取 token
			tokenString := authHeader[7:]

			// 解析 JWT
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// 验证签名算法
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, &gofr.HTTP{
						StatusCode: 401,
						Message:    "invalid signing method",
					}
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				return nil, &gofr.HTTP{
					StatusCode: 401,
					Message:    "invalid or expired token",
				}
			}

			// 提取 claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, &gofr.HTTP{
					StatusCode: 401,
					Message:    "invalid token claims",
				}
			}

			// 注入用户信息到上下文
			ctx.Set("user_id", uint(claims["sub"].(float64)))
			ctx.Set("tenant_id", uint(claims["tenant_id"].(float64)))
			ctx.Set("email", claims["email"].(string))

			// 继续执行下一个处理器
			return next(ctx)
		}
	}
}
