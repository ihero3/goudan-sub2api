package service

// video_task_service.go — 视频任务业务逻辑。
// 负责：调度选 channel → model_mapping 翻译 → 调 adapter → 写表 → 返回。
// 复用现有 GatewayService 调度器和 RateLimitService 故障转移逻辑。

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// VideoTaskRepo 视频任务仓储接口（service 层定义，repository 层实现）。
// 避免直接 import repository 包导致 import cycle。
type VideoTaskRepo interface {
	Create(ctx context.Context, task *VideoTaskRecord) (*VideoTaskRecord, error)
	GetByLocalID(ctx context.Context, localID string) (*VideoTaskRecord, error)
	GetByID(ctx context.Context, id int64) (*VideoTaskRecord, error)
	UpdateStatus(ctx context.Context, id int64, status, errorMsg string) error
	UpdateResult(ctx context.Context, id int64, status, videoURL, thumbnailURL string, durationSec int, costUSD float64) error
	UpdateUpstreamTaskID(ctx context.Context, id int64, upstreamTaskID string) error
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]*VideoTaskRecord, int, error)
	ListProcessingTasks(ctx context.Context, before time.Time, limit int) ([]*VideoTaskRecord, error)
}

// VideoTaskRecord 是 video_tasks 表的 Go 映射。
type VideoTaskRecord struct {
	ID             int64
	LocalID        string
	UserID         int64
	APIKeyID       int64
	PublicModel    string
	UpstreamModel  string
	AccountID      int64
	UpstreamTaskID string
	Status         string
	Resolution     string
	DurationSec    int
	VideoURL       string
	ThumbnailURL   string
	RequestBody    map[string]any
	ErrorMessage   string
	CostUSD        float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	FinishedAt     *time.Time
}

// VideoTaskService 视频任务业务服务。
type VideoTaskService struct {
	videoTaskRepo   VideoTaskRepo
	accountService  *AccountService
	gatewayService  *GatewayService
	rateLimitService *RateLimitService
	adapter         VideoAdapter
	logger          *zap.Logger
}

// NewVideoTaskService 创建视频任务服务实例。
func NewVideoTaskService(
	videoTaskRepo VideoTaskRepo,
	accountService *AccountService,
	gatewayService *GatewayService,
	rateLimitService *RateLimitService,
	adapter VideoAdapter,
) *VideoTaskService {
	return &VideoTaskService{
		videoTaskRepo:    videoTaskRepo,
		accountService:   accountService,
		gatewayService:   gatewayService,
		rateLimitService: rateLimitService,
		adapter:          adapter,
		logger:           logger.L(),
	}
}

// CreateTask 处理用户视频生成请求。
// 流程：调度选 channel → model_mapping 翻译 → 调 adapter → 写 video_tasks → 返回。
func (s *VideoTaskService) CreateTask(c *gin.Context, groupID *int64, userID int64, apiKeyID int64, publicModel string, requestBody map[string]any) (*VideoTaskRecord, error) {
	// 1. 解析用户请求参数
	req, err := parseVideoCreateRequest(publicModel, requestBody)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: parse request: %w", err)
	}

	// 2. 调度选择支持该模型的 channel
	account, err := s.gatewayService.SelectAccountForModel(c.Request.Context(), groupID, "", publicModel)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: select account: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("video_task_service: no available account for model %s", publicModel)
	}

	// 3. model_mapping 翻译
	mapping := account.GetModelMapping()
	upstreamModel := publicModel
	if mapped, ok := mapping[publicModel]; ok && mapped != "" {
		upstreamModel = mapped
	}
	req.PublicModel = publicModel
	req.UpstreamModel = upstreamModel

	// 4. 调用上游 adapter
	createResult, err := s.adapter.Create(c.Request.Context(), account, *req)
	if err != nil {
		s.logger.Error("video_task_service: upstream create failed",
			zap.Int64("account_id", account.ID),
			zap.String("model", upstreamModel),
			zap.Error(err))
		return nil, fmt.Errorf("video_task_service: upstream create: %w", err)
	}

	// 上游返回失败
	if createResult.Status == "failed" {
		// 检查是否需要故障转移（429/401/403/5xx）
		// 这里简单记录错误，后续可加自动重试逻辑
		s.logger.Warn("video_task_service: upstream returned failure",
			zap.Int64("account_id", account.ID),
			zap.String("error", createResult.ErrorMessage))
	}

	// 5. 生成 local_id 并写入 video_tasks 表
	localID := generateVideoLocalID()
	record := &VideoTaskRecord{
		LocalID:        localID,
		UserID:         userID,
		APIKeyID:       apiKeyID,
		PublicModel:    publicModel,
		UpstreamModel:  upstreamModel,
		AccountID:      account.ID,
		UpstreamTaskID: createResult.TaskID,
		Status:         createResult.Status,
		Resolution:     req.Resolution,
		DurationSec:    req.DurationSec,
		RequestBody:    requestBody,
		ErrorMessage:   createResult.ErrorMessage,
	}
	if createResult.InlineVideoURL != "" {
		record.VideoURL = createResult.InlineVideoURL
	}

	saved, err := s.videoTaskRepo.Create(c.Request.Context(), record)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: save task: %w", err)
	}

	return saved, nil
}

// GetTask 查询任务状态。如果任务仍在 processing，尝试向上游刷新状态。
func (s *VideoTaskService) GetTask(ctx context.Context, localID string, userID int64) (*VideoTaskRecord, error) {
	record, err := s.videoTaskRepo.GetByLocalID(ctx, localID)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: get task: %w", err)
	}
	if record.UserID != userID {
		return nil, fmt.Errorf("video_task_service: task not found for user")
	}

	// 如果任务还在 processing，尝试刷新
	if record.Status == "processing" && record.UpstreamTaskID != "" {
		s.refreshTaskStatus(ctx, record)
	}

	return record, nil
}

// PollTask 供 Worker 调用：轮询单个 processing 任务的上游状态。
func (s *VideoTaskService) PollTask(ctx context.Context, record *VideoTaskRecord) error {
	if record.Status != "processing" || record.UpstreamTaskID == "" {
		return nil
	}
	return s.refreshTaskStatus(ctx, record)
}

// refreshTaskStatus 向上游查询并更新本地任务状态。
func (s *VideoTaskService) refreshTaskStatus(ctx context.Context, record *VideoTaskRecord) error {
	// 获取 account
	account, err := s.accountService.GetByID(ctx, record.AccountID)
	if err != nil || account == nil {
		return fmt.Errorf("video_task_service: get account %d: %w", record.AccountID, err)
	}

	result, err := s.adapter.GetResult(ctx, account, record.UpstreamTaskID)
	if err != nil {
		return fmt.Errorf("video_task_service: get upstream result: %w", err)
	}

	switch result.Status {
	case "succeeded":
		cost := calculateVideoCost(record.PublicModel, record.Resolution, result.DurationSec)
		if err := s.videoTaskRepo.UpdateResult(ctx, record.ID, "succeeded", result.VideoURL, result.ThumbnailURL, result.DurationSec, cost); err != nil {
			return fmt.Errorf("video_task_service: update result: %w", err)
		}
	case "failed", "cancelled":
		if err := s.videoTaskRepo.UpdateStatus(ctx, record.ID, result.Status, result.ErrorMessage); err != nil {
			return fmt.Errorf("video_task_service: update status: %w", err)
		}
	default:
		// still processing, no update needed
	}

	return nil
}

// ListTasksByUserID 查询指定用户的视频任务列表（分页）。
func (s *VideoTaskService) ListTasksByUserID(ctx context.Context, userID int64, limit, offset int) ([]*VideoTaskRecord, int, error) {
	return s.videoTaskRepo.ListByUserID(ctx, userID, limit, offset)
}

// GetTaskByLocalID 按 local_id 查询任务（管理后台用，不校验 user_id）。
func (s *VideoTaskService) GetTaskByLocalID(ctx context.Context, localID string) (*VideoTaskRecord, error) {
	record, err := s.videoTaskRepo.GetByLocalID(ctx, localID)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: get task by local_id: %w", err)
	}
	return record, nil
}

// GetTaskByID 按主键 ID 查询任务（管理后台用，不校验 user_id）。
func (s *VideoTaskService) GetTaskByID(ctx context.Context, id int64) (*VideoTaskRecord, error) {
	record, err := s.videoTaskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video_task_service: get task by id: %w", err)
	}
	return record, nil
}

// CancelTask 管理员手动取消任务：仅更新本地状态为 cancelled。
// 不调上游 cancel API（上游 cancel 支持不一致），上游任务自然超时或被忽略。
func (s *VideoTaskService) CancelTask(ctx context.Context, id int64) error {
	if err := s.videoTaskRepo.UpdateStatus(ctx, id, "cancelled", "cancelled by admin"); err != nil {
		return fmt.Errorf("video_task_service: cancel task: %w", err)
	}
	return nil
}

// parseVideoCreateRequest 从请求 body 解析统一参数。
func parseVideoCreateRequest(model string, body map[string]any) (*VideoCreateRequest, error) {
	req := &VideoCreateRequest{
		PublicModel: model,
	}

	if v, ok := body["prompt"].(string); ok {
		req.Prompt = v
	}
	if v, ok := body["negative_prompt"].(string); ok {
		req.NegativePrompt = v
	}
	if v, ok := body["resolution"].(string); ok {
		req.Resolution = v
	}
	if v, ok := body["duration"]; ok {
		switch d := v.(type) {
		case float64:
			req.DurationSec = int(d)
		case int:
			req.DurationSec = d
		}
	}
	if v, ok := body["duration_sec"]; ok {
		switch d := v.(type) {
		case float64:
			req.DurationSec = int(d)
		case int:
			req.DurationSec = d
		}
	}
	// 图片参考
	if v, ok := body["image_url"].(string); ok && v != "" {
		req.ImageRefURLs = []string{v}
	}
	if urls, ok := body["image_urls"].([]any); ok {
		for _, u := range urls {
			if s, ok := u.(string); ok {
				req.ImageRefURLs = append(req.ImageRefURLs, s)
			}
		}
	}
	// 视频参考
	if v, ok := body["video_url"].(string); ok && v != "" {
		req.VideoRefURLs = []string{v}
	}
	// 种子
	if v, ok := body["seed"]; ok {
		switch seed := v.(type) {
		case float64:
			s := int64(seed)
			req.Seed = &s
		case int64:
			req.Seed = &seed
		}
	}

	return req, nil
}

// generateVideoLocalID 生成唯一的本地任务 ID，格式：vid_<16hex>。
func generateVideoLocalID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "vid_" + hex.EncodeToString(b)
}

// calculateVideoCost 简单的计费计算：按模型+分辨率+时长。
// 实际单价从分组配置读取，这里先返回 0，后续接入分组计费。
func calculateVideoCost(model, resolution string, durationSec int) float64 {
	// TODO: 从分组 video_model_prices 读取单价
	// 暂时返回 0，不影响功能运行
	_ = model
	_ = resolution
	_ = durationSec
	return 0
}
