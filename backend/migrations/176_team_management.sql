-- 团队管理模块：创建团队、部门、消费者、审计日志、用量聚合表
-- 并扩展 api_keys 表以支持团队/部门/消费者维度
--
-- 设计要点：
--   - 所有主表（teams, team_members, departments, consumers）支持软删除（deleted_at）
--   - 审计日志和用量聚合表为 append-only，不支持软删除
--   - 用量聚合表按天粒度预聚合，避免对 usage_logs 热路径做全表扫描
--   - api_keys 新增 team_id / department_id / consumer_id 外键，但不修改 usage_logs（热路径）
--   - 遵循 Ent schema 约定：created_at, updated_at, deleted_at（软删除表）

-- ============================================================
-- 1. teams 表（团队/租户）
-- ============================================================
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    owner_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    balance DECIMAL(20, 8) NOT NULL DEFAULT 0,
    concurrency INT NOT NULL DEFAULT 0,
    rpm_limit INT NOT NULL DEFAULT 0,
    total_recharged DECIMAL(20, 8) NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT teams_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_teams_slug ON teams (slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_owner_user_id ON teams (owner_user_id);
CREATE INDEX IF NOT EXISTS idx_teams_status ON teams (status);
CREATE INDEX IF NOT EXISTS idx_teams_deleted_at ON teams (deleted_at);

COMMENT ON TABLE teams IS '团队/租户表，每个团队对应一个独立的资源隔离单元';
COMMENT ON COLUMN teams.slug IS '团队唯一标识 slug，用于 URL 和 API 路径';
COMMENT ON COLUMN teams.status IS '团队状态: active, suspended, deleted';
COMMENT ON COLUMN teams.owner_user_id IS '团队创建者/所有者用户ID';
COMMENT ON COLUMN teams.balance IS '团队账户余额';
COMMENT ON COLUMN teams.concurrency IS '团队并发限制';
COMMENT ON COLUMN teams.rpm_limit IS '团队每分钟请求限制';

-- ============================================================
-- 2. team_members 表（团队成员）
-- ============================================================
CREATE TABLE IF NOT EXISTS team_members (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT team_members_role_check CHECK (role IN ('owner', 'admin', 'member')),
    CONSTRAINT team_members_status_check CHECK (status IN ('active', 'inactive', 'removed')),
    CONSTRAINT team_members_team_user_unique UNIQUE (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members (team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members (user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members (role);
CREATE INDEX IF NOT EXISTS idx_team_members_status ON team_members (status);
CREATE INDEX IF NOT EXISTS idx_team_members_deleted_at ON team_members (deleted_at);

COMMENT ON TABLE team_members IS '团队成员关系表';
COMMENT ON COLUMN team_members.role IS '成员角色: owner, admin, member';
COMMENT ON COLUMN team_members.status IS '成员状态: active, inactive, removed';

-- ============================================================
-- 3. departments 表（部门）
-- ============================================================
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    parent_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50),
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    quota DECIMAL(20, 8) NOT NULL DEFAULT 0,
    quota_used DECIMAL(20, 8) NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT departments_status_check CHECK (status IN ('active', 'inactive', 'deleted')),
    CONSTRAINT departments_team_code_unique UNIQUE (team_id, code)
);

CREATE INDEX IF NOT EXISTS idx_departments_team_id ON departments (team_id);
CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments (parent_id);
CREATE INDEX IF NOT EXISTS idx_departments_status ON departments (status);
CREATE INDEX IF NOT EXISTS idx_departments_deleted_at ON departments (deleted_at);

COMMENT ON TABLE departments IS '团队部门表，支持层级结构';
COMMENT ON COLUMN departments.code IS '部门编码，团队内唯一';
COMMENT ON COLUMN departments.parent_id IS '上级部门ID，支持树形结构';
COMMENT ON COLUMN departments.quota IS '部门配额上限';
COMMENT ON COLUMN departments.quota_used IS '部门已用配额';

-- ============================================================
-- 4. consumers 表（消费者/API 消费者）
-- ============================================================
CREATE TABLE IF NOT EXISTS consumers (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    department_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(16) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    quota DECIMAL(20, 8) NOT NULL DEFAULT 0,
    quota_used DECIMAL(20, 8) NOT NULL DEFAULT 0,
    ip_whitelist TEXT[] NOT NULL DEFAULT '{}',
    ip_blacklist TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT consumers_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

CREATE INDEX IF NOT EXISTS idx_consumers_team_id ON consumers (team_id);
CREATE INDEX IF NOT EXISTS idx_consumers_department_id ON consumers (department_id);
CREATE INDEX IF NOT EXISTS idx_consumers_status ON consumers (status);
CREATE INDEX IF NOT EXISTS idx_consumers_deleted_at ON consumers (deleted_at);

COMMENT ON TABLE consumers IS 'API 消费者表，代表团队下的一个应用/服务/用户';
COMMENT ON COLUMN consumers.key_prefix IS '消费者标识前缀，用于 API key 分组';
COMMENT ON COLUMN consumers.quota IS '消费者配额上限';
COMMENT ON COLUMN consumers.quota_used IS '消费者已用配额';

-- ============================================================
-- 5. team_audit_logs 表（团队审计日志）
-- ============================================================
CREATE TABLE IF NOT EXISTS team_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    actor_type VARCHAR(20) NOT NULL DEFAULT 'user',
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id BIGINT,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_team_audit_logs_team_id ON team_audit_logs (team_id);
CREATE INDEX IF NOT EXISTS idx_team_audit_logs_actor_user_id ON team_audit_logs (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_team_audit_logs_action ON team_audit_logs (action);
CREATE INDEX IF NOT EXISTS idx_team_audit_logs_resource ON team_audit_logs (resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_team_audit_logs_created_at ON team_audit_logs (created_at DESC);

COMMENT ON TABLE team_audit_logs IS '团队审计日志，记录团队层面的操作事件';
COMMENT ON COLUMN team_audit_logs.actor_type IS '操作者类型: user, system, api_key';
COMMENT ON COLUMN team_audit_logs.action IS '操作动作: create, update, delete, login, etc.';
COMMENT ON COLUMN team_audit_logs.resource_type IS '资源类型: team, member, department, consumer, api_key';

-- ============================================================
-- 6. 用量聚合表（按团队/部门/消费者/模型维度，天粒度）
-- ============================================================

-- 团队日用量聚合
CREATE TABLE IF NOT EXISTS team_usage_team_daily (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    bucket_date DATE NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT team_usage_team_daily_unique UNIQUE (team_id, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_team_usage_team_daily_team_id ON team_usage_team_daily (team_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_team_daily_bucket_date ON team_usage_team_daily (bucket_date DESC);

COMMENT ON TABLE team_usage_team_daily IS '团队日用量聚合表';

-- 部门日用量聚合
CREATE TABLE IF NOT EXISTS team_usage_dept_daily (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    bucket_date DATE NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT team_usage_dept_daily_unique UNIQUE (department_id, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_team_usage_dept_daily_team_id ON team_usage_dept_daily (team_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_dept_daily_department_id ON team_usage_dept_daily (department_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_dept_daily_bucket_date ON team_usage_dept_daily (bucket_date DESC);

COMMENT ON TABLE team_usage_dept_daily IS '部门日用量聚合表';

-- 消费者日用量聚合
CREATE TABLE IF NOT EXISTS team_usage_consumer_daily (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    consumer_id BIGINT NOT NULL REFERENCES consumers(id) ON DELETE CASCADE,
    bucket_date DATE NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT team_usage_consumer_daily_unique UNIQUE (consumer_id, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_team_usage_consumer_daily_team_id ON team_usage_consumer_daily (team_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_consumer_daily_consumer_id ON team_usage_consumer_daily (consumer_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_consumer_daily_bucket_date ON team_usage_consumer_daily (bucket_date DESC);

COMMENT ON TABLE team_usage_consumer_daily IS '消费者日用量聚合表';

-- 模型日用量聚合（团队维度）
CREATE TABLE IF NOT EXISTS team_usage_model_daily (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    model VARCHAR(200) NOT NULL,
    bucket_date DATE NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_creation_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    actual_cost DECIMAL(20, 10) NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT team_usage_model_daily_unique UNIQUE (team_id, model, bucket_date)
);

CREATE INDEX IF NOT EXISTS idx_team_usage_model_daily_team_id ON team_usage_model_daily (team_id);
CREATE INDEX IF NOT EXISTS idx_team_usage_model_daily_model ON team_usage_model_daily (model);
CREATE INDEX IF NOT EXISTS idx_team_usage_model_daily_bucket_date ON team_usage_model_daily (bucket_date DESC);

COMMENT ON TABLE team_usage_model_daily IS '团队按模型日用量聚合表';

-- ============================================================
-- 7. 扩展 api_keys 表
-- ============================================================
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS team_id BIGINT REFERENCES teams(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS department_id BIGINT REFERENCES departments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS consumer_id BIGINT REFERENCES consumers(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_api_keys_team_id ON api_keys (team_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_department_id ON api_keys (department_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_consumer_id ON api_keys (consumer_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_team_consumer ON api_keys (team_id, consumer_id) WHERE deleted_at IS NULL;

COMMENT ON COLUMN api_keys.team_id IS '所属团队ID';
COMMENT ON COLUMN api_keys.department_id IS '所属部门ID';
COMMENT ON COLUMN api_keys.consumer_id IS '所属消费者ID';
