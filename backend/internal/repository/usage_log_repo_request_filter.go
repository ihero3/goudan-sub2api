package repository

import (
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// buildRequestTypeFilterCondition 构造 request_type 过滤 SQL，并在 legacy(request_type=0)
// 时为已知请求类型回退到旧的 stream/openai_ws_mode 判断，保证历史数据兼容。
// idx 为下一个可用参数占位符序号（1 起）。
func buildRequestTypeFilterCondition(idx int, requestType int16) (string, []any) {
	return buildRequestTypeFilterConditionWithAlias(idx, requestType, "")
}

// buildRequestTypeFilterConditionWithAlias 同 buildRequestTypeFilterCondition，
// 但可指定表别名（如 "ul"）。
func buildRequestTypeFilterConditionWithAlias(idx int, requestType int16, alias string) (string, []any) {
	rt := service.RequestType(requestType).Normalize()
	col := "request_type"
	streamCol := "stream"
	wsCol := "openai_ws_mode"
	if alias != "" {
		col = alias + ".request_type"
		streamCol = alias + ".stream"
		wsCol = alias + ".openai_ws_mode"
	}

	switch rt {
	case service.RequestTypeSync:
		return fmt.Sprintf("(%s = $%d OR (%s = 0 AND %s = FALSE AND %s = FALSE))", col, idx, col, streamCol, wsCol), []any{requestType}
	case service.RequestTypeStream:
		return fmt.Sprintf("(%s = $%d OR (%s = 0 AND %s = TRUE AND %s = FALSE))", col, idx, col, streamCol, wsCol), []any{requestType}
	case service.RequestTypeWSV2:
		return fmt.Sprintf("(%s = $%d OR (%s = 0 AND %s = TRUE))", col, idx, col, wsCol), []any{requestType}
	default:
		return fmt.Sprintf("%s = $%d", col, idx), []any{int16(service.RequestTypeUnknown)}
	}
}

// appendRequestTypeOrStreamWhereCondition 追加 request_type 或 stream 过滤条件。
// 仅在 request_type 未指定时才使用 legacy stream 过滤。
func appendRequestTypeOrStreamWhereCondition(conditions []string, args []any, requestType *int16, stream *bool) ([]string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterCondition(len(args)+1, *requestType)
		conditions = append(conditions, condition)
		args = append(args, conditionArgs...)
	} else if stream != nil {
		conditions = append(conditions, fmt.Sprintf("stream = $%d", len(args)+1))
		args = append(args, *stream)
	}
	return conditions, args
}

// buildWhere 将由 AND 连接的条件列表拼接为 WHERE 子句；无条件时返回空串。
func buildWhere(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, " AND ")
}

// appendRequestTypeOrStreamQueryFilter 在已有的 SQL 查询字符串上追加
// request_type 或 stream 过滤片段，适用于直接拼接完整 SQL 的场景。
func appendRequestTypeOrStreamQueryFilter(query string, args []any, requestType *int16, stream *bool) (string, []any) {
	if requestType != nil {
		condition, conditionArgs := buildRequestTypeFilterCondition(len(args)+1, *requestType)
		query += " AND " + condition
		args = append(args, conditionArgs...)
	} else if stream != nil {
		query += fmt.Sprintf(" AND stream = $%d", len(args)+1)
		args = append(args, *stream)
	}
	return query, args
}
