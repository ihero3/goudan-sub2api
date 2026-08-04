package service

import (
	"context"
	"fmt"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrDepartmentNotFound   = infraerrors.NotFound("DEPARTMENT_NOT_FOUND", "department not found")
	ErrDepartmentNameEmpty  = infraerrors.BadRequest("DEPARTMENT_NAME_EMPTY", "department name cannot be empty")
	ErrDepartmentNameExists = infraerrors.BadRequest("DEPARTMENT_NAME_EXISTS", "department name already exists in this team")
	ErrInvalidDepartmentID  = infraerrors.BadRequest("INVALID_DEPARTMENT_ID", "invalid department id")
	ErrInvalidTeamIDForDept = infraerrors.BadRequest("INVALID_TEAM_ID", "invalid team id for department")
	ErrParentDeptNotFound   = infraerrors.NotFound("PARENT_DEPARTMENT_NOT_FOUND", "parent department not found")
	ErrParentDeptCycle      = infraerrors.BadRequest("PARENT_DEPARTMENT_CYCLE", "parent department cycle detected")
)

// DepartmentRepository 部门数据访问接口（在 service 包内定义以避免循环依赖）
type DepartmentRepository interface {
	Create(ctx context.Context, d *Department) error
	Update(ctx context.Context, d *Department) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*Department, error)
	ListByTeam(ctx context.Context, teamID int64, parentID *int64) ([]*Department, error)
	ExistsByTeamAndName(ctx context.Context, teamID int64, name string, excludeID int64) (bool, error)
}

// DepartmentService 部门服务接口
type DepartmentService interface {
	CreateDepartment(ctx context.Context, teamID int64, name string, parentID *int64, costCenterCode *string, description *string) (*Department, error)
	UpdateDepartment(ctx context.Context, deptID int64, name string, costCenterCode *string, description *string, parentID *int64, status *string) (*Department, error)
	DeleteDepartment(ctx context.Context, deptID int64) error
	GetDepartment(ctx context.Context, deptID int64) (*Department, error)
	ListDepartments(ctx context.Context, teamID int64, parentID *int64) ([]*Department, error)
	GetDepartmentTree(ctx context.Context, teamID int64) ([]*DepartmentNode, error)
}

// DepartmentNode 部门树节点
// 使用匿名嵌入 *Department，使部门字段（id、name、parent_id 等）直接出现在节点 JSON 中，
// 与前端 DepartmentTreeNode 期望的扁平结构保持一致。
type DepartmentNode struct {
	*Department
	Children []*DepartmentNode `json:"children"`
}

// departmentService 部门服务实现
type departmentService struct {
	deptRepo DepartmentRepository
}

// NewDepartmentService 创建部门服务实例
func NewDepartmentService(deptRepo DepartmentRepository) DepartmentService {
	return &departmentService{
		deptRepo: deptRepo,
	}
}

// CreateDepartment 创建部门
func (s *departmentService) CreateDepartment(ctx context.Context, teamID int64, name string, parentID *int64, costCenterCode *string, description *string) (*Department, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForDept
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrDepartmentNameEmpty
	}

	// 检查同团队下部门名称是否重复
	nameTrimmed := strings.TrimSpace(name)
	exists, err := s.deptRepo.ExistsByTeamAndName(ctx, teamID, nameTrimmed, 0)
	if err != nil {
		return nil, fmt.Errorf("check department name existence: %w", err)
	}
	if exists {
		return nil, ErrDepartmentNameExists
	}

	level := 1
	path := ""

	// 如果有父部门，校验父部门存在并计算层级和路径
	if parentID != nil && *parentID > 0 {
		parent, err := s.deptRepo.GetByID(ctx, *parentID)
		if err != nil {
			return nil, fmt.Errorf("get parent department: %w", err)
		}
		if parent == nil {
			return nil, ErrParentDeptNotFound
		}
		if parent.TeamID != teamID {
			return nil, infraerrors.BadRequest("PARENT_DEPT_TEAM_MISMATCH", "parent department does not belong to the team")
		}
		level = parent.Level + 1
		path = parent.Path + "/" + fmt.Sprintf("%d", parent.ID)
	}

	desc := ""
	if description != nil {
		desc = strings.TrimSpace(*description)
	}

	dept := &Department{
		TeamID:         teamID,
		Name:           strings.TrimSpace(name),
		Description:    desc,
		CostCenterCode: costCenterCode,
		ParentID:       parentID,
		Level:          level,
		Path:           path,
		SortOrder:      0,
		Status:         "active",
		Source:         "manual",
	}

	if err := s.deptRepo.Create(ctx, dept); err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}

	return dept, nil
}

// UpdateDepartment 更新部门信息
func (s *departmentService) UpdateDepartment(ctx context.Context, deptID int64, name string, costCenterCode *string, description *string, parentID *int64, status *string) (*Department, error) {
	if deptID <= 0 {
		return nil, ErrInvalidDepartmentID
	}

	dept, err := s.deptRepo.GetByID(ctx, deptID)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}

	// 更新名称（含重名检查）
	if name != "" {
		newName := strings.TrimSpace(name)
		if newName != dept.Name {
			exists, err := s.deptRepo.ExistsByTeamAndName(ctx, dept.TeamID, newName, dept.ID)
			if err != nil {
				return nil, fmt.Errorf("check department name existence: %w", err)
			}
			if exists {
				return nil, ErrDepartmentNameExists
			}
		}
		dept.Name = newName
	}

	// 更新成本中心编码
	if costCenterCode != nil {
		dept.CostCenterCode = costCenterCode
	}

	// 更新描述
	if description != nil {
		dept.Description = strings.TrimSpace(*description)
	}

	// 更新父部门（含层级和路径重算）
	if parentID != nil {
		newParentID := *parentID
		if newParentID > 0 {
			// 不允许将自身设为父部门
			if newParentID == dept.ID {
				return nil, infraerrors.BadRequest("PARENT_DEPT_SELF", "department cannot be its own parent")
			}
			parent, err := s.deptRepo.GetByID(ctx, newParentID)
			if err != nil {
				return nil, fmt.Errorf("get parent department: %w", err)
			}
			if parent == nil {
				return nil, ErrParentDeptNotFound
			}
			if parent.TeamID != dept.TeamID {
				return nil, infraerrors.BadRequest("PARENT_DEPT_TEAM_MISMATCH", "parent department does not belong to the team")
			}
			// 检查循环引用：如果新父部门是当前部门的子孙，则拒绝
			if isDescendantPath(parent.Path, dept.ID) {
				return nil, ErrParentDeptCycle
			}
			dept.ParentID = &newParentID
			dept.Level = parent.Level + 1
			dept.Path = parent.Path + "/" + fmt.Sprintf("%d", parent.ID)
		} else {
			// parent_id 为 0 或负数，设为根部门
			dept.ParentID = nil
			dept.Level = 1
			dept.Path = ""
		}
	}

	// 更新状态
	if status != nil && *status != "" {
		dept.Status = strings.TrimSpace(*status)
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		return nil, fmt.Errorf("update department: %w", err)
	}

	return dept, nil
}

// isDescendantPath 检查给定 path 是否表明该节点是 deptID 的子孙。
// dept 的 path 格式为 "/2/5"，表示其祖先链为 2 -> 5 -> self。
// 如果 candidatePath 包含 "/{deptID}"，说明 candidate 是 dept 的后代。
func isDescendantPath(candidatePath string, deptID int64) bool {
	if candidatePath == "" {
		return false
	}
	seg := "/" + fmt.Sprintf("%d", deptID)
	return strings.Contains(candidatePath, seg)
}

// DeleteDepartment 删除部门（软删除）
func (s *departmentService) DeleteDepartment(ctx context.Context, deptID int64) error {
	if deptID <= 0 {
		return ErrInvalidDepartmentID
	}

	dept, err := s.deptRepo.GetByID(ctx, deptID)
	if err != nil {
		return fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return ErrDepartmentNotFound
	}

	if err := s.deptRepo.Delete(ctx, deptID); err != nil {
		return fmt.Errorf("delete department: %w", err)
	}

	return nil
}

// GetDepartment 获取部门详情
func (s *departmentService) GetDepartment(ctx context.Context, deptID int64) (*Department, error) {
	if deptID <= 0 {
		return nil, ErrInvalidDepartmentID
	}

	dept, err := s.deptRepo.GetByID(ctx, deptID)
	if err != nil {
		return nil, fmt.Errorf("get department: %w", err)
	}
	if dept == nil {
		return nil, ErrDepartmentNotFound
	}

	return dept, nil
}

// ListDepartments 列出团队下的部门
func (s *departmentService) ListDepartments(ctx context.Context, teamID int64, parentID *int64) ([]*Department, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForDept
	}

	depts, err := s.deptRepo.ListByTeam(ctx, teamID, parentID)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	return depts, nil
}

// GetDepartmentTree 获取部门树结构
func (s *departmentService) GetDepartmentTree(ctx context.Context, teamID int64) ([]*DepartmentNode, error) {
	if teamID <= 0 {
		return nil, ErrInvalidTeamIDForDept
	}

	// 获取团队下所有部门
	allDepts, err := s.deptRepo.ListByTeam(ctx, teamID, nil)
	if err != nil {
		return nil, fmt.Errorf("list all departments: %w", err)
	}

	// 构建部门映射
	deptMap := make(map[int64]*DepartmentNode, len(allDepts))
	for _, d := range allDepts {
		deptMap[d.ID] = &DepartmentNode{
			Department: d,
			Children:   []*DepartmentNode{},
		}
	}

	// 构建树结构
	var roots []*DepartmentNode
	for _, node := range deptMap {
		if node.Department.ParentID != nil {
			if parent, ok := deptMap[*node.Department.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			} else {
				// 父部门不存在（可能已删除），作为根节点
				roots = append(roots, node)
			}
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}
