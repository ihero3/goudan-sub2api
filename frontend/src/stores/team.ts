/**
 * Team Store
 * Manages team state: current team, team list, and team members
 */

import { defineStore } from 'pinia'
import { ref, computed, readonly } from 'vue'
import { teamAPI } from '@/api/team'
import type { Team } from '@/api/team'
import type { BasePaginationResponse } from '@/types'

export type TeamRole = 'owner' | 'admin' | 'member' | 'viewer'

export interface TeamMember {
  id: number
  user_id: number
  team_id: number
  role: TeamRole
  name: string
  email: string
  avatar_url: string | null
  joined_at: string
}

export interface UserTeam {
  team: Team
  role: TeamRole
}

const CURRENT_TEAM_ID_KEY = 'current_team_id'

export const useTeamStore = defineStore('team', () => {
  // ==================== State ====================

  const currentTeam = ref<Team | null>(null)
  const teams = ref<UserTeam[]>([])
  const members = ref<TeamMember[]>([])
  const loading = ref<boolean>(false)
  const membersLoading = ref<boolean>(false)
  const error = ref<string | null>(null)

  // ==================== Computed ====================

  const currentTeamId = computed(() => currentTeam.value?.id ?? null)

  const currentRole = computed<TeamRole | null>(() => {
    if (!currentTeam.value) return null
    const found = teams.value.find((ut) => ut.team.id === currentTeam.value!.id)
    return found?.role ?? null
  })

  const isOwner = computed(() => currentRole.value === 'owner')
  const isAdmin = computed(() => currentRole.value === 'admin' || currentRole.value === 'owner')
  const isMember = computed(() => currentRole.value === 'member' || isAdmin.value)
  const isViewer = computed(() => currentRole.value === 'viewer')

  const teamCount = computed(() => teams.value.length)
  const hasTeams = computed(() => teams.value.length > 0)

  const activeTeams = computed(() =>
    teams.value.filter((ut) => ut.team.status === 'active')
  )

  // ==================== Actions ====================

  /**
   * Fetch list of teams the current user belongs to
   */
  async function fetchTeams(): Promise<UserTeam[]> {
    loading.value = true
    error.value = null
    try {
      const response: BasePaginationResponse<Team> = await teamAPI.listTeams()
      // Backend returns Team list; we wrap with a default role for now.
      // If backend later returns role, update this mapping.
      const userTeams: UserTeam[] = (response.items || []).map((team) => ({
        team,
        role: 'member' as TeamRole, // default; backend may enrich this
      }))
      teams.value = userTeams
      return userTeams
    } catch (err) {
      const message = (err as { message?: string }).message || 'Failed to fetch teams'
      error.value = message
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Set current team by ID
   * @param teamId - Team ID to switch to
   */
  function switchTeam(teamId: number): void {
    const found = teams.value.find((ut) => ut.team.id === teamId)
    if (found) {
      currentTeam.value = found.team
      localStorage.setItem(CURRENT_TEAM_ID_KEY, String(teamId))
      // Clear members cache when switching teams
      members.value = []
    }
  }

  /**
   * Initialize current team from localStorage or first available team
   */
  function initCurrentTeam(): void {
    const savedId = localStorage.getItem(CURRENT_TEAM_ID_KEY)
    if (savedId && teams.value.length > 0) {
      const id = parseInt(savedId, 10)
      const found = teams.value.find((ut) => ut.team.id === id)
      if (found) {
        currentTeam.value = found.team
        return
      }
    }
    // Fallback to first active team
    if (activeTeams.value.length > 0) {
      currentTeam.value = activeTeams.value[0].team
      localStorage.setItem(CURRENT_TEAM_ID_KEY, String(currentTeam.value.id))
    } else if (teams.value.length > 0) {
      currentTeam.value = teams.value[0].team
      localStorage.setItem(CURRENT_TEAM_ID_KEY, String(currentTeam.value.id))
    }
  }

  /**
   * Fetch team details by ID
   * @param teamId - Team ID
   */
  async function fetchTeam(teamId: number): Promise<Team> {
    loading.value = true
    error.value = null
    try {
      const team = await teamAPI.getTeam(teamId)
      // Update in teams list if present
      const idx = teams.value.findIndex((ut) => ut.team.id === teamId)
      if (idx !== -1) {
        teams.value[idx] = { ...teams.value[idx], team }
      }
      // Update current if matching
      if (currentTeam.value?.id === teamId) {
        currentTeam.value = team
      }
      return team
    } catch (err) {
      const message = (err as { message?: string }).message || 'Failed to fetch team'
      error.value = message
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Fetch members for the current team
   */
  async function fetchMembers(): Promise<TeamMember[]> {
    if (!currentTeam.value) {
      return []
    }
    membersLoading.value = true
    try {
      // TODO: Replace with actual members API when available
      // For now, return mock data or empty array
      members.value = []
      return members.value
    } catch (err) {
      console.error('Failed to fetch team members:', err)
      return []
    } finally {
      membersLoading.value = false
    }
  }

  /**
   * Create a new team
   * @param name - Team name
   * @param description - Optional team description
   */
  async function createTeam(name: string, description?: string | null): Promise<Team> {
    loading.value = true
    error.value = null
    try {
      const team = await teamAPI.createTeam({ name, description: description || null })
      teams.value.push({ team, role: 'owner' })
      // Auto-switch to newly created team
      currentTeam.value = team
      localStorage.setItem(CURRENT_TEAM_ID_KEY, String(team.id))
      return team
    } catch (err) {
      const message = (err as { message?: string }).message || 'Failed to create team'
      error.value = message
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Update a team
   * @param teamId - Team ID
   * @param payload - Update data
   */
  async function updateTeam(
    teamId: number,
    payload: { name?: string; description?: string | null; status?: 'active' | 'inactive' }
  ): Promise<Team> {
    loading.value = true
    error.value = null
    try {
      const team = await teamAPI.updateTeam(teamId, payload)
      const idx = teams.value.findIndex((ut) => ut.team.id === teamId)
      if (idx !== -1) {
        teams.value[idx] = { ...teams.value[idx], team }
      }
      if (currentTeam.value?.id === teamId) {
        currentTeam.value = team
      }
      return team
    } catch (err) {
      const message = (err as { message?: string }).message || 'Failed to update team'
      error.value = message
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Delete a team
   * @param teamId - Team ID
   */
  async function deleteTeam(teamId: number): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await teamAPI.deleteTeam(teamId)
      teams.value = teams.value.filter((ut) => ut.team.id !== teamId)
      if (currentTeam.value?.id === teamId) {
        currentTeam.value = null
        localStorage.removeItem(CURRENT_TEAM_ID_KEY)
        initCurrentTeam()
      }
    } catch (err) {
      const message = (err as { message?: string }).message || 'Failed to delete team'
      error.value = message
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Reset store state
   */
  function reset(): void {
    currentTeam.value = null
    teams.value = []
    members.value = []
    loading.value = false
    membersLoading.value = false
    error.value = null
    localStorage.removeItem(CURRENT_TEAM_ID_KEY)
  }

  // ==================== Return Store API ====================

  return {
    // State
    currentTeam: readonly(currentTeam),
    teams: readonly(teams),
    members: readonly(members),
    loading: readonly(loading),
    membersLoading: readonly(membersLoading),
    error: readonly(error),

    // Computed
    currentTeamId,
    currentRole,
    isOwner,
    isAdmin,
    isMember,
    isViewer,
    teamCount,
    hasTeams,
    activeTeams,

    // Actions
    fetchTeams,
    switchTeam,
    initCurrentTeam,
    fetchTeam,
    fetchMembers,
    createTeam,
    updateTeam,
    deleteTeam,
    reset,
  }
})
