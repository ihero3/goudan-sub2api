package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// DepartmentHandler handles department-related requests
type DepartmentHandler struct {
	deptService service.DepartmentService
}

// NewDepartmentHandler creates a new DepartmentHandler
func NewDepartmentHandler(deptService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		deptService: deptService,
	}
}

// CreateDepartmentRequest represents the create department request payload
type CreateDepartmentRequest struct {
	Name           string  `json:"name" binding:"required"`
	ParentID       *int64  `json:"parent_id"`
	CostCenterCode string  `json:"cost_center_code"`
	Description    *string `json:"description"`
}

// UpdateDepartmentRequest represents the update department request payload
type UpdateDepartmentRequest struct {
	Name           string  `json:"name"`
	CostCenterCode string  `json:"cost_center_code"`
	Description    *string `json:"description"`
	ParentID       *int64  `json:"parent_id"`
	Status         *string `json:"status"`
}

// List handles listing departments for a team
// GET /api/v1/teams/:id/departments
func (h *DepartmentHandler) List(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	var parentID *int64
	if pidStr := c.Query("parent_id"); pidStr != "" {
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err == nil {
			parentID = &pid
		}
	}

	depts, err := h.deptService.ListDepartments(c.Request.Context(), teamID, parentID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, depts, int64(len(depts)), 1, len(depts))
}

// Get handles getting a single department
// GET /api/v1/teams/:id/departments/:dept_id
func (h *DepartmentHandler) Get(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("dept_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid department ID")
		return
	}

	dept, err := h.deptService.GetDepartment(c.Request.Context(), deptID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dept)
}

// Create handles creating a new department
// POST /api/v1/teams/:id/departments
func (h *DepartmentHandler) Create(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var costCenterCode *string
	if req.CostCenterCode != "" {
		costCenterCode = &req.CostCenterCode
	}

	dept, err := h.deptService.CreateDepartment(c.Request.Context(), teamID, req.Name, req.ParentID, costCenterCode, req.Description)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dept)
}

// Update handles updating a department
// PUT /api/v1/teams/:id/departments/:dept_id
func (h *DepartmentHandler) Update(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("dept_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid department ID")
		return
	}

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var costCenterCode *string
	if req.CostCenterCode != "" {
		costCenterCode = &req.CostCenterCode
	}

	dept, err := h.deptService.UpdateDepartment(c.Request.Context(), deptID, req.Name, costCenterCode, req.Description, req.ParentID, req.Status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dept)
}

// Delete handles deleting a department
// DELETE /api/v1/teams/:id/departments/:dept_id
func (h *DepartmentHandler) Delete(c *gin.Context) {
	deptID, err := strconv.ParseInt(c.Param("dept_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid department ID")
		return
	}

	if err := h.deptService.DeleteDepartment(c.Request.Context(), deptID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Department deleted successfully"})
}

// GetTree handles getting the department tree for a team
// GET /api/v1/teams/:id/departments/tree
func (h *DepartmentHandler) GetTree(c *gin.Context) {
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid team ID")
		return
	}

	tree, err := h.deptService.GetDepartmentTree(c.Request.Context(), teamID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, tree)
}
