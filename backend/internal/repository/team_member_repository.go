package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/teammember"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// TeamMemberRepository 定义团队成员数据访问接口。
type TeamMemberRepository interface {
	Create(ctx context.Context, m *service.TeamMember) error
	Update(ctx context.Context, m *service.TeamMember) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*service.TeamMember, error)
	GetByTeamAndUser(ctx context.Context, teamID, userID int64) (*service.TeamMember, error)
	ListByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.TeamMember, int, error)
	ListByUser(ctx context.Context, userID int64, p pagination.PaginationParams) ([]*service.TeamMember, int, error)
	UpdateRole(ctx context.Context, id int64, role string) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	Count(ctx context.Context) (int, error)
}

type teamMemberRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewTeamMemberRepository 创建 TeamMemberRepository 实例。
func NewTeamMemberRepository(client *dbent.Client, sqlDB *sql.DB) TeamMemberRepository {
	return newTeamMemberRepositoryWithSQL(client, sqlDB)
}

func newTeamMemberRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *teamMemberRepository {
	return &teamMemberRepository{client: client, sql: sqlq}
}

func entTeamMemberToService(m *dbent.TeamMember) *service.TeamMember {
	if m == nil {
		return nil
	}
	sm := &service.TeamMember{
		ID:          m.ID,
		TeamID:      m.TeamID,
		UserID:      m.UserID,
		Role:        m.Role,
		Status:      m.Status,
		DisplayName: m.DisplayName,
		JoinedAt:    m.JoinedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.DepartmentID != nil {
		sm.DepartmentID = m.DepartmentID
	}
	if m.ConsumerID != nil {
		sm.ConsumerID = m.ConsumerID
	}
	return sm
}

func (r *teamMemberRepository) Create(ctx context.Context, m *service.TeamMember) error {
	builder := r.client.TeamMember.Create().
		SetTeamID(m.TeamID).
		SetUserID(m.UserID).
		SetRole(m.Role).
		SetJoinedAt(m.JoinedAt)

	if m.Status != "" {
		builder.SetStatus(m.Status)
	}
	if m.DisplayName != "" {
		builder.SetDisplayName(m.DisplayName)
	}

	if m.DepartmentID != nil {
		builder.SetDepartmentID(*m.DepartmentID)
	}
	if m.ConsumerID != nil {
		builder.SetConsumerID(*m.ConsumerID)
	}

	entMember, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create team member: %w", err)
	}

	m.ID = entMember.ID
	m.CreatedAt = entMember.CreatedAt
	m.UpdatedAt = entMember.UpdatedAt
	m.JoinedAt = entMember.JoinedAt
	m.DisplayName = entMember.DisplayName
	return nil
}

func (r *teamMemberRepository) Update(ctx context.Context, m *service.TeamMember) error {
	if m == nil || m.ID == 0 {
		return errors.New("team member id required")
	}

	builder := r.client.TeamMember.UpdateOneID(m.ID).
		SetTeamID(m.TeamID).
		SetUserID(m.UserID).
		SetRole(m.Role).
		SetDisplayName(m.DisplayName)

	if m.DepartmentID != nil {
		builder.SetDepartmentID(*m.DepartmentID)
	} else {
		builder.ClearDepartmentID()
	}
	if m.ConsumerID != nil {
		builder.SetConsumerID(*m.ConsumerID)
	} else {
		builder.ClearConsumerID()
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update team member %d: %w", m.ID, err)
	}
	return nil
}

func (r *teamMemberRepository) Delete(ctx context.Context, id int64) error {
	err := r.client.TeamMember.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete team member %d: %w", id, err)
	}
	return nil
}

func (r *teamMemberRepository) GetByID(ctx context.Context, id int64) (*service.TeamMember, error) {
	m, err := r.client.TeamMember.Query().Where(teammember.ID(id)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("team member %d: not found", id)
		}
		return nil, fmt.Errorf("get team member by id %d: %w", id, err)
	}
	return entTeamMemberToService(m), nil
}

func (r *teamMemberRepository) GetByTeamAndUser(ctx context.Context, teamID, userID int64) (*service.TeamMember, error) {
	m, err := r.client.TeamMember.Query().
		Where(
			teammember.TeamID(teamID),
			teammember.UserID(userID),
		).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			// 找不到记录不算错误，返回 nil, nil 表示不存在
			return nil, nil
		}
		return nil, fmt.Errorf("get team member by team %d and user %d: %w", teamID, userID, err)
	}
	return entTeamMemberToService(m), nil
}

func (r *teamMemberRepository) ListByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.TeamMember, int, error) {
	query := r.client.TeamMember.Query().Where(teammember.TeamID(teamID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count team members by team %d: %w", teamID, err)
	}

	list, err := query.
		Order(teammember.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list team members by team %d: %w", teamID, err)
	}

	result := make([]*service.TeamMember, len(list))
	for i, m := range list {
		result[i] = entTeamMemberToService(m)
	}
	return result, total, nil
}

func (r *teamMemberRepository) ListByUser(ctx context.Context, userID int64, p pagination.PaginationParams) ([]*service.TeamMember, int, error) {
	query := r.client.TeamMember.Query().Where(teammember.UserID(userID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count team members by user %d: %w", userID, err)
	}

	list, err := query.
		Order(teammember.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list team members by user %d: %w", userID, err)
	}

	result := make([]*service.TeamMember, len(list))
	for i, m := range list {
		result[i] = entTeamMemberToService(m)
	}
	return result, total, nil
}

func (r *teamMemberRepository) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.client.TeamMember.UpdateOneID(id).
		SetRole(role).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update team member %d role: %w", id, err)
	}
	return nil
}

func (r *teamMemberRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.client.TeamMember.UpdateOneID(id).
		SetStatus(status).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update team member %d status: %w", id, err)
	}
	return nil
}

func (r *teamMemberRepository) Count(ctx context.Context) (int, error) {
	c, err := r.client.TeamMember.Query().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count team members: %w", err)
	}
	return c, nil
}
