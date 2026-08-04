package service

import (
	"context"
	"fmt"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrInvalidAnalyticsDateRange = infraerrors.BadRequest("INVALID_DATE_RANGE", "invalid date range for analytics")
	ErrInvalidTeamIDForAnalytics = infraerrors.BadRequest("INVALID_TEAM_ID", "invalid team id for analytics")
	ErrAnalyticsDataNotFound     = infraerrors.NotFound("ANALYTICS_DATA_NOT_FOUND", "analytics data not found")
)

// DailyTrendItem 每日趋势数据项
type DailyTrendItem struct {
	Date            time.Time `json:"date"`
	TotalRequests   int64     `json:"total_requests"`
	InputTokens     int64     `json:"input_tokens"`
	OutputTokens    int64     `json:"output_tokens"`
	TotalCost       float64   `json:"total_cost"`
	ActualCost      float64   `json:"actual_cost"`
}

// HourlyTrendItem 小时级趋势数据项
type HourlyTrendItem struct {
	Date          time.Time `json:"date"`
	TotalRequests int64     `json:"total_requests"`
	InputTokens   int64     `json:"input_tokens"`
	OutputTokens  int64     `json:"output_tokens"`
	TotalCost     float64   `json:"total_cost"`
	ActualCost    float64   `json:"actual_cost"`
}

// TeamUsageHourly 小时级聚合数据（用于 repository 层返回原始查询结果）
type TeamUsageHourly struct {
	BucketHour    time.Time
	TotalRequests int64
	InputTokens   int64
	OutputTokens  int64
	TotalCost     float64
	ActualCost    float64
}

// ModelDistributionItem 模型分布数据项
type ModelDistributionItem struct {
	ModelName       string  `json:"model_name"`
	TotalRequests   int64   `json:"total_requests"`
	InputTokens     int64   `json:"input_tokens"`
	OutputTokens    int64   `json:"output_tokens"`
	TotalCost       float64 `json:"total_cost"`
	ActualCost      float64 `json:"actual_cost"`
}

// TeamAnalyticsRepository 团队分析数据访问接口（在 service 包内定义以避免循环依赖）
type TeamAnalyticsRepository interface {
	AggregateTeamDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateDeptDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateConsumerDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateModelDaily(ctx context.Context, teamID int64, bucketDate time.Time) error

	QueryTeamDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*TeamUsageTeamDaily, error)
	QueryTeamHourly(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*TeamUsageHourly, error)
	QueryDeptDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*TeamUsageDeptDaily, error)
	QueryConsumerDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*TeamUsageConsumerDaily, error)
	QueryModelDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*TeamUsageModelDaily, error)

	GetTeamOverview(ctx context.Context, teamID int64, startDate, endDate time.Time) (*TeamOverview, error)
	GetDepartmentRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*DeptRankingItem, error)
	GetConsumerRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*ConsumerRankingItem, error)
}

// TeamAnalyticsService 团队分析服务接口
type TeamAnalyticsService interface {
	GetTeamOverview(ctx context.Context, teamID int64, startDate, endDate time.Time) (*TeamOverview, error)
	GetDepartmentRanking(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*DeptRankingItem, error)
	GetConsumerRanking(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*ConsumerRankingItem, error)
	GetDailyTrend(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*DailyTrendItem, error)
	GetHourlyTrend(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*HourlyTrendItem, error)
	GetModelDistribution(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*ModelDistributionItem, error)
	RefreshAggregates(ctx context.Context, teamID int64, date time.Time) error
}

// teamAnalyticsService 团队分析服务实现
type teamAnalyticsService struct {
	analyticsRepo TeamAnalyticsRepository
}

// NewTeamAnalyticsService 创建团队分析服务实例
func NewTeamAnalyticsService(analyticsRepo TeamAnalyticsRepository) TeamAnalyticsService {
	return &teamAnalyticsService{
		analyticsRepo: analyticsRepo,
	}
}

func (s *teamAnalyticsService) validateDateRange(startDate, endDate time.Time) error {
	if startDate.IsZero() || endDate.IsZero() {
		return ErrInvalidAnalyticsDateRange
	}
	if startDate.After(endDate) {
		return infraerrors.BadRequest("START_DATE_AFTER_END_DATE", "start date cannot be after end date")
	}
	maxRange := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxRange {
		return infraerrors.BadRequest("DATE_RANGE_TOO_LARGE", "date range cannot exceed 365 days")
	}
	return nil
}

// GetTeamOverview 获取团队概览统计
func (s *teamAnalyticsService) GetTeamOverview(ctx context.Context, teamID int64, startDate, endDate time.Time) (*TeamOverview, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	overview, err := s.analyticsRepo.GetTeamOverview(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get team overview: %w", err)
	}

	return overview, nil
}

// GetDepartmentRanking 获取部门排名
func (s *teamAnalyticsService) GetDepartmentRanking(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*DeptRankingItem, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	const defaultLimit = 50
	ranking, err := s.analyticsRepo.GetDepartmentRanking(ctx, teamID, startDate, endDate, defaultLimit)
	if err != nil {
		return nil, fmt.Errorf("get department ranking: %w", err)
	}

	return ranking, nil
}

// GetConsumerRanking 获取消费者排名
func (s *teamAnalyticsService) GetConsumerRanking(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*ConsumerRankingItem, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	const defaultLimit = 50
	ranking, err := s.analyticsRepo.GetConsumerRanking(ctx, teamID, startDate, endDate, defaultLimit)
	if err != nil {
		return nil, fmt.Errorf("get consumer ranking: %w", err)
	}

	return ranking, nil
}

// GetDailyTrend 获取每日趋势数据
func (s *teamAnalyticsService) GetDailyTrend(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*DailyTrendItem, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// 查询团队每日聚合数据
	dailyData, err := s.analyticsRepo.QueryTeamDaily(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query team daily: %w", err)
	}

	result := make([]*DailyTrendItem, 0, len(dailyData))
	for _, d := range dailyData {
		result = append(result, &DailyTrendItem{
			Date:          d.BucketDate,
			TotalRequests: d.TotalRequests,
			InputTokens:   d.InputTokens,
			OutputTokens:  d.OutputTokens,
			TotalCost:     d.TotalCost,
			ActualCost:    d.ActualCost,
		})
	}

	return result, nil
}

// GetHourlyTrend 获取小时级趋势数据
func (s *teamAnalyticsService) GetHourlyTrend(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*HourlyTrendItem, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	hourlyData, err := s.analyticsRepo.QueryTeamHourly(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query team hourly: %w", err)
	}

	result := make([]*HourlyTrendItem, 0, len(hourlyData))
	for _, d := range hourlyData {
		result = append(result, &HourlyTrendItem{
			Date:          d.BucketHour,
			TotalRequests: d.TotalRequests,
			InputTokens:   d.InputTokens,
			OutputTokens:  d.OutputTokens,
			TotalCost:     d.TotalCost,
			ActualCost:    d.ActualCost,
		})
	}

	return result, nil
}

// GetModelDistribution 获取模型分布数据
func (s *teamAnalyticsService) GetModelDistribution(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*ModelDistributionItem, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForAnalytics
	}
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// 查询模型每日聚合数据
	modelData, err := s.analyticsRepo.QueryModelDaily(ctx, teamID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query model daily: %w", err)
	}

	// 按模型名称汇总
	modelMap := make(map[string]*ModelDistributionItem)
	for _, d := range modelData {
		item, ok := modelMap[d.ModelName]
		if !ok {
			item = &ModelDistributionItem{
				ModelName: d.ModelName,
			}
			modelMap[d.ModelName] = item
		}
		item.TotalRequests += d.TotalRequests
		item.InputTokens += d.InputTokens
		item.OutputTokens += d.OutputTokens
		item.TotalCost += d.TotalCost
		item.ActualCost += d.ActualCost
	}

	result := make([]*ModelDistributionItem, 0, len(modelMap))
	for _, item := range modelMap {
		result = append(result, item)
	}

	return result, nil
}

// RefreshAggregates 触发聚合数据刷新
func (s *teamAnalyticsService) RefreshAggregates(ctx context.Context, teamID int64, date time.Time) error {
	if teamID <= 0 {
		return ErrInvalidTeamIDForAnalytics
	}
	if date.IsZero() {
		date = time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	}

	// 按顺序执行各类聚合
	if err := s.analyticsRepo.AggregateTeamDaily(ctx, teamID, date); err != nil {
		return fmt.Errorf("aggregate team daily: %w", err)
	}
	if err := s.analyticsRepo.AggregateDeptDaily(ctx, teamID, date); err != nil {
		return fmt.Errorf("aggregate dept daily: %w", err)
	}
	if err := s.analyticsRepo.AggregateConsumerDaily(ctx, teamID, date); err != nil {
		return fmt.Errorf("aggregate consumer daily: %w", err)
	}
	if err := s.analyticsRepo.AggregateModelDaily(ctx, teamID, date); err != nil {
		return fmt.Errorf("aggregate model daily: %w", err)
	}

	return nil
}
