<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-64">
            <input
              v-model.number="filters.user_id"
              type="number"
              :placeholder="t('admin.videoTasks.userIdPlaceholder')"
              class="input"
              min="1"
              @change="handleFilterChange"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="statusFilterOptions"
            class="w-40"
            :placeholder="t('common.status')"
            @change="handleFilterChange"
          />
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button
              @click="loadTasks"
              :disabled="loading || !filters.user_id"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="tasks"
          :loading="loading"
          row-key="id"
          default-sort-key="created_at"
          default-sort-order="desc"
        >
          <template #cell-local_id="{ value, row }">
            <div class="min-w-0">
              <code class="truncate text-xs text-gray-700 dark:text-gray-300">{{ value }}</code>
              <div class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">#{{ row.id }}</div>
            </div>
          </template>

          <template #cell-public_model="{ row }">
            <div class="text-sm">
              <div class="font-medium text-gray-900 dark:text-white">{{ row.public_model }}</div>
              <div class="text-xs text-gray-500 dark:text-dark-400">
                → {{ row.upstream_model }}
              </div>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', statusBadgeClass(value)]">
              {{ statusLabel(value) }}
            </span>
          </template>

          <template #cell-user_id="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">#{{ value }}</span>
          </template>

          <template #cell-account_id="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">#{{ value }}</span>
          </template>

          <template #cell-resolution="{ row }">
            <div class="text-sm text-gray-600 dark:text-gray-300">
              <div>{{ row.resolution || '-' }}</div>
              <div v-if="row.duration_sec" class="text-xs text-gray-500 dark:text-dark-400">
                {{ row.duration_sec }}s
              </div>
            </div>
          </template>

          <template #cell-video_url="{ row }">
            <a
              v-if="row.video_url"
              :href="row.video_url"
              target="_blank"
              rel="noopener noreferrer"
              class="text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400"
            >
              <Icon name="externalLink" size="sm" class="inline" />
              {{ t('admin.videoTasks.openVideo') }}
            </a>
            <span v-else class="text-xs text-gray-400">-</span>
          </template>

          <template #cell-cost_usd="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              ${{ Number(value || 0).toFixed(4) }}
            </span>
          </template>

          <template #cell-error_message="{ value }">
            <span
              v-if="value"
              class="line-clamp-2 max-w-xs text-xs text-red-600 dark:text-red-400"
              :title="value"
            >
              {{ value }}
            </span>
            <span v-else class="text-xs text-gray-400">-</span>
          </template>

          <template #cell-created_at="{ value, row }">
            <div class="text-sm text-gray-700 dark:text-gray-300">
              <div>{{ formatDateTime(value) }}</div>
              <div v-if="row.finished_at" class="text-xs text-gray-500 dark:text-dark-400">
                {{ t('admin.videoTasks.finishedAt') }}: {{ formatDateTime(row.finished_at) }}
              </div>
            </div>
          </template>

          <template #cell-actions="{ row }">
            <button
              v-if="row.status === 'processing'"
              @click="handleCancel(row)"
              :disabled="cancelingIds.has(row.id)"
              class="btn btn-sm btn-secondary"
              :title="t('admin.videoTasks.cancel')"
            >
              <Icon
                name="x"
                size="sm"
                :class="cancelingIds.has(row.id) ? 'animate-spin' : ''"
              />
              {{ t('admin.videoTasks.cancel') }}
            </button>
          </template>

          <template #empty>
            <EmptyState
              v-if="!filters.user_id"
              icon="video"
              :title="t('admin.videoTasks.emptyStateTitle')"
              :description="t('admin.videoTasks.userIdRequired')"
            />
            <EmptyState
              v-else
              icon="video"
              :title="t('admin.videoTasks.noData')"
              :description="t('admin.videoTasks.tryOtherFilters')"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <ConfirmDialog
      :show="confirmDialog.visible"
      :title="t('admin.videoTasks.cancelConfirmTitle')"
      :message="t('admin.videoTasks.cancelConfirmMessage', { id: confirmDialog.taskId })"
      :loading="confirmDialog.loading"
      @confirm="confirmCancel"
      @close="confirmDialog.visible = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import type { VideoTask, VideoTaskStatus } from '@/api/admin/videoTasks'
import { getPersistedPageSize } from '@/composables/usePersistedPageSize'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const tasks = ref<VideoTask[]>([])
const loading = ref(false)
const cancelingIds = ref(new Set<number>())

const filters = reactive<{ user_id: number | undefined; status: VideoTaskStatus | '' }>({
  user_id: undefined,
  status: ''
})

const pagination = reactive({
  page: 1,
  page_size: getPersistedPageSize(),
  total: 0,
  pages: 0
})

const confirmDialog = reactive({
  visible: false,
  loading: false,
  taskId: 0
})

const statusFilterOptions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'processing', label: t('admin.videoTasks.status.processing') },
  { value: 'succeeded', label: t('admin.videoTasks.status.succeeded') },
  { value: 'failed', label: t('admin.videoTasks.status.failed') },
  { value: 'cancelled', label: t('admin.videoTasks.status.cancelled') }
])

const columns = computed(() => [
  { key: 'local_id', label: t('admin.videoTasks.columns.taskId'), sortable: false, width: '180px' },
  { key: 'public_model', label: t('admin.videoTasks.columns.model'), sortable: false, width: '180px' },
  { key: 'status', label: t('admin.videoTasks.columns.status'), sortable: false, width: '110px' },
  { key: 'user_id', label: t('admin.videoTasks.columns.user'), sortable: false, width: '90px' },
  { key: 'account_id', label: t('admin.videoTasks.columns.channel'), sortable: false, width: '90px' },
  { key: 'resolution', label: t('admin.videoTasks.columns.resolution'), sortable: false, width: '120px' },
  { key: 'video_url', label: t('admin.videoTasks.columns.video'), sortable: false, width: '110px' },
  { key: 'cost_usd', label: t('admin.videoTasks.columns.cost'), sortable: false, width: '100px' },
  { key: 'error_message', label: t('admin.videoTasks.columns.error'), sortable: false },
  { key: 'created_at', label: t('admin.videoTasks.columns.createdAt'), sortable: true, width: '180px' },
  { key: 'actions', label: t('common.actions'), sortable: false, width: '90px', align: 'right' as const }
])

let loadController: AbortController | null = null

const statusBadgeClass = (status: VideoTaskStatus) => {
  switch (status) {
    case 'processing':
      return 'badge-info'
    case 'succeeded':
      return 'badge-success'
    case 'failed':
      return 'badge-error'
    case 'cancelled':
      return 'badge-gray'
    default:
      return 'badge-gray'
  }
}

const statusLabel = (status: VideoTaskStatus) => {
  return t(`admin.videoTasks.status.${status}`)
}

const loadTasks = async () => {
  if (!filters.user_id || filters.user_id <= 0) {
    tasks.value = []
    pagination.total = 0
    return
  }
  if (loadController) {
    loadController.abort()
  }
  loadController = new AbortController()
  loading.value = true
  try {
    const response = await adminAPI.videoTasks.list({
      user_id: filters.user_id,
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      signal: loadController.signal
    })
    tasks.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (err: any) {
    if (err?.name === 'CanceledError' || err?.code === 'ERR_CANCELED') {
      return
    }
    console.error('Failed to load video tasks:', err)
  } finally {
    loading.value = false
  }
}

const handleFilterChange = () => {
  pagination.page = 1
  loadTasks()
}

const handlePageChange = (page: number) => {
  const validPage = Math.max(1, Math.min(page, pagination.pages || 1))
  pagination.page = validPage
  loadTasks()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadTasks()
}

const handleCancel = (task: VideoTask) => {
  confirmDialog.taskId = task.id
  confirmDialog.visible = true
}

const confirmCancel = async () => {
  confirmDialog.loading = true
  try {
    cancelingIds.value.add(confirmDialog.taskId)
    await adminAPI.videoTasks.cancel(confirmDialog.taskId)
    confirmDialog.visible = false
    await loadTasks()
  } catch (err: any) {
    console.error('Failed to cancel task:', err)
  } finally {
    cancelingIds.value.delete(confirmDialog.taskId)
    confirmDialog.loading = false
  }
}

onBeforeUnmount(() => {
  if (loadController) {
    loadController.abort()
  }
})
</script>
