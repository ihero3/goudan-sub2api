<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        @click.self="handleClose"
      >
        <Transition
          enter-active-class="transition duration-300 ease-out"
          enter-from-class="opacity-0 scale-95"
          enter-to-class="opacity-100 scale-100"
          leave-active-class="transition duration-200 ease-in"
          leave-from-class="opacity-100 scale-100"
          leave-to-class="opacity-0 scale-95"
        >
          <div
            v-if="modelValue"
            class="mx-4 w-full max-w-2xl rounded-2xl border border-zinc-200 bg-white shadow-2xl dark:border-zinc-700 dark:bg-zinc-800"
          >
            <!-- Header -->
            <div class="flex items-center justify-between border-b border-zinc-100 px-6 py-4 dark:border-zinc-700">
              <h3 class="text-lg font-semibold text-zinc-900 dark:text-zinc-100">
                {{ mode === 'create' ? t('team.consumer.createTitle') : t('team.consumer.editTitle') }}
              </h3>
              <button
                type="button"
                class="rounded-lg p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600 dark:hover:bg-zinc-700 dark:hover:text-zinc-300"
                @click="handleClose"
              >
                <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Form -->
            <form @submit.prevent="handleSubmit" class="space-y-5 p-6">
              <!-- Type (Radio with icons) - first field, full width -->
              <div>
                <label class="mb-2 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                  {{ t('team.consumer.type') }}
                  <span class="text-red-500">*</span>
                </label>
                <div class="flex flex-wrap gap-3">
                  <label
                    v-for="typeOption in typeOptions"
                    :key="typeOption.value"
                    class="flex cursor-pointer items-center gap-2 rounded-lg border px-4 py-2.5 transition-all"
                    :class="[
                      form.type === typeOption.value
                        ? 'border-indigo-500 bg-indigo-50 text-indigo-700 dark:border-indigo-400 dark:bg-indigo-900/20 dark:text-indigo-300'
                        : 'border-zinc-200 bg-white text-zinc-700 hover:border-zinc-300 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-300 dark:hover:border-zinc-500'
                    ]"
                  >
                    <input
                      v-model="form.type"
                      type="radio"
                      :value="typeOption.value"
                      class="sr-only"
                    />
                    <component :is="typeOption.icon" class="h-5 w-5 shrink-0" />
                    <span class="text-sm font-medium">{{ typeOption.label }}</span>
                  </label>
                </div>
                <p v-if="errors.type" class="mt-1 text-xs text-red-500">{{ errors.type }}</p>
              </div>

              <!-- Two-column grid for basic fields -->
              <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <!-- Name -->
                <div>
                  <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                    {{ t('team.consumer.name') }}
                    <span class="text-red-500">*</span>
                  </label>
                  <input
                    v-model="form.name"
                    type="text"
                    :placeholder="t('team.consumer.namePlaceholder')"
                    class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100 dark:placeholder-zinc-500"
                    :class="errors.name ? 'border-red-500 focus:border-red-500 focus:ring-red-500/20' : ''"
                  />
                  <p v-if="errors.name" class="mt-1 text-xs text-red-500">{{ errors.name }}</p>
                </div>

                <!-- Department (Tree Select) -->
                <div>
                  <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                    {{ t('team.consumer.department') }}
                    <span class="text-red-500">*</span>
                  </label>
                  <DepartmentTreeSelect
                    :model-value="form.departmentId"
                    :departments="departments"
                    :placeholder="t('team.consumer.selectDepartment')"
                    :empty-text="t('team.consumer.noDepartments')"
                    :error="errors.departmentId"
                    @update:model-value="(v: number | null) => form.departmentId = v ?? ''"
                  />
                </div>

                <!-- Email -->
                <div>
                  <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                    {{ t('team.consumer.email') }}
                  </label>
                  <input
                    v-model="form.email"
                    type="email"
                    :placeholder="t('team.consumer.emailPlaceholder')"
                    class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100 dark:placeholder-zinc-500"
                  />
                </div>

                <!-- Phone -->
                <div>
                  <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                    {{ t('team.consumer.phone') }}
                  </label>
                  <input
                    v-model="form.phone"
                    type="tel"
                    :placeholder="t('team.consumer.phonePlaceholder')"
                    class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100 dark:placeholder-zinc-500"
                  />
                </div>

                <!-- Title -->
                <div>
                  <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                    {{ t('team.consumer.title') }}
                  </label>
                  <input
                    v-model="form.title"
                    type="text"
                    :placeholder="t('team.consumer.titlePlaceholder')"
                    class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100 dark:placeholder-zinc-500"
                  />
                </div>
              </div>

              <!-- Description (full width) -->
              <div>
                <label class="mb-1.5 block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                  {{ t('admin.team.consumers.descriptionLabel') }}
                </label>
                <textarea
                  v-model="form.description"
                  rows="3"
                  :placeholder="t('admin.team.consumers.descriptionPlaceholder')"
                  class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 placeholder-zinc-400 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100 dark:placeholder-zinc-500"
                ></textarea>
              </div>

              <!-- Status (Toggle) -->
              <div class="flex items-center justify-between rounded-lg border border-zinc-200 bg-zinc-50 px-4 py-3 dark:border-zinc-700 dark:bg-zinc-700/50">
                <div>
                  <p class="text-sm font-medium text-zinc-900 dark:text-zinc-100">{{ t('team.consumer.status') }}</p>
                  <p class="text-xs text-zinc-500 dark:text-zinc-400">
                    {{ form.status === 'active' ? t('team.consumer.statusActive') : t('team.consumer.statusInactive') }}
                  </p>
                </div>
                <button
                  type="button"
                  class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-zinc-800"
                  :class="form.status === 'active' ? 'bg-indigo-600' : 'bg-zinc-300 dark:bg-zinc-600'"
                  @click="toggleStatus"
                >
                  <span
                    class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow transition duration-200 ease-in-out"
                    :class="form.status === 'active' ? 'translate-x-5' : 'translate-x-0.5'"
                  />
                </button>
              </div>

              <!-- Actions -->
              <div class="flex items-center justify-end gap-3 pt-2">
                <button
                  type="button"
                  class="rounded-lg border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-600"
                  @click="handleClose"
                >
                  {{ t('common.cancel') }}
                </button>
                <button
                  type="submit"
                  class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-zinc-800"
                >
                  {{ mode === 'create' ? t('common.create') : t('common.save') }}
                </button>
              </div>
            </form>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { reactive, watch, computed, h, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import DepartmentTreeSelect from '@/components/team/DepartmentTreeSelect.vue'
import type { DepartmentTreeNode } from '@/api/team'

const { t } = useI18n()

// ==================== Types ====================

export type ConsumerType = 'person' | 'application' | 'service_account'
export type ConsumerStatus = 'active' | 'inactive'

export interface ConsumerFormData {
  name: string
  email: string
  phone: string
  title: string
  type: ConsumerType | ''
  departmentId: number | ''
  status: ConsumerStatus
  description: string
}

interface Props {
  modelValue: boolean
  mode: 'create' | 'edit'
  initialData?: Partial<ConsumerFormData>
  departments: DepartmentTreeNode[]
}

// ==================== Props & Emits ====================

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  submit: [data: ConsumerFormData]
}>()

// ==================== Icons ====================

const PersonIcon: Component = {
  render() {
    return h('svg', {
      xmlns: 'http://www.w3.org/2000/svg',
      fill: 'none',
      viewBox: '0 0 24 24',
      'stroke-width': '1.5',
      stroke: 'currentColor',
    }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z',
      }),
    ])
  },
}

const ApplicationIcon: Component = {
  render() {
    return h('svg', {
      xmlns: 'http://www.w3.org/2000/svg',
      fill: 'none',
      viewBox: '0 0 24 24',
      'stroke-width': '1.5',
      stroke: 'currentColor',
    }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M17.25 6.75 22.5 3l-5.25-3.75L12 3l-5.25-3.75L1.5 3l5.25 3.75L1.5 10.5l5.25 3.75L12 10.5l5.25 3.75 5.25-3.75-5.25-3.75ZM12 10.5l-5.25 3.75L12 18l5.25-3.75L12 10.5Z',
      }),
    ])
  },
}

const ServiceAccountIcon: Component = {
  render() {
    return h('svg', {
      xmlns: 'http://www.w3.org/2000/svg',
      fill: 'none',
      viewBox: '0 0 24 24',
      'stroke-width': '1.5',
      stroke: 'currentColor',
    }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.59l7.155 7.155a2.25 2.25 0 0 0 3.182 0l7.155-7.155a2.25 2.25 0 0 0 .659-1.59V5.25A2.25 2.25 0 0 0 18.75 3h-4.318c-.597 0-1.17.237-1.59.659l-.9.9a.75.75 0 0 1-1.06 0l-.9-.9A2.249 2.249 0 0 0 9.568 3Z',
      }),
    ])
  },
}

// ==================== Data ====================

const defaultForm = (): ConsumerFormData => ({
  name: '',
  email: '',
  phone: '',
  title: '',
  type: '',
  departmentId: '',
  status: 'active',
  description: '',
})

const form = reactive<ConsumerFormData>({ ...defaultForm() })
const errors = reactive<Partial<Record<keyof ConsumerFormData, string>>>({})

const typeOptions = computed(() => [
  { value: 'person' as ConsumerType, label: t('team.consumer.types.person'), icon: PersonIcon },
  { value: 'application' as ConsumerType, label: t('team.consumer.types.application'), icon: ApplicationIcon },
  { value: 'service_account' as ConsumerType, label: t('team.consumer.types.serviceAccount'), icon: ServiceAccountIcon },
])

// ==================== Methods ====================

function toggleStatus(): void {
  form.status = form.status === 'active' ? 'inactive' : 'active'
}

function validate(): boolean {
  let isValid = true

  // Clear previous errors
  Object.keys(errors).forEach((key) => {
    delete errors[key as keyof ConsumerFormData]
  })

  if (!form.name.trim()) {
    errors.name = t('team.consumer.errors.nameRequired')
    isValid = false
  }

  if (!form.type) {
    errors.type = t('team.consumer.errors.typeRequired')
    isValid = false
  }

  if (!form.departmentId) {
    errors.departmentId = t('team.consumer.errors.departmentRequired')
    isValid = false
  }

  return isValid
}

function handleSubmit(): void {
  if (!validate()) return
  emit('submit', { ...form })
}

function handleClose(): void {
  emit('update:modelValue', false)
}

function resetForm(): void {
  Object.assign(form, defaultForm())
  Object.keys(errors).forEach((key) => {
    delete errors[key as keyof ConsumerFormData]
  })
}

// ==================== Watchers ====================

watch(
  () => props.modelValue,
  (visible) => {
    if (visible) {
      if (props.mode === 'edit' && props.initialData) {
        Object.assign(form, { ...defaultForm(), ...props.initialData })
      } else {
        resetForm()
      }
    }
  }
)

watch(
  () => props.initialData,
  (data) => {
    if (props.mode === 'edit' && data && props.modelValue) {
      Object.assign(form, { ...defaultForm(), ...data })
    }
  },
  { deep: true }
)
</script>
