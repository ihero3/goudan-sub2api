package admin

// media_task_admin_handler.go — 管理后台媒体任务 handler（图片 / 视频 / 音频）。
// 端点：
//   GET  /admin/media-tasks          列表（分页 + 筛选）
//   GET  /admin/media-tasks/:id      详情
//   POST /admin/media-tasks/:id/cancel  取消任务
// 低耦合：仅依赖 service.MediaTaskService。

import (
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// MediaTaskAdminHandler 管理后台媒体任务 handler。
type MediaTaskAdminHandler struct {
	mediaTaskService *service.MediaTaskService
}

// NewMediaTaskAdminHandler 创建 handler 实例。
func NewMediaTaskAdminHandler(mediaTaskService *service.MediaTaskService) *MediaTaskAdminHandler {
	return &MediaTaskAdminHandler{mediaTaskService: mediaTaskService}
}

// mediaTaskAdminResponse 管理后台响应 DTO。
type mediaTaskAdminResponse struct {
	ID             int64   `json:"id"`
	LocalID        string  `json:"local_id"`
	MediaKind      string  `json:"media_kind"`
	UserID         int64   `json:"user_id"`
	APIKeyID       int64   `json:"api_key_id"`
	PublicModel    string  `json:"public_model"`
	UpstreamModel  string  `json:"upstream_model"`
	AccountID      int64   `json:"account_id"`
	UpstreamTaskID string  `json:"upstream_task_id"`
	Status         string  `json:"status"`
	Resolution     string  `json:"resolution"`
	DurationSec    int     `json:"duration_sec"`
	MediaURL       string  `json:"media_url"`
	ThumbnailURL   string  `json:"thumbnail_url"`
	ErrorMessage   string  `json:"error_message"`
	CostUSD        float64 `json:"cost_usd"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	FinishedAt     string  `json:"finished_at"`
}

func toMediaTaskAdminResponse(t *service.MediaTaskRecord) *mediaTaskAdminResponse {
	if t == nil {
		return nil
	}
	finishedAt := ""
	if t.FinishedAt != nil {
		finishedAt = t.FinishedAt.UTC().Format(time.RFC3339)
	}
	return &mediaTaskAdminResponse{
		ID:             t.ID,
		LocalID:        t.LocalID,
		MediaKind:      string(t.MediaKind),
		UserID:         t.UserID,
		APIKeyID:       t.APIKeyID,
		PublicModel:    t.PublicModel,
		UpstreamModel:  t.UpstreamModel,
		AccountID:      t.AccountID,
		UpstreamTaskID: t.UpstreamTaskID,
		Status:         t.Status,
		Resolution:     t.Resolution,
		DurationSec:    t.DurationSec,
		MediaURL:       t.MediaURL,
		ThumbnailURL:   t.ThumbnailURL,
		ErrorMessage:   t.ErrorMessage,
		CostUSD:        t.CostUSD,
		CreatedAt:      t.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:      t.UpdatedAt.UTC().Format(time.RFC3339),
		FinishedAt:     finishedAt,
	}
}

// List GET /admin/media-tasks?page=1&page_size=20&status=processing&user_id=123
func (h *MediaTaskAdminHandler) List(c *gin.Context) {
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

	if userIDFilter <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("VALIDATION_ERROR", "user_id filter is required for now"))
		return
	}

	items, total, err := h.mediaTaskService.ListTasksByUserID(c.Request.Context(), userIDFilter, pageSize, (page-1)*pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*mediaTaskAdminResponse, 0, len(items))
	for _, t := range items {
		if statusFilter != "" && t.Status != statusFilter {
			continue
		}
		out = append(out, toMediaTaskAdminResponse(t))
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

// Get GET /admin/media-tasks/:id
func (h *MediaTaskAdminHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		record, err := h.mediaTaskService.GetTaskByLocalID(c.Request.Context(), idStr)
		if err != nil {
			response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TASK_ID", "invalid task id"))
			return
		}
		response.Success(c, toMediaTaskAdminResponse(record))
		return
	}

	record, err := h.mediaTaskService.GetTaskByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, toMediaTaskAdminResponse(record))
}

// Cancel POST /admin/media-tasks/:id/cancel
func (h *MediaTaskAdminHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.ErrorFrom(c, infraerrors.BadRequest("INVALID_TASK_ID", "invalid task id"))
		return
	}

	if err := h.mediaTaskService.CancelTask(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": "cancelled"})
}
