package handler

// video_gateway_handler.go — 用户侧视频生成网关 handler。
// 端点：
//   POST /v1/videos/generations  创建视频生成任务
//   GET  /v1/videos/:task_id      查询任务状态
//   GET  /v1/videos/:task_id/content  获取视频内容（302 重定向到 video_url）
// 低耦合：仅依赖 service.VideoTaskService，不直接访问 repository 或 adapter。

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// VideoGatewayHandler 用户侧视频生成网关 handler。
type VideoGatewayHandler struct {
	videoTaskService *service.VideoTaskService
	logger           *zap.Logger
}

// NewVideoGatewayHandler 创建 handler 实例。
func NewVideoGatewayHandler(videoTaskService *service.VideoTaskService) *VideoGatewayHandler {
	return &VideoGatewayHandler{
		videoTaskService: videoTaskService,
		logger:           logger.L(),
	}
}

// CreateTask POST /v1/videos/generations
// 请求体：{ "model": "seedance-2.5", "prompt": "...", "resolution": "720p", "duration": 5, ... }
// 响应体：{ "id": "vid_xxx", "status": "processing", "model": "seedance-2.5" }
func (h *VideoGatewayHandler) CreateTask(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoErrorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		videoErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}

	// 读取请求体
	bodyBytes, err := readVideoRequestBody(c)
	if err != nil {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}
	if len(bodyBytes) == 0 {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Request body is empty")
		return
	}

	var body map[string]any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "Invalid JSON body")
		return
	}

	// 提取 model 字段
	publicModel, _ := body["model"].(string)
	if strings.TrimSpace(publicModel) == "" {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "model is required")
		return
	}

	reqLog := h.logger.With(
		zap.String("component", "handler.video_gateway"),
		zap.Int64("user_id", subject.UserID),
		zap.Int64("api_key_id", apiKey.ID),
		zap.Any("group_id", apiKey.GroupID),
		zap.String("model", publicModel),
	)

	// 调用 service 创建任务
	record, err := h.videoTaskService.CreateTask(c, apiKey.GroupID, subject.UserID, apiKey.ID, publicModel, body)
	if err != nil {
		reqLog.Error("video_gateway.create_task_failed", zap.Error(err))
		// 不向上游用户暴露上游错误细节，只返回平台统一错误
		if strings.Contains(err.Error(), "no available account") {
			videoErrorResponse(c, http.StatusServiceUnavailable, "capacity_error", "No available video generation channels")
			return
		}
		videoErrorResponse(c, http.StatusBadGateway, "api_error", "Video generation request failed")
		return
	}

	reqLog.Info("video_gateway.create_task_succeeded",
		zap.String("local_id", record.LocalID),
		zap.String("status", record.Status),
	)

	// 返回 OpenAI 风格异步任务响应
	response := videoTaskResponse{
		ID:        record.LocalID,
		Status:    record.Status,
		Model:     record.PublicModel,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}
	if record.VideoURL != "" {
		response.URL = record.VideoURL
	}
	if record.ErrorMessage != "" {
		response.Error = record.ErrorMessage
	}

	c.JSON(http.StatusAccepted, response)
}

// GetTask GET /v1/videos/:task_id
// 响应体：{ "id": "vid_xxx", "status": "succeeded", "model": "seedance-2.5", "video_url": "..." }
func (h *VideoGatewayHandler) GetTask(c *gin.Context) {
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok {
		videoErrorResponse(c, http.StatusUnauthorized, "authentication_error", "Invalid API key")
		return
	}
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		videoErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}

	taskID := videoTaskIDParam(c)
	if taskID == "" {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}

	record, err := h.videoTaskService.GetTask(c.Request.Context(), taskID, subject.UserID)
	if err != nil {
		h.logger.Warn("video_gateway.get_task_failed",
			zap.String("task_id", taskID),
			zap.Int64("user_id", subject.UserID),
			zap.Int64("api_key_id", apiKey.ID),
			zap.Error(err),
		)
		videoErrorResponse(c, http.StatusNotFound, "not_found_error", "Video task not found")
		return
	}

	response := videoTaskResponse{
		ID:        record.LocalID,
		Status:    record.Status,
		Model:     record.PublicModel,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}
	if record.VideoURL != "" {
		response.URL = record.VideoURL
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

// GetTaskContent GET /v1/videos/:task_id/content
// 302 重定向到实际视频 URL，或 404 如果任务未完成/无视频 URL。
func (h *VideoGatewayHandler) GetTaskContent(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		videoErrorResponse(c, http.StatusUnauthorized, "authentication_error", "User context not found")
		return
	}

	taskID := videoTaskIDParam(c)
	if taskID == "" {
		videoErrorResponse(c, http.StatusBadRequest, "invalid_request_error", "task_id is required")
		return
	}

	record, err := h.videoTaskService.GetTask(c.Request.Context(), taskID, subject.UserID)
	if err != nil {
		videoErrorResponse(c, http.StatusNotFound, "not_found_error", "Video task not found")
		return
	}

	if record.Status != "succeeded" || record.VideoURL == "" {
		videoErrorResponse(c, http.StatusNotFound, "not_found_error", "Video content not available")
		return
	}

	c.Redirect(http.StatusFound, record.VideoURL)
}

// --- 响应结构 ---

type videoTaskResponse struct {
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

// videoErrorResponse 返回 OpenAI 风格错误响应。
func videoErrorResponse(c *gin.Context, status int, errType, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"type":    errType,
			"message": message,
		},
	})
}

// readVideoRequestBody 读取请求体（最多 10MB）。
func readVideoRequestBody(c *gin.Context) ([]byte, error) {
	const maxVideoBodySize = 10 * 1024 * 1024 // 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxVideoBodySize)
	return io.ReadAll(c.Request.Body)
}

// videoTaskIDParam resolves a local task id from either the dedicated
// /video-tasks/:task_id route or the unified /videos/:request_id route.
func videoTaskIDParam(c *gin.Context) string {
	if taskID := strings.TrimSpace(c.Param("task_id")); taskID != "" {
		return taskID
	}
	return strings.TrimSpace(c.Param("request_id"))
}
