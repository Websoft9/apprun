package config

import (
	"context"
	"fmt"

	"apprun/ent"
	"apprun/ent/configitem"
)

// Repository 实现 ConfigProvider 接口，提供数据库访问层
// 使用反腐层模式，隔离 Ent 实现细节
type Repository struct {
	client *ent.Client
}

// NewRepository 创建配置仓储实例
func NewRepository(client *ent.Client) *Repository {
	return &Repository{client: client}
}

// GetConfig 根据 key 获取配置项
func (r *Repository) GetConfig(ctx context.Context, key string) (value string, isDynamic bool, err error) {
	item, err := r.client.Configitem.
		Query().
		Where(configitem.KeyEQ(key)).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return "", false, fmt.Errorf("config key not found: %s", key)
		}
		return "", false, fmt.Errorf("failed to query config: %w", err)
	}

	return item.Value, item.IsDynamic, nil
}

// SetConfig 设置动态配置项
func (r *Repository) SetConfig(ctx context.Context, key string, value string) error {
	// 检查配置项是否存在
	exists, err := r.client.Configitem.
		Query().
		Where(configitem.KeyEQ(key)).
		Exist(ctx)

	if err != nil {
		return fmt.Errorf("failed to check config existence: %w", err)
	}

	if exists {
		// 更新现有配置
		err = r.client.Configitem.
			Update().
			Where(configitem.KeyEQ(key)).
			SetValue(value).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	} else {
		// 创建新配置项（标记为动态）
		_, err = r.client.Configitem.
			Create().
			SetKey(key).
			SetValue(value).
			SetIsDynamic(true).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
	}

	return nil
}

// ListDynamicConfigs 列出所有动态配置项
func (r *Repository) ListDynamicConfigs(ctx context.Context) (map[string]string, error) {
	items, err := r.client.Configitem.
		Query().
		Where(configitem.IsDynamicEQ(true)).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to list dynamic configs: %w", err)
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.Key] = item.Value
	}

	return result, nil
}

// DeleteConfig 删除动态配置项
func (r *Repository) DeleteConfig(ctx context.Context, key string) error {
	deleted, err := r.client.Configitem.
		Delete().
		Where(configitem.KeyEQ(key)).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete config: %w", err)
	}

	if deleted == 0 {
		return fmt.Errorf("config key not found: %s", key)
	}

	return nil
}
