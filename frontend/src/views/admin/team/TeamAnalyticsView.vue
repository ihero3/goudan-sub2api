<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Header -->
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('admin.team.analytics.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.team.analytics.subtitle') }}
        </p>
      </div>

      <!-- Date Range & Granularity Filter -->
      <div class="flex flex-wrap items-center gap-3">
        <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-600 dark:bg-dark-800">
          <button
            v-for="range in dateRanges"
            :key="range.value"
            @click="selectedRange = range.value"
            :class="[
              'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              selectedRange === range.value
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'
            ]"
          >
            {{ range.label }}
          </button>
        </div>
        <div class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-600 dark:bg-dark-800">
          <button
            @click="granularity = 'day'"
            :class="[
              'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              granularity === 'day'
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'
            ]"
          >
            {{ t('admin.team.analytics.granularityDay') }}
          </button>
          <button
            @click="granularity = 'hour'"
            :class="[
              'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
              granularity === 'hour'
                ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'
            ]"
          >
            {{ t('admin.team.analytics.granularityHour') }}
          </button>
        </div>
      </div>

      <!-- KPI Cards -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.analytics.totalRequests') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ formatNumber(kpi.totalRequests) }}</p>
            </div>
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
              <Icon name="chart" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
          </div>
        </div>

        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.analytics.totalCost') }}</p>
              <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">${{ formatNumber(kpi.totalCost) }}</p>
            </div>
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-amber-100 dark:bg-amber-900/30">
              <Icon name="dollar" size="md" class="text-amber-600 dark:text-amber-400" />
            </div>
          </div>
        </div>
      </div>

      <!-- Charts Row -->
      <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <!-- Usage Trend -->
        <TokenUsageTrend :trend-data="trendData" :loading="loading" />

        <!-- Platform Distribution -->
        <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-600 dark:bg-dark-800">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.team.analytics.platformDistribution') }}
          </h3>
          <div class="space-y-4">
            <div v-for="platform in platformDistribution" :key="platform.name" class="space-y-1">
              <div class="flex items-center justify-between text-sm">
                <span class="font-medium text-gray-700 dark:text-gray-300">{{ platform.name }}</span>
                <span class="text-gray-500 dark:text-dark-400">{{ platform.percent }}% ({{ formatNumber(platform.count) }})</span>
              </div>
              <div class="h-2 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-dark-600">
                <div
                  class="h-full rounded-full transition-all"
                  :class="platform.colorClass"
                  :style="{ width: platform.percent + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Consumers Table -->
      <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 p-4 dark:border-dark-600">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('admin.team.analytics.topConsumers') }}
          </h3>
          <div class="flex flex-wrap items-center gap-2">
            <div class="flex flex-wrap items-center gap-1 rounded-lg border border-gray-200 bg-white p-1 dark:border-dark-600 dark:bg-dark-800">
              <button
                v-for="range in rankingRanges"
                :key="'ranking-' + range.value"
                @click="rankingRange = range.value"
                :class="[
                  'rounded-md px-2.5 py-1.5 text-sm font-medium transition-colors',
                  rankingRange === range.value
                    ? 'bg-primary-100 text-primary-700 dark:bg-primary-900/30 dark:text-primary-300'
                    : 'text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'
                ]"
              >
                {{ range.label }}
              </button>
            </div>
            <div v-if="showCustomRankingPicker" class="flex items-center gap-1">
              <input
                type="date"
                v-model="rankingCustomStart"
                class="rounded-md border border-gray-300 px-2 py-1.5 text-sm text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200"
              />
              <span class="text-gray-400">~</span>
              <input
                type="date"
                v-model="rankingCustomEnd"
                class="rounded-md border border-gray-300 px-2 py-1.5 text-sm text-gray-700 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-200"
              />
            </div>
          </div>
        </div>

        <DataTable
          :columns="topConsumerColumns"
          :data="topConsumers"
          :loading="false"
          :actions-count="1"
        >
          <template #cell-rank="{ row }">
            <div class="flex h-7 w-7 items-center justify-center rounded-full bg-gray-100 text-sm font-bold text-gray-700 dark:bg-dark-700 dark:text-gray-300">
              {{ topConsumers.indexOf(row) + 1 }}
            </div>
          </template>

          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <div class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                {{ getInitials(row.name) }}
              </div>
              <span class="font-medium text-gray-900 dark:text-white">{{ row.name }}</span>
            </div>
          </template>

          <template #cell-requests="{ value }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{ formatNumber(value) }}</span>
          </template>

          <template #cell-cost="{ value }">
            <span class="text-sm font-medium text-gray-900 dark:text-white">${{ formatNumber(value, 2) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <button
              @click="handleViewDetails(row)"
              class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              :title="t('admin.team.analytics.viewConsumerDetails')"
            >
              <Icon name="eye" size="sm" />
            </button>
          </template>
        </DataTable>
      </div>

      <!-- Consumer Details Dialog -->
      <BaseDialog
        :show="showDetailsModal"
        :title="t('admin.team.consumers.detailsTitle')"
        width="normal"
        @close="showDetailsModal = false"
      >
        <div v-if="detailsLoading" class="flex items-center justify-center py-10">
          <span class="text-sm text-gray-500 dark:text-dark-400">{{ t('common.loading') }}</span>
        </div>
        <div v-else-if="detailsConsumer" class="space-y-4">
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center rounded-full bg-primary-100 text-sm font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
              {{ getInitials(detailsConsumer.name) }}
            </div>
            <div>
              <p class="text-lg font-semibold text-gray-900 dark:text-white">{{ detailsConsumer.name }}</p>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.fields.id') }}: {{ detailsConsumer.id }}</p>
            </div>
          </div>

          <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.columns.status') }}</p>
              <div class="mt-1 flex items-center gap-1.5">
                <span :class="['inline-block h-2 w-2 rounded-full', detailsConsumer.status === 'active' ? 'bg-green-500' : 'bg-red-500']"></span>
                <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.team.consumers.status.' + detailsConsumer.status) }}</span>
              </div>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.fields.department') }}</p>
              <p class="mt-1 text-sm text-gray-700 dark:text-gray-300">{{ getDepartmentName(detailsConsumer.department_id) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.columns.createdAt') }}</p>
              <p class="mt-1 text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(detailsConsumer.created_at) }}</p>
            </div>
            <div>
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.fields.updatedAt') }}</p>
              <p class="mt-1 text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(detailsConsumer.updated_at) }}</p>
            </div>
          </div>

          <div>
            <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.descriptionLabel') }}</p>
            <p class="mt-1 whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">
              {{ detailsConsumer.description || t('admin.team.consumers.noDescription') }}
            </p>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end">
            <button type="button" class="btn btn-ghost" @click="showDetailsModal = false">
              {{ t('common.close') }}
            </button>
          </div>
        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTeamContext } from '@/composables/useTeamContext'
import { formatDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import type { Column } from '@/components/common/types'
import type { TrendDataPoint } from '@/types'
import { teamAPI } from '@/api/team'
import type { DailyTrendPoint, ModelDistributionItem, ConsumerRankingItem, AnalyticsDateRange, Consumer, DepartmentTreeNode } from '@/api/team'

const { t } = useI18n()
const appStore = useAppStore()

// Types
interface TopConsumer {
  id: number
  name: string
  requests: number
  cost: number
}

interface PlatformStat {
  name: string
  count: number
  percent: number
  colorClass: string
}

// Data
const loading = ref(false)
const error = ref<string | null>(null)

// Date range selection
const dateRanges = computed(() => [
  { value: '7d', label: t('admin.team.analytics.last7Days') },
  { value: '30d', label: t('admin.team.analytics.last30Days') },
  { value: '90d', label: t('admin.team.analytics.last90Days') }
])
const selectedRange = ref('7d')

// Granularity: day or hour
const granularity = ref<'day' | 'hour'>('day')

// Consumer ranking date range (independent from the main date range)
const rankingRanges = computed(() => [
  { value: '1d', label: t('admin.team.analytics.today') },
  { value: '3d', label: t('admin.team.analytics.last3Days') },
  { value: '7d', label: t('admin.team.analytics.last7Days') },
  { value: '15d', label: t('admin.team.analytics.last15Days') },
  { value: '30d', label: t('admin.team.analytics.last30Days') },
  { value: '90d', label: t('admin.team.analytics.last90Days') },
  { value: 'thisMonth', label: t('admin.team.analytics.thisMonth') },
  { value: 'lastMonth', label: t('admin.team.analytics.lastMonth') },
  { value: 'custom', label: t('admin.team.analytics.custom') },
])
const rankingRange = ref('7d')
const rankingCustomStart = ref('')
const rankingCustomEnd = ref('')

const rankingDateRange = computed<AnalyticsDateRange>(() => {
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  if (rankingRange.value === 'custom') {
    return { start_date: rankingCustomStart.value, end_date: rankingCustomEnd.value }
  }
  if (rankingRange.value === 'thisMonth') {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth(), 1)
    return { start_date: fmt(start), end_date: fmt(now) }
  }
  if (rankingRange.value === 'lastMonth') {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth() - 1, 1)
    const end = new Date(now.getFullYear(), now.getMonth(), 0)
    return { start_date: fmt(start), end_date: fmt(end) }
  }
  const days = parseInt(rankingRange.value)
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - (days - 1))
  return { start_date: fmt(start), end_date: fmt(end) }
})

// Whether to show the custom date inputs
const showCustomRankingPicker = computed(() => rankingRange.value === 'custom')

// Compute start_date / end_date (YYYY-MM-DD) from selectedRange
const dateRange = computed<AnalyticsDateRange>(() => {
  const days = selectedRange.value === '7d' ? 7 : selectedRange.value === '90d' ? 90 : 30
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - days)
  const fmt = (d: Date) => d.toISOString().slice(0, 10)
  return { start_date: fmt(start), end_date: fmt(end) }
})

// KPI Data
const kpi = ref({
  totalRequests: 0,
  totalCost: 0
})

// Usage Trend (line chart data)
const trendData = ref<TrendDataPoint[]>([])

// Platform Distribution
const platformDistribution = ref<PlatformStat[]>([])

// Top Consumers
const topConsumers = ref<TopConsumer[]>([])

// Consumer Detail Dialog
const showDetailsModal = ref(false)
const detailsConsumer = ref<Consumer | null>(null)
const detailsLoading = ref(false)

// Departments for consumer detail lookup
const departments = ref<DepartmentTreeNode[]>([])

const topConsumerColumns: Column[] = [
  { key: 'rank', label: '#', sortable: false },
  { key: 'name', label: t('admin.team.analytics.columns.consumer'), sortable: true },
  { key: 'requests', label: t('admin.team.analytics.columns.requests'), sortable: true },
  { key: 'cost', label: t('admin.team.analytics.columns.cost'), sortable: true },
  { key: 'actions', label: t('admin.team.analytics.columns.actions'), sortable: false }
]

const { teamId, fetchCurrentTeam } = useTeamContext()

const fetchAnalytics = async () => {
  if (!teamId.value) return
  loading.value = true
  error.value = null
  try {
    const dr = dateRange.value
    // Fetch overview
    const overview = await teamAPI.getAnalyticsOverview(teamId.value, dr)
    kpi.value = {
      totalRequests: overview.total_requests,
      totalCost: overview.total_cost
    }

    // Fetch daily trend
    const trend = (await teamAPI.getDailyTrend(teamId.value, dr, granularity.value)) ?? []
    trendData.value = trend.map((point: DailyTrendPoint) => ({
      date: granularity.value === 'hour' ? point.date.slice(0, 16) : point.date.slice(0, 10), // YYYY-MM-DD HH:MM for hour, YYYY-MM-DD for day
      requests: point.total_requests,
      input_tokens: point.input_tokens,
      output_tokens: point.output_tokens,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      total_tokens: point.input_tokens + point.output_tokens,
      cost: point.total_cost,
      actual_cost: point.actual_cost
    }))

    // Fetch model distribution
    const models = (await teamAPI.getModelDistribution(teamId.value, dr)) ?? []
    const totalModelRequests = models.reduce((sum: number, m: ModelDistributionItem) => sum + m.total_requests, 0)
    platformDistribution.value = models.map((m: ModelDistributionItem) => ({
      name: m.model_name,
      count: m.total_requests,
      percent: totalModelRequests > 0 ? Math.round((m.total_requests / totalModelRequests) * 100 * 10) / 10 : 0,
      colorClass: 'bg-primary-500'
    }))

    // Fetch consumer ranking (uses independent date range)
    await fetchConsumerRanking()
  } catch (err: any) {
    error.value = err.message || t('common.error')
    appStore.showError(t('admin.team.analytics.fetchFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchCurrentTeam()
  fetchAnalytics()
  fetchDepartments()
})

// Re-fetch analytics when the date range selection or granularity changes
watch(selectedRange, () => {
  fetchAnalytics()
})
watch(granularity, () => {
  fetchAnalytics()
})

// Re-fetch consumer ranking when its independent date range changes
watch(rankingRange, () => {
  fetchConsumerRanking()
})
watch([rankingCustomStart, rankingCustomEnd], () => {
  if (showCustomRankingPicker.value && rankingCustomStart.value && rankingCustomEnd.value) {
    fetchConsumerRanking()
  }
})

const fetchConsumerRanking = async () => {
  if (!teamId.value) return
  try {
    const rankingDr = rankingDateRange.value
    const ranking = (await teamAPI.getConsumerRanking(teamId.value, rankingDr)) ?? []
    topConsumers.value = ranking.map((c: ConsumerRankingItem) => ({
      id: c.consumer_id,
      name: c.consumer_name,
      requests: c.total_requests,
      cost: c.total_cost,
    }))
  } catch {
    appStore.showError(t('admin.team.analytics.fetchFailed'))
  }
}

const formatNumber = (num: number, decimals = 0): string => {
  if (decimals > 0) {
    return num.toLocaleString(undefined, { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
  }
  return num.toLocaleString()
}

const getInitials = (name: string): string => {
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
}

const fetchDepartments = async () => {
  if (!teamId.value) return
  try {
    departments.value = await teamAPI.getDepartmentTree(teamId.value)
  } catch {
    departments.value = []
  }
}

const findDepartmentInTree = (nodes: DepartmentTreeNode[], id: number | null): DepartmentTreeNode | null => {
  if (id == null) return null
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children && node.children.length > 0) {
      const found = findDepartmentInTree(node.children, id)
      if (found) return found
    }
  }
  return null
}

const getDepartmentName = (departmentId: number | null): string => {
  const dept = findDepartmentInTree(departments.value, departmentId)
  return dept ? dept.name : '-'
}

const handleViewDetails = async (consumer: TopConsumer) => {
  showDetailsModal.value = true
  detailsLoading.value = true
  detailsConsumer.value = null
  try {
    const data = await teamAPI.getConsumer(teamId.value, consumer.id)
    detailsConsumer.value = data
  } catch (err: any) {
    appStore.showError(err?.message || t('admin.team.consumers.detailsFailed'))
    showDetailsModal.value = false
  } finally {
    detailsLoading.value = false
  }
}
</script>
