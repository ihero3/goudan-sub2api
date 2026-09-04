package service

// video_worker.go — 视频任务后台轮询 Worker。
// 定时扫描 status=processing 的任务，向上游查询状态并更新本地记录。
// 模式参考 batch_image_worker.go，但更简单：无队列，直接查库轮询。

import (
	"context"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"go.uber.org/zap"
)

// VideoWorker 视频任务后台轮询 Worker。
type VideoWorker struct {
	videoTaskService *VideoTaskService
	logger           *zap.Logger
	interval         time.Duration
	taskBatchSize    int
	taskMinAge       time.Duration // 任务创建后至少等待这么久才轮询（避免刚创建就查）

	mu     sync.Mutex
	stopCh chan struct{}
	stopped bool
}

// NewVideoWorker 创建视频任务 Worker。
func NewVideoWorker(videoTaskService *VideoTaskService) *VideoWorker {
	return &VideoWorker{
		videoTaskService: videoTaskService,
		logger:           logger.L(),
		interval:         30 * time.Second,
		taskBatchSize:    50,
		taskMinAge:       10 * time.Second,
		stopCh:           make(chan struct{}),
	}
}

// Start 启动后台轮询 goroutine。非阻塞，调用后立即返回。
func (w *VideoWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// Stop 停止 Worker。
func (w *VideoWorker) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	w.stopped = true
	close(w.stopCh)
}

func (w *VideoWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.logger.Info("video_worker: started", zap.Duration("interval", w.interval))

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("video_worker: stopped (context cancelled)")
			return
		case <-w.stopCh:
			w.logger.Info("video_worker: stopped (stop signal)")
			return
		case <-ticker.C:
			w.pollOnce(ctx)
		}
	}
}

// pollOnce 执行一轮轮询。
func (w *VideoWorker) pollOnce(ctx context.Context) {
	// 查询需要轮询的任务：status=processing 且创建时间超过 taskMinAge
	before := time.Now().Add(-w.taskMinAge)
	tasks, err := w.videoTaskService.videoTaskRepo.ListProcessingTasks(ctx, before, w.taskBatchSize)
	if err != nil {
		w.logger.Error("video_worker: list processing tasks failed", zap.Error(err))
		return
	}
	if len(tasks) == 0 {
		return
	}

	w.logger.Debug("video_worker: polling tasks", zap.Int("count", len(tasks)))

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
		}

		if err := w.videoTaskService.PollTask(ctx, task); err != nil {
			w.logger.Warn("video_worker: poll task failed",
				zap.Int64("task_id", task.ID),
				zap.String("local_id", task.LocalID),
				zap.Error(err))
		}
	}
}
