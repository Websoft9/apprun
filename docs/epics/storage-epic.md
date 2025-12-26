# Epic: 文件存储服务
# apprun BaaS Platform

**关联 PRD**: [FR-STORAGE-001](../prd.md#26-文件存储服务)  
**负责人**: Architect Agent  
**状态**: Planning  
**优先级**: P0 (必需)  
**预估工作量**: 2-3 周

---

## 1. Epic 概述

### 1.1 业务目标

提供统一的文件存储服务，支持文件上传、下载、文件夹管理，并可切换存储后端（本地/S3）。

### 1.2 核心价值

- 开发者无需关心底层存储实现
- 支持大文件流式传输
- 文件夹虚拟化管理
- 存储配额控制

### 1.3 验收标准

- [ ] 文件可上传和下载
- [ ] 文件夹结构正确管理
- [ ] 支持本地存储后端
- [ ] 存储后端可切换（为 S3 预留接口）
- [ ] 单文件上传 < 100MB，响应时间 < 5s
- [ ] 存储配额限制生效

---

## 2. 技术规范

> 📖 **通用规范参考**：[API 设计规范](../standards/api-design.md) | [编码规范](../standards/coding-standards.md)

### 2.1 架构设计

#### 存储抽象层
```
Handler → Service → FileStorage Interface
                         ↓
                    LocalStorage / S3Storage
```

#### 文件路径设计
- **虚拟路径**: `/project-1/docs/file.pdf` (用户视角)
- **物理路径**: `/var/apprun/storage/proj-123/abc-def-uuid.pdf` (实际存储)
- **数据库记录**: 虚拟路径与物理路径的映射

### 2.2 API 端点

| 端点 | 方法 | 功能 | 认证 |
|-----|------|------|------|
| `/api/v1/storage/upload` | POST | 上传单文件 | JWT |
| `/api/v1/storage/files` | GET | 列出文件 | JWT |
| `/api/v1/storage/files/{id}/download` | GET | 下载文件 | JWT |
| `/api/v1/storage/files/{id}` | DELETE | 删除文件 | JWT |
| `/api/v1/storage/folders` | POST | 创建文件夹 | JWT |
| `/api/v1/storage/folders/tree` | GET | 获取文件夹树 | JWT |

#### 示例：上传文件

**请求**：
```http
POST /api/v1/storage/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
project_id: proj-123
folder_path: /docs
```

**响应**：
```json
{
  "success": true,
  "code": 201,
  "data": {
    "file_id": "file-456",
    "name": "document.pdf",
    "path": "/docs/document.pdf",
    "size": 1024000,
    "mime_type": "application/pdf",
    "url": "/api/v1/storage/files/file-456/download",
    "created_at": "2025-12-26T10:00:00Z"
  }
}
```

### 2.3 数据模型

#### 文件表（Ent Schema）
```go
// ent/schema/file.go
func (File) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("id").StructTag(`json:"file_id"`),
        field.String("project_id").StorageKey("project_id").StructTag(`json:"project_id"`),
        field.String("name").StorageKey("name").StructTag(`json:"name"`),
        field.String("path").StorageKey("path").StructTag(`json:"path"`),
        field.Int64("size").StorageKey("size").StructTag(`json:"size"`),
        field.String("mime_type").StorageKey("mime_type").StructTag(`json:"mime_type"`),
        field.String("storage_key").StorageKey("storage_key").StructTag(`json:"-"`), // 隐藏
        field.Time("created_at").StorageKey("created_at").StructTag(`json:"created_at"`),
    }
}
```

#### 文件夹表（Ent Schema）
```go
// ent/schema/folder.go
func (Folder) Fields() []ent.Field {
    return []ent.Field{
        field.String("id").StorageKey("id").StructTag(`json:"folder_id"`),
        field.String("project_id").StorageKey("project_id").StructTag(`json:"project_id"`),
        field.String("path").StorageKey("path").StructTag(`json:"path"`),
        field.String("parent_id").Optional().StorageKey("parent_id").StructTag(`json:"parent_id,omitempty"`),
    }
}
```

### 2.4 存储后端接口

```go
// internal/storage/interface.go
type FileStorage interface {
    Upload(ctx context.Context, path string, reader io.Reader) error
    Download(ctx context.Context, path string) (io.ReadCloser, error)
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
}

// 本地存储实现（基于 afero）
type LocalStorage struct {
    fs      afero.Fs
    baseDir string
}

// S3 存储实现（预留）
type S3Storage struct {
    client *s3.Client
    bucket string
}
```

### 2.5 文件类型和大小限制

#### 允许的文件类型（白名单）
```yaml
# config/storage.yaml
storage:
  allowed_types:
    - "image/jpeg"
    - "image/png"
    - "image/gif"
    - "application/pdf"
    - "text/plain"
    - "application/json"
```

#### 大小限制
```yaml
storage:
  limits:
    max_file_size: 104857600      # 100MB
    max_project_quota: 1073741824 # 1GB (免费用户)
```

### 2.6 权限控制

| 操作 | 项目角色 | 说明 |
|-----|---------|------|
| 上传文件 | member+ | 项目成员及以上 |
| 下载文件 | viewer+ | 查看者及以上 |
| 删除文件 | member+ | 仅文件上传者或管理员 |
| 创建文件夹 | member+ | 项目成员及以上 |

### 2.7 配置

```bash
# 环境变量
STORAGE_BACKEND=local                  # "local" or "s3"
STORAGE_LOCAL_BASE_DIR=/var/apprun/storage
STORAGE_MAX_FILE_SIZE=104857600
STORAGE_MAX_PROJECT_QUOTA=1073741824
```

---

## 3. Stories 拆分

### Story 1: 存储后端抽象层
**优先级**: P0  
**工作量**: 2 天
- [ ] 定义 FileStorage 接口
- [ ] 实现 LocalStorage（基于 afero）
- [ ] 编写存储后端单元测试

### Story 2: 文件上传功能
**优先级**: P0  
**工作量**: 3 天
- [ ] 实现文件上传 Handler
- [ ] 文件类型验证
- [ ] 文件大小限制
- [ ] 配额检查
- [ ] 数据库记录文件元数据
- [ ] 编写上传测试

### Story 3: 文件下载功能
**优先级**: P0  
**工作量**: 2 天
- [ ] 实现文件下载 Handler
- [ ] 流式传输支持
- [ ] 正确设置 Content-Type 和 Content-Disposition
- [ ] 编写下载测试

### Story 4: 文件列表与删除
**优先级**: P0  
**工作量**: 2 天
- [ ] 实现文件列表接口（支持分页）
- [ ] 实现文件删除接口
- [ ] 权限验证（仅上传者或管理员可删除）
- [ ] 编写测试

### Story 5: 文件夹管理
**优先级**: P1  
**工作量**: 3 天
- [ ] 实现文件夹创建
- [ ] 实现文件夹树形结构查询
- [ ] 文件夹路径验证
- [ ] 编写文件夹测试

### Story 6: 存储配额管理
**优先级**: P1  
**工作量**: 2 天
- [ ] 实现项目存储用量统计
- [ ] 上传时检查配额
- [ ] 配额告警（80% 使用率）
- [ ] 编写配额测试

---

## 4. 依赖关系

### 技术依赖
- afero (虚拟文件系统)
- go-chi/chi (路由)
- Ent ORM (数据库)

### 模块依赖
- 认证模块（权限验证）
- 配置模块（存储配置）

### 外部依赖
- 文件系统（本地存储）
- PostgreSQL（元数据）

---

## 5. 风险与挑战

| 风险 | 影响 | 缓解措施 |
|-----|------|---------|
| 大文件上传超时 | 中 | 使用流式传输，增加超时配置 |
| 磁盘空间不足 | 高 | 监控磁盘使用率，实施配额控制 |
| 并发上传冲突 | 中 | 使用唯一文件名（UUID） |
| 文件类型伪造 | 中 | 验证文件 MIME 类型和文件头 |

---

## 6. 测试策略

### 单元测试
- 文件上传逻辑
- 文件下载逻辑
- 配额计算
- 文件类型验证

### 集成测试
- 完整上传下载流程
- 文件夹层级结构
- 权限验证场景

### 性能测试
- 10MB 文件上传 < 3s
- 100MB 文件上传 < 30s
- 并发上传 10 个文件

---

## 7. 监控指标

- `storage_upload_total` - 上传总数
- `storage_upload_bytes_total` - 上传字节数
- `storage_quota_usage_bytes` - 配额使用量
- `storage_upload_duration_seconds` - 上传耗时

---

## 附录

### A. 错误码定义

| 错误码 | HTTP 状态码 | 说明 |
|--------|------------|------|
| `STORAGE_FILE_NOT_FOUND` | 404 | 文件不存在 |
| `STORAGE_QUOTA_EXCEEDED` | 400 | 存储配额超限 |
| `STORAGE_FILE_TOO_LARGE` | 400 | 文件过大 |
| `STORAGE_TYPE_NOT_ALLOWED` | 400 | 文件类型不允许 |

### B. 相关文档

- [PRD - 文件存储](../prd.md#26-文件存储服务)
- [API 设计规范](../standards/api-design.md)

---

**文档维护**: Winston (Architect Agent)  
**最后更新**: 2025-12-26
