package auth
package auth

import (
	"gofr.dev/pkg/gofr"

	"github.com/Websoft9/apprun/core/internal/services"
)

// Register 用户注册处理器
func Register(authService *services.AuthService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		var req services.RegisterRequest

		// 绑定请求体
		if err := ctx.Bind(&req); err != nil {
			return nil, &gofr.HTTP{
				StatusCode: 400,
				Message:    "invalid request body",
			}
		}

		// 调用服务层
		resp, err := authService.Register(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// Login 用户登录处理器
func Login(authService *services.AuthService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		var req services.LoginRequest

		// 绑定请求体
		if err := ctx.Bind(&req); err != nil {
			return nil, &gofr.HTTP{
				StatusCode: 400,
				Message:    "invalid request body",
			}
		}

		// 调用服务层
		resp, err := authService.Login(ctx, req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}

// RefreshToken 刷新令牌处理器
func RefreshToken(authService *services.AuthService) func(*gofr.Context) (interface{}, error) {
	return func(ctx *gofr.Context) (interface{}, error) {
		// TODO: 实现刷新令牌逻辑
		return map[string]string{
			"message": "refresh token endpoint - to be implemented",
		}, nil
	}
}
