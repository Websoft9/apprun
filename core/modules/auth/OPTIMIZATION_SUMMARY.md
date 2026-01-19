# Auth Module Optimization - Summary

## 优化内容

### 1. 命名优化：SuperAdmin → Init

**原因**：`Init` 更准确地表达了这是平台初始化配置，而不仅仅是超级管理员配置。

**修改内容**：
- `SuperAdminConfig` → `InitConfig`
- `Config.SuperAdmin` → `Config.Init`
- `DefaultSuperAdminConfig()` → `DefaultInitConfig()`
- 配置文件中 `super_admin` → `init`

**影响文件**：
- `core/modules/auth/config.go`
- `core/modules/auth/service/auth.go`
- `core/internal/bootstrap/server.go`
- `core/config/default.yaml`

### 2. 添加固定的 PlatformProjectUUID

**新增常量**：
```go
// modules/auth/config.go
PlatformProjectUUID = "00000000-0000-0000-0000-000000000000"
```

**优势**：
- ✅ 易于识别：全零 UUID 一眼就能看出是平台项目
- ✅ 固定不变：在数据库中可以直接使用此 UUID 查询平台项目
- ✅ 避免混淆：不会与自动生成的 UUID 冲突

**使用场景**：
```go
// 将来在创建 platform project 时可以使用
project := client.Project.Create().
    SetUUID(authmod.PlatformProjectUUID).  // 使用固定 UUID
    SetName(authmod.PlatformProjectName).
    SetDescription(authmod.PlatformProjectDescription).
    SetOwnerID(superAdminID).
    Save(ctx)
```

## 配置示例

### default.yaml
```yaml
auth:
  init:
    username: "admin"
    email: "admin@example.com"
    password: ""  # 留空自动生成随机密码
```

### 环境变量覆盖
```bash
export INIT_USERNAME="myAdmin"
export INIT_EMAIL="admin@mycompany.com"
export INIT_PASSWORD="MySecurePassword123!"
```

## API 变化

### 方法签名
```go
// 旧
func (s *AuthService) GetOrCreateSuperAdmin(ctx, adminConfig SuperAdminConfig)

// 新
func (s *AuthService) GetOrCreateSuperAdmin(ctx, initConfig InitConfig)
```

### 配置访问
```go
// 旧
authConfig.SuperAdmin.Email

// 新
authConfig.Init.Email
```

## 测试覆盖

已添加测试用例：
- `TestPlatformConstants` - 验证平台常量
- `TestInitConfig` - 验证初始化配置
- `TestDefaultConfig` - 验证默认配置

全部测试通过 ✅

## 向后兼容性

⚠️ **Breaking Change**：配置结构发生变化，需要更新 `default.yaml` 文件。

**迁移指南**：
```yaml
# 旧配置
auth:
  super_admin:
    username: "admin"
    
# 新配置
auth:
  init:
    username: "admin"
```
