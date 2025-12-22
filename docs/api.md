# API Spec

## 按 “领域 / 职责” 拆分 API 路径

将 CRUD 和操作类 API 拆分为不同的路径前缀，明确归属，保持领域模型纯粹：
```
# 1. 核心领域：服务配置的CRUD（静态数据）→ 归属于“配置域”
GET    /api/v1/services/{id}          # 查询服务配置（CRUD-查）
PUT    /api/v1/services/{id}          # 更新服务配置（CRUD-改）
POST   /api/v1/services               # 创建服务配置（CRUD-增）
DELETE /api/v1/services/{id}          # 删除服务配置（CRUD-删）

# 2. 操作域：服务的行为操作 → 归属于“操作/运维域”
POST   /api/v1/services/{id}/actions/restart  # 重启服务（操作类）
POST   /api/v1/services/{id}/actions/stop     # 停止服务（操作类）
POST   /api/v1/services/{id}/actions/upgrade  # 升级服务（操作类）
```

核心设计逻辑：

- 用/actions后缀标识 “操作类 API”，与 CRUD 路径形成明确区分；
- 操作类 API 统一用POST（符合 REST 语义：POST 用于 “触发有副作用的动作”）；
- 核心领域模型（如ServiceConfig结构体）仅承载配置数据，操作逻辑封装在独立的 “操作服务” 中（如ServiceOperationService）。