package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/department"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

// DepartmentRepository 定义部门数据访问接口。
type DepartmentRepository interface {
	Create(ctx context.Context, d *service.Department) error
	Update(ctx context.Context, d *service.Department) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*service.Department, error)
	GetByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.Department, int, error)
	ListByTeam(ctx context.Context, teamID int64, parentID *int64) ([]*service.Department, error)
	Count(ctx context.Context) (int, error)
	ExistsByTeamAndName(ctx context.Context, teamID int64, name string, excludeID int64) (bool, error)
}

type departmentRepository struct {
	client *dbent.Client
	sql    sqlExecutor
}

// NewDepartmentRepository 创建 DepartmentRepository 实例。
func NewDepartmentRepository(client *dbent.Client, sqlDB *sql.DB) DepartmentRepository {
	return newDepartmentRepositoryWithSQL(client, sqlDB)
}

func newDepartmentRepositoryWithSQL(client *dbent.Client, sqlq sqlExecutor) *departmentRepository {
	return &departmentRepository{client: client, sql: sqlq}
}

func (r *departmentRepository) activeQuery() *dbent.DepartmentQuery {
	return r.client.Department.Query().Where(department.DeletedAtIsNil())
}

func entDepartmentToService(d *dbent.Department) *service.Department {
	if d == nil {
		return nil
	}
	sd := &service.Department{
		ID:          d.ID,
		TeamID:      d.TeamID,
		Name:        d.Name,
		Description: d.Description,
		Level:       d.Level,
		Path:        d.Path,
		SortOrder:   d.SortOrder,
		Status:      d.Status,
		Source:      d.Source,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
	if d.CostCenterCode != nil {
		sd.CostCenterCode = d.CostCenterCode
	}
	if d.ParentID != nil {
		sd.ParentID = d.ParentID
	}
	if d.ExternalID != nil {
		sd.ExternalID = d.ExternalID
	}
	if d.DeletedAt != nil {
		sd.DeletedAt = d.DeletedAt
	}
	return sd
}

func (r *departmentRepository) Create(ctx context.Context, d *service.Department) error {
	builder := r.client.Department.Create().
		SetTeamID(d.TeamID).
		SetName(d.Name).
		SetDescription(d.Description).
		SetLevel(d.Level).
		SetPath(d.Path).
		SetSortOrder(d.SortOrder).
		SetStatus(d.Status).
		SetSource(d.Source)

	if d.CostCenterCode != nil {
		builder.SetCostCenterCode(*d.CostCenterCode)
	}
	if d.ParentID != nil {
		builder.SetParentID(*d.ParentID)
	}
	if d.ExternalID != nil {
		builder.SetExternalID(*d.ExternalID)
	}

	entDept, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create department: %w", err)
	}

	d.ID = entDept.ID
	d.CreatedAt = entDept.CreatedAt
	d.UpdatedAt = entDept.UpdatedAt
	return nil
}

func (r *departmentRepository) Update(ctx context.Context, d *service.Department) error {
	if d == nil || d.ID == 0 {
		return errors.New("department id required")
	}

	builder := r.client.Department.UpdateOneID(d.ID).
		SetName(d.Name).
		SetDescription(d.Description).
		SetLevel(d.Level).
		SetPath(d.Path).
		SetSortOrder(d.SortOrder).
		SetStatus(d.Status).
		SetSource(d.Source)

	if d.CostCenterCode != nil {
		builder.SetCostCenterCode(*d.CostCenterCode)
	} else {
		builder.ClearCostCenterCode()
	}
	if d.ParentID != nil {
		builder.SetParentID(*d.ParentID)
	} else {
		builder.ClearParentID()
	}
	if d.ExternalID != nil {
		builder.SetExternalID(*d.ExternalID)
	} else {
		builder.ClearExternalID()
	}

	_, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("update department %d: %w", d.ID, err)
	}
	return nil
}

func (r *departmentRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now()
	_, err := r.client.Department.UpdateOneID(id).
		SetDeletedAt(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("soft delete department %d: %w", id, err)
	}
	return nil
}

func (r *departmentRepository) GetByID(ctx context.Context, id int64) (*service.Department, error) {
	d, err := r.activeQuery().Where(department.ID(id)).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, fmt.Errorf("department %d: not found", id)
		}
		return nil, fmt.Errorf("get department by id %d: %w", id, err)
	}
	return entDepartmentToService(d), nil
}

func (r *departmentRepository) GetByTeam(ctx context.Context, teamID int64, p pagination.PaginationParams) ([]*service.Department, int, error) {
	query := r.activeQuery().Where(department.TeamID(teamID))

	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count departments by team %d: %w", teamID, err)
	}

	list, err := query.
		Order(department.BySortOrder(entsql.OrderAsc()), department.ByCreatedAt(entsql.OrderDesc())).
		Offset(p.Offset()).
		Limit(p.Limit()).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list departments by team %d: %w", teamID, err)
	}

	result := make([]*service.Department, len(list))
	for i, d := range list {
		result[i] = entDepartmentToService(d)
	}
	return result, total, nil
}

func (r *departmentRepository) ListByTeam(ctx context.Context, teamID int64, parentID *int64) ([]*service.Department, error) {
	query := r.activeQuery().Where(department.TeamID(teamID))
	if parentID != nil {
		query = query.Where(department.ParentID(*parentID))
	}
	// parentID 为 nil 时不加过滤，返回该团队下所有部门（含子部门）

	list, err := query.
		Order(department.BySortOrder(entsql.OrderAsc()), department.ByCreatedAt(entsql.OrderDesc())).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list departments by team %d: %w", teamID, err)
	}

	result := make([]*service.Department, len(list))
	for i, d := range list {
		result[i] = entDepartmentToService(d)
	}
	return result, nil
}

func (r *departmentRepository) Count(ctx context.Context) (int, error) {
	c, err := r.activeQuery().Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("count departments: %w", err)
	}
	return c, nil
}

func (r *departmentRepository) ExistsByTeamAndName(ctx context.Context, teamID int64, name string, excludeID int64) (bool, error) {
	query := r.activeQuery().Where(department.TeamID(teamID), department.NameEqualFold(name))
	if excludeID > 0 {
		query = query.Where(department.IDNEQ(excludeID))
	}
	count, err := query.Count(ctx)
	if err != nil {
		return false, fmt.Errorf("check department exists by team %d and name: %w", teamID, err)
	}
	return count > 0, nil
}
