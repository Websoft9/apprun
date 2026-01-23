# Requirements-to-Tests Traceability Matrix
# apprun BaaS Platform

**Generated**: 2026-01-23  
**Test Architect**: Murat (TEA Agent)  
**Status**: Phase 1 Analysis Complete  
**Workflow**: testarch-trace

---

## Executive Summary

### Coverage Overview

| 优先级 | 总需求数 | 已覆盖 | 覆盖率 | 状态 |
|--------|---------|--------|--------|------|
| **P0** | 12 | 5 | **42%** | ❌ FAIL |
| **P1** | 18 | 8 | **44%** | ❌ FAIL |
| **P2** | 15 | 6 | **40%** | ⚠️ WARN |
| **P3** | 8 | 2 | **25%** | ✅ OK |
| **总计** | **53** | **21** | **40%** | ❌ **FAIL** |

### 关键发现

1. **黑盒测试基础设施已搭建**：`tests/` 目录结构完整，Fixture/Factory 模式已实现
2. **白盒测试覆盖良好**：`core/` 模块有 72 个测试文件，覆盖率达标
3. **黑盒测试为模板状态**：集成测试和 E2E 测试大多为 TODO placeholder
4. **Spec 到测试的映射缺失**：测试未显式引用 AC（验收标准）

---

## 1. Spec Source Analysis

### 1.1 PRD Functional Requirements (FR)

| FR ID | 模块 | 验收标准数 | 优先级 |
|-------|------|-----------|--------|
| FR-AUTH-001 | 认证授权 | 7 | P0 |
| FR-DATA-001 | 数据建模 | 5 | P1 |
| FR-CONFIG-001 | 配置中心 | 3 | P1 |
| FR-FUNC-001 | 函数服务 | 4 | P2 |
| FR-STORAGE-001 | 文件存储 | 6 | P1 |
| FR-EVENT-001 | 事件中心 | 3 | P1 |
| FR-I18N-001 | 国际化 | 4 | P2 |
| FR-LOG-001 | 日志服务 | 3 | P2 |
| FR-MON-001 | 系统监控 | 6 | P1 |
| FR-GATEWAY-001 | API网关 | 4 | P1 |

### 1.2 Epic 5 Acceptance Criteria (Authentication)

**来源**: [specs/epics/5-auth-epic.md](epics/5-auth-epic.md)

| AC ID | 描述 | 优先级 | 覆盖状态 |
|-------|------|--------|----------|
| AC-5.1 | 用户可通过 API 注册和登录（Email + Password） | P0 | ⚠️ PARTIAL |
| AC-5.2 | JWT Token 正确签发和验证 | P0 | ⚠️ PARTIAL |
| AC-5.3 | 密码安全存储（bcrypt hashing） | P0 | ✅ FULL |
| AC-5.4 | 项目级权限隔离正常工作 | P0 | ⚠️ PARTIAL |
| AC-5.5 | 资源级权限控制生效 | P0 | ⚠️ PARTIAL |
| AC-5.6 | API 响应时间 P95 < 100ms | P1 | ❌ NONE |
| AC-5.7 | 单元测试覆盖率 > 80% | P1 | ✅ FULL |

---

## 2. Test Inventory

### 2.1 White-box Tests (代码侧 - `core/`)

| 模块路径 | 测试文件数 | 测试类型 | 状态 |
|----------|-----------|---------|------|
| `core/modules/auth/` | 2 | Unit + Integration | ✅ 运行中 |
| `core/modules/config/` | 7 | Unit + Integration | ✅ 运行中 |
| `core/modules/metrics/` | 4 | Unit + Integration | ✅ 运行中 |
| `core/modules/audit/` | 4 | Unit + Integration | ✅ 运行中 |
| `core/modules/admin/` | 1 | Unit | ✅ 运行中 |
| `core/cmd/` | 1 | Unit | ✅ 运行中 |
| `core/ent/` | 1 | Schema | ✅ 运行中 |
| **总计** | **72** | Mixed | ✅ **运行中** |

### 2.2 Black-box Tests (测试侧 - `tests/`)

| 测试文件 | 测试类型 | 测试数 | 状态 |
|----------|---------|--------|------|
| `tests/integration/api/auth_test.go` | Integration | 6 | ⚠️ TODO (模板) |
| `tests/integration/auth/auth_advanced_simple_test.go` | Integration | 3 | ⚠️ SKIP |
| `tests/integration/auth/basic_test.go` | Integration | 1 | ⚠️ Placeholder |
| `tests/e2e/scenarios/auth_flow_test.go` | E2E | 2 | ⚠️ TODO (模板) |
| **总计** | **Mixed** | **12** | ⚠️ **模板状态** |

---

## 3. Detailed Traceability Matrix

### 3.1 FR-AUTH-001: 认证与授权服务

#### AC-5.1: 用户可通过 API 注册和登录

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| AUTH-INT-001 | `TestUserRegistration/successful_registration` | [tests/integration/api/auth_test.go#L28](../tests/integration/api/auth_test.go#L28) | Integration | ⚠️ TODO |
| AUTH-INT-002 | `TestUserRegistration/duplicate_email_rejected` | [tests/integration/api/auth_test.go#L68](../tests/integration/api/auth_test.go#L68) | Integration | ⚠️ TODO |
| AUTH-INT-003 | `TestUserRegistration/invalid_email_format_rejected` | [tests/integration/api/auth_test.go#L85](../tests/integration/api/auth_test.go#L85) | Integration | ⚠️ TODO |
| AUTH-INT-004 | `TestUserRegistration/weak_password_rejected` | [tests/integration/api/auth_test.go#L99](../tests/integration/api/auth_test.go#L99) | Integration | ⚠️ TODO |
| AUTH-INT-005 | `TestUserLogin/successful_login` | [tests/integration/api/auth_test.go#L127](../tests/integration/api/auth_test.go#L127) | Integration | ⚠️ TODO |
| AUTH-INT-006 | `TestUserLogin/incorrect_password_rejected` | [tests/integration/api/auth_test.go#L159](../tests/integration/api/auth_test.go#L159) | Integration | ⚠️ TODO |
| AUTH-UNIT-001 | `TestAuthMiddleware*` | [core/modules/auth/auth_integration_test.go](../core/modules/auth/auth_integration_test.go) | Unit | ✅ FULL |
| AUTH-E2E-001 | `TestCompleteAuthFlow_E2E` | [tests/e2e/scenarios/auth_flow_test.go#L20](../tests/e2e/scenarios/auth_flow_test.go#L20) | E2E | ⚠️ TODO |

**覆盖分析**:
- **实际运行测试**: 1 个 (白盒)
- **模板/TODO 测试**: 7 个
- **覆盖状态**: ⚠️ **PARTIAL** (仅白盒覆盖)

---

#### AC-5.2: JWT Token 正确签发和验证

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| JWT-INT-011 | `TestAuthINT011_JWTExpirationHandling` | [tests/integration/auth/auth_advanced_simple_test.go#L19](../tests/integration/auth/auth_advanced_simple_test.go#L19) | Integration | ⚠️ SKIP |
| JWT-UNIT-001 | JWT 生成/验证 | core/modules/auth/ | Unit | ✅ FULL |

**覆盖分析**:
- **实际运行测试**: 1 个 (白盒)
- **黑盒测试**: 0 个运行
- **覆盖状态**: ⚠️ **PARTIAL**

---

#### AC-5.3: 密码安全存储（bcrypt hashing）

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| PWD-UNIT-001 | bcrypt cost >= 12 | core/modules/auth/ | Unit | ✅ FULL |
| PWD-UNIT-002 | 密码哈希不可逆 | core/modules/auth/ | Unit | ✅ FULL |

**覆盖分析**: ✅ **FULL** - 白盒单元测试已覆盖

---

#### AC-5.4: 项目级权限隔离正常工作

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| PROJ-INT-001 | `TestProjectIsolation` | [tests/integration/api/auth_test.go#L182](../tests/integration/api/auth_test.go#L182) | Integration | ⚠️ TODO |
| RBAC-UNIT-001 | Casbin 策略测试 | core/modules/config/handler_rbac_test.go | Unit | ✅ FULL |

**覆盖分析**:
- **白盒覆盖**: ✅ RBAC 引擎测试
- **黑盒覆盖**: ❌ API 级别隔离测试缺失
- **覆盖状态**: ⚠️ **PARTIAL**

---

#### AC-5.5: 资源级权限控制生效

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| RBAC-INT-001 | 权限矩阵测试 | tests/integration/auth/ | Integration | ❌ NONE |
| RBAC-UNIT-001 | `TestConfigHandlerRBAC*` | [core/modules/config/handler_rbac_test.go](../core/modules/config/handler_rbac_test.go) | Unit | ✅ FULL |

**覆盖分析**: ⚠️ **PARTIAL** - 缺少 API 级别黑盒测试

---

#### AC-5.6: API 响应时间 P95 < 100ms

| 测试 ID | 测试名称 | 测试文件 | 测试级别 | 覆盖状态 |
|---------|---------|---------|---------|---------|
| PERF-E2E-001 | `TestPerformanceBaseline_E2E` | [tests/e2e/scenarios/auth_flow_test.go#L186](../tests/e2e/scenarios/auth_flow_test.go#L186) | E2E | ⚠️ TODO |
| PERF-K6-001 | k6 负载测试脚本 | tests/performance/ | Performance | ❌ NONE |

**覆盖分析**: ❌ **NONE** - 性能测试未实现

---

### 3.2 FR-CONFIG-001: 配置管理服务

| AC ID | 描述 | 测试覆盖 | 状态 |
|-------|------|---------|------|
| CONFIG-AC-1 | 配置可通过 API CRUD | 7 个白盒测试 | ✅ FULL |
| CONFIG-AC-2 | 配置变更实时发布 | 未覆盖 | ❌ NONE |
| CONFIG-AC-3 | 配置变更历史可查 | 部分覆盖 | ⚠️ PARTIAL |

**相关测试文件**:
- [core/modules/config/handler_test.go](../core/modules/config/handler_test.go)
- [core/modules/config/service_test.go](../core/modules/config/service_test.go)
- [core/modules/config/registry_test.go](../core/modules/config/registry_test.go)
- [core/modules/config/loader_test.go](../core/modules/config/loader_test.go)

---

### 3.3 FR-MON-001: 系统监控与指标

| AC ID | 描述 | 测试覆盖 | 状态 |
|-------|------|---------|------|
| MON-AC-1 | 平台管理员可查看用户统计 | 4 个白盒测试 | ✅ FULL |
| MON-AC-2 | 系统健康指标可访问 | 覆盖 | ✅ FULL |
| MON-AC-3 | Metrics API 响应 < 200ms | 未测试 | ❌ NONE |
| MON-AC-4 | 非管理员返回 403 | 覆盖 | ✅ FULL |

**相关测试文件**:
- [core/modules/metrics/handler_test.go](../core/modules/metrics/handler_test.go)
- [core/modules/metrics/service_test.go](../core/modules/metrics/service_test.go)
- [core/modules/metrics/validation_test.go](../core/modules/metrics/validation_test.go)
- [core/modules/metrics/history_test.go](../core/modules/metrics/history_test.go)

---

## 4. Gap Analysis

### 4.1 Critical Gaps (P0 - 阻塞发布)

| Gap ID | 描述 | 关联 AC | 风险等级 | 建议行动 |
|--------|------|---------|---------|---------|
| **GAP-P0-001** | 黑盒 API 测试为 TODO 状态 | AC-5.1, AC-5.2 | 🔴 HIGH | 完成 `auth_test.go` 中所有 TODO |
| **GAP-P0-002** | E2E 认证流程测试为模板 | AC-5.1 | 🔴 HIGH | 实现 `TestCompleteAuthFlow_E2E` |
| **GAP-P0-003** | 项目隔离 API 测试缺失 | AC-5.4 | 🔴 HIGH | 完成 `TestProjectIsolation` |
| **GAP-P0-004** | JWT 过期处理测试 SKIP | AC-5.2 | 🟡 MEDIUM | 启用 `TestAuthINT011_JWTExpirationHandling` |

### 4.2 High Priority Gaps (P1 - PR 阻塞)

| Gap ID | 描述 | 关联 AC | 风险等级 | 建议行动 |
|--------|------|---------|---------|---------|
| **GAP-P1-001** | 性能测试未实现 | AC-5.6 | 🟡 MEDIUM | 配置 k6 负载测试 |
| **GAP-P1-002** | 密码重置测试 SKIP | Epic-5 | 🟡 MEDIUM | 实现密码重置 API 后启用 |
| **GAP-P1-003** | 并发登录限制测试 SKIP | Epic-5 | 🟢 LOW | 实现限流后启用 |

### 4.3 Medium Priority Gaps (P2 - Nightly)

| Gap ID | 描述 | 关联 FR | 建议行动 |
|--------|------|---------|---------|
| GAP-P2-001 | 配置变更实时发布测试 | FR-CONFIG-001 | 添加事件总线测试 |
| GAP-P2-002 | Metrics API 性能测试 | FR-MON-001 | 添加 k6 脚本 |

---

## 5. Coverage Metrics

### 5.1 By Test Level

| 测试级别 | 目标比例 | 实际比例 | 运行状态 |
|----------|---------|---------|---------|
| **Unit** | 40% | ~50% | ✅ 72 个文件运行中 |
| **Integration (白盒)** | 30% | ~30% | ✅ 运行中 |
| **Integration (黑盒)** | 20% | ~5% | ⚠️ 大部分为 TODO |
| **E2E** | 10% | ~2% | ⚠️ 模板状态 |

### 5.2 By Module

| 模块 | 白盒覆盖 | 黑盒覆盖 | 综合状态 |
|------|---------|---------|---------|
| **Auth** | ✅ 运行中 | ⚠️ TODO | ⚠️ PARTIAL |
| **Config** | ✅ 运行中 | ❌ 无 | ⚠️ PARTIAL |
| **Metrics** | ✅ 运行中 | ❌ 无 | ⚠️ PARTIAL |
| **Audit** | ✅ 运行中 | ❌ 无 | ⚠️ PARTIAL |
| **Storage** | ❌ 无 | ❌ 无 | ❌ NONE |
| **Functions** | ❌ 无 | ❌ 无 | ❌ NONE |

---

## 6. Quality Assessment

### 6.1 Test Infrastructure

| 检查项 | 状态 | 备注 |
|--------|------|------|
| 测试目录结构 | ✅ | `tests/` 结构完整 |
| Fixture/Factory 模式 | ✅ | `user_factory.go`, `project_factory.go` 已实现 |
| Test Utilities | ✅ | `testutils/` 工具函数已实现 |
| CI/CD 集成 | ⚠️ | Makefile 命令存在，GitHub Actions 待配置 |
| 测试文档 | ✅ | `tests/README.md` 完整 |

### 6.2 Test Quality Concerns

| 问题 | 严重性 | 建议 |
|------|--------|------|
| 黑盒测试为 TODO 状态 | 🔴 HIGH | 完成所有 TODO 测试实现 |
| 测试未引用 AC ID | 🟡 MEDIUM | 在测试注释中添加 `@spec-ref` |
| 无测试执行报告 | 🟡 MEDIUM | 配置 CI 测试报告上传 |
| 缺少性能基准 | 🟡 MEDIUM | 实现 k6 性能测试 |

---

## 7. Recommendations

### 7.1 Immediate Actions (本 Sprint)

1. **完成黑盒 API 测试**
   - 取消 `tests/integration/api/auth_test.go` 中所有 TODO 注释
   - 连接实际 HTTP 路由器进行测试
   - 预估工作量：2-3 天

2. **启用 E2E 测试**
   - 配置测试服务器启动脚本
   - 完成 `TestCompleteAuthFlow_E2E` 实现
   - 预估工作量：1-2 天

3. **添加 Spec 引用**
   - 在每个测试函数上方添加 AC 引用注释
   - 格式：`// @spec-ref: specs/epics/5-auth-epic.md#AC-5.1`

### 7.2 Short-term Actions (下 Sprint)

1. **性能测试基础设施**
   - 配置 k6 负载测试脚本
   - 设置 CI scheduled job

2. **扩展黑盒测试覆盖**
   - 添加 Config API 黑盒测试
   - 添加 Metrics API 黑盒测试

### 7.3 Long-term Actions (未来)

1. **UI 测试准备**
   - 当 UI 开发完成后，搭建 Playwright 框架
   - 运行 `[TF] Test Framework` 工作流

---

## 8. Gate YAML Snippet

```yaml
# 可供 CI/CD 消费的追溯状态
traceability:
  date: "2026-01-23"
  project: "apprun"
  coverage:
    overall: 40%
    p0: 42%
    p1: 44%
    p2: 40%
  gaps:
    critical: 4
    high: 3
    medium: 2
  status: "FAIL"
  blockers:
    - "GAP-P0-001: Black-box API tests are TODO"
    - "GAP-P0-002: E2E auth flow test is template"
    - "GAP-P0-003: Project isolation API test missing"
  next_steps:
    - "Complete auth_test.go TODO implementations"
    - "Enable E2E test with running server"
    - "Add @spec-ref comments to all tests"
```

---

## References

- **PRD**: [specs/prd.md](prd.md)
- **Epic 5**: [specs/epics/5-auth-epic.md](epics/5-auth-epic.md)
- **API Spec**: [specs/api.md](api.md)
- **Test Design**: [_bmad-output/test-design-apprun-system.md](../_bmad-output/test-design-apprun-system.md)
- **Test README**: [tests/README.md](../tests/README.md)

---

<!-- Generated by TEA Agent (BMAD v6) -->
