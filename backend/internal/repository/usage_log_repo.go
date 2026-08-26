package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbapikey "github.com/Wei-Shaw/sub2api/ent/apikey"
	dbconsumer "github.com/Wei-Shaw/sub2api/ent/consumer"
	dbdept "github.com/Wei-Shaw/sub2api/ent/department"
	"github.com/Wei-Shaw/sub2api/ent/predicate"
	"github.com/Wei-Shaw/sub2api/internal/service"
	gocache "github.com/patrickmn/go-cache"
)

const rawUsageLogModelColumn = "model"

// rawUsageLogModelColumn preserves the exact stored usage_logs.model semantics for direct filters.
// Historical rows may contain upstream/billing model values, while newer rows store requested_model.
// Requested/upstream/mapping analytics must use resolveModelDimensionExpression instead.

// usageLogSuccessFilterUL 用于把"失败请求 usage log"（tokens=0、cost=0、不计费的占位记录）
// 从统计性聚合中排除，避免污染 Dashboard / 用量拆分等指标。
//
// schema 中没有 success bool 列；新增列要做迁移，风险大；这里用 actual_cost > 0 作为代理：
// 任何成功落账的请求都会产生 actual_cost（包括 token 计费、纯图片 token 计费、按次/按图计费），
// 反之 failed-request usage log 的 actual_cost 为 0。
// 早期版本用 4 项 token 和 > 0 判定会把"按次/按图计费"与"image_output_tokens 独立计费"的纯图片
// 请求误判为失败，导致这部分请求从用量统计里消失，故改用 actual_cost。
// 配合 `FROM usage_logs ul` JOIN 查询使用。
const usageLogSuccessFilterUL = "ul.actual_cost > 0"

// usageLogEffectivePlatformExpr 用于按"有效平台"维度聚合 usage_logs：
// 优先取请求实际走的分组 platform，若分组未设置 platform 再 fallback 到 account.platform。
// Composite groups are a routing layer, so platform analytics must use the
// resolved concrete account platform instead of grouping spend under "composite".
// 配套要求查询里 LEFT JOIN groups g ON g.id = ul.group_id 与 LEFT JOIN accounts a ON a.id = ul.account_id。
const usageLogEffectivePlatformExpr = "CASE WHEN g.platform = 'composite' THEN a.platform ELSE COALESCE(NULLIF(g.platform,''), a.platform) END"

// dateFormatWhitelist 将 granularity 参数映射为 PostgreSQL TO_CHAR 格式字符串，防止外部输入直接拼入 SQL
var dateFormatWhitelist = map[string]string{
	"hour":  "YYYY-MM-DD HH24:00",
	"day":   "YYYY-MM-DD",
	"week":  "IYYY-IW",
	"month": "YYYY-MM",
}

// safeDateFormat 根据白名单获取 dateFormat，未匹配时返回默认值
func safeDateFormat(granularity string) string {
	if f, ok := dateFormatWhitelist[granularity]; ok {
		return f
	}
	return "YYYY-MM-DD"
}

// appendRawUsageLogModelWhereCondition keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelWhereCondition(conditions []string, args []any, model string) ([]string, []any) {
	if strings.TrimSpace(model) == "" {
		return conditions, args
	}
	conditions = append(conditions, fmt.Sprintf("%s = $%d", rawUsageLogModelColumn, len(args)+1))
	args = append(args, model)
	return conditions, args
}

func appendUsageLogBillingModeWhereCondition(conditions []string, args []any, billingMode string) ([]string, []any) {
	return appendUsageLogBillingModeWhereConditionWithAlias(conditions, args, billingMode, "")
}

func appendUsageLogBillingModeWhereConditionWithAlias(conditions []string, args []any, billingMode string, alias string) ([]string, []any) {
	mode := strings.TrimSpace(billingMode)
	if mode == "" {
		return conditions, args
	}
	column := func(name string) string {
		if alias == "" {
			return name
		}
		return alias + "." + name
	}
	placeholder := fmt.Sprintf("$%d", len(args)+1)
	switch service.BillingMode(mode) {
	case service.BillingModeImage:
		conditions = append(conditions, fmt.Sprintf("(%s = %s OR ((%s IS NULL OR %s = '') AND COALESCE(%s, 0) > 0))", column("billing_mode"), placeholder, column("billing_mode"), column("billing_mode"), column("image_count")))
	case service.BillingModeVideo:
		conditions = append(conditions, fmt.Sprintf("%s = %s", column("billing_mode"), placeholder))
	case service.BillingModeToken:
		conditions = append(conditions, fmt.Sprintf("(%s = %s OR ((%s IS NULL OR %s = '') AND COALESCE(%s, 0) <= 0))", column("billing_mode"), placeholder, column("billing_mode"), column("billing_mode"), column("image_count")))
	default:
		conditions = append(conditions, fmt.Sprintf("%s = %s", column("billing_mode"), placeholder))
	}
	args = append(args, mode)
	return conditions, args
}

func appendUsageLogBillingModeQueryFilter(query string, args []any, billingMode string, alias string) (string, []any) {
	conditions, args := appendUsageLogBillingModeWhereConditionWithAlias(nil, args, billingMode, alias)
	if len(conditions) == 0 {
		return query, args
	}
	return query + " AND " + conditions[0], args
}

func appendUsageLogModelWhereCondition(conditions []string, args []any, model string, source string) ([]string, []any) {
	if strings.TrimSpace(source) == "" {
		return appendRawUsageLogModelWhereCondition(conditions, args, model)
	}
	if strings.TrimSpace(model) == "" {
		return conditions, args
	}
	conditions = append(conditions, fmt.Sprintf("%s = $%d", resolveModelDimensionExpression(source), len(args)+1))
	args = append(args, model)
	return conditions, args
}

// appendRawUsageLogModelQueryFilter keeps direct model filters on the raw model column for backward
// compatibility with historical rows. Requested/upstream analytics must use
// resolveModelDimensionExpression instead.
func appendRawUsageLogModelQueryFilter(query string, args []any, model string) (string, []any) {
	if strings.TrimSpace(model) == "" {
		return query, args
	}
	query += fmt.Sprintf(" AND %s = $%d", rawUsageLogModelColumn, len(args)+1)
	args = append(args, model)
	return query, args
}

func appendUsageLogModelQueryFilter(query string, args []any, model string, source string) (string, []any) {
	if strings.TrimSpace(source) == "" {
		return appendRawUsageLogModelQueryFilter(query, args, model)
	}
	if strings.TrimSpace(model) == "" {
		return query, args
	}
	query += fmt.Sprintf(" AND %s = $%d", resolveModelDimensionExpression(source), len(args)+1)
	args = append(args, model)
	return query, args
}

type usageLogRepository struct {
	client *dbent.Client
	sql    sqlExecutor
	db     *sql.DB

	createBatchOnce     sync.Once
	createBatchCh       chan usageLogCreateRequest
	bestEffortBatchOnce sync.Once
	bestEffortBatchCh   chan usageLogBestEffortRequest
	bestEffortRecent    *gocache.Cache
}

func NewUsageLogRepository(client *dbent.Client, sqlDB *sql.DB) service.UsageLogRepository {
	return newUsageLogRepositoryWithSQL(client, sqlDB)
}

func newUsageLogRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *usageLogRepository {
	// 使用 scanSingleRow 替代 QueryRowContext，保证 ent.Tx 作为 sqlExecutor 可用。
	repo := &usageLogRepository{client: client, sql: sqlq}
	if db, ok := sqlq.(*sql.DB); ok {
		repo.db = db
	}
	repo.bestEffortRecent = gocache.New(usageLogBestEffortRecentTTL, time.Minute)
	return repo
}

// findAPIKeyIDsByDeptOrConsumer returns api_key IDs filtered by department_id and/or consumer_id.
func (r *usageLogRepository) findAPIKeyIDsByDeptOrConsumer(ctx context.Context, deptID, consumerID int64) ([]int64, error) {
	q := r.client.APIKey.Query()
	var predicates []predicate.APIKey
	if deptID > 0 {
		predicates = append(predicates, dbapikey.DepartmentID(deptID))
	}
	if consumerID > 0 {
		predicates = append(predicates, dbapikey.ConsumerID(consumerID))
	}
	if len(predicates) > 0 {
		q = q.Where(predicates...)
	}
	// Select only IDs
	intIDs, err := q.Select(dbapikey.FieldID).Ints(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(intIDs))
	for _, id := range intIDs {
		ids = append(ids, int64(id))
	}
	return ids, nil
}

func (r *usageLogRepository) findAPIKeyIDsByDeptOrConsumerName(ctx context.Context, deptName, consumerName string) ([]int64, error) {
	// Build subquery to find api_key_ids matching department/consumer names
	var apiKeyIDs []int64
	seen := make(map[int64]struct{})

	if deptName != "" {
		// Find departments matching name (case-insensitive LIKE)
		depts, err := r.client.Department.Query().
			Where(dbdept.NameContains(deptName)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		if len(depts) > 0 {
			deptIDs := make([]int64, len(depts))
			for i, d := range depts {
				deptIDs[i] = d.ID
			}
			keys, err := r.client.APIKey.Query().
				Where(dbapikey.DepartmentIDIn(deptIDs...)).
				Select(dbapikey.FieldID).
				Ints(ctx)
			if err != nil {
				return nil, err
			}
			for _, id := range keys {
				id64 := int64(id)
				if _, ok := seen[id64]; !ok {
					seen[id64] = struct{}{}
					apiKeyIDs = append(apiKeyIDs, id64)
				}
			}
		}
	}

	if consumerName != "" {
		// Find consumers matching name (case-insensitive LIKE)
		consumers, err := r.client.Consumer.Query().
			Where(dbconsumer.NameContains(consumerName)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		if len(consumers) > 0 {
			consumerIDs := make([]int64, len(consumers))
			for i, c := range consumers {
				consumerIDs[i] = c.ID
			}
			keys, err := r.client.APIKey.Query().
				Where(dbapikey.ConsumerIDIn(consumerIDs...)).
				Select(dbapikey.FieldID).
				Ints(ctx)
			if err != nil {
				return nil, err
			}
			for _, id := range keys {
				id64 := int64(id)
				if _, ok := seen[id64]; !ok {
					seen[id64] = struct{}{}
					apiKeyIDs = append(apiKeyIDs, id64)
				}
			}
		}
	}

	return apiKeyIDs, nil
}

// hydrateDeptConsumerNames loads department and consumer names for usage logs.
func (r *usageLogRepository) hydrateDeptConsumerNames(ctx context.Context, logs []service.UsageLog) error {
	deptIDs := make(map[int64]struct{})
	consumerIDs := make(map[int64]struct{})
	for i := range logs {
		if logs[i].DepartmentID != nil {
			deptIDs[*logs[i].DepartmentID] = struct{}{}
		}
		if logs[i].ConsumerID != nil {
			consumerIDs[*logs[i].ConsumerID] = struct{}{}
		}
	}

	// Load department names
	deptNames := make(map[int64]string)
	if len(deptIDs) > 0 {
		ids := make([]int64, 0, len(deptIDs))
		for id := range deptIDs {
			ids = append(ids, id)
		}
		models, err := r.client.Department.Query().Where(dbdept.IDIn(ids...)).All(ctx)
		if err != nil {
			return err
		}
		for _, m := range models {
			deptNames[m.ID] = m.Name
		}
	}

	// Load consumer names
	consumerNames := make(map[int64]string)
	if len(consumerIDs) > 0 {
		ids := make([]int64, 0, len(consumerIDs))
		for id := range consumerIDs {
			ids = append(ids, id)
		}
		models, err := r.client.Consumer.Query().Where(dbconsumer.IDIn(ids...)).All(ctx)
		if err != nil {
			return err
		}
		for _, m := range models {
			consumerNames[m.ID] = m.Name
		}
	}

	// Assign names to logs
	for i := range logs {
		if logs[i].DepartmentID != nil {
			if name, ok := deptNames[*logs[i].DepartmentID]; ok {
				logs[i].DepartmentName = &name
			}
		}
		if logs[i].ConsumerID != nil {
			if name, ok := consumerNames[*logs[i].ConsumerID]; ok {
				logs[i].ConsumerName = &name
			}
		}
	}
	return nil
}
