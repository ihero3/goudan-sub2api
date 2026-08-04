-- 为 team_members 表补充 Ent schema 中定义但迁移 176 遗漏的列
-- department_id / consumer_id 在 Ent schema 中为 Optional + Nillable

ALTER TABLE team_members ADD COLUMN IF NOT EXISTS department_id BIGINT;
ALTER TABLE team_members ADD COLUMN IF NOT EXISTS consumer_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_team_members_department_id ON team_members (department_id);
CREATE INDEX IF NOT EXISTS idx_team_members_consumer_id ON team_members (consumer_id);
