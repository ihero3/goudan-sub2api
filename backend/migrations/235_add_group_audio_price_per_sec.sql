-- 235: 补齐 groups.audio_price_per_sec 列。
-- 该列由 ent schema group.go 的 media 音频按秒定价字段定义，但此前 218 号迁移只补了
-- audio_realtime/tts/stt 三列，漏掉此列；SchedulerSnapshotService 读取该列时因缺列
-- 在 outbox group_changed 处理中报 "column groups.audio_price_per_sec does not exist"。
-- NULL = 使用代码默认单价；显式 0 = 免费；>0 = 分组覆盖价。
ALTER TABLE groups ADD COLUMN IF NOT EXISTS audio_price_per_sec DECIMAL(20,8);
