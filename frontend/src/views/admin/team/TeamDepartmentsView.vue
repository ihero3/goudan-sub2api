<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('admin.team.departments.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('admin.team.departments.subtitle') }}
          </p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          <Icon name="plus" size="md" class="mr-2" />
          {{ t('admin.team.departments.addDepartment') }}
        </button>
      </div>

      <!-- Search & Filter -->
      <div class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800">
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-64">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="searchQuery"
              type="text"
              :placeholder="t('admin.team.departments.searchPlaceholder')"
              class="input pl-10"
            />
          </div>
          <div class="w-full sm:w-40">
            <Select
              v-model="statusFilter"
              :options="statusOptions"
            />
          </div>
          <!-- View Mode Toggle -->
          <div class="ml-auto flex items-center gap-1 rounded-lg border border-gray-200 bg-gray-50 p-1 dark:border-dark-600 dark:bg-dark-700/50">
            <button
              type="button"
              @click="viewMode = 'tree'"
              :class="[
                'inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                viewMode === 'tree'
                  ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-400'
                  : 'text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-gray-300'
              ]"
              :title="t('admin.team.departments.treeView')"
            >
              <Icon name="users" size="sm" />
              {{ t('admin.team.departments.treeView') }}
            </button>
            <button
              type="button"
              @click="viewMode = 'grid'"
              :class="[
                'inline-flex items-center gap-1.5 rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                viewMode === 'grid'
                  ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-400'
                  : 'text-gray-500 hover:text-gray-700 dark:text-dark-400 dark:hover:text-gray-300'
              ]"
              :title="t('admin.team.departments.gridView')"
            >
              <Icon name="grid" size="sm" />
              {{ t('admin.team.departments.gridView') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex justify-center py-12">
        <Icon name="refresh" size="lg" class="animate-spin text-gray-400" />
      </div>

      <!-- Tree View -->
      <DepartmentTree
        v-else-if="viewMode === 'tree' && departmentTree.length > 0"
        :departments="departmentTree"
        @add="handleTreeAdd"
        @edit="handleTreeEdit"
        @delete="handleTreeDelete"
        @reorder="handleReorder"
      />

      <!-- Departments Grid -->
      <div
        v-else-if="viewMode === 'grid' && filteredDepartments.length > 0"
        class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
      >
        <div
          v-for="dept in filteredDepartments"
          :key="dept.id"
          class="group rounded-xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md dark:border-dark-600 dark:bg-dark-800"
        >
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <div
                class="flex h-12 w-12 items-center justify-center rounded-lg"
                :class="dept.colorClass"
              >
                <Icon :name="(dept.icon as any)" size="md" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">{{ dept.name }}</h3>
                <p class="text-xs text-gray-500 dark:text-dark-400">{{ dept.code }}</p>
              </div>
            </div>
            <div class="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100">
              <button
                @click="handleEdit(dept)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleDelete(dept)"
                class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-2 gap-3">
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700/50">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('common.status') }}</p>
              <p class="mt-0.5">
                <span
                  :class="['badge', dept.status === 'active' ? 'badge-success' : 'badge-gray']"
                >
                  {{ t('admin.team.departments.status.' + dept.status) }}
                </span>
              </p>
            </div>
            <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-700/50">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.departments.parentDepartment') }}</p>
              <p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">{{ dept.parentName }}</p>
            </div>
          </div>

          <div class="mt-3">
            <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.departments.description') }}</p>
            <p class="mt-0.5 text-sm font-medium text-gray-700 dark:text-gray-300 line-clamp-2">
              {{ dept.description || '-' }}
            </p>
          </div>

          <div class="mt-3 border-t border-gray-100 pt-3 dark:border-dark-700">
            <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.team.departments.createdAt') }}</p>
            <p class="mt-0.5 text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(dept.created_at) }}</p>
          </div>
        </div>

        <!-- Add New Placeholder Card -->
        <button
          @click="showCreateModal = true"
          class="flex flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed border-gray-300 p-5 transition-colors hover:border-primary-400 hover:bg-primary-50/50 dark:border-dark-600 dark:hover:border-primary-500 dark:hover:bg-primary-900/10"
        >
          <div class="flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
            <Icon name="plus" size="md" class="text-gray-500 dark:text-dark-400" />
          </div>
          <span class="text-sm font-medium text-gray-600 dark:text-dark-300">{{ t('admin.team.departments.addDepartment') }}</span>
        </button>
      </div>

      <!-- Empty State -->
      <EmptyState
        v-else
        :title="hasFilters ? t('admin.team.departments.noResults') : t('admin.team.departments.noDepartments')"
        :description="hasFilters ? t('admin.team.departments.noResultsDescription') : t('admin.team.departments.addFirstDepartment')"
        :action-text="hasFilters ? undefined : t('admin.team.departments.addDepartment')"
        @action="showCreateModal = true"
      />
    </div>

    <!-- Create Dialog -->
    <DepartmentForm
      :show="showCreateModal"
      mode="create"
      :departments="departmentTree"
      :is-submitting="createLoading"
      :default-parent-id="defaultParentId"
      @close="handleCreateClose"
      @submit="handleCreateDepartment"
    />

    <!-- Edit Dialog -->
    <DepartmentForm
      :show="showEditModal"
      mode="edit"
      :department="editingDepartment"
      :departments="departmentTree"
      :is-submitting="updateLoading"
      @close="handleEditClose"
      @submit="handleUpdateDepartment"
    />

    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.team.departments.deleteDepartment')"
      :message="t('admin.team.departments.deleteConfirm', { name: deletingDepartment?.name })"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTeamContext } from '@/composables/useTeamContext'
import { formatDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import DepartmentForm from '@/components/team/DepartmentForm.vue'
import DepartmentTree from '@/components/team/DepartmentTree.vue'
import { teamAPI } from '@/api/team'
import type { Department as ApiDepartment, DepartmentTreeNode } from '@/api/team'

const { t } = useI18n()
const appStore = useAppStore()

// ==================== Types ====================

interface DisplayDepartment extends ApiDepartment {
  code: string
  icon: string
  colorClass: string
  parentName: string
}

const iconMap: Record<string, string> = {
  Engineering: 'terminal',
  Product: 'cube',
  Design: 'sparkles',
  Marketing: 'chatBubble',
  Sales: 'trendingUp',
}

const colorClasses: Record<string, string> = {
  Engineering: 'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
  Product: 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
  Design: 'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
  Marketing: 'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
  Sales: 'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
}

const defaultColorClass = 'bg-gray-100 text-gray-600 dark:bg-gray-900/30 dark:text-gray-400'

// ==================== Data ====================

const departments = ref<ApiDepartment[]>([])
const departmentTree = ref<DepartmentTreeNode[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

const searchQuery = ref('')
const statusFilter = ref<string>('')
const viewMode = ref<'tree' | 'grid'>('tree')
const defaultParentId = ref<number | null>(null)

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteDialog = ref(false)
const editingDepartment = ref<(ApiDepartment & { code?: string }) | null>(null)
const deletingDepartment = ref<DisplayDepartment | null>(null)
const createLoading = ref(false)
const updateLoading = ref(false)
const deleteLoading = ref(false)

const { teamId, fetchCurrentTeam, loadError } = useTeamContext()

// ==================== Computed ====================

const statusOptions = computed(() => [
  { value: '', label: t('admin.team.departments.allStatus') },
  { value: 'active', label: t('admin.team.departments.status.active') },
  { value: 'inactive', label: t('admin.team.departments.status.inactive') },
])

const hasFilters = computed(
  () => searchQuery.value.trim() !== '' || statusFilter.value !== ''
)

const getParentName = (parentId: number | null): string => {
  if (parentId === null) return t('admin.team.departments.noParent')
  const parent = departments.value.find((d) => d.id === parentId)
  return parent?.name || '-'
}

const displayDepartments = computed<DisplayDepartment[]>(() =>
  departments.value.map((dept) => ({
    ...dept,
    code: dept.cost_center_code || dept.name.slice(0, 3).toUpperCase(),
    icon: iconMap[dept.name] || 'users',
    colorClass: colorClasses[dept.name] || defaultColorClass,
    parentName: getParentName(dept.parent_id),
  }))
)

const filteredDepartments = computed<DisplayDepartment[]>(() => {
  let result = displayDepartments.value
  const query = searchQuery.value.trim().toLowerCase()
  if (query) {
    result = result.filter(
      (d) =>
        d.name.toLowerCase().includes(query) ||
        (d.description || '').toLowerCase().includes(query)
    )
  }
  if (statusFilter.value) {
    result = result.filter((d) => d.status === statusFilter.value)
  }
  return result
})

// ==================== Fetch ====================

const fetchDepartments = async () => {
  if (!teamId.value) return
  loading.value = true
  error.value = null
  try {
    const response = await teamAPI.listDepartments(teamId.value)
    departments.value = response.items ?? []
  } catch (err: any) {
    error.value = err.message || t('common.error')
    appStore.showError(t('admin.team.departments.fetchFailed'))
  } finally {
    loading.value = false
  }
}

const fetchDepartmentTree = async () => {
  if (!teamId.value) return
  try {
    const data = await teamAPI.getDepartmentTree(teamId.value)
    departmentTree.value = data || []
  } catch {
    departmentTree.value = []
  }
}

onMounted(async () => {
  await fetchCurrentTeam()
  if (!teamId.value) {
    // 团队上下文加载失败时不阻塞页面渲染，但提示用户
    if (loadError.value) {
      appStore.showError(loadError.value)
    }
    return
  }
  fetchDepartments()
  fetchDepartmentTree()
})

// ==================== Handlers ====================

const handleCreateDepartment = async (payload: any) => {
  if (!teamId.value) {
    appStore.showError(t('admin.team.departments.createFailed'))
    await fetchCurrentTeam()
    return
  }
  createLoading.value = true
  try {
    await teamAPI.createDepartment(teamId.value, {
      name: payload.name,
      cost_center_code: payload.cost_center_code,
      parent_id: payload.parent_id,
      description: payload.description || null,
    })
    appStore.showSuccess(t('admin.team.departments.createSuccess'))
    showCreateModal.value = false
    defaultParentId.value = null
    await fetchDepartments()
    await fetchDepartmentTree()
  } catch (err: any) {
    appStore.showError(err.message || t('admin.team.departments.createFailed'))
  } finally {
    createLoading.value = false
  }
}

const handleCreateClose = () => {
  showCreateModal.value = false
  defaultParentId.value = null
}

const handleEdit = (dept: DisplayDepartment) => {
  // API doesn't return a code field; derive one for the form display
  editingDepartment.value = {
    ...dept,
    code: dept.code,
  }
  showEditModal.value = true
}

const handleEditClose = () => {
  showEditModal.value = false
  editingDepartment.value = null
}

const handleUpdateDepartment = async (payload: any) => {
  if (!editingDepartment.value) return
  updateLoading.value = true
  try {
    await teamAPI.updateDepartment(teamId.value, editingDepartment.value.id, {
      name: payload.name,
      cost_center_code: payload.cost_center_code,
      // Backend treats parent_id=0 as "move to root" (null parent);
      // a nil pointer (JSON null) means "skip update", so convert null → 0.
      parent_id: payload.parent_id ?? 0,
      description: payload.description || null,
      status: payload.status,
    })
    appStore.showSuccess(t('admin.team.departments.editSuccess'))
    showEditModal.value = false
    editingDepartment.value = null
    await fetchDepartments()
    await fetchDepartmentTree()
  } catch (err: any) {
    appStore.showError(err.message || t('admin.team.departments.updateFailed'))
  } finally {
    updateLoading.value = false
  }
}

const handleDelete = (dept: DisplayDepartment) => {
  deletingDepartment.value = dept
  showDeleteDialog.value = true
}

const confirmDelete = async () => {
  if (!deletingDepartment.value || deleteLoading.value) return
  deleteLoading.value = true
  try {
    await teamAPI.deleteDepartment(teamId.value, deletingDepartment.value.id)
    appStore.showSuccess(t('admin.team.departments.departmentDeleted'))
    showDeleteDialog.value = false
    deletingDepartment.value = null
    await fetchDepartments()
    await fetchDepartmentTree()
  } catch (err: any) {
    appStore.showError(err.message || t('admin.team.departments.deleteFailed'))
  } finally {
    deleteLoading.value = false
  }
}

// ==================== Tree Handlers ====================

const handleTreeAdd = (parentId?: number) => {
  defaultParentId.value = parentId ?? null
  showCreateModal.value = true
}

const handleTreeEdit = (dept: DepartmentTreeNode) => {
  editingDepartment.value = {
    ...dept,
    code: dept.cost_center_code || dept.name.slice(0, 3).toUpperCase(),
  }
  showEditModal.value = true
}

const handleTreeDelete = async (dept: DepartmentTreeNode) => {
  if (!teamId.value) return
  try {
    await teamAPI.deleteDepartment(teamId.value, dept.id)
    appStore.showSuccess(t('admin.team.departments.departmentDeleted'))
    await fetchDepartments()
    await fetchDepartmentTree()
  } catch (err: any) {
    appStore.showError(err.message || t('admin.team.departments.deleteFailed'))
  }
}

const findNodeById = (nodes: DepartmentTreeNode[], id: number): DepartmentTreeNode | null => {
  for (const node of nodes) {
    if (node.id === id) return node
    if (node.children) {
      const found = findNodeById(node.children, id)
      if (found) return found
    }
  }
  return null
}

const isDescendantOrSelf = (rootId: number, targetId: number): boolean => {
  if (rootId === targetId) return true
  const root = findNodeById(departmentTree.value, rootId)
  if (!root || !root.children) return false
  for (const child of root.children) {
    if (isDescendantOrSelf(child.id, targetId)) return true
  }
  return false
}

const handleReorder = async (
  draggedId: number,
  targetParentId: number | null,
  _targetIndex: number
) => {
  if (!teamId.value) return
  // Prevent moving a department under itself or its descendants (cycle)
  if (targetParentId !== null && isDescendantOrSelf(draggedId, targetParentId)) {
    appStore.showError(t('admin.team.departments.reorderCycle'))
    return
  }
  // Skip if parent hasn't changed
  const dragged = findNodeById(departmentTree.value, draggedId)
  if (dragged && dragged.parent_id === targetParentId) return
  try {
    // Backend treats parent_id=0 as "move to root" (null parent);
    // a nil pointer (JSON null) means "skip update", so convert null → 0.
    await teamAPI.updateDepartment(teamId.value, draggedId, {
      parent_id: targetParentId ?? 0,
    })
    appStore.showSuccess(t('admin.team.departments.reorderSuccess'))
    await fetchDepartments()
    await fetchDepartmentTree()
  } catch (err: any) {
    appStore.showError(err.message || t('admin.team.departments.reorderFailed'))
  }
}
</script>
