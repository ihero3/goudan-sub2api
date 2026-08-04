package repository

import (
	"context"
	"database/sql"
	"fmt"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/teamauditlog"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// TeamAuditLogRepository 定义团队审计日志数据访问接口。
type TeamAuditLogRepository interface {
	Create(ctx context.Context, log *service.TeamAuditLog) error
	ListByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.TeamAuditLog, int, error)
	CountByTeam(ctx context.Context, teamID int64) (int, error)
}

type teamAuditLogRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewTeamAuditLogRepository 创建 TeamAuditLogRepository 实例。
func NewTeamAuditLogRepository(client *dbent.Client, sqlDB *sql.DB) TeamAuditLogRepository {
	return newTeamAuditLogRepositoryWithSQL(client, sqlDB)
}

func newTeamAuditLogRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *teamAuditLogRepository {
	return &teamAuditLogRepository{client: client, sql: sqlq}
}

func entTeamAuditLogToService(l *dbent.TeamAuditLog) *service.TeamAuditLog {
	if l == nil {
		return nil
	}
	sl := &service.TeamAuditLog{
		ID:           l.ID,
		TeamID:       l.TeamID,
		Action:       l.Action,
		ResourceType: l.ResourceType,
		CreatedAt:    l.CreatedAt,
	}
	if l.UserID != nil {
		sl.UserID = l.UserID
	}
	if l.OperationType != nil {
		sl.OperationType = l.OperationType
	}
	if l.ResourceID != nil {
		sl.ResourceID = l.ResourceID
	}
	if l.Changes != nil {
		sl.Changes = l.Changes
	}
	if l.IP != nil {
		sl.IP = l.IP
	}
	if l.UserAgent != nil {
		sl.UserAgent = l.UserAgent
	}
	return sl
}

func (r *teamAuditLogRepository) Create(ctx context.Context, log *service.TeamAuditLog) error {
	builder := r.client.TeamAuditLog.Create().
		SetTeamID(log.TeamID).
		SetAction(log.Action).
		SetResourceType(log.ResourceType)

	if log.UserID != nil {
		builder.SetUserID(*log.UserID)
	}
	if log.OperationType != nil {
		builder.SetOperationType(*log.OperationType)
	}
	if log.ResourceID != nil {
		builder.SetResourceID(*log.ResourceID)
	}
	if log.Changes != nil {
		builder.SetChanges(log.Changes)
	}
	if log.IP != nil {
		builder.SetIP(*log.IP)
	}
	if log.UserAgent != nil {
		builder.SetUserAgent(*log.UserAgent)
	}

	entLog, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create team audit log: %w", err)
	}

	log.ID = entLog.ID
	log.CreatedAt = entLog.CreatedAt
	return nil
}

func (r *teamAuditLogRepository) ListByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.TeamAuditLog, int, error) {
	query := r.client.TeamAuditLog.Query().
		Where(teamauditlog.TeamID(teamID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count team audit logs for team %d: %w", teamID, err)
	}

	list, err := query.
		Order(teamauditlog.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list team audit logs for team %d: %w", teamID, err)
	}

	result := make([]*service.TeamAuditLog, len(list))
	for i, l := range list {
		result[i] = entTeamAuditLogToService(l)
	}
	return result, total, nil
}

func (r *teamAuditLogRepository) CountByTeam(ctx context.Context, teamID int64) (int, error) {
	c, err := r.client.TeamAuditLog.Query().
		Where(teamauditlog.TeamID(teamID)).
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count team audit logs for team %d: %w", teamID, err)
	}
	return c, nil
}
