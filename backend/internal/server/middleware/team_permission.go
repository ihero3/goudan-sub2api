package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// TeamRole 定义团队角色
type TeamRole string

const (
	// TeamRoleOwner 团队所有者
	TeamRoleOwner TeamRole = "owner"
	// TeamRoleAdmin 团队管理员
	TeamRoleAdmin TeamRole = "admin"
	// TeamRoleMember 团队成员
	TeamRoleMember TeamRole = "member"
	// TeamRoleViewer 团队观察者
	TeamRoleViewer TeamRole = "viewer"
)

// TeamDataScope 定义数据访问范围
type TeamDataScope string

const (
	// TeamDataScopeAll 所有数据
	TeamDataScopeAll TeamDataScope = "all"
	// TeamDataScopeDepartment 部门级别数据
	TeamDataScopeDepartment TeamDataScope = "department"
	// TeamDataScopeSelf 仅自己的数据
	TeamDataScopeSelf TeamDataScope = "self"
)

// TeamPermissionChecker 团队权限检查器接口
type TeamPermissionChecker interface {
	CheckTeamPermission(ctx *gin.Context, userID int64, teamID int64, requiredRole TeamRole) (bool, TeamRole, TeamDataScope, error)
}

// TeamPermissionMiddleware 团队权限中间件
// 检查用户是否对指定团队具有所需权限
type TeamPermissionMiddleware struct {
	checker TeamPermissionChecker
}

// NewTeamPermissionMiddleware 创建团队权限中间件
func NewTeamPermissionMiddleware(checker TeamPermissionChecker) *TeamPermissionMiddleware {
	return &TeamPermissionMiddleware{
		checker: checker,
	}
}

// RequireRole 返回要求指定角色的中间件
func (m *TeamPermissionMiddleware) RequireRole(requiredRole TeamRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		subject, ok := GetAuthSubjectFromContext(c)
		if !ok {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}

		teamID, err := extractTeamID(c)
		if err != nil {
			response.BadRequest(c, "Invalid team ID")
			c.Abort()
			return
		}

		if m.checker == nil {
			// 如果没有权限检查器，允许通过（开发模式）
			c.Set(string(TeamPermissionCtxKeyTeamID), teamID)
			c.Set(string(TeamPermissionCtxKeyTeamRole), TeamRoleOwner)
			c.Set(string(TeamPermissionCtxKeyTeamDataScope), TeamDataScopeAll)
			c.Next()
			return
		}

		allowed, role, scope, err := m.checker.CheckTeamPermission(c, subject.UserID, teamID, requiredRole)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "You do not have permission to access this team resource",
			})
			c.Abort()
			return
		}

		// 将权限信息存入上下文
		c.Set(string(TeamPermissionCtxKeyTeamID), teamID)
		c.Set(string(TeamPermissionCtxKeyTeamRole), role)
		c.Set(string(TeamPermissionCtxKeyTeamDataScope), scope)
		c.Next()
	}
}

// RequireOwner 要求 owner 角色
func (m *TeamPermissionMiddleware) RequireOwner() gin.HandlerFunc {
	return m.RequireRole(TeamRoleOwner)
}

// RequireAdminOrAbove 要求 admin 或更高角色
func (m *TeamPermissionMiddleware) RequireAdminOrAbove() gin.HandlerFunc {
	return m.RequireRole(TeamRoleAdmin)
}

// RequireMemberOrAbove 要求 member 或更高角色
func (m *TeamPermissionMiddleware) RequireMemberOrAbove() gin.HandlerFunc {
	return m.RequireRole(TeamRoleMember)
}

// RequireViewerOrAbove 要求 viewer 或更高角色
func (m *TeamPermissionMiddleware) RequireViewerOrAbove() gin.HandlerFunc {
	return m.RequireRole(TeamRoleViewer)
}

// TeamPermissionCtxKey 定义团队权限中间件上下文键类型
type TeamPermissionCtxKey string

const (
	// TeamPermissionCtxKeyTeamID 团队ID上下文键
	TeamPermissionCtxKeyTeamID TeamPermissionCtxKey = "team_id"
	// TeamPermissionCtxKeyTeamRole 团队角色上下文键
	TeamPermissionCtxKeyTeamRole TeamPermissionCtxKey = "team_role"
	// TeamPermissionCtxKeyTeamDataScope 数据范围上下文键
	TeamPermissionCtxKeyTeamDataScope TeamPermissionCtxKey = "team_data_scope"
)

// GetTeamIDFromContext 从上下文获取团队ID
func GetTeamIDFromContext(c *gin.Context) (int64, bool) {
	value, exists := c.Get(string(TeamPermissionCtxKeyTeamID))
	if !exists {
		return 0, false
	}
	teamID, ok := value.(int64)
	return teamID, ok
}

// GetTeamRoleFromContext 从上下文获取团队角色
func GetTeamRoleFromContext(c *gin.Context) (TeamRole, bool) {
	value, exists := c.Get(string(TeamPermissionCtxKeyTeamRole))
	if !exists {
		return "", false
	}
	role, ok := value.(TeamRole)
	return role, ok
}

// GetTeamDataScopeFromContext 从上下文获取数据范围
func GetTeamDataScopeFromContext(c *gin.Context) (TeamDataScope, bool) {
	value, exists := c.Get(string(TeamPermissionCtxKeyTeamDataScope))
	if !exists {
		return "", false
	}
	scope, ok := value.(TeamDataScope)
	return scope, ok
}

// HasTeamRole 检查当前用户是否具有指定角色或更高角色
func HasTeamRole(c *gin.Context, requiredRole TeamRole) bool {
	role, ok := GetTeamRoleFromContext(c)
	if !ok {
		return false
	}
	return rolePriority(role) >= rolePriority(requiredRole)
}

// HasTeamDataScope 检查当前用户是否具有指定数据范围或更宽范围
func HasTeamDataScope(c *gin.Context, requiredScope TeamDataScope) bool {
	scope, ok := GetTeamDataScopeFromContext(c)
	if !ok {
		return false
	}
	return scopePriority(scope) >= scopePriority(requiredScope)
}

// rolePriority 返回角色优先级数值（越大权限越高）
func rolePriority(role TeamRole) int {
	switch role {
	case TeamRoleOwner:
		return 4
	case TeamRoleAdmin:
		return 3
	case TeamRoleMember:
		return 2
	case TeamRoleViewer:
		return 1
	default:
		return 0
	}
}

// scopePriority 返回数据范围优先级数值（越大范围越广）
func scopePriority(scope TeamDataScope) int {
	switch scope {
	case TeamDataScopeAll:
		return 3
	case TeamDataScopeDepartment:
		return 2
	case TeamDataScopeSelf:
		return 1
	default:
		return 0
	}
}

// extractTeamID 从请求路径或查询参数中提取团队ID
func extractTeamID(c *gin.Context) (int64, error) {
	// 优先从路径参数获取
	if teamIDStr := c.Param("team_id"); teamIDStr != "" {
		return strconv.ParseInt(teamIDStr, 10, 64)
	}
	// 其次从查询参数获取
	if teamIDStr := c.Query("team_id"); teamIDStr != "" {
		return strconv.ParseInt(teamIDStr, 10, 64)
	}
	// 最后从路径中的 id 参数获取（用于 /teams/:id 路由）
	if idStr := c.Param("id"); idStr != "" {
		// 检查路径是否以 /teams/ 开头
		if strings.Contains(c.Request.URL.Path, "/teams/") {
			return strconv.ParseInt(idStr, 10, 64)
		}
	}
	return 0, nil
}
