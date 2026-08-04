<template>
  <div>
    <!-- Node Row -->
    <div
      :class="[
        'group flex items-center gap-2 px-4 py-3 transition-colors',
        isDragOver ? 'bg-primary-50 dark:bg-primary-900/20' : 'hover:bg-gray-50 dark:hover:bg-dark-700/50',
        level > 0 ? 'border-t border-gray-100 dark:border-dark-700' : ''
      ]"
      :style="{ paddingLeft: `${16 + level * 24}px` }"
      draggable="true"
      @dragstart="onDragStart"
      @dragover.prevent="onDragOver"
      @dragleave="onDragLeave"
      @drop.prevent="onDrop"
    >
      <!-- Expand/Collapse Toggle -->
      <button
        v-if="hasChildren"
        @click="toggleExpand"
        class="flex h-5 w-5 shrink-0 items-center justify-center rounded transition-colors hover:bg-gray-200 dark:hover:bg-dark-600"
        :title="isExpanded ? t('common.collapse') : t('common.expand')"
      >
        <Icon
          name="chevronRight"
          size="xs"
          :class="[
            'transition-transform duration-200',
            isExpanded ? 'rotate-90' : ''
          ]"
        />
      </button>
      <div v-else class="h-5 w-5 shrink-0" />

      <!-- Department Icon -->
      <div
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
        :class="iconColorClass"
      >
        <Icon name="users" size="sm" />
      </div>

      <!-- Department Info -->
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <span class="truncate text-sm font-medium text-gray-900 dark:text-white">
            {{ node.name }}
          </span>
          <span
            v-if="node.code"
            class="shrink-0 rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-mono font-medium text-gray-500 dark:bg-dark-600 dark:text-dark-400"
          >
            {{ node.code }}
          </span>
          <span
            v-if="node.status === 'inactive'"
            class="shrink-0 rounded-md bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium text-gray-500 dark:bg-dark-600 dark:text-dark-400"
          >
            {{ t('common.inactive') }}
          </span>
        </div>
        <p
          v-if="node.description"
          class="mt-0.5 truncate text-xs text-gray-500 dark:text-dark-400"
        >
          {{ node.description }}
        </p>
      </div>

      <!-- Member Count -->
      <div class="flex shrink-0 items-center gap-1.5">
        <div
          class="flex items-center gap-1 rounded-full bg-gray-100 px-2 py-0.5 dark:bg-dark-700"
        >
          <Icon name="users" size="xs" class="text-gray-400 dark:text-dark-500" />
          <span class="text-xs font-medium text-gray-600 dark:text-dark-300">
            {{ memberCount }}
          </span>
        </div>
      </div>

      <!-- Actions -->
      <div
        class="flex shrink-0 items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100"
      >
        <button
          @click="handleAdd"
          class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-primary-50 hover:text-primary-600 dark:hover:bg-primary-900/20 dark:hover:text-primary-400"
          :title="t('admin.team.departments.addChild')"
        >
          <Icon name="plus" size="sm" />
        </button>
        <button
          @click="handleEdit"
          class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
          :title="t('common.edit')"
        >
          <Icon name="edit" size="sm" />
        </button>
        <button
          @click="handleDelete"
          class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
          :title="t('common.delete')"
        >
          <Icon name="trash" size="sm" />
        </button>
      </div>
    </div>

    <!-- Children -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="max-h-0 opacity-0"
      enter-to-class="max-h-[2000px] opacity-100"
      leave-active-class="transition-all duration-200 ease-in"
      leave-from-class="max-h-[2000px] opacity-100"
      leave-to-class="max-h-0 opacity-0"
    >
      <div
        v-if="hasChildren && isExpanded"
        class="overflow-hidden"
      >
        <DepartmentTreeNode
          v-for="child in node.children"
          :key="child.id"
          :node="child"
          :level="level + 1"
          :expanded-ids="expandedIds"
          :dragging-id="draggingId"
          :drag-over-id="dragOverId"
          @toggle="$emit('toggle', ($event as unknown) as number)"
          @add="$emit('add', ($event as unknown) as number | undefined)"
          @edit="$emit('edit', ($event as unknown) as DeptTreeNode)"
          @delete="$emit('delete', ($event as unknown) as DeptTreeNode)"
          @drag-start="$emit('drag-start', ($event as unknown) as number)"
          @drag-over="$emit('drag-over', ($event as unknown) as number)"
          @drag-leave="$emit('drag-leave')"
          @drop="$emit('drop', ...(($event as unknown) as [number, number | null, number]))"
        />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DepartmentTreeNode as DeptTreeNode } from '@/api/team'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

// ==================== Props & Emits ====================

interface Props {
  node: DeptTreeNode
  level: number
  expandedIds: Set<number>
  draggingId: number | null
  dragOverId: number | null
}

interface Emits {
  (e: 'toggle', id: number): void
  (e: 'add', parentId?: number): void
  (e: 'edit', dept: DeptTreeNode): void
  (e: 'delete', dept: DeptTreeNode): void
  (e: 'drag-start', id: number): void
  (e: 'drag-over', id: number): void
  (e: 'drag-leave'): void
  (e: 'drop', draggedId: number, targetParentId: number | null, targetIndex: number): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// ==================== Computed ====================

const hasChildren = computed(() => props.node.children && props.node.children.length > 0)

const isExpanded = computed(() => props.expandedIds.has(props.node.id))

const isDragOver = computed(() => props.dragOverId === props.node.id)

const memberCount = computed(() => {
  // If the node has a memberCount property, use it
  // Otherwise return 0 (backend can provide this)
  return (props.node as any).memberCount ?? 0
})

const iconColorClass = computed(() => {
  const colors = [
    'bg-blue-100 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400',
    'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400',
    'bg-pink-100 text-pink-600 dark:bg-pink-900/30 dark:text-pink-400',
    'bg-amber-100 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400',
    'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400',
    'bg-teal-100 text-teal-600 dark:bg-teal-900/30 dark:text-teal-400',
    'bg-indigo-100 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-400',
    'bg-rose-100 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400',
  ]
  // Use department id to deterministically pick a color
  return colors[(props.node.id % colors.length)]
})

// ==================== Methods ====================

function toggleExpand() {
  if (hasChildren.value) {
    emit('toggle', props.node.id)
  }
}

function handleAdd() {
  emit('add', props.node.id)
}

function handleEdit() {
  emit('edit', props.node)
}

function handleDelete() {
  emit('delete', props.node)
}

// ==================== Drag & Drop ====================

function onDragStart(event: DragEvent) {
  if (event.dataTransfer) {
    event.dataTransfer.setData('text/plain', String(props.node.id))
    event.dataTransfer.effectAllowed = 'move'
  }
  emit('drag-start', props.node.id)
}

function onDragOver(event: DragEvent) {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  emit('drag-over', props.node.id)
}

function onDragLeave() {
  emit('drag-leave')
}

function onDrop(event: DragEvent) {
  const draggedId = parseInt(event.dataTransfer?.getData('text/plain') || '0', 10)
  if (draggedId && draggedId !== props.node.id) {
    emit('drop', draggedId, props.node.parent_id, 0)
  }
}
</script>
