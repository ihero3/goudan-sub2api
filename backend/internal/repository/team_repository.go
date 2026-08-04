package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/team"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// TeamRepository 定义团队数据访问接口。
type TeamRepository interface {
	Create(ctx context.Context, t *service.Team) error
	Update(ctx context.Context, t *service.Team) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*service.Team, error)
	GetByOwnerID(ctx context.Context, ownerID int64) ([]*service.Team, error)
	List(ctx context.Context, p pagination.PaginationParams) ([]*service.Team, int, error)
	Count(ctx context.Context) (int, error)
}

type teamRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewTeamRepository 创建 TeamRepository 实例。
func NewTeamRepository(client *dbent.Client, sqlDB *sql.DB) TeamRepository {
	return newTeamRepositoryWithSQL(client, sqlDB)
}

func newTeamRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *teamRepository {
	return &teamRepository{client: client, sql: sqlq}
}

func (r *teamRepository) activeQuery() *dbent.TeamQuery {
	return r.client.Team.Query().Where(team.DeletedAtIsNil())
}

func entTeamToService(t *dbent.Team) *service.Team {
	if t == nil {
		return nil
	}
	st := &service.Team{
		ID:           t.ID,
		Name:         t.Name,
		Slug:         t.Slug,
		Description:  t.Description,
		Timezone:     t.Timezone,
		Language:     t.Language,
		OwnerID:      t.OwnerID,
		BillingEmail: t.BillingEmail,
		Status:       t.Status,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if t.Settings != nil {
		st.Settings = t.Settings
	}
	if t.DeletedAt != nil {
		st.DeletedAt = t.DeletedAt
	}
	return st
}

func (r *teamRepository) Create(ctx context.Context, t *service.Team) error {
	builder := r.client.Team.Create().
		SetName(t.Name).
		SetSlug(t.Slug).
		SetDescription(t.Description).
		SetTimezone(t.Timezone).
		SetLanguage(t.Language).
		SetOwnerID(t.OwnerID).
		SetStatus(t.Status)

	if t.BillingEmail != nil {
		builder.SetBillingEmail(*t.BillingEmail)
	}
	if t.Settings != nil {
		builder.SetSettings(t.Settings)
	}

	entTeam, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create team: %w", err)
	}

	t.ID = entTeam.ID
	t.CreatedAt = entTeam.CreatedAt
	t.UpdatedAt = entTeam.UpdatedAt
	return nil
}

func (r *teamRepository) Update(ctx context.Context, t *service.Team) error {
	if t == nil || t.ID == 0 {
		return errors.New("team id required")
	}

	builder := r.client.Team.UpdateOneID(t.ID).
		SetName(t.Name).
		SetSlug(t.Slug).
		SetDescription(t.Description).
		SetTimezone(t.Timezone).
		SetLanguage(t.Language).
		SetOwnerID(t.OwnerID).
		SetStatus(t.Status)

	if t.BillingEmail != nil {
		builder.SetBillingEmail(*t.BillingEmail)
	} else {
		builder.ClearBillingEmail()
	}
	if t.Settings != nil {
		builder.SetSettings(t.Settings)
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update team %d: %w", t.ID, err)
	}
	return nil
}

func (r *teamRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.client.Team.UpdateOneID(id).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("soft delete team %d: %w", id, err)
	}
	return nil
}

func (r *teamRepository) GetByID(ctx context.Context, id int64) (*service.Team, error) {
	t, err := r.activeQuery().Where(team.ID(id)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("team %d: not found", id)
		}
		return nil, fmt.Errorf("get team by id %d: %w", id, err)
	}
	return entTeamToService(t), nil
}

func (r *teamRepository) GetByOwnerID(ctx context.Context, ownerID int64) ([]*service.Team, error) {
	list, err := r.activeQuery().
		Where(team.OwnerID(ownerID)).
		Order(team.ByCreatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("get teams by owner %d: %w", ownerID, err)
	}

	result := make([]*service.Team, len(list))
	for i, t := range list {
		result[i] = entTeamToService(t)
	}
	return result, nil
}

func (r *teamRepository) List(ctx context.Context, p pagination.PaginationParams) ([]*service.Team, int, error) {
	query := r.activeQuery()

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count teams: %w", err)
	}

	list, err := query.
		Order(team.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list teams: %w", err)
	}

	result := make([]*service.Team, len(list))
	for i, t := range list {
		result[i] = entTeamToService(t)
	}
	return result, total, nil
}

func (r *teamRepository) Count(ctx context.Context) (int, error) {
	c, err := r.activeQuery().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count teams: %w", err)
	}
	return c, nil
}
