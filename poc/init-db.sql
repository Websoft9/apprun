-- apprun POC 数据库初始化脚本
-- 用途：创建测试数据表和权限配置

-- ==========================================
-- 1. 创建角色
-- ==========================================

-- PostgREST 匿名角色
CREATE ROLE web_anon NOLOGIN;

-- PostgREST 认证用户角色
CREATE ROLE authenticated NOLOGIN;

-- 授予权限
GRANT web_anon TO apprun;
GRANT authenticated TO apprun;

-- ==========================================
-- 2. 创建测试表
-- ==========================================

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50) DEFAULT 'free',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 用户表（简化版，主要认证由 Kratos 负责）
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    role VARCHAR(50) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 产品表（测试 PostgREST + RLS）
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 工作流定义表
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    definition JSONB NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 工作流执行历史
CREATE TABLE IF NOT EXISTS workflow_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID REFERENCES workflows(id),
    status VARCHAR(50) NOT NULL,
    input JSONB,
    output JSONB,
    error TEXT,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- 函数定义表
CREATE TABLE IF NOT EXISTS functions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    runtime VARCHAR(50) NOT NULL,
    code TEXT,
    config JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID,
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB,
    ip_address INET,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ==========================================
-- 3. 创建索引
-- ==========================================

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_products_tenant ON products(tenant_id);
CREATE INDEX idx_workflows_tenant ON workflows(tenant_id);
CREATE INDEX idx_workflow_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);
CREATE INDEX idx_functions_tenant ON functions(tenant_id);
CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id, created_at DESC);

-- ==========================================
-- 4. 启用行级安全（RLS）
-- ==========================================

-- 产品表 RLS
ALTER TABLE products ENABLE ROW LEVEL SECURITY;

-- 创建策略：只能访问自己租户的数据
CREATE POLICY tenant_isolation_products ON products
    FOR ALL
    TO authenticated
    USING (tenant_id::text = current_setting('request.jwt.claims', true)::json->>'tenant_id');

-- 允许匿名用户读取（用于测试）
CREATE POLICY anon_read_products ON products
    FOR SELECT
    TO web_anon
    USING (true);

-- 工作流表 RLS
ALTER TABLE workflows ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_workflows ON workflows
    FOR ALL
    TO authenticated
    USING (tenant_id::text = current_setting('request.jwt.claims', true)::json->>'tenant_id');

-- 函数表 RLS
ALTER TABLE functions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_functions ON functions
    FOR ALL
    TO authenticated
    USING (tenant_id::text = current_setting('request.jwt.claims', true)::json->>'tenant_id');

-- ==========================================
-- 5. 授予表权限
-- ==========================================

-- 匿名用户只读权限
GRANT SELECT ON ALL TABLES IN SCHEMA public TO web_anon;

-- 认证用户完整权限
GRANT ALL ON ALL TABLES IN SCHEMA public TO authenticated;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO authenticated;

-- ==========================================
-- 6. 插入测试数据
-- ==========================================

-- 测试租户
INSERT INTO tenants (id, name, plan) VALUES 
    ('11111111-1111-1111-1111-111111111111', 'Test Tenant 1', 'free'),
    ('22222222-2222-2222-2222-222222222222', 'Test Tenant 2', 'pro');

-- 测试用户
INSERT INTO users (id, email, tenant_id, role) VALUES 
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'alice@test.com', '11111111-1111-1111-1111-111111111111', 'admin'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'bob@test.com', '11111111-1111-1111-1111-111111111111', 'user'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'charlie@test.com', '22222222-2222-2222-2222-222222222222', 'admin');

-- 测试产品
INSERT INTO products (tenant_id, name, description, price) VALUES 
    ('11111111-1111-1111-1111-111111111111', 'Product A', 'Test product A', 19.99),
    ('11111111-1111-1111-1111-111111111111', 'Product B', 'Test product B', 29.99),
    ('22222222-2222-2222-2222-222222222222', 'Product C', 'Test product C', 39.99);

-- 测试工作流
INSERT INTO workflows (tenant_id, name, definition, enabled) VALUES 
    ('11111111-1111-1111-1111-111111111111', 'User Registration', 
     '{"steps": [{"name": "send_email", "type": "function"}, {"name": "create_project", "type": "http"}]}', 
     true);

-- ==========================================
-- 7. 创建视图（用于监控）
-- ==========================================

-- 租户统计视图
CREATE OR REPLACE VIEW tenant_stats AS
SELECT 
    t.id as tenant_id,
    t.name as tenant_name,
    COUNT(DISTINCT u.id) as user_count,
    COUNT(DISTINCT p.id) as product_count,
    COUNT(DISTINCT w.id) as workflow_count
FROM tenants t
LEFT JOIN users u ON t.id = u.tenant_id
LEFT JOIN products p ON t.id = p.tenant_id
LEFT JOIN workflows w ON t.id = w.tenant_id
GROUP BY t.id, t.name;

-- 工作流执行统计
CREATE OR REPLACE VIEW workflow_execution_stats AS
SELECT 
    DATE(started_at) as date,
    status,
    COUNT(*) as count,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration_seconds
FROM workflow_executions
WHERE started_at IS NOT NULL
GROUP BY DATE(started_at), status
ORDER BY date DESC, status;

GRANT SELECT ON tenant_stats TO web_anon, authenticated;
GRANT SELECT ON workflow_execution_stats TO web_anon, authenticated;

-- ==========================================
-- 8. 创建函数（用于测试）
-- ==========================================

-- 获取当前租户ID
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
BEGIN
    RETURN (current_setting('request.jwt.claims', true)::json->>'tenant_id')::uuid;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

-- 记录审计日志
CREATE OR REPLACE FUNCTION log_audit(
    p_action VARCHAR,
    p_resource_type VARCHAR,
    p_resource_id UUID,
    p_details JSONB
)
RETURNS void AS $$
BEGIN
    INSERT INTO audit_logs (
        tenant_id, 
        user_id, 
        action, 
        resource_type, 
        resource_id, 
        details
    ) VALUES (
        current_tenant_id(),
        (current_setting('request.jwt.claims', true)::json->>'user_id')::uuid,
        p_action,
        p_resource_type,
        p_resource_id,
        p_details
    );
END;
$$ LANGUAGE plpgsql;

-- ==========================================
-- 9. 触发器（自动更新 updated_at）
-- ==========================================

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_workflows_updated_at
    BEFORE UPDATE ON workflows
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_functions_updated_at
    BEFORE UPDATE ON functions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ==========================================
-- 完成
-- ==========================================

\echo 'Database initialization completed successfully!'
\echo 'Test tenants, users, and products have been created.'
\echo 'PostgREST API available at http://localhost:3000'
