package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/consumer"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// ConsumerRepository 定义消费者数据访问接口。
type ConsumerRepository interface {
	Create(ctx context.Context, c *service.Consumer) error
	Update(ctx context.Context, c *service.Consumer) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*service.Consumer, error)
	GetByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.Consumer, int, error)
	ListByTeam(ctx context.Context, teamID int64) ([]*service.Consumer, error)
	ListByDepartment(ctx context.Context, departmentID int64) ([]*service.Consumer, error)
	Count(ctx context.Context) (int, error)
}

type consumerRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewConsumerRepository 创建 ConsumerRepository 实例。
func NewConsumerRepository(client *dbent.Client, sqlDB *sql.DB) ConsumerRepository {
	return newConsumerRepositoryWithSQL(client, sqlDB)
}

func newConsumerRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *consumerRepository {
	return &consumerRepository{client: client, sql: sqlq}
}

func (r *consumerRepository) activeQuery() *dbent.ConsumerQuery {
	return r.client.Consumer.Query().Where(consumer.DeletedAtIsNil())
}

func entConsumerToService(c *dbent.Consumer) *service.Consumer {
	if c == nil {
		return nil
	}
	sc := &service.Consumer{
		ID:          c.ID,
		TeamID:      c.TeamID,
		Type:        c.Type,
		Name:        c.Name,
		Status:      c.Status,
		Source:      c.Source,
		Settings:    c.Settings,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.DepartmentID != nil {
		sc.DepartmentID = c.DepartmentID
	}
	if c.Email != nil {
		sc.Email = c.Email
	}
	if c.Phone != nil {
		sc.Phone = c.Phone
	}
	if c.Title != nil {
		sc.Title = c.Title
	}
	if c.AppID != nil {
		sc.AppID = c.AppID
	}
	if c.AppDescription != nil {
		sc.AppDescription = c.AppDescription
	}
	if c.ExternalID != nil {
		sc.ExternalID = c.ExternalID
	}
	if c.DeactivatedAt != nil {
		sc.DeactivatedAt = c.DeactivatedAt
	}
	if c.DeletedAt != nil {
		sc.DeletedAt = c.DeletedAt
	}
	return sc
}

func (r *consumerRepository) Create(ctx context.Context, c *service.Consumer) error {
	builder := r.client.Consumer.Create().
		SetTeamID(c.TeamID).
		SetType(c.Type).
		SetName(c.Name).
		SetStatus(c.Status).
		SetSource(c.Source)

	if c.DepartmentID != nil {
		builder.SetDepartmentID(*c.DepartmentID)
	}
	if c.Email != nil {
		builder.SetEmail(*c.Email)
	}
	if c.Phone != nil {
		builder.SetPhone(*c.Phone)
	}
	if c.Title != nil {
		builder.SetTitle(*c.Title)
	}
	if c.AppID != nil {
		builder.SetAppID(*c.AppID)
	}
	if c.AppDescription != nil {
		builder.SetAppDescription(*c.AppDescription)
	}
	if c.ExternalID != nil {
		builder.SetExternalID(*c.ExternalID)
	}
	if c.Settings != nil {
		builder.SetSettings(c.Settings)
	}

	entConsumer, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}

	c.ID = entConsumer.ID
	c.CreatedAt = entConsumer.CreatedAt
	c.UpdatedAt = entConsumer.UpdatedAt
	c.DepartmentID = entConsumer.DepartmentID
	c.Email = entConsumer.Email
	c.Phone = entConsumer.Phone
	c.Title = entConsumer.Title
	c.AppDescription = entConsumer.AppDescription
	c.ExternalID = entConsumer.ExternalID
	return nil
}

func (r *consumerRepository) Update(ctx context.Context, c *service.Consumer) error {
	if c == nil || c.ID == 0 {
		return errors.New("consumer id required")
	}

	builder := r.client.Consumer.UpdateOneID(c.ID).
		SetType(c.Type).
		SetName(c.Name).
		SetStatus(c.Status).
		SetSource(c.Source)

	if c.DepartmentID != nil {
		builder.SetDepartmentID(*c.DepartmentID)
	} else {
		builder.ClearDepartmentID()
	}
	if c.Email != nil {
		builder.SetEmail(*c.Email)
	} else {
		builder.ClearEmail()
	}
	if c.Phone != nil {
		builder.SetPhone(*c.Phone)
	} else {
		builder.ClearPhone()
	}
	if c.Title != nil {
		builder.SetTitle(*c.Title)
	} else {
		builder.ClearTitle()
	}
	if c.AppID != nil {
		builder.SetAppID(*c.AppID)
	} else {
		builder.ClearAppID()
	}
	if c.AppDescription != nil {
		builder.SetAppDescription(*c.AppDescription)
	} else {
		builder.ClearAppDescription()
	}
	if c.ExternalID != nil {
		builder.SetExternalID(*c.ExternalID)
	} else {
		builder.ClearExternalID()
	}
	if c.Settings != nil {
		builder.SetSettings(c.Settings)
	}
	if c.DeactivatedAt != nil {
		builder.SetDeactivatedAt(*c.DeactivatedAt)
	} else {
		builder.ClearDeactivatedAt()
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update consumer %d: %w", c.ID, err)
	}
	return nil
}

func (r *consumerRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.client.Consumer.UpdateOneID(id).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("soft delete consumer %d: %w", id, err)
	}
	return nil
}

func (r *consumerRepository) GetByID(ctx context.Context, id int64) (*service.Consumer, error) {
	c, err := r.activeQuery().Where(consumer.ID(id)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("consumer %d: not found", id)
		}
		return nil, fmt.Errorf("get consumer by id %d: %w", id, err)
	}
	return entConsumerToService(c), nil
}

func (r *consumerRepository) GetByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.Consumer, int, error) {
	query := r.activeQuery().Where(consumer.TeamID(teamID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count consumers by team %d: %w", teamID, err)
	}

	list, err := query.
		Order(consumer.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list consumers by team %d: %w", teamID, err)
	}

	result := make([]*service.Consumer, len(list))
	for i, c := range list {
		result[i] = entConsumerToService(c)
	}
	return result, total, nil
}

func (r *consumerRepository) ListByTeam(ctx context.Context, teamID int64) ([]*service.Consumer, error) {
	list, err := r.activeQuery().
		Where(consumer.TeamID(teamID)).
		Order(consumer.ByCreatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list consumers by team %d: %w", teamID, err)
	}

	result := make([]*service.Consumer, len(list))
	for i, c := range list {
		result[i] = entConsumerToService(c)
	}
	return result, nil
}

func (r *consumerRepository) ListByDepartment(ctx context.Context, departmentID int64) ([]*service.Consumer, error) {
	list, err := r.activeQuery().
		Where(consumer.DepartmentID(departmentID)).
		Order(consumer.ByCreatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list consumers by department %d: %w", departmentID, err)
	}

	result := make([]*service.Consumer, len(list))
	for i, c := range list {
		result[i] = entConsumerToService(c)
	}
	return result, nil
}

func (r *consumerRepository) Count(ctx context.Context) (int, error) {
	c, err := r.activeQuery().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count consumers: %w", err)
	}
	return c, nil
}
