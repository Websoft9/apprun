# 配置中心设计与实现

## 概述

apprun的配置中心采用**混合策略**（文件系统 + 数据库），结合Viper和Ent，实现灵活、可扩展的配置管理。支持静态默认配置和动态可修改配置，通过API提供查询和部分修改功能。

## 架构

- **文件系统**：存储默认/静态配置（YAML），作为fallback。
- **数据库**：存储动态配置（Ent表），支持API修改。
- **优先级**：DB > 文件 > 默认值。
- **反射机制**：基于Go结构体标签自动处理配置项。

## 核心组件

- **Viper**：配置加载、环境变量、默认值。
- **Ent**：数据库ORM，静态schema（configitem表）。
- **反射**：遍历结构体，提取`db`标签决定存储策略。

## 结构体定义（types.go）
```go
type Config struct {
    App struct {
        Name    string `validate:"required" default:"apprun" db:"false"`
        Version string `validate:"required" default:"1.0.0" db:"false"`
    }
    Database struct {
        Host string `validate:"required" db:"false"`
        Port int    `validate:"min=1,max=65535" default:"5432" db:"false"`
    }
    POC struct {
        Enabled bool   `default:"true" db:"true"`
        APIKey  string `validate:"min=10" db:"true"`
    }
}
```
- **validate**：校验规则。
- **default**：默认值。
- **db**：是否存储DB（true=可修改）。

## Ent Schema（静态不变）
```go
type Configitem struct {
    ent.Schema
}
func (Configitem) Fields() []ent.Field {
    return []ent.Field{
        field.String("key").Unique(),
        field.String("value"),
        field.Bool("is_dynamic").Default(false),
    }
}
```

## API设计
- **GET /config**：返回所有配置项（JSON数组），并标记哪些项是可更改的（`db:"true"`，即存储到数据库）。
  - 响应示例：
    ```json
    [
      {"path": "app.name", "value": "apprun", "dbStorable": false},
      {"path": "poc.enabled", "value": true, "dbStorable": true},
      {"path": "poc.apikey", "value": "secret123", "dbStorable": true}
    ]
    ```
- **PUT /config**：修改`db:"true"`的项，存储到DB。支持批量修改。
  - 请求体：`{"poc.enabled": false, "poc.apikey": "newsecret"}`
  - 响应：更新后的配置（数组格式，同GET）。
  - 错误处理：
    - 403：尝试修改`db:"false"`的项。
    - 400：校验失败（如值类型错误、违反validate规则）。
    - 500：DB写入失败，回滚内存修改，返回错误信息。

## Ent与Viper集成

- **PUT接口**：通过Ent存储修改值到数据库，确保持久化。支持事务，失败时回滚。
- **GET接口**：从内存（Viper）获取，无需关注来源（文件或DB），统一接口。
- **DB作为Viper外部配置源**：实现自定义`RemoteProvider`，从DB加载配置，作为Viper的外部源。
  ```go
  // 自定义DBProvider实现
  type DBProvider struct {
      client *ent.Client
  }
  func (p *DBProvider) Get(rp viper.RemoteProvider) ([]byte, error) {
      items, err := p.client.Configitem.Query().Where(configitem.IsDynamic(true)).All(context.Background())
      if err != nil {
          return nil, err
      }
      configMap := make(map[string]interface{})
      for _, item := range items {
          configMap[item.Key] = item.Value
      }
      return json.Marshal(configMap)  // 返回JSON格式
  }
  // 使用
  viper.AddRemoteProvider("db", "config", &DBProvider{client: entClient})
  viper.ReadRemoteConfig()  // 从DB加载，覆盖文件配置
  ```
  - **优势**：DB配置统一为Viper源，支持热重载（通过`WatchRemoteConfig`）。
  - **错误处理**：DB不可用时，降级使用文件配置。

## 实现流程
1. **加载**：
   - Viper加载文件（default.yaml → 领域文件（动态扫描，按字母排序）→ conf_d/*.yaml）。
   - 从DB加载动态配置（通过RemoteProvider），覆盖文件配置。
   - 优先级：DB > 文件 > 默认值（SetDefault）。
   - **默认值处理**：默认值在文件和DB都缺失时生效，通过`viper.SetDefault`设置。
   - **领域文件动态扫描**：扫描config/目录下的所有.yaml文件（排除default.yaml和conf_d/），按文件名字母排序加载，确保一致顺序。
2. **校验**：
   - LoadConfig时：使用validator库校验所有字段，不符合规则则启动失败。
   - PUT时：校验修改值，不符合则返回400错误。
3. **反射**：遍历结构体，提取字段和标签。
4. **API**：
   - GET查询：反射遍历，返回所有字段及`dbStorable`标记。
   - PUT修改：检查`db`标签，更新Viper内存，写入DB（事务）。
5. **持久化**：修改写入DB，重启时从DB加载。
6. **错误处理与回滚**：
   - DB写入失败：回滚Viper内存中的修改，返回500错误。
   - DB加载失败：降级使用文件配置，记录警告日志。

## 使用示例
- **查询**：`GET /config` → 
  ```json
  [
    {"path": "app.name", "value": "apprun", "dbStorable": false},
    {"path": "poc.enabled", "value": true, "dbStorable": true}
  ]
  ```
- **修改**：`PUT /config` body `{"poc.enabled": false}` → DB更新，内存刷新，返回更新后的配置数组。

## 性能优化
- **DB连接池**：Ent使用连接池，减少查询延迟。
- **批量操作**：PUT支持批量修改，减少DB事务次数。

## 优势
- **动态**：反射自动适应新配置项，添加字段无需改API代码。
- **安全**：标签控制可修改项，防止敏感配置被修改。
- **轻量**：混合存储，无复杂依赖，适合BaaS轻量级部署。
- **可靠**：事务回滚、错误降级，确保配置一致性。

## 注意事项
- **校验规则**：validate标签在LoadConfig和PUT时执行，确保配置始终有效。
- **默认值冲突**：DB中存储的值始终覆盖默认值和文件配置。
- **性能**：反射遍历在大配置项下性能可控。
- **安全性**：建议为PUT接口添加认证（JWT/API Key），防止未授权修改。

这为apprun提供了强大的配置中心，支持BaaS的扩展需求。