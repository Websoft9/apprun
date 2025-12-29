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

### 1. 配置优先级实现（6层）
- [ ] 实现配置优先级（从高到低）：
  1. 环境变量（最高优先级）
  2. 数据库配置（`configitems` 表）
  3. 用户配置目录（`config/conf_d/*.yaml`，按字母序）
  4. 专用配置文件（`config/database.yaml`, `config/server.yaml`，按字母序）
  5. 基础配置文件（`config/default.yaml`）
  6. 结构体 tag 默认值（`default:"value"`，最低优先级）
- [ ] 通过 `db:"false"` tag 控制配置项不可存储到数据库（如 `database.*`）

> 覆盖规则：高优先级覆盖低优先级，同级文件按字母序加载（后覆盖前）

### 2. 结构体 Tag 支持
- [ ] 支持 `default` tag：自动设置默认值（`default:"apprun"`）
- [ ] 支持 `db` tag：控制配置可否存储到数据库（`db:"false"` 禁止存储）
- [ ] 支持 `validate` tag：自动校验配置值（`validate:"required,min=1"`）
- [ ] 使用反射自动处理 tag（启动时一次性遍历）

### 3. 环境变量自动映射
- [ ] 无环境变量前缀
- [ ] 映射规则：`database.host` → `DATABASE_HOST`（`.` → `_`，全大写）
- [ ] 使用 Viper 自动映射，无需手动注册

### 4. 模块化设计
- [ ] `internal/config/` - 唯一配置结构体定义（带 tag）
- [ ] `modules/config/` - 所有配置逻辑（Loader、Repository、Service、Handler）
- [ ] Loader 通过 ConfigProvider 接口获取数据库配置（解耦）
- [ ] Repository 实现 ConfigProvider 接口（防腐层，隔离 Ent）
- [ ] 反射处理 tag（启动时遍历，运行时无开销）

### 5. API 接口
- [ ] `GET /api/config` - 返回所有配置项（含 `dbStorable` 元数据）
- [ ] `PUT /api/config` - 批量更新配置（带 `db` tag 验证和事务）
- [ ] 自动拒绝修改 `db:"false"` 的配置项（403 Forbidden）

### 6. 测试验证
- [ ] 单元测试通过（Loader、Service、Repository）
- [ ] 集成测试通过（API 端到端）
- [ ] 配置优先级验证通过

---

## 📦 Deliverables

### 1. 基础设施层（Internal）

**目录**: `core/internal/config/`

**文件**:
- `types.go` - **唯一配置结构体定义**（Config, AppConfig, DatabaseConfig, ServerConfig）
  - 支持 `default` tag：默认值
  - 支持 `db` tag：控制是否可存储到数据库（`db:"false"` 禁止）
  - 支持 `validate` tag：配置验证规则

**职责**: 全局配置结构体定义（单一来源），通过 tag 声明配置元数据

**示例**:
```go
type Config struct {
    App      AppConfig      `yaml:"app"`
    Database DatabaseConfig `yaml:"database" db:"false"` // 不可存DB
}

type AppConfig struct {
    Name    string `yaml:"name" default:"apprun" db:"false"`
    Theme   string `yaml:"theme" default:"light" db:"true"` // 可存DB
    Timeout int    `yaml:"timeout" default:"30" validate:"min=1,max=300"`
}
```

---

### 2. 配置模块（Modules）

**目录**: `core/modules/config/`

**文件**:
- `types.go` - ConfigProvider 接口 + API 模型（ConfigItem, UpdateConfigRequest, ConfigResponse）
- `loader.go` - 配置加载器（6层优先级，反射处理 tag，依赖 ConfigProvider 接口）
- `repository.go` - 数据访问层（实现 ConfigProvider 接口，防腐层）
- `service.go` - 业务逻辑（反射验证 `db` tag，配置校验，事务管理）
- `handler.go` - HTTP 接口（GET/PUT /api/config）

**职责**: 启动时加载配置 + 运行时配置管理（自动处理 tag 元数据）

---

### 3. 数据模型

**文件**: `core/ent/schema/configitem.go`

**字段**: key (unique), value, is_dynamic, created_at, updated_at

---

### 4. 测试

**单元测试**:
- `core/modules/config/loader_test.go` - 配置加载逻辑
- `core/modules/config/service_test.go` - 业务逻辑验证
- `core/modules/config/repository_test.go` - 数据访问层测试

**集成测试**:
- `tests/integration/config/test-priority.sh` - 配置优先级验证
- `tests/integration/config/test-api.sh` - API 端到端测试

---

### 5. 文档

**开发者规范**: `docs/standards/coding-standards.md` Section 14
- 配置优先级说明
- 环境变量映射规则
- 模块化架构说明

**用户指南**: `docs/product/setup/configuration.md`
- 环境变量使用示例
- 配置文件说明
- API 使用方法

---

## 🔧 Technical Design

### 架构总览

```
core/
├── internal/config/
│   └── types.go              # 👑 唯一配置结构体（带 tag）
│
├── modules/config/
│   ├── types.go              # ConfigProvider 接口 + API 模型
│   ├── loader.go             # 配置加载器（反射处理 tag）
│   ├── repository.go         # 数据访问（防腐层）
│   ├── service.go            # 业务逻辑（tag 验证）
│   └── handler.go            # HTTP 接口
│
└── ent/schema/
    └── configitem.go         # Ent Schema (key, value, is_dynamic)
```

**核心原则**:
- 结构体 tag 声明配置元数据（`default`, `db`, `validate`）
- 反射自动处理 tag（启动时一次性，无运行时开销）
- 减少硬编码，添加新配置无需修改业务逻辑

---

### 配置加载流程（6层优先级）

```
Tag 默认值 → default.yaml → 专用文件 → conf_d/ → 数据库 → 环境变量
```

**详细说明**:
1. **Tag 默认值**: 反射读取 `default:"value"` tag
2. **基础配置**: `config/default.yaml`
3. **专用配置**: `config/database.yaml`, `config/server.yaml`（按字母序）
4. **用户配置**: `config/conf_d/*.yaml`（按字母序）
5. **数据库配置**: `configitems` 表（仅 `db:"true"` 的字段）
6. **环境变量**: `DATABASE_HOST`（最高优先级，自动映射）

**Loader 实现**:
```go
func LoadGlobalConfig(provider ConfigProvider) (*Config, error) {
    cfg := &Config{}
    
    // 1. 反射设置 tag 默认值
    setDefaultsByTag(cfg)
    
    // 2-4. Viper 加载文件配置
    viper.SetConfigName("default")
    viper.ReadInConfig()
    viper.Unmarshal(cfg)
    
    // 5. 从数据库加载（仅 db:"true" 字段）
    dbConfigs, _ := provider.GetAll()
    applyDBConfigsByTag(cfg, dbConfigs) // 反射检查 db tag
    
    // 6. 环境变量自动覆盖（Viper 自动绑定）
    
    // 7. 验证配置（读取 validate tag）
    validate.Struct(cfg)
    
    return cfg, nil
}
```

---

### Tag 控制机制

#### **1. `db` Tag - 控制数据库存储**

```go
// internal/config/types.go
type Config struct {
    Database DatabaseConfig `yaml:"database" db:"false"` // 不可存DB
    App      AppConfig      `yaml:"app"`
}

type AppConfig struct {
    Name  string `yaml:"name" db:"false"`  // 静态配置
    Theme string `yaml:"theme" db:"true"`   // 动态配置（可运行时修改）
}
```

**Service 层自动验证**:
```go
// modules/config/service.go
func (s *Service) UpdateBatch(updates map[string]string) error {
    for key := range updates {
        if !isDBStorableByTag(key) { // 反射检查 db tag
            return fmt.Errorf("config '%s' cannot be stored in database", key)
        }
    }
    return s.repo.SetBatch(updates)
}
```

#### **2. `default` Tag - 默认值**

```go
type AppConfig struct {
    Timeout int `yaml:"timeout" default:"30"` // 启动时自动设置
}
```

#### **3. `validate` Tag - 配置验证**

```go
type DatabaseConfig struct {
    Port int `yaml:"port" default:"5432" validate:"min=1,max=65535"`
}
```

### API 接口设计

**GET /api/config** - 返回所有配置项（含元数据）

```json
[
  {
    "path": "database.host",
    "value": "localhost",
    "dbStorable": false,
    "source": "file"
  },
  {
    "path": "app.theme",
    "value": "dark",
    "dbStorable": true,
    "source": "database"
  }
]
```

**PUT /api/config** - 批量更新配置

```bash
curl -X PUT http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{"app.theme": "light", "app.timeout": "60"}'
```

**自动验证**:
- ✅ 反射检查 `db` tag（拒绝 `db:"false"` 配置）
- ✅ 使用 `validate` tag 校验值
- ✅ 事务保证原子性（全部成功或全部回滚）

---

## 🧪 Testing Strategy

**单元测试**:
- Loader: 6层优先级、tag 默认值、环境变量覆盖
- Service: `db` tag 验证、`validate` tag 校验、事务回滚
- Repository: Ent 查询、防腐层转换

**集成测试**:
- API: GET/PUT 接口、错误处理（403/400/500）
- 优先级: 环境变量 > DB > 文件 > tag 默认值

**验证清单**:
- [ ] Tag 默认值自动设置
- [ ] `db:"false"` 配置禁止通过 API 修改
- [ ] `validate` tag 校验生效
- [ ] 配置优先级正确
- [ ] 事务回滚正常

---

## 📝 Notes

### 设计原则
- **结构体 tag 声明元数据**: 通过 `default`, `db`, `validate` tag 控制配置行为
- **反射自动处理**: 启动时遍历 tag，无需硬编码，运行时无开销
- **单一配置来源**: `internal/config/types.go` 唯一定义结构体
- **高内聚**: 所有配置逻辑集中在 `modules/config/` 模块
- **防腐层**: Repository 隔离 Ent，便于替换持久化技术
- **约定优于配置**: 环境变量自动映射，配置文件按字母序加载

### 反射性能说明
- **启动开销**: ~1-2ms（一次性遍历结构体）
- **运行时**: 零开销（tag 信息缓存后直接使用）
- **结论**: 性能影响可忽略，可维护性提升显著

### 依赖关系
- **依赖**: Story 1 (Docker环境)
- **被依赖**: 所有需要配置管理的 Story

---

## ✅ Definition of Done

- [ ] `core/internal/config/types.go` 定义**唯一**配置结构体（带 `default`, `db`, `validate` tag）
- [ ] `core/modules/config/types.go` 定义 ConfigProvider 接口 + API 模型
- [ ] `core/modules/config/loader.go` 实现加载器（6层优先级，反射处理 tag）
- [ ] `core/modules/config/repository.go` 实现 ConfigProvider 接口（防腐层）
- [ ] `core/modules/config/service.go` 实现业务逻辑（反射验证 `db` tag，配置校验）
- [ ] `core/modules/config/handler.go` 实现 HTTP 接口（GET/PUT /api/config）
- [ ] `core/ent/schema/configitem.go` Ent Schema 定义
- [ ] 单元测试通过（Loader、Service、Repository）
- [ ] 集成测试通过（API、优先级、tag 验证）
- [ ] `docs/standards/coding-standards.md` Section 14 添加配置管理规范
- [ ] `docs/product/setup/configuration.md` 完善用户指南
- [ ] Code Review 通过
- [ ] ✅ 验证配置结构体仅在 `internal/config/types.go` 定义一次
- [ ] ✅ 验证 `db` tag 控制机制生效（无硬编码）

---

**Created**: 2025-12-28  
**Updated**: 2025-12-29  
**Author**: Winston (Architect Agent)
