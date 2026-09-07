-- 237_add_video_tasks_reserved_cost.sql
-- 补齐 video_tasks.reserved_cost 列。
-- 234 号建表迁移漏掉了该列，VideoWorker 轮询时报
-- column video_tasks.reserved_cost does not exist。
-- NULL = 创建时未预扣；>0 = 创建任务时按估算费用预扣。
ALTER TABLE video_tasks ADD COLUMN IF NOT EXISTS reserved_cost DECIMAL(20,8);
