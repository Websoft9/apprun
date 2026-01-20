# System-Level Test Design
# apprun BaaS Platform

**Generated**: 2026-01-20  
**Test Architect**: Murat (TEA Agent)  
**Project**: apprun  
**Phase**: Phase 4 (Implementation)  
**Status**: Active

---

## 执行摘要

这是apprun BaaS平台的**系统级测试设计文档**，为Go后端API项目定义comprehensive testing strategy。文档包含：

- **风险评估**: 基于6类风险分类的全面风险分析
- **测试覆盖策略**: API测试为主，E2E测试为辅的分层测试架构
- **测试优先级**: P0-P3优先级矩阵，聚焦关键业务路径
- **NFR测试方法**: 安全、性能、可靠性测试方案
- **执行计划**: Sprint 0测试框架搭建路线图

---

## 1. 项目概况

### 1.1 技术栈

| 层级 | 技术 | 用途 |
|------|------|------|
| **语言** | Go 1.24+ | 后端开发 |
| **框架** | Chi Router v5 | HTTP路由 |
| **ORM** | Ent + Atlas | 类型安全ORM + Schema迁移 |
| **数据库** | PostgreSQL 14+ | 主数据库 |
| **缓存** | Redis 7+ | 可选缓存/事件 |
| **认证** | bcrypt + JWT | Go Native Auth |
| **授权** | Casbin v2 | RBAC权限引擎 |
| **测试** | Testify | 单元测试 |

### 1.2 架构模式

**模块化单体架构 (Modular Monolith)**
- 单进程部署，模块清晰分离
- API Gateway (Chi Router + Reverse Proxy)
- 核心模块：Auth, Data, Storage, Functions, Config, Workflow, Events
- 中间件层：RBAC, Logging, Metrics
- 数据访问层：Ent ORM + Repository Pattern

### 1.3 当前状态

**已完成Epic**:
- Epic 1: Infrastructure & Foundation (77% - 14/18 stories完成)
- Epic 3: Configuration Management (100% - 核心功能完成)
- Epic 4: API Documentation (100%)

**进行中Epic**:
- Epic 5: Authentication & Authorization (33% - 1/3 stories完成)
- Epic 2: i18n & Localization (0%)

**关键缺失**:
- ❌ Story 1-8: 测试框架与工具集 (ready-for-dev)
- ❌ 缺少完整的E2E测试基础设施
- ❌ 缺少API集成测试框架
- ❌ 缺少性能测试工具链

---

## 2. 风险评估与分类

### 2.1 风险评分方法

**评分公式**: `Risk Score = Probability × Impact`

**概率评级 (1-3)**:
- **1 (不太可能)**: <10% 概率，边缘情况
- **2 (可能)**: 10-50% 概率，已知场景
- **3 (很可能)**: >50% 概率，常见情况

**影响评级 (1-3)**:
- **1 (轻微)**: 外观问题，存在变通方案，影响少数用户
- **2 (降级)**: 功能受损，变通困难，影响多数用户
- **3 (严重)**: 系统故障，数据丢失，无变通方案，阻塞使用

**风险阈值**:
- **Score 1-2**: 低风险（监控）
- **Score 3-4**: 中风险（计划缓解）
- **Score 6-9**: 高风险（立即缓解）

### 2.2 风险分类矩阵

| Risk ID | 类别 | 描述 | 概率 | 影响 | 分数 | 优先级 | 缓解措施 | 负责人 |
|---------|------|------|------|------|------|--------|----------|--------|
| **TECH风险** |||||||
| R-TECH-001 | TECH | 数据库迁移失败导致schema不一致 | 2 | 3 | **6** | 🔴 HIGH | Atlas schema validation + 回滚测试 | DevOps |
| R-TECH-002 | TECH | Ent ORM代码生成与手写代码冲突 | 2 | 2 | 4 | 🟡 MEDIUM | 单元测试覆盖生成代码 | Backend |
| R-TECH-003 | TECH | Go 1.24+新特性引入的兼容性问题 | 1 | 2 | 2 | 🟢 LOW | CI多版本测试 | Backend |
| R-TECH-004 | TECH | 模块间循环依赖导致编译失败 | 1 | 2 | 2 | 🟢 LOW | 静态分析工具检测 | Architect |
| **SEC风险** |||||||
| R-SEC-001 | SEC | JWT Token泄露导致身份伪造 | 2 | 3 | **6** | 🔴 HIGH | Token短期有效+Refresh机制+安全存储测试 | Security |
| R-SEC-002 | SEC | bcrypt密码哈希配置不当（成本因子过低） | 1 | 3 | 3 | 🟡 MEDIUM | 配置验证测试（cost ≥12） | Security |
| R-SEC-003 | SEC | Casbin RBAC策略绕过（权限检查缺失） | 2 | 3 | **6** | 🔴 HIGH | 权限矩阵全覆盖测试 | Security |
| R-SEC-004 | SEC | SQL注入（虽然Ent提供保护，但自定义查询风险） | 1 | 3 | 3 | 🟡 MEDIUM | 自定义SQL查询审计+测试 | Security |
| R-SEC-005 | SEC | 项目间数据泄露（project_id过滤缺失） | 2 | 3 | **6** | 🔴 HIGH | Project隔离测试（负面测试） | Backend |
| **PERF风险** |||||||
| R-PERF-001 | PERF | Auth API响应时间超过100ms P95 | 2 | 2 | 4 | 🟡 MEDIUM | 性能测试+Redis缓存优化 | Performance |
| R-PERF-002 | PERF | 数据库查询N+1问题导致性能下降 | 3 | 2 | **6** | 🔴 HIGH | Query分析+Eager loading测试 | Backend |
| R-PERF-003 | PERF | 并发登录导致数据库连接池耗尽 | 2 | 3 | **6** | 🔴 HIGH | 负载测试（1000并发用户） | Performance |
| **DATA风险** |||||||
| R-DATA-001 | DATA | 用户删除时级联删除失败（孤儿数据） | 2 | 2 | 4 | 🟡 MEDIUM | 级联删除测试+数据完整性校验 | Backend |
| R-DATA-002 | DATA | 并发创建用户导致邮箱唯一约束冲突 | 2 | 2 | 4 | 🟡 MEDIUM | 并发测试+事务隔离测试 | Backend |
| R-DATA-003 | DATA | 迁移失败后数据不一致（部分表更新） | 1 | 3 | 3 | 🟡 MEDIUM | 迁移事务测试+回滚验证 | DevOps |
| **BUS风险** |||||||
| R-BUS-001 | BUS | 用户注册后无法登录（状态不一致） | 2 | 3 | **6** | 🔴 HIGH | 端到端注册+登录流程测试 | Backend |
| R-BUS-002 | BUS | 项目Owner转移后权限丢失 | 1 | 3 | 3 | 🟡 MEDIUM | 权限转移测试 | Backend |
| R-BUS-003 | BUS | 密码重置链接过期处理不当 | 2 | 2 | 4 | 🟡 MEDIUM | 过期Token测试 | Backend |
| **OPS风险** |||||||
| R-OPS-001 | OPS | Docker Compose启动顺序错误（DB未就绪） | 2 | 2 | 4 | 🟡 MEDIUM | 健康检查+depends_on配置测试 | DevOps |
| R-OPS-002 | OPS | 配置文件缺失导致启动失败 | 2 | 3 | **6** | 🔴 HIGH | 配置验证测试+默认值fallback | DevOps |
| R-OPS-003 | OPS | 日志输出过多导致存储耗尽 | 1 | 2 | 2 | 🟢 LOW | 日志轮转配置+监控 | DevOps |

### 2.3 高优先级风险汇总

**🔴 CRITICAL (Score = 9)**: 0个  
**🔴 HIGH (Score 6-8)**: **10个**

高优先级风险清单：
1. R-TECH-001: 数据库迁移失败
2. R-SEC-001: JWT Token泄露
3. R-SEC-003: RBAC策略绕过
4. R-SEC-005: 项目间数据泄露
5. R-PERF-002: N+1查询问题
6. R-PERF-003: 连接池耗尽
7. R-BUS-001: 注册后无法登录
8. R-OPS-002: 配置缺失

**缓解策略**:
- ✅ 所有Score≥6的风险必须在Sprint 0完成缓解测试
- ✅ 每个风险分配明确负责人和截止日期
- ✅ 缓解测试纳入CI/CD流程

---

## 3. 测试分层策略

### 3.1 测试金字塔（API项目优化版）

```
        ▲
       / E\          E2E Tests (10%)
      /  2 \         - 关键用户流程
     / End \         - 跨系统集成
    /───────\        - Smoke tests
   / API/Int \       API/Integration Tests (50%)
  / egration \       - REST API契约
 /   Tests    \      - Service层逻辑
/─────────────\      - 数据库操作
\ Unit Tests  /      Unit Tests (40%)
 \   (40%)   /       - 业务逻辑
  \         /        - 工具函数
   \───────/         - 数据验证
    \     /
     \   /
      \ /
       V
```

**理由**:
- apprun是**API-first**后端项目，不是UI项目
- **API测试**提供最佳ROI（速度快+覆盖广+稳定）
- **E2E测试**仅限关键流程（避免维护负担）
- **单元测试**覆盖复杂业务逻辑

### 3.2 测试级别决策矩阵

| 测试场景 | 推荐级别 | 理由 |
|----------|---------|------|
| **密码哈希算法** | Unit | 纯函数，无依赖 |
| **JWT Token生成/验证** | Unit + Integration | 算法单元测试 + 中间件集成测试 |
| **用户注册API** | Integration | 涉及数据库+业务逻辑 |
| **登录→访问受保护资源** | E2E | 跨多个端点的完整流程 |
| **RBAC权限检查** | Integration | Casbin引擎+数据库策略 |
| **数据库迁移** | Integration | Schema变更验证 |
| **配置加载** | Unit | 配置解析逻辑 |
| **API响应格式** | Integration | 统一响应包装 |
| **并发安全** | Integration | 数据库事务隔离 |
| **性能SLA** | E2E | 完整环境性能测试 |

### 3.3 避免重复覆盖

**反模式**:
- ❌ 在E2E中测试单个字段验证（应该用Unit）
- ❌ 在Integration中测试第三方库行为（应该用Unit Mock）
- ❌ 在Unit中测试数据库查询（应该用Integration）

**正确做法**:
- ✅ Unit测试：业务逻辑纯函数
- ✅ Integration测试：API端点+数据库操作
- ✅ E2E测试：关键业务流程（注册→登录→操作→登出）

---

## 4. NFR（非功能需求）测试方法

### 4.1 安全测试

#### 4.1.1 认证测试

**测试场景**:
```go
// 测试：密码哈希安全性
func TestPasswordHashingSecurity(t *testing.T) {
    password := "SecurePass123!"
    
    // 验证bcrypt cost ≥ 12
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    assert.NoError(t, err)
    
    // 验证哈希不可逆
    assert.NotContains(t, string(hash), password)
    
    // 验证相同密码产生不同哈希（salt）
    hash2, _ := bcrypt.GenerateFromPassword([]byte(password), 12)
    assert.NotEqual(t, hash, hash2)
}

// 测试：JWT Token过期
func TestJWTTokenExpiration(t *testing.T) {
    // 创建过期Token（-1小时）
    token := generateExpiredToken(time.Now().Add(-1 * time.Hour))
    
    // 验证Token被拒绝
    _, err := validateToken(token)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "token expired")
}
```

#### 4.1.2 授权测试

**RBAC权限矩阵测试**:
```go
// 测试：项目隔离（负面测试）
func TestProjectIsolation(t *testing.T) {
    // 用户A属于Project 1
    userA := createTestUser("userA@example.com")
    projectA := createTestProject("Project A")
    assignUserToProject(userA.ID, projectA.ID, "member")
    
    // 用户B属于Project 2
    userB := createTestUser("userB@example.com")
    projectB := createTestProject("Project B")
    assignUserToProject(userB.ID, projectB.ID, "member")
    
    // 用户A尝试访问Project B的资源
    token := loginAsUser(userA)
    resp := makeAuthenticatedRequest(
        "GET",
        fmt.Sprintf("/api/projects/%s/resources", projectB.ID),
        token,
    )
    
    // 验证：403 Forbidden
    assert.Equal(t, 403, resp.StatusCode)
    assertErrorCode(t, resp, "INSUFFICIENT_PERMISSION")
}

// 测试：权限升级攻击
func TestPrivilegeEscalation(t *testing.T) {
    // Viewer角色尝试执行Admin操作
    viewer := createUserWithRole("viewer@example.com", "viewer")
    token := loginAsUser(viewer)
    
    // 尝试删除项目（需要owner权限）
    resp := makeAuthenticatedRequest(
        "DELETE",
        "/api/projects/123",
        token,
    )
    
    assert.Equal(t, 403, resp.StatusCode)
}
```

#### 4.1.3 安全扫描工具

**工具链**:
- `gosec`: 静态代码安全扫描
- `govulncheck`: Go漏洞检测
- `nancy`: 依赖安全扫描（Sonatype）

**CI集成**:
```yaml
# .github/workflows/security.yml
- name: Run gosec
  run: gosec -fmt json -out gosec-report.json ./...
  
- name: Check vulnerabilities
  run: govulncheck ./...
```

### 4.2 性能测试

#### 4.2.1 SLA目标

| 端点类别 | P95延迟 | P99延迟 | 吞吐量 |
|----------|---------|---------|--------|
| **认证API** | <100ms | <200ms | 500 req/s |
| **数据查询** | <50ms | <100ms | 1000 req/s |
| **数据写入** | <200ms | <500ms | 200 req/s |

#### 4.2.2 性能测试工具

**推荐**: `k6` (Go原生，适合API负载测试)

**测试脚本示例**:
```javascript
// tests/performance/auth-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up
    { duration: '5m', target: 500 },  // Stay at 500
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    'http_req_duration': ['p(95)<100'], // 95% < 100ms
    'http_req_failed': ['rate<0.01'],   // <1% errors
  },
};

export default function () {
  let loginRes = http.post('http://localhost:8080/api/auth/login', JSON.stringify({
    email: `user${__VU}@example.com`,
    password: 'TestPass123!',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });
  
  check(loginRes, {
    'login status 200': (r) => r.status === 200,
    'login duration < 100ms': (r) => r.timings.duration < 100,
  });
  
  sleep(1);
}
```

#### 4.2.3 性能回归检测

**CI集成策略**:
- 每日scheduled run（避免阻塞PR）
- 基线对比：当前 vs 上一次成功build
- 告警阈值：P95延迟增加>20% → 失败

### 4.3 可靠性测试

#### 4.3.1 错误处理测试

**测试场景**:
```go
// 测试：数据库连接失败
func TestDatabaseConnectionFailure(t *testing.T) {
    // 关闭数据库连接
    db.Close()
    
    // 尝试创建用户
    user := CreateUserRequest{
        Email: "test@example.com",
        Password: "Pass123!",
    }
    
    resp := postJSON("/api/auth/register", user)
    
    // 验证：500错误+友好错误消息
    assert.Equal(t, 500, resp.StatusCode)
    assertErrorMessage(t, resp, "Service temporarily unavailable")
    
    // 验证：错误日志记录
    assertLogContains(t, "database connection failed")
}

// 测试：超时处理
func TestRequestTimeout(t *testing.T) {
    // 模拟慢查询（超过5秒）
    mockSlowQuery()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := queryUsers(ctx)
    
    // 验证：超时错误
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "context deadline exceeded")
}
```

#### 4.3.2 重试与降级测试

**Resilience Patterns**:
```go
// 测试：指数退避重试
func TestExponentialBackoffRetry(t *testing.T) {
    attemptCount := 0
    maxRetries := 3
    
    err := retryWithBackoff(func() error {
        attemptCount++
        if attemptCount < maxRetries {
            return errors.New("transient error")
        }
        return nil
    }, maxRetries)
    
    assert.NoError(t, err)
    assert.Equal(t, maxRetries, attemptCount)
}

// 测试：Circuit Breaker
func TestCircuitBreaker(t *testing.T) {
    cb := NewCircuitBreaker(5, 10*time.Second) // 5次失败后开启
    
    // 触发5次失败
    for i := 0; i < 5; i++ {
        cb.Call(func() error { return errors.New("failure") })
    }
    
    // 验证：Circuit Breaker开启
    assert.True(t, cb.IsOpen())
    
    // 第6次调用应该快速失败
    err := cb.Call(func() error { return nil })
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circuit breaker open")
}
```

### 4.4 可维护性测试

#### 4.4.1 代码质量门禁

**Linter配置** (`.golangci.yml`):
```yaml
linters:
  enable:
    - gofmt
    - goimports
    - golint
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - gocritic
    - cyclop       # 圈复杂度检测
    - gocyclo      # 圈复杂度阈值
    - nestif       # 嵌套深度检测

linters-settings:
  cyclop:
    max-complexity: 15  # 单函数最大圈复杂度
  gocyclo:
    min-complexity: 15
  nestif:
    min-complexity: 5   # 最大嵌套层级
```

#### 4.4.2 测试覆盖率目标

| 模块 | 单元测试覆盖率 | 集成测试覆盖率 |
|------|---------------|---------------|
| **Auth** | >90% | >80% |
| **RBAC** | >90% | >90% |
| **Data** | >80% | >70% |
| **Config** | >85% | >60% |
| **Storage** | >70% | >80% |
| **总体目标** | **>80%** | **>70%** |

**CI门禁**:
```yaml
# .github/workflows/test.yml
- name: Run tests with coverage
  run: go test -coverprofile=coverage.out ./...

- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$coverage < 80" | bc -l) )); then
      echo "Coverage $coverage% is below 80% threshold"
      exit 1
    fi
```

---

## 5. 测试场景设计（Epic 5: 认证与授权）

### 5.1 Story 5-1: User Registration

#### 5.1.1 测试场景清单

| 场景ID | 描述 | 测试级别 | 优先级 | 风险关联 |
|--------|------|---------|--------|----------|
| **正向场景** ||||||
| REG-P-001 | 有效数据注册成功 | Integration | P0 | R-BUS-001 |
| REG-P-002 | 密码安全存储（bcrypt验证） | Unit | P0 | R-SEC-002 |
| REG-P-003 | 注册后用户状态正确（is_active=true） | Integration | P1 | R-BUS-001 |
| **负向场景** ||||||
| REG-N-001 | 邮箱格式无效被拒绝 | Unit | P0 | R-DATA-002 |
| REG-N-002 | 密码强度不足被拒绝（<8字符） | Unit | P0 | R-SEC-002 |
| REG-N-003 | 邮箱重复注册被拒绝（409 Conflict） | Integration | P0 | R-DATA-002 |
| REG-N-004 | 缺少必填字段被拒绝（400 Bad Request） | Integration | P1 | - |
| **边界场景** ||||||
| REG-E-001 | 邮箱长度边界（1-255字符） | Unit | P2 | - |
| REG-E-002 | 密码长度边界（8-128字符） | Unit | P2 | - |
| REG-E-003 | 并发注册相同邮箱（唯一约束测试） | Integration | P1 | R-DATA-002 |
| **性能场景** ||||||
| REG-PERF-001 | 100并发注册（P95<200ms） | E2E | P1 | R-PERF-003 |

#### 5.1.2 测试实现示例

**Unit Test** (密码哈希):
```go
// tests/unit/auth/password_test.go
func TestPasswordHashing(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {"valid strong password", "SecurePass123!", false},
        {"too short", "Pass1!", true},
        {"no special char", "Password123", true},
        {"no number", "Password!@#", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            hash, err := HashPassword(tt.password)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.NotEmpty(t, hash)
            
            // 验证哈希cost
            cost, _ := bcrypt.Cost([]byte(hash))
            assert.GreaterOrEqual(t, cost, 12)
        })
    }
}
```

**Integration Test** (注册API):
```go
// tests/integration/auth/register_test.go
func TestUserRegistration(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    router := setupRouter(db)
    
    t.Run("successful registration", func(t *testing.T) {
        reqBody := map[string]string{
            "email":    "newuser@example.com",
            "password": "SecurePass123!",
            "name":     "New User",
        }
        
        resp := postJSON(router, "/api/auth/register", reqBody)
        
        assert.Equal(t, 200, resp.StatusCode)
        
        var result map[string]interface{}
        json.Unmarshal(resp.Body, &result)
        
        assert.True(t, result["success"].(bool))
        assert.NotEmpty(t, result["data"].(map[string]interface{})["user"].(map[string]interface{})["id"])
        
        // 验证数据库记录
        user := getUserByEmail(db, "newuser@example.com")
        assert.NotNil(t, user)
        assert.True(t, user.IsActive)
        assert.NotEqual(t, "SecurePass123!", user.PasswordHash) // 验证密码已哈希
    })
    
    t.Run("duplicate email rejected", func(t *testing.T) {
        // 创建已存在用户
        createTestUser(db, "existing@example.com")
        
        reqBody := map[string]string{
            "email":    "existing@example.com",
            "password": "SecurePass123!",
        }
        
        resp := postJSON(router, "/api/auth/register", reqBody)
        
        assert.Equal(t, 409, resp.StatusCode)
        assertErrorCode(t, resp, "EMAIL_ALREADY_EXISTS")
    })
}
```

### 5.2 Story 5-2: User Login

#### 5.2.1 测试场景清单

| 场景ID | 描述 | 测试级别 | 优先级 | 风险关联 |
|--------|------|---------|--------|----------|
| **正向场景** ||||||
| LOGIN-P-001 | 有效凭证登录成功 | Integration | P0 | R-BUS-001 |
| LOGIN-P-002 | 返回JWT Token（Access + Refresh） | Integration | P0 | R-SEC-001 |
| LOGIN-P-003 | Token包含正确Claims（user_id, email） | Unit | P0 | R-SEC-001 |
| LOGIN-P-004 | 更新last_login_at时间戳 | Integration | P2 | - |
| **负向场景** ||||||
| LOGIN-N-001 | 错误密码被拒绝（401） | Integration | P0 | R-SEC-003 |
| LOGIN-N-002 | 不存在的邮箱被拒绝（401） | Integration | P0 | - |
| LOGIN-N-003 | 空邮箱/密码被拒绝（400） | Integration | P1 | - |
| LOGIN-N-004 | 未激活用户被拒绝（403） | Integration | P1 | - |
| **安全场景** ||||||
| LOGIN-SEC-001 | 暴力破解防护（5次失败后锁定） | Integration | P1 | R-SEC-003 |
| LOGIN-SEC-002 | Token过期后被拒绝 | Integration | P0 | R-SEC-001 |
| LOGIN-SEC-003 | 伪造Token被拒绝 | Unit | P0 | R-SEC-001 |
| **性能场景** ||||||
| LOGIN-PERF-001 | 并发1000登录（P95<100ms） | E2E | P0 | R-PERF-001, R-PERF-003 |

### 5.3 Story 5-3: JWT Middleware

#### 5.3.1 测试场景清单

| 场景ID | 描述 | 测试级别 | 优先级 | 风险关联 |
|--------|------|---------|--------|----------|
| **正向场景** ||||||
| JWT-P-001 | 有效Token通过验证 | Integration | P0 | R-SEC-001 |
| JWT-P-002 | Token解析到Context（user_id） | Integration | P0 | R-SEC-003 |
| JWT-P-003 | Refresh Token刷新Access Token | Integration | P1 | R-SEC-001 |
| **负向场景** ||||||
| JWT-N-001 | 缺少Token被拒绝（401） | Integration | P0 | R-SEC-003 |
| JWT-N-002 | 过期Token被拒绝（401） | Integration | P0 | R-SEC-001 |
| JWT-N-003 | 无效签名Token被拒绝（401） | Integration | P0 | R-SEC-001 |
| JWT-N-004 | 格式错误Token被拒绝（400） | Integration | P1 | - |
| **RBAC场景** ||||||
| JWT-RBAC-001 | Project隔离：访问非成员项目被拒绝 | Integration | P0 | R-SEC-005 |
| JWT-RBAC-002 | 权限检查：Viewer执行Admin操作被拒绝 | Integration | P0 | R-SEC-003 |
| JWT-RBAC-003 | 权限矩阵：所有角色+资源组合测试 | Integration | P0 | R-SEC-003 |

### 5.4 E2E关键流程

#### 5.4.1 场景：完整认证流程

```go
// tests/e2e/auth_flow_test.go
func TestCompleteAuthFlow(t *testing.T) {
    // 准备：清理测试数据
    cleanupTestData(t)
    
    // Step 1: 注册新用户
    registerReq := map[string]string{
        "email":    "e2e@example.com",
        "password": "SecurePass123!",
        "name":     "E2E Test User",
    }
    
    registerResp := apiClient.POST("/api/auth/register", registerReq)
    assert.Equal(t, 200, registerResp.StatusCode)
    
    // Step 2: 登录获取Token
    loginReq := map[string]string{
        "email":    "e2e@example.com",
        "password": "SecurePass123!",
    }
    
    loginResp := apiClient.POST("/api/auth/login", loginReq)
    assert.Equal(t, 200, loginResp.StatusCode)
    
    var loginData map[string]interface{}
    json.Unmarshal(loginResp.Body, &loginData)
    
    accessToken := loginData["data"].(map[string]interface{})["access_token"].(string)
    assert.NotEmpty(t, accessToken)
    
    // Step 3: 使用Token访问受保护资源
    apiClient.SetAuthToken(accessToken)
    
    meResp := apiClient.GET("/api/users/me")
    assert.Equal(t, 200, meResp.StatusCode)
    
    var meData map[string]interface{}
    json.Unmarshal(meResp.Body, &meData)
    
    user := meData["data"].(map[string]interface{})["user"].(map[string]interface{})
    assert.Equal(t, "e2e@example.com", user["email"])
    assert.Equal(t, "E2E Test User", user["name"])
    
    // Step 4: 创建项目（测试RBAC）
    createProjectReq := map[string]string{
        "name": "Test Project",
    }
    
    projectResp := apiClient.POST("/api/projects", createProjectReq)
    assert.Equal(t, 200, projectResp.StatusCode)
    
    // Step 5: 验证项目隔离（负面测试）
    // 创建另一个用户+项目
    otherUser := createTestUserAndProject("other@example.com", "Other Project")
    
    // 尝试访问其他用户的项目
    unauthorizedResp := apiClient.GET(fmt.Sprintf("/api/projects/%s", otherUser.ProjectID))
    assert.Equal(t, 403, unauthorizedResp.StatusCode)
    
    // Step 6: 登出（可选）
    logoutResp := apiClient.POST("/api/auth/logout", nil)
    assert.Equal(t, 200, logoutResp.StatusCode)
    
    // 验证Token失效
    meRespAfterLogout := apiClient.GET("/api/users/me")
    assert.Equal(t, 401, meRespAfterLogout.StatusCode)
}
```

---

## 6. 测试执行策略

### 6.1 测试优先级与执行顺序

**Smoke Tests** (<5分钟):
- 健康检查：`GET /health`
- 基础认证：注册+登录
- 数据库连接验证

**P0 Tests** (<15分钟):
- 所有认证API（注册、登录、Token验证）
- RBAC权限矩阵（所有角色+资源组合）
- 数据库CRUD操作
- 配置加载与验证

**P1 Tests** (<45分钟):
- 错误处理场景
- 边界值测试
- 并发安全测试
- 性能基准测试

**P2/P3 Tests** (<2小时):
- 高级功能测试
- 性能回归测试（k6负载测试）
- 安全扫描（gosec, govulncheck）

### 6.2 CI/CD集成

#### 6.2.1 GitHub Actions工作流

```yaml
# .github/workflows/test.yml
name: Test Suite

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run unit tests
        run: make test-unit
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:14
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run migrations
        run: make db-migrate
      
      - name: Run integration tests
        run: make test-integration
        env:
          DATABASE_URL: postgresql://postgres:postgres@localhost:5432/test
          REDIS_URL: redis://localhost:6379

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Start services
        run: docker-compose -f docker-compose.test.yml up -d
      
      - name: Wait for services
        run: ./scripts/wait-for-services.sh
      
      - name: Run E2E tests
        run: make test-e2e
      
      - name: Cleanup
        if: always()
        run: docker-compose -f docker-compose.test.yml down

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run gosec
        run: |
          go install github.com/securego/gosec/v2/cmd/gosec@latest
          gosec -fmt json -out gosec-report.json ./...
      
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...
```

### 6.3 测试数据管理

#### 6.3.1 Fixture与Factory模式

**Factory示例**:
```go
// tests/fixtures/user_factory.go
type UserFactory struct {
    db *ent.Client
}

func (f *UserFactory) CreateUser(opts ...UserOption) *ent.User {
    // 默认值
    user := &CreateUserRequest{
        Email:    fmt.Sprintf("user-%d@example.com", time.Now().UnixNano()),
        Password: "TestPass123!",
        Name:     "Test User",
        IsActive: true,
    }
    
    // 应用选项覆盖
    for _, opt := range opts {
        opt(user)
    }
    
    // 创建用户
    created, err := f.db.User.Create().
        SetEmail(user.Email).
        SetPasswordHash(HashPassword(user.Password)).
        SetName(user.Name).
        SetIsActive(user.IsActive).
        Save(context.Background())
    
    if err != nil {
        panic(err)
    }
    
    return created
}

// 选项模式
type UserOption func(*CreateUserRequest)

func WithEmail(email string) UserOption {
    return func(u *CreateUserRequest) {
        u.Email = email
    }
}

func WithRole(role string) UserOption {
    return func(u *CreateUserRequest) {
        u.Role = role
    }
}

// 使用示例
func TestUserWithCustomRole(t *testing.T) {
    factory := NewUserFactory(db)
    
    admin := factory.CreateUser(
        WithEmail("admin@example.com"),
        WithRole("admin"),
    )
    
    assert.Equal(t, "admin@example.com", admin.Email)
}
```

#### 6.3.2 数据库清理策略

**每个测试隔离**:
```go
func setupTestDB(t *testing.T) *ent.Client {
    client, err := ent.Open("postgres", "postgresql://localhost/test")
    require.NoError(t, err)
    
    // 运行迁移
    err = client.Schema.Create(context.Background())
    require.NoError(t, err)
    
    // 清理钩子
    t.Cleanup(func() {
        truncateAllTables(client)
        client.Close()
    })
    
    return client
}

func truncateAllTables(client *ent.Client) {
    tables := []string{"users", "projects", "project_members", "casbin_rule"}
    for _, table := range tables {
        client.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
    }
}
```

---

## 7. 质量门禁标准

### 7.1 Gate Decision Criteria

**PASS条件（全部满足）**:
- ✅ P0测试100%通过
- ✅ P1测试≥95%通过
- ✅ 单元测试覆盖率≥80%
- ✅ 集成测试覆盖率≥70%
- ✅ 所有Score≥6风险已缓解
- ✅ 无Critical安全漏洞（gosec, govulncheck）
- ✅ 性能SLA达标（P95<100ms）

**CONCERNS条件（任一满足）**:
- ⚠️ P1测试通过率90-95%
- ⚠️ 覆盖率75-80%
- ⚠️ 存在Medium风险（Score 3-5）有缓解计划

**FAIL条件（任一满足）**:
- ❌ P0测试失败
- ❌ Critical风险（Score 9）未缓解
- ❌ High风险（Score 6-8）无缓解计划
- ❌ 覆盖率<75%
- ❌ 性能SLA严重超标（P95>200ms）

### 7.2 发布检查清单

**Sprint 0完成标准**:
- [ ] 测试框架搭建完成（Story 1-8）
- [ ] 单元测试框架（Testify）
- [ ] 集成测试框架（httptest + Testcontainers）
- [ ] E2E测试框架（Go http client）
- [ ] CI/CD流水线配置
- [ ] 测试数据工厂（Fixture/Factory）
- [ ] 性能测试工具（k6）
- [ ] 安全扫描工具（gosec, govulncheck）
- [ ] 代码覆盖率报告（Codecov集成）
- [ ] 测试文档与最佳实践（README）

**Epic 5完成标准**:
- [ ] 所有认证API测试通过（注册、登录、Token）
- [ ] RBAC权限矩阵全覆盖测试
- [ ] 项目隔离测试（负面测试）
- [ ] 性能测试达标（Auth API P95<100ms）
- [ ] 安全测试通过（Token安全、密码哈希、权限绕过）
- [ ] E2E完整流程测试（注册→登录→操作→登出）

---

## 8. 工具链与基础设施

### 8.1 测试工具矩阵

| 工具 | 用途 | 配置文件 | CI集成 |
|------|------|---------|--------|
| **Testify** | 单元测试断言 | - | ✅ |
| **httptest** | HTTP测试服务器 | - | ✅ |
| **Testcontainers** | 集成测试容器 | - | ✅ |
| **k6** | 性能负载测试 | `tests/performance/*.js` | ⚠️ Scheduled |
| **gosec** | 安全静态扫描 | `.gosec.json` | ✅ |
| **govulncheck** | 漏洞检测 | - | ✅ |
| **golangci-lint** | 代码质量检查 | `.golangci.yml` | ✅ |
| **Codecov** | 覆盖率报告 | `.codecov.yml` | ✅ |

### 8.2 测试环境配置

**环境变量** (`.env.test`):
```bash
# Database
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/apprun_test
DB_MAX_CONN=10
DB_IDLE_CONN=5

# Redis (Optional)
REDIS_URL=redis://localhost:6379/1

# JWT
JWT_SECRET=test-secret-key-DO-NOT-USE-IN-PRODUCTION
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=7d

# Bcrypt
BCRYPT_COST=10  # Lower for faster tests (production: 12)

# Server
SERVER_PORT=8081
LOG_LEVEL=debug

# Feature Flags
ENABLE_RATE_LIMIT=false  # Disable for tests
```

### 8.3 Docker Compose测试环境

**docker-compose.test.yml**:
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_DB: apprun_test
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5433:5432"  # 避免与本地冲突
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  apprun-test:
    build:
      context: .
      dockerfile: docker/Dockerfile
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/apprun_test
      - REDIS_URL=redis://redis:6379/1
    ports:
      - "8081:8080"
    command: ["./bin/apprun", "serve"]
```

---

## 9. Sprint 0实施计划

### 9.1 Story 1-8: 测试框架与工具集

**工作量估算**: 3-5天

#### 阶段1: 基础框架搭建 (1天)

**任务清单**:
- [ ] 创建测试目录结构
  ```
  tests/
  ├── unit/               # 单元测试
  │   ├── auth/
  │   ├── config/
  │   └── utils/
  ├── integration/        # 集成测试
  │   ├── api/
  │   └── db/
  ├── e2e/               # 端到端测试
  │   └── scenarios/
  ├── performance/       # 性能测试
  │   └── k6/
  ├── fixtures/          # 测试数据工厂
  │   ├── user_factory.go
  │   └── project_factory.go
  └── testutils/         # 测试工具函数
      ├── db_helper.go
      ├── http_helper.go
      └── assert_helper.go
  ```

- [ ] 配置Testify + httptest
- [ ] 创建测试数据库设置脚本
- [ ] 编写测试辅助函数（HTTP client, DB setup）

#### 阶段2: CI/CD集成 (1天)

**任务清单**:
- [ ] 配置GitHub Actions工作流
  - unit-tests job
  - integration-tests job
  - e2e-tests job
  - security-scan job
- [ ] 集成Codecov覆盖率报告
- [ ] 配置测试失败通知（Slack/Email）

#### 阶段3: 示例测试编写 (1天)

**任务清单**:
- [ ] 编写认证模块示例测试
  - Unit: 密码哈希测试
  - Integration: 注册API测试
  - E2E: 完整认证流程测试
- [ ] 编写配置模块示例测试
- [ ] 验证所有测试在CI中通过

#### 阶段4: 性能与安全测试工具 (1天)

**任务清单**:
- [ ] 安装并配置k6
- [ ] 编写Auth API性能测试脚本
- [ ] 配置gosec静态扫描
- [ ] 配置govulncheck漏洞检测
- [ ] 集成到CI流水线

#### 阶段5: 文档与培训 (0.5天)

**任务清单**:
- [ ] 编写测试框架README
  - 如何运行测试
  - 如何编写新测试
  - 最佳实践
- [ ] 创建测试模板（复制粘贴即用）
- [ ] 团队培训Session

### 9.2 验收标准

**Definition of Done**:
- ✅ 测试目录结构完整
- ✅ 单元测试示例通过
- ✅ 集成测试示例通过
- ✅ E2E测试示例通过
- ✅ CI/CD流水线成功运行
- ✅ 覆盖率报告生成
- ✅ 性能测试工具配置完成
- ✅ 安全扫描无Critical漏洞
- ✅ 测试文档完整

---

## 10. 下一步行动

### 10.1 立即行动（本周）

1. **启动Story 1-8**: 测试框架搭建
   - 负责人：Backend Team
   - 截止日期：本周五
   - 依赖：无

2. **风险缓解计划**:
   - R-SEC-001, R-SEC-003, R-SEC-005: 认证授权测试（P0）
   - R-PERF-002, R-PERF-003: 性能测试工具配置（P1）
   - R-BUS-001: E2E完整流程测试（P0）

### 10.2 Sprint 1行动（下周）

1. **Epic 5完整测试覆盖**:
   - Story 5-1: 用户注册测试（Unit + Integration + E2E）
   - Story 5-2: 用户登录测试（Unit + Integration + E2E）
   - Story 5-3: JWT中间件测试（Integration）

2. **质量门禁验证**:
   - 运行完整测试套件
   - 验证覆盖率达标（>80%）
   - 性能基准测试
   - 安全扫描无Critical漏洞

### 10.3 持续改进

- **每日**: 运行Smoke tests (<5分钟)
- **每次PR**: 运行P0+P1 tests (<45分钟)
- **每晚**: 运行完整测试套件 + 性能回归 (<2小时)
- **每周**: 安全扫描 + 依赖更新
- **每月**: 测试覆盖率报告 + 测试债务清理

---

## 11. 附录

### 11.1 测试命令快速参考

```bash
# 单元测试
make test-unit

# 集成测试
make test-integration

# E2E测试
make test-e2e

# 完整测试套件
make test

# 带覆盖率
make test-cover

# 性能测试
make test-perf

# 安全扫描
make security

# 完整质量检查（lint + test + security）
make check
```

### 11.2 相关文档

- [PRD文档](../docs/prd.md)
- [技术架构](../docs/architecture/tech-architecture.md)
- [Epic 5: 认证与授权](../docs/epics/5-auth-epic.md)
- [API设计规范](../docs/standards/api-design.md)
- [编码规范](../docs/standards/coding-standards.md)

### 11.3 联系人

| 角色 | 负责人 | 联系方式 |
|------|--------|---------|
| **Test Architect** | Murat (TEA Agent) | - |
| **Backend Lead** | TBD | - |
| **DevOps Lead** | TBD | - |
| **Security Lead** | TBD | - |

---

**文档状态**: ✅ 完成  
**下次审查**: Sprint 0结束后  
**批准人**: Websoft9
