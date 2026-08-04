package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/teamusageteamdaily"
	"github.com/Wei-Shaw/sub2api/ent/teamusagedeptdaily"
	"github.com/Wei-Shaw/sub2api/ent/teamusageconsumerdaily"
	"github.com/Wei-Shaw/sub2api/ent/teamusagemodeldaily"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// TeamAnalyticsRepository 定义团队分析数据访问接口。
type TeamAnalyticsRepository interface {
	AggregateTeamDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateDeptDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateConsumerDaily(ctx context.Context, teamID int64, bucketDate time.Time) error
	AggregateModelDaily(ctx context.Context, teamID int64, bucketDate time.Time) error

	QueryTeamDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageTeamDaily, error)
	QueryTeamHourly(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageHourly, error)
	QueryDeptDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageDeptDaily, error)
	QueryConsumerDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageConsumerDaily, error)
	QueryModelDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageModelDaily, error)

	GetTeamOverview(ctx context.Context, teamID int64, startDate, endDate time.Time) (*service.TeamOverview, error)
	GetDepartmentRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*service.DeptRankingItem, error)
	GetConsumerRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*service.ConsumerRankingItem, error)
}

type teamAnalyticsRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewTeamAnalyticsRepository 创建 TeamAnalyticsRepository 实例。
func NewTeamAnalyticsRepository(client *dbent.Client, sqlDB *sql.DB) TeamAnalyticsRepository {
	return newTeamAnalyticsRepositoryWithSQL(client, sqlDB)
}

func newTeamAnalyticsRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *teamAnalyticsRepository {
	return &teamAnalyticsRepository{client: client, sql: sqlq}
}

// AggregateTeamDaily 聚合团队每日数据（使用原生 SQL join usage_logs + api_keys）。
func (r *teamAnalyticsRepository) AggregateTeamDaily(ctx context.Context, teamID int64, bucketDate time.Time) error {
	query := `
		INSERT INTO team_usage_team_daily (team_id, bucket_date, total_requests, input_tokens, output_tokens,
			cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, computed_at, created_at, updated_at)
		SELECT
			$1 AS team_id,
			$2::date AS bucket_date,
			COUNT(*) AS total_requests,
			COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
			COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
			COALESCE(SUM(ul.cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(ul.cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(ul.total_cost), 0) AS total_cost,
			COALESCE(SUM(ul.actual_cost), 0) AS actual_cost,
			NOW() AS computed_at,
			NOW() AS created_at,
			NOW() AS updated_at
		FROM usage_logs ul
		INNER JOIN api_keys ak ON ak.id = ul.api_key_id
		WHERE ak.team_id = $1
			AND ul.created_at >= $2::date
			AND ul.created_at < ($2::date + INTERVAL '1 day')
		GROUP BY $1, $2::date
		ON CONFLICT (team_id, bucket_date) DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_creation_tokens = EXCLUDED.cache_creation_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			total_cost = EXCLUDED.total_cost,
			actual_cost = EXCLUDED.actual_cost,
			computed_at = EXCLUDED.computed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.sql.ExecContext(ctx, query, teamID, bucketDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("aggregate team daily for team %d date %s: %w", teamID, bucketDate.Format("2006-01-02"), err)
	}
	return nil
}

// AggregateDeptDaily 聚合部门每日数据（使用原生 SQL join usage_logs + api_keys）。
func (r *teamAnalyticsRepository) AggregateDeptDaily(ctx context.Context, teamID int64, bucketDate time.Time) error {
	query := `
		INSERT INTO team_usage_dept_daily (team_id, department_id, department_name, cost_center_code, bucket_date,
			total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
			total_cost, actual_cost, computed_at, created_at, updated_at)
		SELECT
			$1 AS team_id,
			ak.department_id AS department_id,
			MAX(d.name) AS department_name,
			MAX(d.cost_center_code) AS cost_center_code,
			$2::date AS bucket_date,
			COUNT(*) AS total_requests,
			COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
			COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
			COALESCE(SUM(ul.cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(ul.cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(ul.total_cost), 0) AS total_cost,
			COALESCE(SUM(ul.actual_cost), 0) AS actual_cost,
			NOW() AS computed_at,
			NOW() AS created_at,
			NOW() AS updated_at
		FROM usage_logs ul
		INNER JOIN api_keys ak ON ak.id = ul.api_key_id
		LEFT JOIN departments d ON d.id = ak.department_id
		WHERE ak.team_id = $1
			AND ak.department_id IS NOT NULL
			AND ul.created_at >= $2::date
			AND ul.created_at < ($2::date + INTERVAL '1 day')
		GROUP BY ak.department_id
		ON CONFLICT (team_id, department_id, bucket_date) DO UPDATE SET
			department_name = EXCLUDED.department_name,
			cost_center_code = EXCLUDED.cost_center_code,
			total_requests = EXCLUDED.total_requests,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_creation_tokens = EXCLUDED.cache_creation_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			total_cost = EXCLUDED.total_cost,
			actual_cost = EXCLUDED.actual_cost,
			computed_at = EXCLUDED.computed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.sql.ExecContext(ctx, query, teamID, bucketDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("aggregate dept daily for team %d date %s: %w", teamID, bucketDate.Format("2006-01-02"), err)
	}
	return nil
}

// AggregateConsumerDaily 聚合消费者每日数据（使用原生 SQL join usage_logs + api_keys）。
func (r *teamAnalyticsRepository) AggregateConsumerDaily(ctx context.Context, teamID int64, bucketDate time.Time) error {
	query := `
		INSERT INTO team_usage_consumer_daily (team_id, consumer_id, consumer_name, consumer_type, bucket_date,
			total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
			total_cost, actual_cost, computed_at, created_at, updated_at)
		SELECT
			$1 AS team_id,
			ak.consumer_id AS consumer_id,
			MAX(c.name) AS consumer_name,
			MAX(c.type) AS consumer_type,
			$2::date AS bucket_date,
			COUNT(*) AS total_requests,
			COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
			COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
			COALESCE(SUM(ul.cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(ul.cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(ul.total_cost), 0) AS total_cost,
			COALESCE(SUM(ul.actual_cost), 0) AS actual_cost,
			NOW() AS computed_at,
			NOW() AS created_at,
			NOW() AS updated_at
		FROM usage_logs ul
		INNER JOIN api_keys ak ON ak.id = ul.api_key_id
		LEFT JOIN consumers c ON c.id = ak.consumer_id
		WHERE ak.team_id = $1
			AND ak.consumer_id IS NOT NULL
			AND ul.created_at >= $2::date
			AND ul.created_at < ($2::date + INTERVAL '1 day')
		GROUP BY ak.consumer_id
		ON CONFLICT (team_id, consumer_id, bucket_date) DO UPDATE SET
			consumer_name = EXCLUDED.consumer_name,
			consumer_type = EXCLUDED.consumer_type,
			total_requests = EXCLUDED.total_requests,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_creation_tokens = EXCLUDED.cache_creation_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			total_cost = EXCLUDED.total_cost,
			actual_cost = EXCLUDED.actual_cost,
			computed_at = EXCLUDED.computed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.sql.ExecContext(ctx, query, teamID, bucketDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("aggregate consumer daily for team %d date %s: %w", teamID, bucketDate.Format("2006-01-02"), err)
	}
	return nil
}

// AggregateModelDaily 聚合模型每日数据（使用原生 SQL join usage_logs + api_keys）。
func (r *teamAnalyticsRepository) AggregateModelDaily(ctx context.Context, teamID int64, bucketDate time.Time) error {
	query := `
		INSERT INTO team_usage_model_daily (team_id, department_id, consumer_id, bucket_date, model_name,
			total_requests, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens,
			total_cost, actual_cost, computed_at, created_at, updated_at)
		SELECT
			$1 AS team_id,
			ak.department_id AS department_id,
			ak.consumer_id AS consumer_id,
			$2::date AS bucket_date,
			COALESCE(NULLIF(ul.model, ''), 'unknown') AS model_name,
			COUNT(*) AS total_requests,
			COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
			COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
			COALESCE(SUM(ul.cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(ul.cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(ul.total_cost), 0) AS total_cost,
			COALESCE(SUM(ul.actual_cost), 0) AS actual_cost,
			NOW() AS computed_at,
			NOW() AS created_at,
			NOW() AS updated_at
		FROM usage_logs ul
		INNER JOIN api_keys ak ON ak.id = ul.api_key_id
		WHERE ak.team_id = $1
			AND ak.department_id IS NOT NULL
			AND ak.consumer_id IS NOT NULL
			AND ul.created_at >= $2::date
			AND ul.created_at < ($2::date + INTERVAL '1 day')
		GROUP BY ak.department_id, ak.consumer_id, COALESCE(NULLIF(ul.model, ''), 'unknown')
		ON CONFLICT (team_id, department_id, consumer_id, bucket_date, model_name) DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens,
			cache_creation_tokens = EXCLUDED.cache_creation_tokens,
			cache_read_tokens = EXCLUDED.cache_read_tokens,
			total_cost = EXCLUDED.total_cost,
			actual_cost = EXCLUDED.actual_cost,
			computed_at = EXCLUDED.computed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.sql.ExecContext(ctx, query, teamID, bucketDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("aggregate model daily for team %d date %s: %w", teamID, bucketDate.Format("2006-01-02"), err)
	}
	return nil
}

func entTeamUsageTeamDailyToService(d *dbent.TeamUsageTeamDaily) *service.TeamUsageTeamDaily {
	if d == nil {
		return nil
	}
	return &service.TeamUsageTeamDaily{
		ID:                  d.ID,
		TeamID:              d.TeamID,
		BucketDate:          d.BucketDate,
		TotalRequests:       d.TotalRequests,
		InputTokens:         d.InputTokens,
		OutputTokens:        d.OutputTokens,
		CacheCreationTokens: d.CacheCreationTokens,
		CacheReadTokens:     d.CacheReadTokens,
		TotalCost:           d.TotalCost,
		ActualCost:          d.ActualCost,
		ComputedAt:          d.ComputedAt,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

func entTeamUsageDeptDailyToService(d *dbent.TeamUsageDeptDaily) *service.TeamUsageDeptDaily {
	if d == nil {
		return nil
	}
	dd := &service.TeamUsageDeptDaily{
		ID:                  d.ID,
		TeamID:              d.TeamID,
		DepartmentID:        d.DepartmentID,
		BucketDate:          d.BucketDate,
		TotalRequests:       d.TotalRequests,
		InputTokens:         d.InputTokens,
		OutputTokens:        d.OutputTokens,
		CacheCreationTokens: d.CacheCreationTokens,
		CacheReadTokens:     d.CacheReadTokens,
		TotalCost:           d.TotalCost,
		ActualCost:          d.ActualCost,
		ComputedAt:          d.ComputedAt,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
	if d.DepartmentName != nil {
		dd.DepartmentName = d.DepartmentName
	}
	if d.CostCenterCode != nil {
		dd.CostCenterCode = d.CostCenterCode
	}
	return dd
}

func entTeamUsageConsumerDailyToService(d *dbent.TeamUsageConsumerDaily) *service.TeamUsageConsumerDaily {
	if d == nil {
		return nil
	}
	cd := &service.TeamUsageConsumerDaily{
		ID:                  d.ID,
		TeamID:              d.TeamID,
		ConsumerID:          d.ConsumerID,
		BucketDate:          d.BucketDate,
		TotalRequests:       d.TotalRequests,
		InputTokens:         d.InputTokens,
		OutputTokens:        d.OutputTokens,
		CacheCreationTokens: d.CacheCreationTokens,
		CacheReadTokens:     d.CacheReadTokens,
		TotalCost:           d.TotalCost,
		ActualCost:          d.ActualCost,
		ComputedAt:          d.ComputedAt,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
	if d.ConsumerName != nil {
		cd.ConsumerName = d.ConsumerName
	}
	if d.ConsumerType != nil {
		cd.ConsumerType = d.ConsumerType
	}
	return cd
}

func entTeamUsageModelDailyToService(d *dbent.TeamUsageModelDaily) *service.TeamUsageModelDaily {
	if d == nil {
		return nil
	}
	md := &service.TeamUsageModelDaily{
		ID:                  d.ID,
		TeamID:              d.TeamID,
		BucketDate:          d.BucketDate,
		ModelName:           d.ModelName,
		TotalRequests:       d.TotalRequests,
		InputTokens:         d.InputTokens,
		OutputTokens:        d.OutputTokens,
		CacheCreationTokens: d.CacheCreationTokens,
		CacheReadTokens:     d.CacheReadTokens,
		TotalCost:           d.TotalCost,
		ActualCost:          d.ActualCost,
		ComputedAt:          d.ComputedAt,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
	if d.DepartmentID != nil {
		md.DepartmentID = d.DepartmentID
	}
	if d.ConsumerID != nil {
		md.ConsumerID = d.ConsumerID
	}
	return md
}

func (r *teamAnalyticsRepository) QueryTeamDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageTeamDaily, error) {
	list, err := r.client.TeamUsageTeamDaily.Query().
		Where(
			teamusageteamdaily.TeamID(teamID),
			teamusageteamdaily.BucketDateGTE(startDate),
			teamusageteamdaily.BucketDateLTE(endDate),
		).
		Order(teamusageteamdaily.ByBucketDate(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query team daily for team %d: %w", teamID, err)
	}

	result := make([]*service.TeamUsageTeamDaily, len(list))
	for i, d := range list {
		result[i] = entTeamUsageTeamDailyToService(d)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) QueryTeamHourly(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageHourly, error) {
	query := `
		SELECT
			DATE_TRUNC('hour', ul.created_at) AS bucket_hour,
			COUNT(*) AS total_requests,
			COALESCE(SUM(ul.input_tokens), 0) AS input_tokens,
			COALESCE(SUM(ul.output_tokens), 0) AS output_tokens,
			COALESCE(SUM(ul.total_cost), 0) AS total_cost,
			COALESCE(SUM(ul.actual_cost), 0) AS actual_cost
		FROM usage_logs ul
		INNER JOIN api_keys ak ON ak.id = ul.api_key_id
		WHERE ak.team_id = $1
			AND ul.created_at >= $2::timestamp
			AND ul.created_at < $3::timestamp
		GROUP BY DATE_TRUNC('hour', ul.created_at)
		ORDER BY bucket_hour ASC
	`
	rows, err := r.sql.QueryContext(ctx, query, teamID, startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, fmt.Errorf("query team hourly for team %d: %w", teamID, err)
	}
	defer rows.Close()

	var result []*service.TeamUsageHourly
	for rows.Next() {
		var item service.TeamUsageHourly
		err := rows.Scan(
			&item.BucketHour,
			&item.TotalRequests,
			&item.InputTokens,
			&item.OutputTokens,
			&item.TotalCost,
			&item.ActualCost,
		)
		if err != nil {
			return nil, fmt.Errorf("scan team hourly for team %d: %w", teamID, err)
		}
		result = append(result, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("team hourly rows for team %d: %w", teamID, err)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) QueryDeptDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageDeptDaily, error) {
	list, err := r.client.TeamUsageDeptDaily.Query().
		Where(
			teamusagedeptdaily.TeamID(teamID),
			teamusagedeptdaily.BucketDateGTE(startDate),
			teamusagedeptdaily.BucketDateLTE(endDate),
		).
		Order(teamusagedeptdaily.ByBucketDate(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query dept daily for team %d: %w", teamID, err)
	}

	result := make([]*service.TeamUsageDeptDaily, len(list))
	for i, d := range list {
		result[i] = entTeamUsageDeptDailyToService(d)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) QueryConsumerDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageConsumerDaily, error) {
	list, err := r.client.TeamUsageConsumerDaily.Query().
		Where(
			teamusageconsumerdaily.TeamID(teamID),
			teamusageconsumerdaily.BucketDateGTE(startDate),
			teamusageconsumerdaily.BucketDateLTE(endDate),
		).
		Order(teamusageconsumerdaily.ByBucketDate(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query consumer daily for team %d: %w", teamID, err)
	}

	result := make([]*service.TeamUsageConsumerDaily, len(list))
	for i, d := range list {
		result[i] = entTeamUsageConsumerDailyToService(d)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) QueryModelDaily(ctx context.Context, teamID int64, startDate, endDate time.Time) ([]*service.TeamUsageModelDaily, error) {
	list, err := r.client.TeamUsageModelDaily.Query().
		Where(
			teamusagemodeldaily.TeamID(teamID),
			teamusagemodeldaily.BucketDateGTE(startDate),
			teamusagemodeldaily.BucketDateLTE(endDate),
		).
		Order(teamusagemodeldaily.ByBucketDate(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query model daily for team %d: %w", teamID, err)
	}

	result := make([]*service.TeamUsageModelDaily, len(list))
	for i, d := range list {
		result[i] = entTeamUsageModelDailyToService(d)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) GetTeamOverview(ctx context.Context, teamID int64, startDate, endDate time.Time) (*service.TeamOverview, error) {
	query := `
		SELECT
			COALESCE(SUM(total_requests), 0) AS total_requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(total_cost), 0) AS total_cost,
			COALESCE(SUM(actual_cost), 0) AS actual_cost
		FROM team_usage_team_daily
		WHERE team_id = $1
			AND bucket_date >= $2::date
			AND bucket_date <= $3::date
	`
	var ov service.TeamOverview
	err := scanSingleRow(ctx, r.sql, query, []any{teamID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")},
		&ov.TotalRequests, &ov.InputTokens, &ov.OutputTokens, &ov.TotalCost, &ov.ActualCost)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("get team overview for team %d: %w", teamID, err)
	}
	ov.TeamID = teamID
	return &ov, nil
}

func (r *teamAnalyticsRepository) GetDepartmentRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*service.DeptRankingItem, error) {
	query := `
		SELECT
			department_id,
			MAX(department_name) AS department_name,
			SUM(total_requests) AS total_requests,
			SUM(input_tokens) AS input_tokens,
			SUM(output_tokens) AS output_tokens,
			SUM(total_cost) AS total_cost,
			SUM(actual_cost) AS actual_cost
		FROM team_usage_dept_daily
		WHERE team_id = $1
			AND bucket_date >= $2::date
			AND bucket_date <= $3::date
		GROUP BY department_id
		ORDER BY total_cost DESC
		LIMIT $4
	`
	rows, err := r.sql.QueryContext(ctx, query, teamID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), limit)
	if err != nil {
		return nil, fmt.Errorf("get department ranking for team %d: %w", teamID, err)
	}
	defer rows.Close()

	var result []*service.DeptRankingItem
	result = make([]*service.DeptRankingItem, 0)
	for rows.Next() {
		var item service.DeptRankingItem
		var deptID sql.NullInt64
		var deptName sql.NullString
		err := rows.Scan(&deptID, &deptName, &item.TotalRequests, &item.InputTokens, &item.OutputTokens, &item.TotalCost, &item.ActualCost)
		if err != nil {
			return nil, fmt.Errorf("scan department ranking: %w", err)
		}
		if deptID.Valid {
			item.DepartmentID = deptID.Int64
		}
		if deptName.Valid {
			item.DepartmentName = deptName.String
		}
		result = append(result, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("department ranking rows: %w", err)
	}
	return result, nil
}

func (r *teamAnalyticsRepository) GetConsumerRanking(ctx context.Context, teamID int64, startDate, endDate time.Time, limit int) ([]*service.ConsumerRankingItem, error) {
	query := `
		SELECT
			consumer_id,
			MAX(consumer_name) AS consumer_name,
			MAX(consumer_type) AS consumer_type,
			SUM(total_requests) AS total_requests,
			SUM(input_tokens) AS input_tokens,
			SUM(output_tokens) AS output_tokens,
			SUM(total_cost) AS total_cost,
			SUM(actual_cost) AS actual_cost
		FROM team_usage_consumer_daily
		WHERE team_id = $1
			AND bucket_date >= $2::date
			AND bucket_date <= $3::date
		GROUP BY consumer_id
		ORDER BY total_cost DESC
		LIMIT $4
	`
	rows, err := r.sql.QueryContext(ctx, query, teamID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), limit)
	if err != nil {
		return nil, fmt.Errorf("get consumer ranking for team %d: %w", teamID, err)
	}
	defer rows.Close()

	var result []*service.ConsumerRankingItem
	result = make([]*service.ConsumerRankingItem, 0)
	for rows.Next() {
		var item service.ConsumerRankingItem
		var consumerID sql.NullInt64
		var consumerName, consumerType sql.NullString
		err := rows.Scan(&consumerID, &consumerName, &consumerType, &item.TotalRequests, &item.InputTokens, &item.OutputTokens, &item.TotalCost, &item.ActualCost)
		if err != nil {
			return nil, fmt.Errorf("scan consumer ranking: %w", err)
		}
		if consumerID.Valid {
			item.ConsumerID = consumerID.Int64
		}
		if consumerName.Valid {
			item.ConsumerName = consumerName.String
		}
		if consumerType.Valid {
			item.ConsumerType = consumerType.String
		}
		result = append(result, &item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("consumer ranking rows: %w", err)
	}
	return result, nil
}
