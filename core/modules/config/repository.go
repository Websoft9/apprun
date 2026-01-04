package config

import (
	"context"

	"apprun/ent"
	"apprun/ent/configitem"
	"apprun/pkg/errors"
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

// GetConfig 根据 key 获取配置项（仅返回 active 状态）
func (r *Repository) GetConfig(ctx context.Context, key string) (value string, isDynamic bool, err error) {
	item, err := r.client.Configitem.
		Query().
		Where(
			configitem.KeyEQ(key),
			configitem.StatusEQ(configitem.StatusActive),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return "", false, errors.New(errors.ErrCodeConfigNotFound, "Config key not found").
				WithContext("key", key)
		}
		return "", false, errors.Wrap(err, errors.ErrCodeConfigQueryFailed, "Failed to query config").
			WithContext("key", key)
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
		return errors.Wrap(err, errors.ErrCodeConfigQueryFailed, "Failed to check config existence").
			WithContext("key", key)
	}

	if exists {
		// 更新现有配置
		err = r.client.Configitem.
			Update().
			Where(configitem.KeyEQ(key)).
			SetValue(value).
			SetStatus(configitem.StatusActive).
			Exec(ctx)

		if err != nil {
			return errors.Wrap(err, errors.ErrCodeConfigUpdateFailed, "Failed to update config").
				WithContext("key", key).
				WithContext("value", value)
		}
	} else {
		// 创建新配置项（标记为动态，状态为active）
		_, err = r.client.Configitem.
			Create().
			SetKey(key).
			SetValue(value).
			SetIsDynamic(true).
			SetStatus(configitem.StatusActive).
			Save(ctx)

		if err != nil {
			return errors.Wrap(err, errors.ErrCodeConfigCreateFailed, "Failed to create config").
				WithContext("key", key).
				WithContext("value", value)
		}
	}

	return nil
}

// ListDynamicConfigs 列出所有动态配置项（仅返回 active 状态）
func (r *Repository) ListDynamicConfigs(ctx context.Context) (map[string]string, error) {
	items, err := r.client.Configitem.
		Query().
		Where(
			configitem.IsDynamicEQ(true),
			configitem.StatusEQ(configitem.StatusActive),
		).
		All(ctx)

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigQueryFailed, "Failed to list dynamic configs")
	}

	result := make(map[string]string, len(items))
	for _, item := range items {
		result[item.Key] = item.Value
	}

	return result, nil
}

// DeleteConfig 删除动态配置项（软删除：设置为 inactive）
func (r *Repository) DeleteConfig(ctx context.Context, key string) error {
	affected, err := r.client.Configitem.
		Update().
		Where(configitem.KeyEQ(key)).
		SetStatus(configitem.StatusInactive).
		Save(ctx)

	if err != nil {
		return errors.Wrap(err, errors.ErrCodeConfigDeleteFailed, "Failed to delete config").
			WithContext("key", key)
	}

	if affected == 0 {
		return errors.New(errors.ErrCodeConfigNotFound, "Config key not found").
			WithContext("key", key)
	}

	return nil
}
