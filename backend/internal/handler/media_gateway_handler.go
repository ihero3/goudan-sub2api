package handler

// media_gateway_handler.go — 用户侧统一媒体生成网关（图片 / 视频 / 音频）。
// 端点：
//   POST /v1/media/generations   创建媒体生成任务
//   GET  /v1/media/:id           查询任务状态
//   GET  /v1/media/:id/content   获取产物内容（302）
//
// 低耦合：仅依赖 service.MediaTaskService，不直接访问 repository 或 adapter。

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MediaGatewayHandler 用户侧统一媒体生成网关 handler。
type MediaGatewayHandler struct {
	mediaTaskService    *service.MediaTaskService
	billingCacheService *service.BillingCacheService
	logger              *zap.Logger
}

// NewMediaGatewayHandler 创建 handler 实例。
func NewMediaGatewayHandler(mediaTaskService *service.MediaTaskService, billingCacheService *service.BillingCacheService) *MediaGatewayHandler {
	return &MediaGatewayHandler{
		mediaTaskService:    mediaTaskService,
		billingCacheService: billingCacheService,
		logger:              logger.L(),
	}
}

// Create POST /v1/media/generations
// 请求体：{ "model": "wan3.0-video", "prompt": "...", ... }
// 响应体：{ "id": "vid_xxx", "status": "processing", "model": "wan3.0-video" }
func (h *MediaGatewayHandler) Create(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		mediaErrorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		mediaErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}
	subscription, _ := middleware2.GetSubscriptionFromContext(c)

	bodyBytes, err := readMediaRequestBody(c)
	if err != nil {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(bodyBytes) == 0 {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	var body map[string]any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid JSON body")
		return
	}

	publicModel, _ := body["model"].(string)
	if strings.TrimSpace(publicModel) == "" {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}
	kind := service.MediaKindFromModel(publicModel, body)

	// 调用前额度/余额检查：余额或平台配额不足时直接拒绝，避免白白调用上游。
	if h.billingCacheService != nil {
		if err := h.billingCacheService.CheckBillingEligibility(
			c.Request.Context(), apiKey.User, apiKey, apiKey.Group, subscription,
			service.QuotaPlatform(c.Request.Context(), apiKey),
		); err != nil {
			reqLog := h.logger.With(
				zap.Int64("user_id", subject.UserID),
				zap.Int64("api_key_id", apiKey.ID),
				zap.String("model", publicModel),
				zap.String("kind", string(kind)),
			)
			reqLog.Info("media_gateway.billing_eligibility_check_failed", zap.Error(err))
			status, code, message, retryAfter := billingErrorDetails(err)
			if retryAfter > 0 {
				c.Header("Retry-After", strconv.Itoa(retryAfter))
			}
			mediaErrorResponse(c, status, code, message)
			return
		}
	}

	reqLog := h.logger.With(
		zap.String("component", "handler.media_gateway"),
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
		zap.String("model", publicModel),
		zap.String("kind", string(kind)),
	)

	record, err := h.mediaTaskService.CreateTask(c, kind, apiKey.GroupID, subject.UserID, apiKey.ID, publicModel, body)
	if err != nil {
		reqLog.Error("media_gateway.create_task_failed", zap.Error(err))
		if strings.Contains(err.Error(), "no available account") {
			mediaErrorResponse(c, http.StatusServiceUnavailable, "capacity_error", "No available media generation channels")
			return
		}
		mediaErrorResponse(c, http.StatusBadGateway, "api_error", "Media generation request failed")
		return
	}

	reqLog.Info("media_gateway.create_task_succeeded",
		zap.String("local_id", record.LocalID),
		zap.String("status", record.Status),
	)

	response := mediaTaskResponse{
		ID:        record.LocalID,
		Status:    record.Status,
		Model:     record.PublicModel,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}
	if record.MediaURL != "" {
		response.URL = record.MediaURL
	}
	if record.ErrorMessage != "" {
		response.Error = record.ErrorMessage
	}

	c.JSON(http.StatusAccepted, response)
}

// Get GET /v1/media/:id
func (h *MediaGatewayHandler) Get(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		mediaErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}

	taskID := mediaTaskIDParam(c)
	if taskID == "" {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}

	record, err := h.mediaTaskService.GetTask(c.Request.Context(), taskID, subject.UserID)
	if err != nil {
		h.logger.Warn("media_gateway.get_task_failed",
			zap.String("task_id", taskID),
			zap.Int64("user_id", subject.UserID),
			zap.Error(err),
		)
		mediaErrorResponse(c, http.StatusNotFound, "not_found_error", "Media task not found")
		return
	}

	response := mediaTaskResponse{
		ID:        record.LocalID,
		Status:    record.Status,
		Model:     record.PublicModel,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}
	if record.MediaURL != "" {
		response.URL = record.MediaURL
	}
	if record.ThumbnailURL != "" {
		response.ThumbnailURL = record.ThumbnailURL
	}
	if record.DurationSec > 0 {
		response.DurationSec = record.DurationSec
	}
	if record.Resolution != "" {
		response.Resolution = record.Resolution
	}
	if record.ErrorMessage != "" {
		response.Error = record.ErrorMessage
	}
	if record.FinishedAt != nil {
		response.FinishedAt = record.FinishedAt.UTC().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, response)
}

// GetContent GET /v1/media/:id/content
func (h *MediaGatewayHandler) GetContent(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		mediaErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}

	taskID := mediaTaskIDParam(c)
	if taskID == "" {
		mediaErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}

	record, err := h.mediaTaskService.GetTask(c.Request.Context(), taskID, subject.UserID)
	if err != nil {
		mediaErrorResponse(c, http.StatusNotFound, "not_found_error", "Media task not found")
		return
	}

	if record.Status != "succeeded" || record.MediaURL == "" {
		mediaErrorResponse(c, http.StatusNotFound, "not_found_error", "Media content not available")
		return
	}

	c.Redirect(http.StatusFound, record.MediaURL)
}

// --- 响应结构 ---

type mediaTaskResponse struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Model        string `json:"model"`
	URL          string `json:"url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Resolution   string `json:"resolution,omitempty"`
	DurationSec  int    `json:"duration_sec,omitempty"`
	Error        string `json:"error,omitempty"`
	CreatedAt    string `json:"created_at"`
	FinishedAt   string `json:"finished_at,omitempty"`
}

// --- 辅助函数 ---

func mediaErrorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

func readMediaRequestBody(c *gin.Context) ([]byte, error) {
	const maxMediaBodySize = 10 * 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxMediaBodySize)
	return io.ReadAll(c.Request.Body)
}

func mediaTaskIDParam(c *gin.Context) string {
	if taskID := strings.TrimSpace(c.Param("task_id")); taskID != "" {
		return taskID
	}
	return strings.TrimSpace(c.Param("id"))
}
