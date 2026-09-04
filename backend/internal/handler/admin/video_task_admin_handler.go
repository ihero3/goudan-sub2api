package admin

// video_task_admin_handler.go — 管理后台视频任务 handler。
// 端点：
//   GET  /admin/video-tasks           列表（分页 + 筛选）
//   GET  /admin/video-tasks/:id       详情
//   POST /admin/video-tasks/:id/cancel  取消任务
// 低耦合：仅依赖 service.VideoTaskService，通过接口方法访问。

import (
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// VideoTaskAdminHandler 管理后台视频任务 handler。
type VideoTaskAdminHandler struct {
	videoTaskService *service.VideoTaskService
}

// NewVideoTaskAdminHandler 创建 handler 实例。
func NewVideoTaskAdminHandler(videoTaskService *service.VideoTaskService) *VideoTaskAdminHandler {
	return &VideoTaskAdminHandler{videoTaskService: videoTaskService}
}

// --- 响应 DTO ---

type videoTaskAdminResponse struct {
	ID             int64  `json:"id"`
	LocalID        string `json:"local_id"`
	UserID         int64  `json:"user_id"`
	APIKeyID       int64  `json:"api_key_id"`
	PublicModel    string `json:"public_model"`
	UpstreamModel  string `json:"upstream_model"`
	AccountID      int64  `json:"account_id"`
	UpstreamTaskID string `json:"upstream_task_id"`
	Status         string `json:"status"`
	Resolution     string `json:"resolution"`
	DurationSec    int    `json:"duration_sec"`
	VideoURL       string `json:"video_url"`
	ThumbnailURL   string `json:"thumbnail_url"`
	ErrorMessage   string `json:"error_message"`
	CostUSD        float64 `json:"cost_usd"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	FinishedAt     string `json:"finished_at"`
}

func toVideoTaskAdminResponse(t *service.VideoTaskRecord) *videoTaskAdminResponse {
	if t == nil {
		return nil
	}
	finishedAt := ""
	if t.FinishedAt != nil {
		finishedAt = t.FinishedAt.UTC().Format(time.RFC3339)
	}
	return &videoTaskAdminResponse{
		ID:             t.ID,
		LocalID:        t.LocalID,
		UserID:         t.UserID,
		APIKeyID:       t.APIKeyID,
		PublicModel:    t.PublicModel,
		UpstreamModel:  t.UpstreamModel,
		AccountID:      t.AccountID,
		UpstreamTaskID: t.UpstreamTaskID,
		Status:         t.Status,
		Resolution:     t.Resolution,
		DurationSec:    t.DurationSec,
		VideoURL:       t.VideoURL,
		ThumbnailURL:   t.ThumbnailURL,
		ErrorMessage:   t.ErrorMessage,
		CostUSD:        t.CostUSD,
		CreatedAt:      t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.UTC().Format(time.RFC3339),
		FinishedAt:     finishedAt,
	}
}

// --- Handlers ---

// List GET /admin/video-tasks?page=1&page_size=20&status=processing&user_id=123
func (h *VideoTaskAdminHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	statusFilter := strings.TrimSpace(c.Query("status"))
	userIDFilter, _ := strconv.ParseInt(c.Query("user_id"), 10, 64)

	// 当前 VideoTaskRepo.ListByUserID 仅按 user_id 查询。
	// 管理后台需要全量列表 + 多条件筛选，这里先通过 user_id 路由：
	// - user_id > 0：按用户查
	// - user_id == 0：暂不支持全量列表（需后续扩展 repo 方法）
	if userIDFilter <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", "user_id filter is required for now"))
		return
	}

	items, total, err := h.videoTaskService.ListTasksByUserID(c.Request.Context(), userIDFilter, pageSize, (page-1)*pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*videoTaskAdminResponse, 0, len(items))
	for _, t := range items {
		// 可选 status 过滤（内存过滤，数据量大时应下推到 repo）
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}
		out = append(out, toVideoTaskAdminResponse(t))
	}

	totalInt64 := int64(total)
	pages := total / pageSize
	if total%pageSize > 0 {
		pages++
	}

	response.Success(c, response.PaginatedData{
		Items:    out,
		Total:    totalInt64,
		Page:     page,
		PageSize: pageSize,
		Pages:    pages,
	})
}

// Get GET /admin/video-tasks/:id
func (h *VideoTaskAdminHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		// 也支持按 local_id 查询
		record, err := h.videoTaskService.GetTaskByLocalID(c.Request.Context(), idStr)
		if err != nil {
			response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TASK_ID", "invalid task id"))
			return
		}
		response.Success(c, toVideoTaskAdminResponse(record))
		return
	}

	record, err := h.videoTaskService.GetTaskByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, toVideoTaskAdminResponse(record))
}

// Cancel POST /admin/video-tasks/:id/cancel
// 管理员手动取消任务（仅更新本地状态，不调上游 cancel API）。
func (h *VideoTaskAdminHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TASK_ID", "invalid task id"))
		return
	}

	if err := h.videoTaskService.CancelTask(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": "cancelled"})
}
