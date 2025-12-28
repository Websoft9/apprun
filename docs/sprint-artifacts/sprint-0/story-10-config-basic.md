# Story 10: Configuration Center Foundation

**Epic**: Sprint-0 基础设施  
**Priority**: High  
**Points**: 5  
**Status**: Ready  
**Sprint**: Sprint-0

---

## 📋 User Story

**As a** Platform Developer  
**I want** 统一的配置管理系统，支持多种配置源和自动环境变量映射  
**So that** 配置灵活可控，敏感信息安全，运维简单

---

## 🎯 Acceptance Criteria

### 1. 配置优先级实现（6层完整版）
- [ ] 实现完整配置优先级（从高到低）：
  1. 环境变量（最高优先级）
  2. 数据库配置（`configitems` 表）
  3. 用户配置目录（`config/conf_d/*.yaml`）
  4. 领域配置文件（`config/database.yaml`, `config/server.yaml` 等）
  5. 默认配置文件（`config/default.yaml`）
  6. 结构体默认值（`default` tag，最低优先级）
- [ ] 环境变量能够覆盖所有其他配置源
- [ ] 环境变量为空值视为"未设置"，将回退到下一级配置源
- [ ] 数据库配置保护机制：数据库连接配置不从数据库加载（防止循环依赖）

### 2. 环境变量自动映射
- [ ] 无环境变量前缀
- [ ] 自动映射规则：`配置组.配置项（小写） → 配置组_配置项（大写）`
- [ ] 示例：`database.host` → `DATABASE_HOST`
- [ ] 无需手动注册，新增配置项自动支持环境变量

### 3. 配置结构统一
- [ ] `core/internal/config/types.go` 作为唯一配置定义来源
- [ ] 支持 struct tag：`default`, `validate`, `db`
- [ ] 配置文件格式：YAML

### 4. 测试验证
- [ ] 环境变量覆盖测试通过
- [ ] 配置优先级测试通过
- [ ] Docker Compose 环境变量配置正确

---

## 📦 Deliverables

### 1. 核心实现

#### `core/internal/config/loader.go` 
```go
package config

import (
    "os"
    "strings"
    "github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
    v := viper.New()
    
    // 1. 结构体默认值（从 default tag 提取）
    applyStructDefaults(v, &Config{})
    
    // 2. 默认配置文件
    v.SetConfigFile(configPath)
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    
    // 3. 领域配置文件（如 database.yaml, server.yaml）
    loadDomainConfigs(v, "config")
    
    // 4. 用户配置目录（conf_d/*.yaml）
    loadConfD(v, "config/conf_d")
    
    // 5. 数据库配置（configitems 表）
    loadFromDB(v)
    
    // 6. 环境变量自动映射（最高优先级）
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    cfg := &Config{}
    if err := v.Unmarshal(cfg); err != nil {
        return nil, err
    }
    
    return cfg, nil
}

// 从结构体 default tag 提取默认值
func applyStructDefaults(v *viper.Viper, cfg *Config) {
    // 使用反射提取 default tag 并设置
}

// 加载领域配置文件
func loadDomainConfigs(v *viper.Viper, configDir string) {
    // 加载 database.yaml, server.yaml 等
}

// 加载 conf_d 目录下的配置文件
func loadConfD(v *viper.Viper, confDDir string) {
    // 按字母顺序加载 *.yaml 文件
}

// 从数据库加载配置（排除数据库连接配置本身）
func loadFromDB(v *viper.Viper) {
    // 连接数据库，查询 configitems 表
    // 注意：不加载 database.* 配置（防止循环依赖）
}
```

#### `core/internal/config/types.go` 

- 添加所有配置项的 `default` tag
- 清晰的配置结构体定义
- 注释说明环境变量映射规则

**示例**:
```go
type Config struct {
    Database struct {
        Host     string `yaml:"host" default:"localhost" db:"false"`     // 环境变量: DATABASE_HOST
        Port     int    `yaml:"port" default:"5432" db:"false"`          // 环境变量: DATABASE_PORT
        User     string `yaml:"user" default:"postgres" db:"false"`      // 环境变量: DATABASE_USER
        Password string `yaml:"password" default:"" db:"false"`          // 环境变量: DATABASE_PASSWORD (数据库配置不从DB加载)
        DBName   string `yaml:"dbname" default:"apprun" db:"false"`      // 环境变量: DATABASE_DBNAME
    } `yaml:"database"`
    
    App struct {
        Name    string `yaml:"name" default:"apprun" db:"true"`         // 环境变量: APP_NAME (可从数据库加载)
        Version string `yaml:"version" default:"1.0.0" db:"true"`       // 环境变量: APP_VERSION
    } `yaml:"app"`
}
```

> **注意**: `db:"false"` 标记的配置项不会从数据库加载，防止循环依赖

#### `docker-compose.yml`

- 文件头部注释说明环境变量映射规则
- 提供默认值（使用 `${VAR:-default}` 语法）

### 2. 测试

#### 单元测试：`core/internal/config/loader_test.go`
- 测试配置加载逻辑
- 测试环境变量映射规则
- 测试优先级覆盖

#### 集成测试：`tests/integration/config/test-priority.sh`
- 测试环境变量 > DB > 文件的优先级
- 测试自动映射规则
- 输出详细日志

### 3. 文档

#### `docs/standards/coding-standards.md` (Section 14)

- 配置优先级规范
- 环境变量映射规则
- 结构体标签使用说明
- 数据库配置保护机制

#### `docs/product/setup/configuration.md`

- 用户配置指南
- 环境变量使用示例
- 配置文件说明
- 配置 API 使用方法

---

## 🔧 Technical Design

### 配置加载流程（完整6层）

```
1. 结构体默认值 (types.go `default` tag)
   ↓
2. 默认配置文件 (config/default.yaml)
   ↓
3. 领域配置文件 (config/database.yaml, config/server.yaml 等)
   ↓
4. 用户配置目录 (config/conf_d/*.yaml，按字母顺序加载)
   ↓
5. 数据库配置 (configitems 表，排除 db:"false" 标记的配置)
   ↓
6. 环境变量 (无前缀，自动映射，最高优先级)
```

### 数据库配置保护机制

为防止循环依赖（加载配置需要数据库连接，但数据库连接配置本身在数据库中），使用 `db` tag 标记：

```go
type Config struct {
    Database struct {
        Host string `yaml:"host" default:"localhost" db:"false"`  // 不从数据库加载
        // ... 其他数据库连接配置
    }
    
    App struct {
        Name string `yaml:"name" default:"apprun" db:"true"`  // 可从数据库加载
    }
}
```

**加载逻辑**:
```go
func loadFromDB(v *viper.Viper) {
    // 使用已加载的数据库配置建立连接
    db := connectWithCurrentConfig(v)
    
    // 查询 configitems 表
    rows := db.Query("SELECT key, value FROM configitems")
    for rows.Next() {
        var key, value string
        rows.Scan(&key, &value)
        
        // 检查 struct tag，跳过 db:"false" 的配置
        if !isDBLoadable(key) {
            continue
        }
        
        v.Set(key, value)
    }
}
```

### 环境变量映射规则

| 配置路径 | 环境变量名 | 示例值 |
|---------|-----------|--------|
| `app.name` | `APP_NAME` | `apprun` |
| `database.host` | `DATABASE_HOST` | `postgres` |
| `database.dbname` | `DATABASE_DBNAME` | `apprun` |
| `server.http_port` | `SERVER_HTTP_PORT` | `8080` |

**规则**: 
- 无前缀（保持简洁）
- 路径中的 `.` 转为 `_`
- 全部大写

---

## 🧪 Testing Strategy

### 单元测试
```go
// core/internal/config/loader_test.go
func TestConfigPriority(t *testing.T) {
    // 测试环境变量 > 文件
    os.Setenv("APP_NAME", "env-app")
    cfg, _ := Load("testdata/config.yaml")
    assert.Equal(t, "env-app", cfg.App.Name)
}
```

### 集成测试
```bash
# 启动环境并测试
make test-config-priority
```

### 验证清单
- [ ] 无环境变量时使用 default.yaml
- [ ] 领域配置文件覆盖 default.yaml
- [ ] conf_d/ 配置覆盖领域配置
- [ ] 数据库配置覆盖文件配置
- [ ] 环境变量覆盖所有配置源
- [ ] 环境变量为空时回退到下一级配置源
- [ ] 数据库连接配置不从数据库加载（db:"false" 生效）
- [ ] 新增配置项无需代码修改即可支持环境变量

---

## 📝 Notes

### 设计原则
- **完整性优先**: 实现完整的 6 层配置优先级
- **循环依赖保护**: `db` tag 防止数据库配置从数据库加载
- **约定优于配置**: 自动映射，无需手动注册
- **灵活性**: 支持多种配置源，满足不同场景需求

### 依赖关系
- **依赖**: Story 1 (Docker环境)
- **被依赖**: 所有需要配置管理的Story

### 风险
- ⚠️ 环境变量名称与现有系统冲突 → **缓解**: 使用明确的命名规范（大写+下划线）
- ⚠️ 配置优先级混淆 → **缓解**: 清晰的文档和日志

---

## ✅ Definition of Done

- [ ] `core/internal/config/loader.go` 实现完成
- [ ] `core/internal/config/types.go` 添加所有标签
- [ ] `core/ent/schema/configitem.go` Schema 定义
- [ ] 配置加载器通过单元测试
- [ ] 环境变量自动映射工作正常
- [ ] 配置 API (`GET/PUT /config`) 实现完成
- [ ] `docs/standards/coding-standards.md` Section 14 添加
- [ ] `docs/product/setup/configuration.md` 完善
- [ ] 集成测试通过
- [ ] Code Review 通过

---

**Created**: 2025-12-28  
**Updated**: 2025-12-28  
**Author**: Winston (Architect Agent)
