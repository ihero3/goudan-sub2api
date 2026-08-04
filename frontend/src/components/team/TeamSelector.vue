<template>
  <div ref="dropdownRef" class="relative inline-block text-left">
    <!-- Trigger Button -->
    <button
      type="button"
      class="inline-flex items-center gap-2 rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm font-medium text-zinc-900 shadow-sm transition-colors hover:bg-zinc-50 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100 dark:hover:bg-zinc-700 dark:focus:ring-indigo-400 dark:focus:ring-offset-zinc-900"
      @click="toggleDropdown"
    >
      <!-- Team Avatar -->
      <div
        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-xs font-bold text-white"
        :class="avatarColorClass"
      >
        {{ currentTeamInitials }}
      </div>

      <!-- Team Name -->
      <span class="max-w-[140px] truncate">
        {{ teamStore.currentTeam?.name || t('team.noTeam') }}
      </span>

      <!-- Role Badge -->
      <span
        v-if="teamStore.currentRole"
        class="hidden rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide sm:inline-flex"
        :class="roleBadgeClass"
      >
        {{ t(`team.role.${teamStore.currentRole}`) }}
      </span>

      <!-- Chevron -->
      <svg
        :class="[
          'h-4 w-4 shrink-0 text-zinc-400 transition-transform duration-200 dark:text-zinc-500',
          isOpen ? 'rotate-180' : ''
        ]"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M19 9l-7 7-7-7"
        />
      </svg>
    </button>

    <!-- Dropdown Panel -->
    <Transition
      enter-active-class="transition ease-out duration-150"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition ease-in duration-100"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isOpen"
        class="absolute right-0 z-50 mt-2 w-72 origin-top-right rounded-xl border border-zinc-200 bg-white shadow-xl ring-1 ring-black/5 dark:border-zinc-700 dark:bg-zinc-800 dark:ring-white/5"
      >
        <!-- Header -->
        <div
          class="border-b border-zinc-100 px-4 py-3 dark:border-zinc-700"
        >
          <p class="text-xs font-semibold uppercase tracking-wider text-zinc-500 dark:text-zinc-400">
            {{ t('team.yourTeams') }}
          </p>
        </div>

        <!-- Team List -->
        <div class="max-h-64 overflow-y-auto py-1">
          <button
            v-for="userTeam in teamStore.activeTeams"
            :key="userTeam.team.id"
            type="button"
            class="group flex w-full items-center gap-3 px-4 py-2.5 text-left transition-colors hover:bg-zinc-50 dark:hover:bg-zinc-700/50"
            :class="{
              'bg-indigo-50 dark:bg-indigo-900/20':
                teamStore.currentTeamId === userTeam.team.id
            }"
            @click="handleSelectTeam(userTeam.team.id)"
          >
            <!-- Team Avatar -->
            <div
              class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-xs font-bold text-white"
              :class="getAvatarColorClass(userTeam.team.name)"
            >
              {{ getInitials(userTeam.team.name) }}
            </div>

            <!-- Team Info -->
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <span
                  class="truncate text-sm font-medium text-zinc-900 dark:text-zinc-100"
                  :class="{
                    'text-indigo-700 dark:text-indigo-300':
                      teamStore.currentTeamId === userTeam.team.id
                  }"
                >
                  {{ userTeam.team.name }}
                </span>
                <!-- Checkmark for active team -->
                <svg
                  v-if="teamStore.currentTeamId === userTeam.team.id"
                  class="h-4 w-4 shrink-0 text-indigo-600 dark:text-indigo-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M5 13l4 4L19 7"
                  />
                </svg>
              </div>
              <!-- Role Badge -->
              <span
                class="mt-0.5 inline-flex rounded-md px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide"
                :class="getRoleBadgeClass(userTeam.role)"
              >
                {{ t(`team.role.${userTeam.role}`) }}
              </span>
            </div>
          </button>

          <!-- Empty State -->
          <div
            v-if="!teamStore.hasTeams"
            class="px-4 py-6 text-center"
          >
            <svg
              class="mx-auto h-10 w-10 text-zinc-300 dark:text-zinc-600"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="1.5"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
            <p class="mt-2 text-sm text-zinc-500 dark:text-zinc-400">
              {{ t('team.noTeams') }}
            </p>
          </div>
        </div>

        <!-- Divider -->
        <div class="border-t border-zinc-100 dark:border-zinc-700" />

        <!-- Create New Team Button -->
        <div class="p-2">
          <button
            type="button"
            class="flex w-full items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 dark:text-zinc-300 dark:hover:bg-zinc-700/50"
            @click="handleCreateTeam"
          >
            <svg
              class="h-4 w-4 shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4v16m8-8H4"
              />
            </svg>
            {{ t('team.createNewTeam') }}
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useTeamStore, type TeamRole } from '@/stores/team'

const { t } = useI18n()
const router = useRouter()
const teamStore = useTeamStore()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

// ==================== Computed ====================

const currentTeamInitials = computed(() => {
  return getInitials(teamStore.currentTeam?.name || '')
})

const avatarColorClass = computed(() => {
  return getAvatarColorClass(teamStore.currentTeam?.name || '')
})

const roleBadgeClass = computed(() => {
  return getRoleBadgeClass(teamStore.currentRole)
})

// ==================== Methods ====================

function toggleDropdown(): void {
  isOpen.value = !isOpen.value
}

function closeDropdown(): void {
  isOpen.value = false
}

function handleSelectTeam(teamId: number): void {
  teamStore.switchTeam(teamId)
  closeDropdown()
}

function handleCreateTeam(): void {
  closeDropdown()
  router.push('/teams/create')
}

function getInitials(name: string): string {
  if (!name) return '?'
  const words = name.trim().split(/\s+/)
  if (words.length === 1) {
    return words[0].slice(0, 2).toUpperCase()
  }
  return (words[0][0] + words[words.length - 1][0]).toUpperCase()
}

function getAvatarColorClass(name: string): string {
  const colors = [
    'bg-rose-500',
    'bg-orange-500',
    'bg-amber-500',
    'bg-emerald-500',
    'bg-teal-500',
    'bg-cyan-500',
    'bg-sky-500',
    'bg-blue-500',
    'bg-indigo-500',
    'bg-violet-500',
    'bg-purple-500',
    'bg-fuchsia-500',
    'bg-pink-500',
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

function getRoleBadgeClass(role: TeamRole | null): string {
  switch (role) {
    case 'owner':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300'
    case 'admin':
      return 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-300'
    case 'member':
      return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300'
    case 'viewer':
      return 'bg-zinc-100 text-zinc-600 dark:bg-zinc-700/60 dark:text-zinc-300'
    default:
      return 'bg-zinc-100 text-zinc-600 dark:bg-zinc-700/60 dark:text-zinc-300'
  }
}

// ==================== Click Outside ====================

function handleClickOutside(event: MouseEvent): void {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    closeDropdown()
  }
}

// ==================== Lifecycle ====================

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
