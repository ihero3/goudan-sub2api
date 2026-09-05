package service

// media_worker.go — 媒体任务（图片 / 视频 / 音频）后台轮询 Worker。
// 定时扫描 status=processing 的媒体任务，向上游查询状态并更新本地记录。
// 与 VideoWorker 对齐，但操作 MediaTaskService + media_tasks 表。

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

// MediaWorker 媒体任务后台轮询 Worker。
type MediaWorker struct {
	mediaTaskService *MediaTaskService
	logger           *zap.Logger
	interval         time.Duration
	taskBatchSize    int
	taskMinAge       time.Duration

	leaderLockCache LeaderLockCache
	leaderLockKey   string
	instanceID      string
	leaderLockTTL   time.Duration

	mu      sync.Mutex
	stopCh  chan struct{}
	stopped bool
}

// NewMediaWorker 创建媒体任务 Worker。
func NewMediaWorker(mediaTaskService *MediaTaskService) *MediaWorker {
	return &MediaWorker{
		mediaTaskService: mediaTaskService,
		logger:           logger.L(),
		interval:         30 * time.Second,
		taskBatchSize:    50,
		taskMinAge:       10 * time.Second,
		leaderLockKey:    "sub2api:media_worker:leader",
		instanceID:       uuid.NewString(),
		leaderLockTTL:    45 * time.Second,
		stopCh:           make(chan struct{}),
	}
}

// SetLeaderLock injects the leader-lock cache used to elect a single worker instance.
// Call before Start. When nil or Redis is unavailable, the worker runs ungated.
func (w *MediaWorker) SetLeaderLock(lockCache LeaderLockCache) {
	if w == nil {
		return
	}
	w.leaderLockCache = lockCache
}

// Start 启动后台轮询 goroutine。非阻塞，调用后立即返回。
func (w *MediaWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// Stop 停止 Worker。
func (w *MediaWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	w.stopped = true
	close(w.stopCh)
}

func (w *MediaWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("media_worker: started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("media_worker: stopped (context cancelled)")
			return
		case <-w.stopCh:
			w.logger.Info("media_worker: stopped (stop signal)")
			return
		case <-ticker.C:
			w.pollOnce(ctx)
		}
	}
}

// pollOnce 执行一轮轮询。
func (w *MediaWorker) pollOnce(ctx context.Context) {
	// 分布式 leader 锁：仅 leader 实例执行轮询，避免多副本重复处理。
	release, ok := tryAcquireSingletonLeaderLock(
		ctx, w.leaderLockCache, nil, w.leaderLockKey, w.instanceID, w.leaderLockTTL,
	)
	if release != nil {
		defer release()
	}
	if !ok {
		return
	}
	before := time.Now().Add(-w.taskMinAge)
	tasks, err := w.mediaTaskService.mediaTaskRepo.ListProcessingTasks(ctx, before, w.taskBatchSize)
	if err != nil {
		w.logger.Error("media_worker: list processing tasks failed", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	w.logger.Debug("media_worker: polling tasks", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
		}

		if err := w.mediaTaskService.PollTask(ctx, task); err != nil {
			w.logger.Warn("media_worker: poll task failed",
				zap.Int64("task_id", task.ID),
				zap.String("local_id", task.LocalID),
				zap.Error(err))
		}
	}
}
