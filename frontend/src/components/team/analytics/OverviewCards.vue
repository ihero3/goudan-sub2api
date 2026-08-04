<script setup lang="ts">
import { computed } from 'vue'

export interface KPICard {
  label: string
  value: string | number
  trend?: number
  trendLabel?: string
  icon: string
  iconColorClass: string
  iconBgClass: string
  prefix?: string
  suffix?: string
}

interface Props {
  cards: KPICard[]
  loading?: boolean
}

const props = defineProps<Props>()

const safeCards = computed(() => props.cards ?? [])

const formatTrend = (val: number | undefined): string => {
  if (val === undefined || val === null) return ''
  const sign = val > 0 ? '+' : ''
  return `${sign}${val}%`
}

const getTrendColor = (val: number | undefined): string => {
  if (val === undefined || val === null) return 'text-gray-400 dark:text-dark-400'
  if (val > 0) return 'text-emerald-600 dark:text-emerald-400'
  if (val < 0) return 'text-rose-600 dark:text-rose-400'
  return 'text-gray-400 dark:text-dark-400'
}


</script>

<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
    <template v-if="loading">
      <div
        v-for="i in 6"
        :key="i"
        class="animate-pulse rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800"
      >
        <div class="flex items-center justify-between">
          <div class="space-y-2">
            <div class="h-4 w-24 rounded bg-gray-200 dark:bg-dark-700"></div>
            <div class="h-8 w-20 rounded bg-gray-200 dark:bg-dark-700"></div>
          </div>
          <div class="h-10 w-10 rounded-lg bg-gray-200 dark:bg-dark-700"></div>
        </div>
        <div class="mt-3 h-4 w-16 rounded bg-gray-200 dark:bg-dark-700"></div>
      </div>
    </template>

    <template v-else>
      <div
        v-for="(card, index) in safeCards"
        :key="index"
        class="group relative overflow-hidden rounded-xl border border-gray-200 bg-white p-5 transition-all duration-300 hover:shadow-card-hover dark:border-dark-700 dark:bg-dark-800"
      >
        <!-- Subtle gradient overlay on hover -->
        <div
          class="absolute inset-0 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
          :class="card.iconBgClass.replace('bg-', 'bg-').replace('100', '50').replace('900/30', '500/5')"
        ></div>

        <div class="relative">
          <div class="flex items-start justify-between">
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-gray-500 dark:text-dark-400">
                {{ card.label }}
              </p>
              <p class="mt-1.5 text-2xl font-bold tracking-tight text-gray-900 dark:text-white">
                <template v-if="card.prefix">{{ card.prefix }}</template>
                {{ card.value }}
                <template v-if="card.suffix">{{ card.suffix }}</template>
              </p>
            </div>
            <div
              class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg transition-transform duration-300 group-hover:scale-110"
              :class="card.iconBgClass"
            >
              <svg
                class="h-5 w-5"
                :class="card.iconColorClass"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" :d="card.icon" />
              </svg>
            </div>
          </div>

          <!-- Trend indicator -->
          <div v-if="card.trend !== undefined" class="mt-3 flex items-center gap-1.5">
            <span class="inline-flex items-center gap-0.5 text-xs font-semibold" :class="getTrendColor(card.trend)">
              <svg
                class="h-3.5 w-3.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path
                  v-if="card.trend !== 0"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  :d="
                    card.trend > 0
                      ? 'M7 17l4-4 4 4M7 7l4-4 4 4'
                      : 'M7 7l4 4 4-4M7 17l4-4 4 4'
                  "
                />
              </svg>
              <span>{{ formatTrend(card.trend) }}</span>
            </span>
            <span v-if="card.trendLabel" class="text-xs text-gray-400 dark:text-dark-500">
              {{ card.trendLabel }}
            </span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
