-- 236_create_media_tasks.sql
-- 统一媒体任务表（图片 / 视频 / 音频）。
-- 对应 ent schema media_task.go；此前仅生成了 ent 代码，漏掉了建表迁移，
-- 导致 MediaWorker 轮询时报 relation "media_tasks" does not exist。
CREATE TABLE IF NOT EXISTS media_tasks (
    id               BIGSERIAL PRIMARY KEY,
    local_id         VARCHAR(128) UNIQUE NOT NULL,
    media_kind       VARCHAR(20) NOT NULL DEFAULT 'video',
    user_id          BIGINT NOT NULL,
    api_key_id       BIGINT,
    public_model     VARCHAR(100) NOT NULL,
    upstream_model   VARCHAR(100) NOT NULL,
    account_id       BIGINT NOT NULL,
    upstream_task_id VARCHAR(255),
    status           VARCHAR(20) NOT NULL DEFAULT 'processing',
    resolution       VARCHAR(20),
    duration_sec     INT,
    media_url        TEXT,
    thumbnail_url    TEXT,
    request_body     JSONB,
    error_message    TEXT,
    cost_usd         DECIMAL(20,8) DEFAULT 0,
    reserved_cost    DECIMAL(20,8),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_media_tasks_status_created ON media_tasks(media_kind, status, created_at);
CREATE INDEX IF NOT EXISTS idx_media_tasks_user_created   ON media_tasks(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_media_tasks_account         ON media_tasks(account_id);
