-- 183: Align team_usage_*_daily tables with Ent schema (TimeMixin)
-- 181 added the ent-specific columns (department_name, consumer_type, model_name, etc.)
-- but forgot the created_at / updated_at columns that the TimeMixin mandates.
-- Without them, Ent-generated SELECT statements reference non-existent columns
-- and the /teams/:id/analytics/trend endpoint returns HTTP 500.
-- Also drop the legacy total_duration_ms column which is not in the Ent schema.

-- ============================================================
-- 1. team_usage_team_daily
--    Ent (TimeMixin): created_at, updated_at
--    DB legacy: total_duration_ms (not in Ent)
-- ============================================================
ALTER TABLE team_usage_team_daily
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_team_daily
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_team_daily DROP COLUMN IF EXISTS total_duration_ms;

-- ============================================================
-- 2. team_usage_dept_daily
-- ============================================================
ALTER TABLE team_usage_dept_daily
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_dept_daily
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_dept_daily DROP COLUMN IF EXISTS total_duration_ms;

-- ============================================================
-- 3. team_usage_consumer_daily
-- ============================================================
ALTER TABLE team_usage_consumer_daily
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_consumer_daily
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_consumer_daily DROP COLUMN IF EXISTS total_duration_ms;

-- ============================================================
-- 4. team_usage_model_daily
-- ============================================================
ALTER TABLE team_usage_model_daily
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_model_daily
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
ALTER TABLE team_usage_model_daily DROP COLUMN IF EXISTS total_duration_ms;
