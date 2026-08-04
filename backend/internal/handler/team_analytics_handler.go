package handler

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// TeamAnalyticsHandler handles team analytics-related requests
type TeamAnalyticsHandler struct {
	analyticsService service.TeamAnalyticsService
}

// NewTeamAnalyticsHandler creates a new TeamAnalyticsHandler
func NewTeamAnalyticsHandler(analyticsService service.TeamAnalyticsService) *TeamAnalyticsHandler {
	return &TeamAnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// parseDateRange extracts start_date and end_date from query parameters.
// If not provided, defaults to the last 30 days (endDate = today, startDate = today - 30d).
func parseDateRange(c *gin.Context) (time.Time, time.Time, bool) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr == "" && endDateStr == "" {
		return startDate, endDate, true
	}

	if startDateStr != "" {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid start_date format, expected YYYY-MM-DD")
			return time.Time{}, time.Time{}, false
		}
		startDate = parsed
	}

	if endDateStr != "" {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			response.BadRequest(c, "Invalid end_date format, expected YYYY-MM-DD")
			return time.Time{}, time.Time{}, false
		}
		endDate = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 0, parsed.Location())
	}

	return startDate, endDate, true
}

// Overview handles getting team overview analytics
// GET /api/v1/teams/:team_id/analytics/overview
func (h *TeamAnalyticsHandler) Overview(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	startDate, endDate, ok := parseDateRange(c)
	if !ok {
		return
	}

	overview, err := h.analyticsService.GetTeamOverview(c.Request.Context(), teamID, startDate, endDate)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, overview)
}

// DepartmentRanking handles getting department ranking
// GET /api/v1/teams/:team_id/analytics/departments/ranking
func (h *TeamAnalyticsHandler) DepartmentRanking(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	startDate, endDate, ok := parseDateRange(c)
	if !ok {
		return
	}

	ranking, err := h.analyticsService.GetDepartmentRanking(c.Request.Context(), teamID, startDate, endDate)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ranking)
}

// ConsumerRanking handles getting consumer ranking
// GET /api/v1/teams/:team_id/analytics/consumers/ranking
func (h *TeamAnalyticsHandler) ConsumerRanking(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	startDate, endDate, ok := parseDateRange(c)
	if !ok {
		return
	}

	ranking, err := h.analyticsService.GetConsumerRanking(c.Request.Context(), teamID, startDate, endDate)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, ranking)
}

// DailyTrend handles getting daily/hourly trend data
// GET /api/v1/teams/:team_id/analytics/daily-trend
func (h *TeamAnalyticsHandler) DailyTrend(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	startDate, endDate, ok := parseDateRange(c)
	if !ok {
		return
	}

	granularity := c.DefaultQuery("granularity", "day")

	var result any
	if granularity == "hour" {
		result, err = h.analyticsService.GetHourlyTrend(c.Request.Context(), teamID, startDate, endDate)
	} else {
		result, err = h.analyticsService.GetDailyTrend(c.Request.Context(), teamID, startDate, endDate)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, result)
}

// ModelDistribution handles getting model distribution data
// GET /api/v1/teams/:team_id/analytics/models
func (h *TeamAnalyticsHandler) ModelDistribution(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	startDate, endDate, ok := parseDateRange(c)
	if !ok {
		return
	}

	distribution, err := h.analyticsService.GetModelDistribution(c.Request.Context(), teamID, startDate, endDate)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, distribution)
}
