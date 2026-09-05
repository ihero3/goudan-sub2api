package service

// media_audio_file.go — OpenAI 兼容音频文件端点（/audio/transcriptions、/audio/translations）。
// 这些端点是 multipart 文件上传（语音转文本），与 /audio/speech（文本转音频）不同。
// 这里复用账号调度，直接用账号 base_url/api_key 转发到上游同名端点。

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// ResolveAudioTranscription 转发 multipart 音频文件到上游 /audio/transcriptions。
// 返回上游响应体（OpenAI 兼容 JSON，通常为 {"text":"..."}）。
func (s *MediaTaskService) ResolveAudioTranscription(ctx context.Context, groupID *int64, publicModel, endpoint string, fileBytes []byte, filename, contentType string, extraForm map[string]string) ([]byte, error) {
	if len(fileBytes) == 0 {
		return nil, fmt.Errorf("media_audio_file: empty audio file")
	}
	excluded := make(map[int64]struct{})
	const maxAttempts = 100

	for attempt := 0; attempt < maxAttempts; attempt++ {
		account, selectErr := s.gatewayService.SelectAccountForModelWithExclusions(ctx, groupID, "", publicModel, excluded)
		if selectErr != nil || account == nil {
			if selectErr == nil {
				selectErr = fmt.Errorf("media_audio_file: no available account for model %s", publicModel)
			}
			return nil, fmt.Errorf("media_audio_file: select account: %w", selectErr)
		}
		upstreamModel := publicModel
		if mapped := account.GetMappedModel(publicModel); mapped != "" {
			upstreamModel = mapped
		}

		baseURL := strings.TrimRight(account.GetCredential("base_url"), "/")
		if baseURL == "" {
			return nil, fmt.Errorf("media_audio_file: account %d has no base_url", account.ID)
		}
		apiKey := account.GetCredential("api_key")
		if apiKey == "" {
			return nil, fmt.Errorf("media_audio_file: account %d has no api_key", account.ID)
		}
		if endpoint == "" {
			endpoint = "/v1/audio/transcriptions"
		}
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/" + endpoint
		}
		url := baseURL + endpoint

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			return nil, fmt.Errorf("media_audio_file: create form file: %w", err)
		}
		if _, err := part.Write(fileBytes); err != nil {
			return nil, fmt.Errorf("media_audio_file: write file part: %w", err)
		}
		_ = writer.WriteField("model", upstreamModel)
		for k, v := range extraForm {
			_ = writer.WriteField(k, v)
		}
		_ = writer.Close()

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
		if err != nil {
			return nil, fmt.Errorf("media_audio_file: new request: %w", err)
		}
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
		account.ApplyHeaderOverrides(httpReq.Header)

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			excluded[account.ID] = struct{}{}
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("media_audio_file: upstream returned %d: %s", resp.StatusCode, string(respBody))
		}
		return respBody, nil
	}
	return nil, fmt.Errorf("media_audio_file: upstream account switches exhausted")
}
