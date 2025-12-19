package datamodel
package datamodel

import (
	"strconv"

	"gofr.dev/pkg/gofr"

	"github.com/Websoft9/apprun/core/internal/models"
)

// GetUsers 获取用户列表
func GetUsers(ctx *gofr.Context) (interface{}, error) {
	var users []models.User

	// 多租户自动过滤（由 TenantMiddleware 注入）
	tenantID := ctx.Get("tenant_id").(uint)

	result := ctx.GORM().Where("tenant_id = ?", tenantID).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// GetUser 获取单个用户
func GetUser(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 400,
			Message:    "invalid user id",
		}
	}

	var user models.User
	tenantID := ctx.Get("tenant_id").(uint)

	result := ctx.GORM().Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if result.Error != nil {
		return nil, &gofr.HTTP{
			StatusCode: 404,
			Message:    "user not found",
		}
	}

	return user, nil
}

// CreateUser 创建用户
func CreateUser(ctx *gofr.Context) (interface{}, error) {
	var user models.User

	if err := ctx.Bind(&user); err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 400,
			Message:    "invalid request body",
		}
	}

	// 注入租户 ID
	tenantID := ctx.Get("tenant_id").(uint)
	user.TenantID = tenantID

	if err := ctx.GORM().Create(&user).Error; err != nil {
		return nil, err
	}

	// 发布用户创建事件
	ctx.PubSub.Publish(ctx, "user.created", user)

	return user, nil
}

// UpdateUser 更新用户
func UpdateUser(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 400,
			Message:    "invalid user id",
		}
	}

	var user models.User
	tenantID := ctx.Get("tenant_id").(uint)

	// 查找用户
	result := ctx.GORM().Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user)
	if result.Error != nil {
		return nil, &gofr.HTTP{
			StatusCode: 404,
			Message:    "user not found",
		}
	}

	// 绑定更新数据
	if err := ctx.Bind(&user); err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 400,
			Message:    "invalid request body",
		}
	}

	// 更新
	if err := ctx.GORM().Save(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteUser 删除用户（软删除）
func DeleteUser(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, &gofr.HTTP{
			StatusCode: 400,
			Message:    "invalid user id",
		}
	}

	tenantID := ctx.Get("tenant_id").(uint)

	// 软删除
	result := ctx.GORM().Where("id = ? AND tenant_id = ?", userID, tenantID).Delete(&models.User{})
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, &gofr.HTTP{
			StatusCode: 404,
			Message:    "user not found",
		}
	}

	return map[string]string{
		"message": "user deleted successfully",
	}, nil
}
