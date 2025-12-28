# 配置指南

## 配置优先级

从高到低：**环境变量 > 数据库 > 用户配置文件 > 领域配置 > 默认配置 > 程序默认值**

## 环境变量配置

### 命名规则

直接使用大写字母加下划线，无需前缀：

```bash
export DATABASE_HOST=prod-db.example.com
export DATABASE_PORT=5432
export APP_NAME=myapp
```

### Docker Compose 示例

```yaml
services:
  app:
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - APP_NAME=apprun
```

## 配置文件

### 默认配置 (`config/default.yaml`)

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  dbname: apprun

app:
  name: apprun
  version: 1.0.0
```

### 用户自定义配置 (`config/conf_d/custom.yaml`)

创建 `config/conf_d/` 目录并添加配置文件，按字母顺序加载：

```yaml
# config/conf_d/01-custom.yaml
app:
  name: my-custom-app
```

## 数据库动态配置

### 查询配置 API

```bash
curl http://localhost:8080/config
```

返回示例：

```json
[
  {"path": "app.name", "value": "apprun", "dbStorable": true},
  {"path": "database.host", "value": "localhost", "dbStorable": false}
]
```

### 修改配置 API

```bash
curl -X PUT http://localhost:8080/config \
  -H "Content-Type: application/json" \
  -d '{"app.name": "new-app-name"}'
```

**注意**: 仅 `dbStorable: true` 的配置可通过 API 修改。

## 常见配置项

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|---------|--------|------|
| `database.host` | `DATABASE_HOST` | `localhost` | 数据库地址 |
| `database.port` | `DATABASE_PORT` | `5432` | 数据库端口 |
| `app.name` | `APP_NAME` | `apprun` | 应用名称 |

## 安全建议

- 敏感信息（密码、密钥）使用环境变量
- 生产环境不要提交 `.env` 文件到 Git
- 定期轮换数据库密码和 API 密钥

