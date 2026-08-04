<template>
  <div class="rounded-xl border border-zinc-200 bg-white p-5 dark:border-zinc-700 dark:bg-zinc-800">
    <!-- Header -->
    <div class="mb-4 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-100 dark:bg-indigo-900/30">
          <svg class="h-4 w-4 text-indigo-600 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182 1.105-.88 2.996-.88 4.101 0l.88.659M12 18a.75.75 0 0 1 .75.75v.008c0 .414-.336.75-.75.75h-.008a.75.75 0 0 1-.75-.75v-.008c0-.414.336-.75.75-.75H12Z" />
          </svg>
        </div>
        <h3 class="text-base font-semibold text-zinc-900 dark:text-zinc-100">{{ t('team.consumer.quota.title') }}</h3>
      </div>
      <button
        v-if="canEdit"
        type="button"
        class="rounded-lg p-1.5 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-indigo-600 dark:hover:bg-zinc-700 dark:hover:text-indigo-400"
        :title="t('team.consumer.quota.edit')"
        @click="startEditing"
      >
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.53-3.395a4.5 4.5 0 011.13-1.897l8.899-8.898zM16.862 4.487L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
        </svg>
      </button>
    </div>

    <!-- Quota Stats -->
    <div class="mb-4 grid grid-cols-3 gap-3">
      <div class="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-700/50">
        <p class="text-xs text-zinc-500 dark:text-zinc-400">{{ t('team.consumer.quota.monthlyLimit') }}</p>
        <p class="mt-1 text-lg font-semibold text-zinc-900 dark:text-zinc-100">{{ formatNumber(quota.monthlyLimit) }}</p>
      </div>
      <div class="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-700/50">
        <p class="text-xs text-zinc-500 dark:text-zinc-400">{{ t('team.consumer.quota.used') }}</p>
        <p class="mt-1 text-lg font-semibold text-zinc-900 dark:text-zinc-100">{{ formatNumber(quota.used) }}</p>
      </div>
      <div class="rounded-lg bg-zinc-50 p-3 dark:bg-zinc-700/50">
        <p class="text-xs text-zinc-500 dark:text-zinc-400">{{ t('team.consumer.quota.remaining') }}</p>
        <p class="mt-1 text-lg font-semibold" :class="remainingClass">{{ formatNumber(remaining) }}</p>
      </div>
    </div>

    <!-- Progress Bar -->
    <div class="mb-4">
      <div class="mb-1.5 flex items-center justify-between text-xs">
        <span class="text-zinc-500 dark:text-zinc-400">{{ t('team.consumer.quota.usageRate') }}</span>
        <span class="font-medium" :class="usagePercent > 90 ? 'text-red-600 dark:text-red-400' : 'text-zinc-700 dark:text-zinc-300'">
          {{ usagePercent.toFixed(1) }}%
        </span>
      </div>
      <div class="h-2.5 w-full overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-700">
        <div
          class="h-full rounded-full transition-all duration-500 ease-out"
          :class="progressBarColorClass"
          :style="{ width: `${Math.min(usagePercent, 100)}%` }"
        />
      </div>
      <div class="mt-1 flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400">
        <span>{{ formatNumber(quota.used) }} / {{ formatNumber(quota.monthlyLimit) }}</span>
        <span v-if="usagePercent > 90" class="font-medium text-red-600 dark:text-red-400">
          {{ t('team.consumer.quota.nearLimit') }}
        </span>
      </div>
    </div>

    <!-- Edit Quota Form -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-96"
      leave-active-class="transition-all duration-150 ease-in"
      leave-from-class="opacity-100 max-h-96"
      leave-to-class="opacity-0 max-h-0"
    >
      <div v-if="isEditing" class="overflow-hidden">
        <div class="rounded-lg border border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-700/30">
          <h4 class="mb-3 text-sm font-medium text-zinc-900 dark:text-zinc-100">
            {{ t('team.consumer.quota.editTitle') }}
          </h4>
          <div class="space-y-3">
            <div>
              <label class="mb-1 block text-xs font-medium text-zinc-700 dark:text-zinc-300">
                {{ t('team.consumer.quota.monthlyLimitLabel') }}
              </label>
              <input
                v-model.number="editForm.monthlyLimit"
                type="number"
                min="0"
                class="w-full rounded-lg border border-zinc-300 bg-white px-3 py-2 text-sm text-zinc-900 transition-colors focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-100"
                :placeholder="t('team.consumer.quota.monthlyLimitPlaceholder')"
              />
            </div>
            <div class="flex items-center gap-3 pt-1">
              <button
                type="button"
                class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 dark:focus:ring-offset-zinc-800"
                @click="saveQuota"
              >
                {{ t('common.save') }}
              </button>
              <button
                type="button"
                class="rounded-lg border border-zinc-300 bg-white px-4 py-2 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-50 dark:border-zinc-600 dark:bg-zinc-700 dark:text-zinc-300 dark:hover:bg-zinc-600"
                @click="cancelEditing"
              >
                {{ t('common.cancel') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Quota History (Optional) -->
    <div v-if="showHistory && quotaHistory.length > 0" class="mt-4 border-t border-zinc-100 pt-4 dark:border-zinc-700">
      <button
        type="button"
        class="mb-2 flex items-center gap-1 text-sm font-medium text-zinc-700 transition-colors hover:text-indigo-600 dark:text-zinc-300 dark:hover:text-indigo-400"
        @click="historyExpanded = !historyExpanded"
      >
        <svg
          class="h-4 w-4 transition-transform duration-200"
          :class="historyExpanded ? 'rotate-90' : ''"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
        {{ t('team.consumer.quota.history') }}
      </button>
      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 max-h-0"
        enter-to-class="opacity-100 max-h-96"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 max-h-96"
        leave-to-class="opacity-0 max-h-0"
      >
        <div v-if="historyExpanded" class="overflow-hidden">
          <div class="space-y-2">
            <div
              v-for="item in quotaHistory"
              :key="item.id"
              class="flex items-center justify-between rounded-lg bg-zinc-50 px-3 py-2 dark:bg-zinc-700/30"
            >
              <div class="flex items-center gap-2">
                <span
                  class="inline-flex h-2 w-2 shrink-0 rounded-full"
                  :class="item.action === 'increase' ? 'bg-green-500' : 'bg-amber-500'"
                />
                <span class="text-sm text-zinc-700 dark:text-zinc-300">{{ item.description }}</span>
              </div>
              <div class="text-right">
                <span
                  class="text-sm font-medium"
                  :class="item.action === 'increase' ? 'text-green-600 dark:text-green-400' : 'text-amber-600 dark:text-amber-400'"
                >
                  {{ item.action === 'increase' ? '+' : '-' }}{{ formatNumber(item.amount) }}
                </span>
                <p class="text-xs text-zinc-400 dark:text-zinc-500">{{ item.date }}</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

// ==================== Types ====================

export interface QuotaData {
  monthlyLimit: number
  used: number
}

export interface QuotaHistoryItem {
  id: number
  action: 'increase' | 'decrease'
  amount: number
  description: string
  date: string
}

interface Props {
  quota: QuotaData
  canEdit?: boolean
  showHistory?: boolean
  quotaHistory?: QuotaHistoryItem[]
}

// ==================== Props & Emits ====================

const props = withDefaults(defineProps<Props>(), {
  canEdit: false,
  showHistory: false,
  quotaHistory: () => [],
})

const emit = defineEmits<{
  'update:quota': [quota: QuotaData]
}>()

// ==================== State ====================

const isEditing = ref(false)
const historyExpanded = ref(false)

const editForm = reactive<QuotaData>({
  monthlyLimit: 0,
  used: 0,
})

// ==================== Computed ====================

const remaining = computed(() => Math.max(0, props.quota.monthlyLimit - props.quota.used))

const usagePercent = computed(() => {
  if (props.quota.monthlyLimit <= 0) return 0
  return (props.quota.used / props.quota.monthlyLimit) * 100
})

const remainingClass = computed(() => {
  if (remaining.value === 0) return 'text-red-600 dark:text-red-400'
  if (remaining.value < props.quota.monthlyLimit * 0.1) return 'text-amber-600 dark:text-amber-400'
  return 'text-green-600 dark:text-green-400'
})

const progressBarColorClass = computed(() => {
  if (usagePercent.value > 90) return 'bg-red-500'
  if (usagePercent.value > 70) return 'bg-amber-500'
  return 'bg-green-500'
})

// ==================== Methods ====================

function formatNumber(n: number): string {
  if (n >= 1_000_000) {
    return (n / 1_000_000).toFixed(1) + 'M'
  }
  if (n >= 1_000) {
    return (n / 1_000).toFixed(1) + 'K'
  }
  return String(n)
}

function startEditing(): void {
  editForm.monthlyLimit = props.quota.monthlyLimit
  editForm.used = props.quota.used
  isEditing.value = true
}

function cancelEditing(): void {
  isEditing.value = false
}

function saveQuota(): void {
  emit('update:quota', {
    monthlyLimit: Math.max(0, editForm.monthlyLimit),
    used: props.quota.used,
  })
  isEditing.value = false
}

// ==================== Watchers ====================

watch(() => props.quota, (newQuota) => {
  if (!isEditing.value) {
    editForm.monthlyLimit = newQuota.monthlyLimit
    editForm.used = newQuota.used
  }
}, { deep: true, immediate: true })
</script>
