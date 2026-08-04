import { ref, computed } from 'vue'
import { listTeams } from '@/api/team'
import type { Team } from '@/api/team'

// Module-level cache so all team pages share the same team context
const cachedTeam = ref<Team | null>(null)
const loading = ref(false)
const loaded = ref(false)
const loadError = ref<string | null>(null)

/**
 * Team context composable.
 * Fetches the current user's first team and exposes its ID for use
 * in team-scoped API calls (departments, consumers, analytics, etc.).
 */
export function useTeamContext() {
  const fetchCurrentTeam = async (): Promise<Team | null> => {
    if (loaded.value && cachedTeam.value) {
      return cachedTeam.value
    }
    loading.value = true
    loadError.value = null
    try {
      const response = await listTeams()
      cachedTeam.value = response.items?.[0] ?? null
      // 只在成功获取到团队时才缓存，避免空结果被缓存导致后续无法重试
      if (cachedTeam.value) {
        loaded.value = true
      } else {
        loadError.value = 'No team available for the current user'
      }
      return cachedTeam.value
    } catch (err: any) {
      // 捕获错误而不是抛出，避免阻塞 onMounted 后续逻辑导致 teamId 永远为 0
      loadError.value = err?.message || 'Failed to load team context'
      console.error('[useTeamContext] fetchCurrentTeam failed:', err)
      cachedTeam.value = null
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    currentTeam: cachedTeam,
    teamId: computed(() => cachedTeam.value?.id ?? 0),
    loading,
    loaded,
    loadError,
    fetchCurrentTeam,
  }
}
