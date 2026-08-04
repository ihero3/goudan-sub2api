<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('admin.team.consumers.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.team.consumers.subtitle') }}
          </p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          <Icon name="plus" size="md" class="mr-2" />
          {{ t('admin.team.consumers.addConsumer') }}
        </button>
      </div>

      <!-- Stats Cards -->
      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary-100 dark:bg-primary-900/30">
              <Icon name="users" size="md" class="text-primary-600 dark:text-primary-400" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.totalConsumers') }}</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ stats.total }}</p>
            </div>
          </div>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-green-100 dark:bg-green-900/30">
              <Icon name="checkCircle" size="md" class="text-green-600 dark:text-green-400" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.activeConsumers') }}</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ stats.active }}</p>
            </div>
          </div>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-900/30">
              <Icon name="chart" size="md" class="text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.newThisMonth') }}</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ stats.newThisMonth }}</p>
            </div>
          </div>
        </div>
        <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-red-100 dark:bg-red-900/30">
              <Icon name="x" size="md" class="text-red-600 dark:text-red-400" />
            </div>
            <div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('admin.team.consumers.inactive') }}</p>
              <p class="text-xl font-semibold text-gray-900 dark:text-white">{{ stats.inactive }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Consumers Table -->
      <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
        <div class="border-b border-gray-200 p-4 dark:border-dark-600">
          <div class="flex flex-wrap items-center gap-3">
            <div class="relative w-full md:w-64">
              <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.team.consumers.searchPlaceholder')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>
            <div class="w-full sm:w-40">
              <Select
                v-model="filters.type"
                :options="typeOptions"
                @change="applyFilter"
              />
            </div>
            <div class="w-full sm:w-40">
              <Select
                v-model="filters.status"
                :options="statusOptions"
                @change="applyFilter"
              />
            </div>
          </div>
        </div>

        <DataTable
          :columns="columns"
          :data="displayedConsumers"
          :loading="loading"
          :actions-count="3"
        >
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-full bg-primary-100 text-sm font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                {{ getInitials(row.name) }}
              </div>
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ row.name }}</p>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ row.email }}</p>
              </div>
            </div>
          </template>

          <template #cell-type="{ value }">
            <span :class="['badge', value === 'person' ? 'badge-blue' : value === 'application' ? 'badge-purple' : 'badge-gray']">
              {{ t('team.consumer.types.' + value) }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <div class="flex items-center gap-1.5">
              <span :class="['inline-block h-2 w-2 rounded-full', value === 'active' ? 'bg-green-500' : 'bg-red-500']"></span>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.team.consumers.status.' + value) }}</span>
            </div>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="handleEdit(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleViewDetails(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('common.view')"
              >
                <Icon name="eye" size="sm" />
              </button>
              <button
                @click="handleDelete(row)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.team.consumers.noConsumers')"
              :description="t('admin.team.consumers.addFirstConsumer')"
              :action-text="t('admin.team.consumers.addConsumer')"
              @action="showCreateModal = true"
            />
          </template>
        </DataTable>

        <div class="border-t border-gray-200 p-4 dark:border-dark-600">
          <Pagination
            v-if="pagination.total > 0"
            :page="pagination.page"
            :total="pagination.total"
            :page-size="pagination.pageSize"
            @update:page="handlePageChange"
            @update:pageSize="handlePageSizeChange"
          />
        </div>
      </div>
    </div>

    <ConsumerForm
      v-model="showCreateModal"
      mode="create"
      :departments="departments"
      @submit="handleCreateConsumer"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.team.consumers.deleteConsumer')"
      :message="t('admin.team.consumers.deleteConfirm', { name: deletingConsumer?.name })"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />

    <!-- Edit Consumer Dialog (uses ConsumerForm for field consistency) -->
    <ConsumerForm
      v-model="showEditModal"
      mode="edit"
      :initial-data="editingConsumerFormData"
      :departments="departments"
      @submit="handleUpdateConsumer"
    />

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
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTeamContext } from '@/composables/useTeamContext'
import { formatDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import ConsumerForm from '@/components/team/ConsumerForm.vue'
import type { ConsumerFormData, ConsumerType } from '@/components/team/ConsumerForm.vue'
import type { Column } from '@/components/common/types'
import { teamAPI } from '@/api/team'
import type { Consumer as ApiConsumer, DepartmentTreeNode, UpdateConsumerRequest } from '@/api/team'

const { t } = useI18n()
const appStore = useAppStore()

// Types
interface Consumer {
  id: number
  name: string
  description: string
  email: string | null
  phone: string | null
  title: string | null
  type: string
  department_id: number | null
  status: 'active' | 'inactive'
  created_at: string
}

// Data
const consumers = ref<Consumer[]>([])
const displayedConsumers = ref<Consumer[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const searchQuery = ref('')

const stats = computed(() => {
  const now = new Date()
  const currentYear = now.getFullYear()
  const currentMonth = now.getMonth()
  return {
    total: consumers.value.length,
    active: consumers.value.filter(c => c.status === 'active').length,
    newThisMonth: consumers.value.filter(c => {
      const d = new Date(c.created_at)
      return d.getFullYear() === currentYear && d.getMonth() === currentMonth
    }).length,
    inactive: consumers.value.filter(c => c.status === 'inactive').length
  }
})

const columns: Column[] = [
  { key: 'name', label: t('admin.team.consumers.columns.name'), sortable: true },
  { key: 'type', label: t('admin.team.consumers.columns.type'), sortable: true },
  { key: 'status', label: t('admin.team.consumers.columns.status'), sortable: true },
  { key: 'created_at', label: t('admin.team.consumers.columns.createdAt'), sortable: true },
  { key: 'actions', label: t('admin.team.consumers.columns.actions'), sortable: false }
]

const filters = reactive({
  type: '',
  status: ''
})

const typeOptions = computed(() => [
  { value: '', label: t('admin.team.consumers.allTypes') },
  { value: 'person', label: t('team.consumer.types.person') },
  { value: 'application', label: t('team.consumer.types.application') },
  { value: 'service_account', label: t('team.consumer.types.serviceAccount') }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.team.consumers.allStatus') },
  { value: 'active', label: t('admin.team.consumers.status.active') },
  { value: 'inactive', label: t('admin.team.consumers.status.inactive') }
])

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
  pages: 1
})

const showCreateModal = ref(false)
const showDeleteDialog = ref(false)
const deletingConsumer = ref<Consumer | null>(null)
const departments = ref<DepartmentTreeNode[]>([])

const { teamId, fetchCurrentTeam } = useTeamContext()

const fetchDepartments = async () => {
  if (!teamId.value) return
  try {
    departments.value = await teamAPI.getDepartmentTree(teamId.value)
  } catch {
    departments.value = []
  }
}

const applyLocalFilters = () => {
  let result = consumers.value
  const query = searchQuery.value.trim().toLowerCase()
  if (query) {
    result = result.filter(c => c.name.toLowerCase().includes(query))
  }
  if (filters.type) {
    result = result.filter(c => c.type === filters.type)
  }
  if (filters.status) {
    result = result.filter(c => c.status === filters.status)
  }
  displayedConsumers.value = result
}

const fetchConsumers = async () => {
  if (!teamId.value) return
  loading.value = true
  error.value = null
  try {
    const response = await teamAPI.listConsumers(teamId.value, undefined, {
      page: pagination.page,
      page_size: pagination.pageSize,
    })
    const items = response.items ?? []
    consumers.value = items.map((c: ApiConsumer) => ({
      id: c.id,
      name: c.name,
      description: c.description ?? '',
      email: c.email ?? null,
      phone: c.phone ?? null,
      title: c.title ?? null,
      type: c.type ?? '',
      department_id: c.department_id ?? null,
      status: c.status,
      created_at: c.created_at,
    }))
    pagination.total = response.total ?? consumers.value.length
    pagination.pages = response.pages ?? 1
    if (response.page) pagination.page = response.page
    if (response.page_size) pagination.pageSize = response.page_size
    applyLocalFilters()
  } catch (err: any) {
    error.value = err.message || t('common.error')
    appStore.showError(t('admin.team.consumers.fetchFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchCurrentTeam()
  fetchConsumers()
  fetchDepartments()
})

const handleCreateConsumer = async (payload: any) => {
  try {
    await teamAPI.createConsumer(teamId.value, payload)
    appStore.showSuccess(t('admin.team.consumers.createSuccess'))
    showCreateModal.value = false
    await fetchConsumers()
  } catch (err: any) {
    appStore.showError(err?.message || t('admin.team.consumers.createFailed'))
  }
}

const getInitials = (name: string): string => {
  return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
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

let searchTimeout: ReturnType<typeof setTimeout>
const handleSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    applyLocalFilters()
  }, 300)
}

const applyFilter = () => {
  applyLocalFilters()
}

const handlePageChange = (page: number) => {
  pagination.page = page
  fetchConsumers()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchConsumers()
}

// Edit dialog state
const showEditModal = ref(false)
const editingConsumer = ref<Consumer | null>(null)
const editSaving = ref(false)
const editingConsumerFormData = ref<Partial<ConsumerFormData>>({})

const handleEdit = (consumer: Consumer) => {
  editingConsumer.value = consumer
  editingConsumerFormData.value = {
    name: consumer.name,
    email: consumer.email ?? '',
    phone: consumer.phone ?? '',
    title: consumer.title ?? '',
    type: (consumer.type as ConsumerType) || '',
    departmentId: consumer.department_id ?? '',
    status: consumer.status,
    description: consumer.description ?? '',
  }
  showEditModal.value = true
}

const handleUpdateConsumer = async (formData: ConsumerFormData) => {
  if (!editingConsumer.value) return
  editSaving.value = true
  try {
    const payload: UpdateConsumerRequest = {
      name: formData.name.trim(),
      email: formData.email.trim() || undefined,
      phone: formData.phone.trim() || null,
      title: formData.title.trim() || null,
      type: formData.type || undefined,
      description: formData.description.trim() || null,
      status: formData.status,
    }
    if (formData.departmentId) {
      payload.dept_id = Number(formData.departmentId)
    }
    await teamAPI.updateConsumer(teamId.value, editingConsumer.value.id, payload)
    appStore.showSuccess(t('admin.team.consumers.editSuccess'))
    showEditModal.value = false
    editingConsumer.value = null
    await fetchConsumers()
  } catch (err: any) {
    appStore.showError(err?.message || t('admin.team.consumers.editFailed'))
  } finally {
    editSaving.value = false
  }
}

// Details dialog state
const showDetailsModal = ref(false)
const detailsConsumer = ref<ApiConsumer | null>(null)
const detailsLoading = ref(false)

const handleViewDetails = async (consumer: Consumer) => {
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

const handleDelete = (consumer: Consumer) => {
  deletingConsumer.value = consumer
  showDeleteDialog.value = true
}

const deleting = ref(false)
const confirmDelete = async () => {
  if (!deletingConsumer.value || deleting.value) return
  deleting.value = true
  try {
    await teamAPI.deleteConsumer(teamId.value, deletingConsumer.value.id)
    appStore.showSuccess(t('admin.team.consumers.consumerDeleted'))
    showDeleteDialog.value = false
    deletingConsumer.value = null
    await fetchConsumers()
  } catch (err: any) {
    appStore.showError(err?.message || t('admin.team.consumers.deleteFailed'))
    showDeleteDialog.value = false
    deletingConsumer.value = null
  } finally {
    deleting.value = false
  }
}
</script>
