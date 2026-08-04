export type TeamRole = 'owner' | 'admin' | 'member' | 'viewer'

export enum Permission {
  MANAGE_MEMBERS = 'manage_members',
  MANAGE_DEPARTMENTS = 'manage_departments',
  MANAGE_CONSUMERS = 'manage_consumers',
  VIEW_ANALYTICS = 'view_analytics',
  MANAGE_SETTINGS = 'manage_settings',
  MANAGE_BILLING = 'manage_billing',
  MANAGE_API_KEYS = 'manage_api_keys',
  VIEW_LOGS = 'view_logs',
  MANAGE_ROLES = 'manage_roles',
  DELETE_TEAM = 'delete_team',
}

const ROLE_HIERARCHY: Record<TeamRole, number> = {
  viewer: 0,
  member: 1,
  admin: 2,
  owner: 3,
}

const PERMISSION_MATRIX: Record<TeamRole, Permission[]> = {
  owner: Object.values(Permission),
  admin: [
    Permission.MANAGE_MEMBERS,
    Permission.MANAGE_DEPARTMENTS,
    Permission.MANAGE_CONSUMERS,
    Permission.VIEW_ANALYTICS,
    Permission.MANAGE_SETTINGS,
    Permission.MANAGE_API_KEYS,
    Permission.VIEW_LOGS,
    Permission.MANAGE_ROLES,
  ],
  member: [
    Permission.MANAGE_CONSUMERS,
    Permission.VIEW_ANALYTICS,
    Permission.VIEW_LOGS,
  ],
  viewer: [
    Permission.VIEW_ANALYTICS,
    Permission.VIEW_LOGS,
  ],
}

export interface UseTeamPermissionReturn {
  role: TeamRole
  roleLevel: number
  canManageMembers: () => boolean
  canManageDepartments: () => boolean
  canManageConsumers: () => boolean
  canViewAnalytics: () => boolean
  canManageSettings: () => boolean
  canManageBilling: () => boolean
  canManageApiKeys: () => boolean
  canViewLogs: () => boolean
  canManageRoles: () => boolean
  canDeleteTeam: () => boolean
  hasPermission: (permission: Permission) => boolean
  isAtLeast: (role: TeamRole) => boolean
}

/**
 * 团队权限管理 Composable
 * 根据团队角色提供细粒度的权限检查方法
 */
export function useTeamPermission(teamRole: TeamRole): UseTeamPermissionReturn {
  const roleLevel = ROLE_HIERARCHY[teamRole]
  const permissions = new Set(PERMISSION_MATRIX[teamRole] || [])

  const hasPermission = (permission: Permission): boolean => {
    return permissions.has(permission)
  }

  const isAtLeast = (role: TeamRole): boolean => {
    return roleLevel >= ROLE_HIERARCHY[role]
  }

  return {
    role: teamRole,
    roleLevel,
    canManageMembers: () => hasPermission(Permission.MANAGE_MEMBERS),
    canManageDepartments: () => hasPermission(Permission.MANAGE_DEPARTMENTS),
    canManageConsumers: () => hasPermission(Permission.MANAGE_CONSUMERS),
    canViewAnalytics: () => hasPermission(Permission.VIEW_ANALYTICS),
    canManageSettings: () => hasPermission(Permission.MANAGE_SETTINGS),
    canManageBilling: () => hasPermission(Permission.MANAGE_BILLING),
    canManageApiKeys: () => hasPermission(Permission.MANAGE_API_KEYS),
    canViewLogs: () => hasPermission(Permission.VIEW_LOGS),
    canManageRoles: () => hasPermission(Permission.MANAGE_ROLES),
    canDeleteTeam: () => hasPermission(Permission.DELETE_TEAM),
    hasPermission,
    isAtLeast,
  }
}
