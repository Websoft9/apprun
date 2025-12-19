package models
package models

import (
	"time"

	"gorm.io/gorm"
)

// Tenant 租户模型（多租户隔离）
type Tenant struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"uniqueIndex;not null" json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Domain      string `gorm:"uniqueIndex" json:"domain"`
	Status      string `gorm:"default:'active'" json:"status"` // active, suspended, deleted
	Plan        string `gorm:"default:'free'" json:"plan"`     // free, pro, enterprise

	// 关联关系
	Users []User `gorm:"foreignKey:TenantID" json:"users,omitempty"`
}

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 多租户字段
	TenantID uint   `gorm:"index;not null" json:"tenant_id"`
	Tenant   Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	// 用户信息
	Email        string `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	Name         string `json:"name" validate:"required"`
	PasswordHash string `json:"-"` // 不返回给客户端
	Avatar       string `json:"avatar"`
	Status       string `gorm:"default:'active'" json:"status"` // active, suspended, deleted

	// 关联关系
	Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// Role 角色模型（RBAC）
type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 多租户字段
	TenantID uint   `gorm:"index;not null" json:"tenant_id"`
	Tenant   Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	// 角色信息
	Name        string `gorm:"index:idx_tenant_role,unique" json:"name" validate:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	IsSystem    bool   `gorm:"default:false" json:"is_system"` // 系统内置角色不可删除

	// 关联关系
	Users       []User       `gorm:"many2many:user_roles" json:"users,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// Permission 权限模型（RBAC）
type Permission struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 权限信息
	Resource    string `gorm:"index:idx_permission,unique" json:"resource" validate:"required"` // 资源路径，如 /api/v1/users
	Action      string `gorm:"index:idx_permission,unique" json:"action" validate:"required"`   // 操作，如 GET, POST, PUT, DELETE
	Description string `json:"description"`

	// 关联关系
	Roles []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}

// Workflow 工作流模型
type Workflow struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// 多租户字段
	TenantID uint   `gorm:"index;not null" json:"tenant_id"`
	Tenant   Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	// 工作流信息
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	WorkflowID    string `gorm:"uniqueIndex;not null" json:"workflow_id"` // Temporal Workflow ID
	RunID         string `json:"run_id"`                                  // Temporal Run ID
	Status        string `gorm:"default:'running'" json:"status"`         // running, completed, failed, cancelled
	StartedAt     time.Time
	CompletedAt   *time.Time
	Input         string `gorm:"type:jsonb" json:"input"`   // JSON 格式的输入参数
	Output        string `gorm:"type:jsonb" json:"output"`  // JSON 格式的输出结果
	ErrorMessage  string `json:"error_message,omitempty"`
	
	// 创建者
	CreatedBy uint `gorm:"index" json:"created_by"`
	Creator   User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// AutoMigrate 自动迁移所有数据模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tenant{},
		&User{},
		&Role{},
		&Permission{},
		&Workflow{},
	)
}
