-- 团队管理模块：存量数据迁移
-- 将现有用户迁移为团队，并为每个用户创建默认消费者和团队成员关系
--
-- 迁移策略：
--   1. 每个现有用户创建一个同名团队（teams.id = users.id，保持主键一致）
--   2. 每个团队创建一个默认消费者（consumers.id = users.id，保持主键一致）
--   3. 每个用户作为其团队的 owner 加入 team_members
--   4. 更新 api_keys.team_id = api_keys.user_id（将现有 API key 归属到用户对应的团队）
--   5. 跳过 usage_logs 迁移（热路径，不修改）
--   6. 跳过 consumer_id 匹配（过于脆弱，后续按需手动关联）
--   7. 同步序列，确保后续插入的 ID 不冲突
--
-- 幂等性：使用 WHERE NOT EXISTS / ON CONFLICT 确保可重复执行

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '10min';

-- ============================================================
-- 1. 为每个现有用户创建团队（teams.id = users.id）
-- ============================================================
INSERT INTO teams (
    id, name, slug, description, status, owner_user_id,
    balance, concurrency, rpm_limit, total_recharged,
    metadata, created_at, updated_at
)
SELECT
    u.id,
    COALESCE(u.username, u.email, 'team_' || u.id::text) AS name,
    'team_' || u.id::text AS slug,
    'Auto-migrated from user ' || u.id::text AS description,
    'active' AS status,
    u.id AS owner_user_id,
    u.balance,
    u.concurrency,
    u.rpm_limit,
    u.total_recharged,
    '{}'::jsonb AS metadata,
    COALESCE(u.created_at, NOW()) AS created_at,
    COALESCE(u.updated_at, NOW()) AS updated_at
FROM users u
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM teams t WHERE t.id = u.id
  )
ON CONFLICT (id) DO NOTHING;

-- 处理 slug 冲突（理论上不应发生，因为 id 唯一）
UPDATE teams t
SET slug = 'team_' || t.id::text || '_' || EXTRACT(EPOCH FROM NOW())::bigint::text
WHERE t.slug IN (
    SELECT slug
    FROM teams
    GROUP BY slug
    HAVING COUNT(*) > 1
);

-- ============================================================
-- 2. 为每个团队创建默认消费者（consumers.id = users.id）
-- ============================================================
INSERT INTO consumers (
    id, team_id, department_id, name, key_prefix, description,
    status, quota, quota_used, metadata, created_at, updated_at
)
SELECT
    u.id AS id,
    t.id AS team_id,
    NULL AS department_id,
    COALESCE(u.username, u.email, 'default-consumer-' || u.id::text) AS name,
    'default' AS key_prefix,
    'Default consumer for team ' || t.id::text AS description,
    'active' AS status,
    0 AS quota,
    0 AS quota_used,
    '{}'::jsonb AS metadata,
    COALESCE(u.created_at, NOW()) AS created_at,
    COALESCE(u.updated_at, NOW()) AS updated_at
FROM users u
JOIN teams t ON t.id = u.id
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM consumers c WHERE c.id = u.id
  )
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- 3. 将每个用户作为 owner 加入其团队
-- ============================================================
INSERT INTO team_members (
    team_id, user_id, role, status, joined_at, created_at, updated_at
)
SELECT
    t.id AS team_id,
    u.id AS user_id,
    'owner' AS role,
    'active' AS status,
    COALESCE(u.created_at, NOW()) AS joined_at,
    COALESCE(u.created_at, NOW()) AS created_at,
    COALESCE(u.updated_at, NOW()) AS updated_at
FROM users u
JOIN teams t ON t.id = u.id
WHERE u.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM team_members tm
      WHERE tm.team_id = t.id AND tm.user_id = u.id
  )
ON CONFLICT (team_id, user_id) DO NOTHING;

-- ============================================================
-- 4. 更新 api_keys：将 team_id 设置为 user_id（归属到用户对应的团队）
--    注意：consumer_id 不自动匹配（过于脆弱，后续按需手动关联）
-- ============================================================
UPDATE api_keys
SET team_id = user_id
WHERE team_id IS NULL
  AND user_id IS NOT NULL
  AND deleted_at IS NULL;

-- ============================================================
-- 5. 同步序列，确保后续插入的 ID 不冲突
-- ============================================================
-- teams 序列
DO $$
DECLARE
    max_team_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_team_id FROM teams;
    PERFORM setval('teams_id_seq', GREATEST(max_team_id, 1));
EXCEPTION WHEN undefined_table THEN
    -- 序列不存在，忽略
    NULL;
END $$;

-- team_members 序列
DO $$
DECLARE
    max_member_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_member_id FROM team_members;
    PERFORM setval('team_members_id_seq', GREATEST(max_member_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- departments 序列
DO $$
DECLARE
    max_dept_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_dept_id FROM departments;
    PERFORM setval('departments_id_seq', GREATEST(max_dept_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- consumers 序列
DO $$
DECLARE
    max_consumer_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_consumer_id FROM consumers;
    PERFORM setval('consumers_id_seq', GREATEST(max_consumer_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- team_audit_logs 序列
DO $$
DECLARE
    max_audit_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_audit_id FROM team_audit_logs;
    PERFORM setval('team_audit_logs_id_seq', GREATEST(max_audit_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- team_usage_team_daily 序列
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM team_usage_team_daily;
    PERFORM setval('team_usage_team_daily_id_seq', GREATEST(max_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- team_usage_dept_daily 序列
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM team_usage_dept_daily;
    PERFORM setval('team_usage_dept_daily_id_seq', GREATEST(max_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- team_usage_consumer_daily 序列
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM team_usage_consumer_daily;
    PERFORM setval('team_usage_consumer_daily_id_seq', GREATEST(max_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;

-- team_usage_model_daily 序列
DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) + 1 INTO max_id FROM team_usage_model_daily;
    PERFORM setval('team_usage_model_daily_id_seq', GREATEST(max_id, 1));
EXCEPTION WHEN undefined_table THEN
    NULL;
END $$;
