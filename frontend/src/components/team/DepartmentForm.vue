<template>
  <BaseDialog
    :show="show"
    :title="dialogTitle"
    width="normal"
    @close="handleClose"
  >
    <form @submit.prevent="handleSubmit" class="space-y-5">
      <!-- Name -->
      <div>
        <label
          for="dept-name"
          class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {{ t('admin.team.departments.name') }}
          <span class="text-red-500">*</span>
        </label>
        <input
          id="dept-name"
          v-model="form.name"
          type="text"
          :placeholder="t('admin.team.departments.namePlaceholder')"
          :class="[
            'w-full rounded-lg border bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:bg-dark-700 dark:text-white dark:placeholder-dark-400',
            errors.name
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500/20 dark:border-red-500/50'
              : 'border-gray-300 dark:border-dark-600'
          ]"
        />
        <p v-if="errors.name" class="mt-1 text-xs text-red-500">{{ errors.name }}</p>
      </div>

      <!-- Cost Center Code -->
      <div>
        <label
          for="dept-cost-center-code"
          class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {{ t('admin.team.departments.costCenterCode') }}
        </label>
        <input
          id="dept-cost-center-code"
          v-model="form.cost_center_code"
          type="text"
          :placeholder="t('admin.team.departments.costCenterCodePlaceholder')"
          :class="[
            'w-full rounded-lg border bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:bg-dark-700 dark:text-white dark:placeholder-dark-400',
            errors.cost_center_code
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500/20 dark:border-red-500/50'
              : 'border-gray-300 dark:border-dark-600'
          ]"
        />
        <p v-if="errors.cost_center_code" class="mt-1 text-xs text-red-500">{{ errors.cost_center_code }}</p>
      </div>

      <!-- Description -->
      <div>
        <label
          for="dept-description"
          class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {{ t('admin.team.departments.descriptionLabel') }}
        </label>
        <textarea
          id="dept-description"
          v-model="form.description"
          rows="3"
          :placeholder="t('admin.team.departments.descriptionPlaceholder')"
          class="w-full resize-none rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-700 dark:text-white dark:placeholder-dark-400"
        />
      </div>

      <!-- Parent Department -->
      <div>
        <label
          for="dept-parent"
          class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {{ t('admin.team.departments.parentDepartment') }}
        </label>
        <div class="relative">
          <button
            type="button"
            @click="showParentDropdown = !showParentDropdown"
            :class="[
              'flex w-full items-center justify-between rounded-lg border bg-white px-3 py-2 text-sm transition-colors focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:bg-dark-700 dark:text-white',
              errors.parent_id
                ? 'border-red-300 focus:border-red-500 focus:ring-red-500/20 dark:border-red-500/50'
                : 'border-gray-300 dark:border-dark-600'
            ]"
          >
            <span :class="selectedParentName ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-dark-400'">
              {{ selectedParentName || t('admin.team.departments.selectParent') }}
            </span>
            <Icon
              name="chevronDown"
              size="sm"
              :class="[
                'shrink-0 text-gray-400 transition-transform duration-200 dark:text-dark-400',
                showParentDropdown ? 'rotate-180' : ''
              ]"
            />
          </button>

          <!-- Parent Dropdown -->
          <Transition
            enter-active-class="transition ease-out duration-150"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition ease-in duration-100"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              v-if="showParentDropdown"
              class="absolute z-50 mt-1 w-full rounded-xl border border-gray-200 bg-white py-1 shadow-lg dark:border-dark-600 dark:bg-dark-800"
            >
              <div class="max-h-60 overflow-y-auto py-1">
                <button
                  type="button"
                  @click="selectParent(null)"
                  class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/50"
                  :class="form.parent_id === null ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300' : 'text-gray-700 dark:text-gray-300'"
                >
                  <Icon name="home" size="sm" class="text-gray-400 dark:text-dark-500" />
                  {{ t('admin.team.departments.noParent') }}
                </button>
                <div
                  v-for="option in parentOptions"
                  :key="option.id"
                  class="contents"
                >
                  <button
                    type="button"
                    @click="selectParent(option.id)"
                    class="flex w-full items-center gap-2 text-left text-sm transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/50"
                    :class="[
                      form.parent_id === option.id
                        ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300'
                        : 'text-gray-700 dark:text-gray-300',
                    ]"
                    :style="{ paddingLeft: `${12 + option.level * 20}px` }"
                  >
                    <span class="text-gray-300 dark:text-dark-500">
                      {{ '│ '.repeat(option.level) }}{{ option.level > 0 ? '├─ ' : '' }}
                    </span>
                    <span class="truncate">{{ option.name }}</span>
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
        <p v-if="errors.parent_id" class="mt-1 text-xs text-red-500">{{ errors.parent_id }}</p>
      </div>

      <!-- Status (edit mode only) -->
      <div v-if="mode === 'edit'">
        <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('common.status') }}
        </label>
        <div class="flex gap-3">
          <label class="flex cursor-pointer items-center gap-2">
            <input
              v-model="form.status"
              type="radio"
              value="active"
              class="h-4 w-4 text-primary-600 focus:ring-primary-500"
            />
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('common.active') }}</span>
          </label>
          <label class="flex cursor-pointer items-center gap-2">
            <input
              v-model="form.status"
              type="radio"
              value="inactive"
              class="h-4 w-4 text-primary-600 focus:ring-primary-500"
            />
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ t('common.inactive') }}</span>
          </label>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          @click="handleClose"
          class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 dark:focus:ring-offset-dark-800"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          @click="handleSubmit"
          :disabled="isSubmitting"
          class="inline-flex items-center gap-2 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 dark:focus:ring-offset-dark-800"
        >
          <Icon v-if="isSubmitting" name="refresh" size="sm" class="animate-spin" />
          {{ submitButtonText }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Department, DepartmentTreeNode } from '@/api/team'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

// ==================== Types ====================

interface ParentOption {
  id: number
  name: string
  level: number
}

interface FormData {
  name: string
  cost_center_code: string
  description: string
  parent_id: number | null
  status: 'active' | 'inactive'
}

// Submit payload allows null for optional fields (empty cost_center_code -> null)
interface DepartmentSubmitPayload {
  name: string
  cost_center_code: string | null
  description: string
  parent_id: number | null
  status: 'active' | 'inactive'
}

// ==================== Props & Emits ====================

interface Props {
  show: boolean
  mode: 'create' | 'edit'
  department?: Department | DepartmentTreeNode | null
  departments: DepartmentTreeNode[]
  isSubmitting?: boolean
  defaultParentId?: number | null
}

interface Emits {
  (e: 'close'): void
  (e: 'submit', data: DepartmentSubmitPayload): void
}

const props = withDefaults(defineProps<Props>(), {
  department: null,
  isSubmitting: false,
  defaultParentId: null,
})

const emit = defineEmits<Emits>()

// ==================== State ====================

const form = ref<FormData>({
  name: '',
  cost_center_code: '',
  description: '',
  parent_id: null,
  status: 'active',
})

const errors = ref<Partial<Record<keyof FormData, string>>>({})
const showParentDropdown = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)

// ==================== Computed ====================

const dialogTitle = computed(() => {
  return props.mode === 'create'
    ? t('admin.team.departments.createDepartment')
    : t('admin.team.departments.editDepartment')
})

const submitButtonText = computed(() => {
  if (props.isSubmitting) {
    return t('common.saving')
  }
  return props.mode === 'create'
    ? t('common.create')
    : t('common.save')
})

const parentOptions = computed<ParentOption[]>(() => {
  const options: ParentOption[] = []
  const collect = (nodes: DepartmentTreeNode[], level: number) => {
    for (const node of nodes) {
      // Don't allow selecting self or descendants of self as parent in edit mode
      if (props.mode === 'edit' && props.department) {
        if (node.id === props.department.id) continue
        if (isInSubtree(node, props.department.id)) continue
      }
      options.push({ id: node.id, name: node.name, level })
      if (node.children) {
        collect(node.children, level + 1)
      }
    }
  }
  collect(props.departments, 0)
  return options
})

const selectedParentName = computed(() => {
  if (form.value.parent_id === null) return ''
  const option = parentOptions.value.find((o) => o.id === form.value.parent_id)
  return option?.name || ''
})

// ==================== Methods ====================

function isInSubtree(node: DepartmentTreeNode, targetId: number): boolean {
  if (node.id === targetId) return true
  if (!node.children) return false
  for (const child of node.children) {
    if (isInSubtree(child, targetId)) return true
  }
  return false
}

function selectParent(id: number | null) {
  form.value.parent_id = id
  showParentDropdown.value = false
  if (errors.value.parent_id) {
    delete errors.value.parent_id
  }
}

function validate(): boolean {
  const newErrors: Partial<Record<keyof FormData, string>> = {}

  if (!form.value.name.trim()) {
    newErrors.name = t('validation.required', { field: t('admin.team.departments.name') })
  } else if (form.value.name.trim().length > 100) {
    newErrors.name = t('validation.maxLength', { field: t('admin.team.departments.name'), max: 100 })
  }

  if (form.value.cost_center_code.trim().length > 50) {
    newErrors.cost_center_code = t('validation.maxLength', { field: t('admin.team.departments.costCenterCode'), max: 50 })
  }

  errors.value = newErrors
  return Object.keys(newErrors).length === 0
}

function handleSubmit() {
  if (!validate()) return

  emit('submit', {
    name: form.value.name.trim(),
    cost_center_code: form.value.cost_center_code.trim() || null,
    description: form.value.description.trim() || '',
    parent_id: form.value.parent_id,
    status: form.value.status,
  })
}

function handleClose() {
  emit('close')
}

function resetForm() {
  if (props.mode === 'edit' && props.department) {
    form.value = {
      name: props.department.name,
      cost_center_code: (props.department as any).cost_center_code || '',
      description: props.department.description || '',
      parent_id: props.department.parent_id ?? null,
      status: props.department.status,
    }
  } else {
    form.value = {
      name: '',
      cost_center_code: '',
      description: '',
      parent_id: props.defaultParentId,
      status: 'active',
    }
  }
  errors.value = {}
  showParentDropdown.value = false
}

// ==================== Click Outside ====================

function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (dropdownRef.value && !dropdownRef.value.contains(target)) {
    showParentDropdown.value = false
  }
}

// ==================== Watch ====================

watch(
  () => props.show,
  (isOpen) => {
    if (isOpen) {
      resetForm()
    }
  },
  { immediate: true }
)

watch(
  () => props.department,
  () => {
    if (props.show) {
      resetForm()
    }
  }
)

// ==================== Lifecycle ====================

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
