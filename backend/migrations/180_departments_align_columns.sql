-- 180: Align departments table columns with Ent schema
-- Ent schema defines: cost_center_code, external_id, level, path, source, sort_order
-- DB table has: code, quota, quota_used, metadata (legacy)

-- 1. Rename code -> cost_center_code (if exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'departments' AND column_name = 'code') THEN
        ALTER TABLE departments RENAME COLUMN code TO cost_center_code;
    END IF;
END $$;

-- 2. Drop old unique constraint on (team_id, code)
ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_team_code_unique;

-- 3. Add missing columns
ALTER TABLE departments ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE departments ADD COLUMN IF NOT EXISTS level INT NOT NULL DEFAULT 0;
ALTER TABLE departments ADD COLUMN IF NOT EXISTS path VARCHAR(500) NOT NULL DEFAULT '/';
ALTER TABLE departments ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT 'manual';
ALTER TABLE departments ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

-- 4. Drop legacy columns (no longer in Ent schema)
ALTER TABLE departments DROP COLUMN IF EXISTS quota;
ALTER TABLE departments DROP COLUMN IF EXISTS quota_used;
ALTER TABLE departments DROP COLUMN IF EXISTS metadata;

-- 5. Add new unique constraint matching Ent schema (team_id, name)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'departments_team_id_name_key'
    ) THEN
        ALTER TABLE departments ADD CONSTRAINT departments_team_id_name_key UNIQUE (team_id, name);
    END IF;
END $$;

-- 6. Add index on external_id (Ent schema defines this index)
CREATE INDEX IF NOT EXISTS idx_departments_external_id ON departments (external_id);
CREATE INDEX IF NOT EXISTS idx_departments_path ON departments (path);

-- 7. Soft delete: ensure deleted_at column exists (already in original migration)


-- ============================================================
-- Part 2: Align consumers table columns with Ent schema
-- Ent schema: type, email, phone, title, app_id, app_description,
--             external_id, source, deactivated_at, settings(JSONB)
-- DB table: key_prefix, quota, quota_used, ip_whitelist, ip_blacklist, metadata (legacy)
-- ============================================================

-- 8. Drop legacy columns not in Ent schema
ALTER TABLE consumers DROP COLUMN IF EXISTS key_prefix;
ALTER TABLE consumers DROP COLUMN IF EXISTS quota;
ALTER TABLE consumers DROP COLUMN IF EXISTS quota_used;
ALTER TABLE consumers DROP COLUMN IF EXISTS ip_whitelist;
ALTER TABLE consumers DROP COLUMN IF EXISTS ip_blacklist;
ALTER TABLE consumers DROP COLUMN IF EXISTS metadata;

-- 9. Add missing columns per Ent schema
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'person';
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS email VARCHAR(255);
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS phone VARCHAR(50);
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS title VARCHAR(100);
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS app_id VARCHAR(100);
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS app_description TEXT;
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS external_id VARCHAR(255);
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS source VARCHAR(50) NOT NULL DEFAULT 'manual';
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;
ALTER TABLE consumers ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb;

-- 10. Add unique constraint for (team_id, app_id) per Ent schema
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'consumers_team_id_app_id_key'
    ) THEN
        ALTER TABLE consumers ADD CONSTRAINT consumers_team_id_app_id_key UNIQUE (team_id, app_id);
    END IF;
END $$;

-- 11. Add new indexes per Ent schema
CREATE INDEX IF NOT EXISTS idx_consumers_type ON consumers (type);
CREATE INDEX IF NOT EXISTS idx_consumers_external_id ON consumers (external_id);
