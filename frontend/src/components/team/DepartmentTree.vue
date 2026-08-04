<template>
  <div class="department-tree">
    <!-- Tree Header -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <button
          @click="expandAll"
          class="rounded-md px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-primary-400"
        >
          {{ t('common.expandAll') }}
        </button>
        <button
          @click="collapseAll"
          class="rounded-md px-3 py-1.5 text-xs font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:text-gray-400 dark:hover:bg-dark-700 dark:hover:text-primary-400"
        >
          {{ t('common.collapseAll') }}
        </button>
      </div>
      <button
        @click="handleAdd()"
        class="inline-flex items-center gap-1.5 rounded-lg bg-primary-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-primary-700"
      >
        <Icon name="plus" size="sm" />
        {{ t('admin.team.departments.addDepartment') }}
      </button>
    </div>

    <!-- Tree Content -->
    <div class="rounded-xl border border-gray-200 dark:border-dark-600">
      <div v-if="treeData.length === 0" class="flex flex-col items-center justify-center py-12">
        <div
          class="flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="users" size="lg" class="text-gray-400 dark:text-dark-400" />
        </div>
        <p class="mt-4 text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.team.departments.noDepartments') }}
        </p>
        <button
          @click="handleAdd()"
          class="mt-3 inline-flex items-center gap-1.5 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
        >
          <Icon name="plus" size="sm" />
          {{ t('admin.team.departments.addDepartment') }}
        </button>
      </div>

      <div v-else class="divide-y divide-gray-100 dark:divide-dark-700">
        <DepartmentTreeNode
          v-for="node in treeData"
          :key="node.id"
          :node="node"
          :level="0"
          :expanded-ids="expandedIds"
          :dragging-id="draggingId"
          :drag-over-id="dragOverId"
          @toggle="toggleExpand"
          @add="handleAdd"
          @edit="handleEdit"
          @delete="handleDelete"
          @drag-start="handleDragStart"
          @drag-over="handleDragOver"
          @drag-leave="handleDragLeave"
          @drop="handleDrop"
        />
      </div>
    </div>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.team.departments.deleteDepartment')"
      :message="
        t('admin.team.departments.deleteConfirm', { name: deletingDepartment?.name })
      "
      :danger="true"
      @confirm="confirmDelete"
      @cancel="showDeleteDialog = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DepartmentTreeNode as DeptTreeNode } from '@/api/team'
import Icon from '@/components/icons/Icon.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import DepartmentTreeNode from './DepartmentTreeNode.vue'

const { t } = useI18n()

// ==================== Props & Emits ====================

interface Props {
  departments: DeptTreeNode[]
}

interface Emits {
  (e: 'add', parentId?: number): void
  (e: 'edit', dept: DeptTreeNode): void
  (e: 'delete', dept: DeptTreeNode): void
  (e: 'reorder', draggedId: number, targetParentId: number | null, targetIndex: number): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// ==================== State ====================

const expandedIds = ref<Set<number>>(new Set())
const draggingId = ref<number | null>(null)
const dragOverId = ref<number | null>(null)
const showDeleteDialog = ref(false)
const deletingDepartment = ref<DeptTreeNode | null>(null)

// ==================== Computed ====================

const treeData = computed(() => props.departments)

// ==================== Methods ====================

function toggleExpand(id: number) {
  if (expandedIds.value.has(id)) {
    expandedIds.value.delete(id)
  } else {
    expandedIds.value.add(id)
  }
}

function expandAll() {
  const ids = new Set<number>()
  const collectIds = (nodes: DeptTreeNode[]) => {
    for (const node of nodes) {
      if (node.children && node.children.length > 0) {
        ids.add(node.id)
        collectIds(node.children)
      }
    }
  }
  collectIds(treeData.value)
  expandedIds.value = ids
}

function collapseAll() {
  expandedIds.value.clear()
}

function handleAdd(parentId?: number) {
  emit('add', parentId)
}

function handleEdit(dept: DeptTreeNode) {
  emit('edit', dept)
}

function handleDelete(dept: DeptTreeNode) {
  deletingDepartment.value = dept
  showDeleteDialog.value = true
}

function confirmDelete() {
  if (deletingDepartment.value) {
    emit('delete', deletingDepartment.value)
    showDeleteDialog.value = false
    deletingDepartment.value = null
  }
}

// ==================== Drag & Drop ====================

function handleDragStart(id: number) {
  draggingId.value = id
}

function handleDragOver(id: number) {
  if (draggingId.value !== null && draggingId.value !== id) {
    dragOverId.value = id
  }
}

function handleDragLeave() {
  dragOverId.value = null
}

function handleDrop(draggedId: number, targetParentId: number | null, targetIndex: number) {
  if (draggedId !== dragOverId.value) {
    emit('reorder', draggedId, targetParentId, targetIndex)
  }
  draggingId.value = null
  dragOverId.value = null
}

// ==================== Watch ====================

// Auto-expand nodes with children on initial load
watch(
  () => props.departments,
  (newDepts) => {
    if (newDepts.length > 0 && expandedIds.value.size === 0) {
      const collectIds = (nodes: DeptTreeNode[]) => {
        for (const node of nodes) {
          if (node.children && node.children.length > 0) {
            expandedIds.value.add(node.id)
            collectIds(node.children)
          }
        }
      }
      collectIds(newDepts)
    }
  },
  { immediate: true }
)
</script>
