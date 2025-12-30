# Logger Package

统一日志接口（Anti-Corruption Layer），隔离第三方日志库依赖。

## 特性

- 🔌 **防腐层设计**：隔离 zap 等第三方库，可无缝切换实现
- 📊 **结构化日志**：支持 key-value 字段
- 🆔 **自动 request_id**：从 chi middleware 自动提取
- ⚙️ **配置驱动**：支持日志级别和多目标输出
- 🧪 **易于测试**：提供 NopLogger 用于测试

## 快速开始

### 基本使用

```go
package main

import (
	"apprun/pkg/logger"
)

func main() {
	// 配置 logger
	cfg := logger.Config{
		Level: logger.LevelInfo,
		Output: logger.OutputConfig{
			Targets: []string{"stdout"},
		},
	}

	// 初始化
	log, err := logger.NewZapLogger(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close() // 确保资源清理
	logger.SetLogger(log)

	// 使用全局 logger
	logger.Info("Server started", logger.Field{"port", 8080})
	logger.Error("Failed to connect", logger.Field{"error", "timeout"})
}
```

### HTTP Handler 中使用（自动注入 request_id）

```go
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// 从 context 自动提取 request_id
	log := logger.L().WithContext(r.Context())
	
	log.Info("Processing request", 
		logger.Field{"method", r.Method},
		logger.Field{"path", r.URL.Path})
	
	// 业务逻辑...
	
	log.Info("Request completed")
}
```

### 固定字段（服务标识）

```go
// 为整个服务模块添加固定字段
serviceLog := logger.L().With(logger.Field{"service", "user-service"})

serviceLog.Info("User created", logger.Field{"user_id", 123})
serviceLog.Info("User updated", logger.Field{"user_id", 123})
// 所有日志都会带上 service="user-service"
```

## API 文档

### Logger 接口

```go
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}
```

### 日志级别

| 级别 | 用途 | 示例 |
|------|------|------|
| `Debug` | 开发调试信息 | `logger.Debug("Cache hit", logger.Field{"key", cacheKey})` |
| `Info` | 常规操作记录 | `logger.Info("User logged in", logger.Field{"user_id", 123})` |
| `Warn` | 警告（不影响功能） | `logger.Warn("Cache miss", logger.Field{"key", cacheKey})` |
| `Error` | 错误（影响功能） | `logger.Error("DB query failed", logger.Field{"error", err})` |
| `Fatal` | 致命错误（程序退出） | `logger.Fatal("Cannot start server", logger.Field{"error", err})` |

### 配置

```go
type Config struct {
	Level  Level        // 日志级别
	Output OutputConfig // 输出配置
}

type OutputConfig struct {
	Targets []string // 输出目标列表
}
```

**支持的输出目标**：
- `"stdout"` - 标准输出
- `"stderr"` - 标准错误
- `"file:/path/to/file.log"` - 文件输出

**多目标输出示例**：
```go
cfg := logger.Config{
	Level: logger.LevelInfo,
	Output: logger.OutputConfig{
		Targets: []string{"stdout", "file:/var/log/app.log"},
	},
}
```

## 最佳实践

### 1. 生产环境配置

```go
// 生产：Info 级别，输出到 stdout（容器环境）
cfg := logger.Config{
	Level: logger.LevelInfo,
	Output: logger.OutputConfig{
		Targets: []string{"stdout"},
	},
}
```

### 2. 开发环境配置

```go
// 开发：Debug 级别，输出到 stdout
cfg := logger.Config{
	Level: logger.LevelDebug,
	Output: logger.OutputConfig{
		Targets: []string{"stdout"},
	},
}
```

### 3. 日志级别选择指南

- ❌ **避免滥用 Debug**：生产环境应禁用，避免高噪声
- ⚠️ **慎用 Fatal**：仅用于启动期或不可恢复错误，业务逻辑不应使用
- 🔒 **敏感信息处理**：密码、token、隐私字段必须脱敏或不记录

### 4. 结构化字段建议

```go
// ✅ Good: 使用结构化字段
logger.Info("User action", 
	logger.Field{"user_id", userID},
	logger.Field{"action", "login"},
	logger.Field{"ip", clientIP})

// ❌ Bad: 字符串拼接
logger.Info(fmt.Sprintf("User %d logged in from %s", userID, clientIP))
```

## 测试

### 在测试中关闭日志输出

```go
func TestMyFunction(t *testing.T) {
	// 使用 NopLogger 静默日志
	logger.SetLogger(&logger.NopLogger{})
	
	// 测试代码...
}
```

### Mock Logger

```go
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	m.logs = append(m.logs, msg)
}

// ... 实现其他方法

func TestWithMock(t *testing.T) {
	mock := &MockLogger{}
	logger.SetLogger(mock)
	
	// 执行业务逻辑
	MyBusinessLogic()
	
	// 验证日志
	if len(mock.logs) == 0 {
		t.Error("Expected log output")
	}
}
```

## 与 pkg/response 集成

如需将 Story 2 的 `pkg/response` 迁移到使用此 logger：

```go
import "apprun/pkg/logger"

// 替换原有的 zap 调用
logger.Error("Failed to encode response", 
	logger.Field{"error", err},
	logger.Field{"status", statusCode})
```

## 架构说明

此包是 **Anti-Corruption Layer**（防腐层）设计，业务代码依赖 `Logger` 接口而非具体实现：

```
业务代码 → Logger 接口 ← zapLogger (当前实现)
                        ← zerologLogger (未来可选)
                        ← logrusLogger (未来可选)
```

**优势**：
- 隔离依赖：可切换到 zerolog、logrus 等
- 简化测试：注入 NopLogger 避免 I/O
- 统一接口：降低学习成本
- 可扩展：支持钩子、字段过滤等自定义

## License

Copyright © 2025 Websoft9
