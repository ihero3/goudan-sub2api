-- 185: Add description, timezone, language columns to teams table
-- Aligns with Ent schema fields added for team settings.

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS description VARCHAR(500) DEFAULT '',
    ADD COLUMN IF NOT EXISTS timezone VARCHAR(50) DEFAULT 'Asia/Shanghai',
    ADD COLUMN IF NOT EXISTS language VARCHAR(10) DEFAULT 'zh-CN';
