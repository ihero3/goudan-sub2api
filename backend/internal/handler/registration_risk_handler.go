package handler

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegistrationClickCaptchaHandler 自建顺序点击验证码接口。
type RegistrationClickCaptchaHandler struct {
	clickCaptcha *service.RegistrationClickCaptchaService
}

// NewRegistrationClickCaptchaHandler 创建 handler。
func NewRegistrationClickCaptchaHandler(clickCaptcha *service.RegistrationClickCaptchaService) *RegistrationClickCaptchaHandler {
	return &RegistrationClickCaptchaHandler{clickCaptcha: clickCaptcha}
}

// clickCaptchaRequestBody 请求体公共字段。
type clickCaptchaRequestBody struct {
	ChallengeID string   `json:"challenge_id"`
	Clicks      []string `json:"clicks"`
}

// Challenge POST /api/v1/auth/captcha/challenge
func (h *RegistrationClickCaptchaHandler) Challenge(c *gin.Context) {
	if h == nil || h.clickCaptcha == nil {
		response.InternalError(c, "Click captcha service not configured")
		return
	}
	ipHash := service.HashFingerprint(ip.GetClientIP(c), c.Request.UserAgent())
	challenge, err := h.clickCaptcha.CreateChallenge(c.Request.Context(), ipHash, ipHash)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, challenge)
}

// Verify POST /api/v1/auth/captcha/verify
func (h *RegistrationClickCaptchaHandler) Verify(c *gin.Context) {
	if h == nil || h.clickCaptcha == nil {
		response.InternalError(c, "Click captcha service not configured")
		return
	}
	var req clickCaptchaRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if strings.TrimSpace(req.ChallengeID) == "" || len(req.Clicks) == 0 {
		response.BadRequest(c, "Invalid request: challenge_id and clicks are required")
		return
	}
	ipHash := service.HashFingerprint(ip.GetClientIP(c), c.Request.UserAgent())
	token, ttl, err := h.clickCaptcha.VerifyChallenge(c.Request.Context(), req.ChallengeID, req.Clicks, ipHash, ipHash)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"captcha_token": token,
		"expires_in":    ttl,
	})
}
