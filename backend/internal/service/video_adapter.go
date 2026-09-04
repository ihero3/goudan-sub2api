package service

// video_adapter.go — 视频生成上游适配器。
// 设计原则（Experience 100012072）：
//   - 不同上游模型的请求体字段契约不同，按模型 builder 构造请求体，不复用 payload 结构。
//   - 通用 OpenAI-compatible 上游走 OpenAIVideoAdapter。
//   - 如某家 API 差异巨大，可新增 Adapter 实现而不影响现有逻辑。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VideoCreateRequest 是视频生成请求的统一内部表示。
type VideoCreateRequest struct {
	PublicModel    string         // 用户传的统一模型名，如 seedance-2.5
	UpstreamModel  string         // model_mapping 翻译后上游实际模型名
	Prompt         string         // 文本提示词
	NegativePrompt string         // 负面提示词
	ImageRefURLs   []string       // 参考图 URL 列表（图生视频）
	VideoRefURLs   []string       // 参考视频 URL 列表（视频续生）
	Resolution     string         // 分辨率：480p / 720p / 1080p
	DurationSec    int            // 时长（秒）
	Seed           *int64         // 随机种子
	Extra          map[string]any // 其他上游专有参数透传
}

// VideoCreateResult 是创建任务后上游的即时响应。
type VideoCreateResult struct {
	TaskID         string // 上游返回的异步任务 ID
	InlineVideoURL string // 如果同步直接出 URL（无异步）
	Status         string // processing / succeeded / failed
	ErrorMessage   string
}

// VideoTaskResult 是查询任务状态/结果的响应。
type VideoTaskResult struct {
	Status       string // processing / succeeded / failed / cancelled
	VideoURL     string
	ThumbnailURL string
	DurationSec  int
	ErrorMessage string
	StatusCode   int // 上游 HTTP 状态码（用于故障转移判定）
}

// VideoAdapter 视频生成上游适配器接口。
type VideoAdapter interface {
	// Create 向上游提交视频生成任务。
	Create(ctx context.Context, account *Account, req VideoCreateRequest) (*VideoCreateResult, error)
	// GetResult 查询上游任务状态和结果。
	GetResult(ctx context.Context, account *Account, upstreamTaskID string) (*VideoTaskResult, error)
}

// OpenAIVideoAdapter 是通用 OpenAI-compatible 视频上游适配器。
// 读取 Account 的 credentials["base_url"] + credentials["api_key"]，
// 向 {base_url}/videos/generations POST 创建任务，
// 向 {base_url}/videos/{task_id} GET 查询状态。
type OpenAIVideoAdapter struct {
	httpClient *http.Client
}

// NewOpenAIVideoAdapter 创建通用视频适配器实例。
func NewOpenAIVideoAdapter() *OpenAIVideoAdapter {
	return &OpenAIVideoAdapter{
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Create 向上游提交视频生成任务。
func (a *OpenAIVideoAdapter) Create(ctx context.Context, account *Account, req VideoCreateRequest) (*VideoCreateResult, error) {
	baseURL := strings.TrimRight(account.GetCredential("base_url"), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("video_adapter: account %d has no base_url", account.ID)
	}
	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("video_adapter: account %d has no api_key", account.ID)
	}

	// 按模型 builder 构造请求体（不同模型字段契约不同）
	body := buildVideoCreateBody(req)
	url := baseURL + "/videos/generations"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("video_adapter: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("video_adapter: create do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return &VideoCreateResult{
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("upstream returned %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	// 解析上游响应
	var upstreamResp struct {
		ID      string `json:"id"`
		TaskID  string `json:"task_id"`
		Status  string `json:"status"`
		Error   string `json:"error"`
		Video   string `json:"video_url"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &upstreamResp); err != nil {
		return nil, fmt.Errorf("video_adapter: unmarshal create response: %w", err)
	}

	taskID := upstreamResp.ID
	if taskID == "" {
		taskID = upstreamResp.TaskID
	}
	status := upstreamResp.Status
	if status == "" {
		status = "processing"
	}

	result := &VideoCreateResult{
		TaskID:  taskID,
		Status:  status,
	}
	// 同步返回结果的情况
	if status == "succeeded" {
		result.InlineVideoURL = upstreamResp.Video
		if result.InlineVideoURL == "" {
			result.InlineVideoURL = upstreamResp.URL
		}
	}
	if upstreamResp.Error != "" {
		result.ErrorMessage = upstreamResp.Error
	}

	return result, nil
}

// GetResult 查询上游任务状态和结果。
func (a *OpenAIVideoAdapter) GetResult(ctx context.Context, account *Account, upstreamTaskID string) (*VideoTaskResult, error) {
	baseURL := strings.TrimRight(account.GetCredential("base_url"), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("video_adapter: account %d has no base_url", account.ID)
	}
	apiKey := account.GetCredential("api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("video_adapter: account %d has no api_key", account.ID)
	}

	url := fmt.Sprintf("%s/videos/%s", baseURL, upstreamTaskID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("video_adapter: get request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("video_adapter: get do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	result := &VideoTaskResult{
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode >= 400 {
		result.Status = "failed"
		result.ErrorMessage = fmt.Sprintf("upstream returned %d: %s", resp.StatusCode, string(respBody))
		return result, nil
	}

	var upstreamResp struct {
		Status       string `json:"status"`
		VideoURL     string `json:"video_url"`
		URL          string `json:"url"`
		ThumbnailURL string `json:"thumbnail_url"`
		Duration     int    `json:"duration_sec"`
		DurationSec  int    `json:"durationSec"`
		Error        string `json:"error"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.Unmarshal(respBody, &upstreamResp); err != nil {
		return nil, fmt.Errorf("video_adapter: unmarshal get response: %w", err)
	}

	result.Status = upstreamResp.Status
	result.VideoURL = upstreamResp.VideoURL
	if result.VideoURL == "" {
		result.VideoURL = upstreamResp.URL
	}
	result.ThumbnailURL = upstreamResp.ThumbnailURL
	result.DurationSec = upstreamResp.Duration
	if result.DurationSec == 0 {
		result.DurationSec = upstreamResp.DurationSec
	}
	if upstreamResp.Error != "" {
		result.ErrorMessage = upstreamResp.Error
	} else if upstreamResp.ErrorMessage != "" {
		result.ErrorMessage = upstreamResp.ErrorMessage
	}

	return result, nil
}

// buildVideoCreateBody 按模型构造请求体。
// 不同上游模型的字段命名可能不同，这里用通用 OpenAI 格式，
// 额外字段通过 Extra 透传。如某模型需要特殊字段名，在此分支处理。
func buildVideoCreateBody(req VideoCreateRequest) []byte {
	body := map[string]any{
		"model":  req.UpstreamModel,
		"prompt": req.Prompt,
	}
	if req.NegativePrompt != "" {
		body["negative_prompt"] = req.NegativePrompt
	}
	if len(req.ImageRefURLs) > 0 {
		body["image_url"] = req.ImageRefURLs[0]   // 单图
		body["image_urls"] = req.ImageRefURLs      // 多图
	}
	if len(req.VideoRefURLs) > 0 {
		body["video_url"] = req.VideoRefURLs[0]
	}
	if req.Resolution != "" {
		body["resolution"] = req.Resolution
	}
	if req.DurationSec > 0 {
		body["duration"] = req.DurationSec
		body["duration_sec"] = req.DurationSec
	}
	if req.Seed != nil {
		body["seed"] = *req.Seed
	}
	// 透传上游专有参数
	for k, v := range req.Extra {
		body[k] = v
	}

	data, _ := json.Marshal(body)
	return data
}
