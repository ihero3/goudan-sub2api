-- 181: Align remaining team-management tables with Ent schema
-- 180 already fixed departments & consumers; this fixes the rest.
-- See: ent/schema/team.go, team_member.go, team_audit_log.go,
--       team_usage_team_daily.go, team_usage_dept_daily.go,
--       team_usage_consumer_daily.go, team_usage_model_daily.go

-- ============================================================
-- 1. teams: drop legacy columns not in Ent schema
--    Ent: name, slug, owner_id(StorageKey=owner_user_id), billing_email,
--         settings(StorageKey=metadata), status
--    DB legacy: description, balance, concurrency, rpm_limit, total_recharged
-- ============================================================
ALTER TABLE teams DROP COLUMN IF EXISTS description;
ALTER TABLE teams DROP COLUMN IF EXISTS balance;
ALTER TABLE teams DROP COLUMN IF EXISTS concurrency;
ALTER TABLE teams DROP COLUMN IF EXISTS rpm_limit;
ALTER TABLE teams DROP COLUMN IF EXISTS total_recharged;
-- 'metadata' column is kept (Ent uses StorageKey("metadata") for 'settings' field)

-- ============================================================
-- 2. team_members: align CHECK constraints with Ent schema
--    Ent role comment: owner|admin|member|viewer (4 values)
--    DB CHECK: only owner, admin, member (3 values) -> missing 'viewer'
--    Ent status comment: active|inactive|removed
--    DB CHECK: active, inactive, removed (matches)
-- ============================================================
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_check;
ALTER TABLE team_members ADD CONSTRAINT team_members_role_check
    CHECK (role IN ('owner', 'admin', 'member', 'viewer'));

-- ============================================================
-- 3. team_audit_logs: align columns with Ent schema
--    Ent: user_id(nullable), action, operation_type(nullable), resource_type,
--         resource_id(nullable), changes(JSONB nullable), ip(nullable),
--         user_agent(nullable text), created_at
--    DB legacy: actor_user_id, actor_type, details, ip_address
-- ============================================================
-- Rename actor_user_id -> user_id
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'team_audit_logs' AND column_name = 'actor_user_id')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'team_audit_logs' AND column_name = 'user_id') THEN
        ALTER TABLE team_audit_logs RENAME COLUMN actor_user_id TO user_id;
    END IF;
END $$;

-- Drop legacy columns
ALTER TABLE team_audit_logs DROP COLUMN IF EXISTS actor_type;
ALTER TABLE team_audit_logs DROP COLUMN IF EXISTS details;
ALTER TABLE team_audit_logs DROP COLUMN IF EXISTS ip_address;

-- Add missing columns
ALTER TABLE team_audit_logs ADD COLUMN IF NOT EXISTS operation_type VARCHAR(50);
ALTER TABLE team_audit_logs ADD COLUMN IF NOT EXISTS changes JSONB;
ALTER TABLE team_audit_logs ADD COLUMN IF NOT EXISTS ip VARCHAR(45);
ALTER TABLE team_audit_logs ADD COLUMN IF NOT EXISTS user_agent TEXT;

-- 'action', 'resource_type', 'resource_id', 'created_at' already exist in DB.

-- ============================================================
-- 4. team_usage_team_daily: align with Ent schema
--    Ent: team_id, bucket_date, total_requests, input_tokens, output_tokens,
--         cache_creation_tokens, cache_read_tokens, total_cost, actual_cost,
--         computed_at
--    DB matches except: DB has UNIQUE (team_id, bucket_date) which matches Ent.
--    No changes needed.
-- ============================================================

-- ============================================================
-- 5. team_usage_dept_daily: align columns with Ent schema
--    Ent adds: department_name, cost_center_code (nullable)
--    DB unique: (department_id, bucket_date)
--    Ent unique: (team_id, department_id, bucket_date)
-- ============================================================
ALTER TABLE team_usage_dept_daily ADD COLUMN IF NOT EXISTS department_name VARCHAR(100);
ALTER TABLE team_usage_dept_daily ADD COLUMN IF NOT EXISTS cost_center_code VARCHAR(50);

-- Replace old unique constraint with new one
ALTER TABLE team_usage_dept_daily DROP CONSTRAINT IF EXISTS team_usage_dept_daily_unique;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'team_usage_dept_daily_team_id_department_id_bucket_date_key'
    ) THEN
        ALTER TABLE team_usage_dept_daily
            ADD CONSTRAINT team_usage_dept_daily_team_id_department_id_bucket_date_key
            UNIQUE (team_id, department_id, bucket_date);
    END IF;
END $$;

-- ============================================================
-- 6. team_usage_consumer_daily: align columns with Ent schema
--    Ent adds: consumer_name, consumer_type (nullable)
--    DB unique: (consumer_id, bucket_date)
--    Ent unique: (team_id, consumer_id, bucket_date)
-- ============================================================
ALTER TABLE team_usage_consumer_daily ADD COLUMN IF NOT EXISTS consumer_name VARCHAR(100);
ALTER TABLE team_usage_consumer_daily ADD COLUMN IF NOT EXISTS consumer_type VARCHAR(20);

ALTER TABLE team_usage_consumer_daily DROP CONSTRAINT IF EXISTS team_usage_consumer_daily_unique;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'team_usage_consumer_daily_team_id_consumer_id_bucket_date_key'
    ) THEN
        ALTER TABLE team_usage_consumer_daily
            ADD CONSTRAINT team_usage_consumer_daily_team_id_consumer_id_bucket_date_key
            UNIQUE (team_id, consumer_id, bucket_date);
    END IF;
END $$;

-- ============================================================
-- 7. team_usage_model_daily: align columns with Ent schema
--    Ent: team_id, department_id(nullable), consumer_id(nullable), bucket_date,
--         model_name, total_requests, ..., computed_at
--    DB has: model (not model_name), no department_id/consumer_id
--    DB unique: (team_id, model, bucket_date)
--    Ent unique: (team_id, department_id, consumer_id, bucket_date, model_name)
-- ============================================================
-- Rename model -> model_name
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'team_usage_model_daily' AND column_name = 'model')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'team_usage_model_daily' AND column_name = 'model_name') THEN
        ALTER TABLE team_usage_model_daily RENAME COLUMN model TO model_name;
    END IF;
END $$;

ALTER TABLE team_usage_model_daily ADD COLUMN IF NOT EXISTS department_id BIGINT;
ALTER TABLE team_usage_model_daily ADD COLUMN IF NOT EXISTS consumer_id BIGINT;

ALTER TABLE team_usage_model_daily DROP CONSTRAINT IF EXISTS team_usage_model_daily_unique;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'team_usage_model_daily_team_id_department_id_consumer_id_bucket_date_model_name_key'
    ) THEN
        ALTER TABLE team_usage_model_daily
            ADD CONSTRAINT team_usage_model_daily_team_id_department_id_consumer_id_bucket_date_model_name_key
            UNIQUE (team_id, department_id, consumer_id, bucket_date, model_name);
    END IF;
END $$;
