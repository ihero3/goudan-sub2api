-- 234_create_video_tasks.sql
-- 视频生成异步任务表：记录用户请求、上游 channel、任务状态及最终结果。
CREATE TABLE IF NOT EXISTS video_tasks (
    id              BIGSERIAL PRIMARY KEY,
    local_id        VARCHAR(128) UNIQUE NOT NULL,
    user_id         BIGINT NOT NULL,
    api_key_id      BIGINT,
    public_model    VARCHAR(100) NOT NULL,
    upstream_model  VARCHAR(100) NOT NULL,
    account_id      BIGINT NOT NULL,
    upstream_task_id VARCHAR(255),
    status          VARCHAR(20) NOT NULL DEFAULT 'processing',
    resolution      VARCHAR(20),
    duration_sec    INT,
    video_url       TEXT,
    thumbnail_url   TEXT,
    request_body    JSONB,
    error_message   TEXT,
    cost_usd        DECIMAL(20,8) DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_video_tasks_status_created ON video_tasks(status, created_at);
CREATE INDEX IF NOT EXISTS idx_video_tasks_user_created   ON video_tasks(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_video_tasks_account         ON video_tasks(account_id);
