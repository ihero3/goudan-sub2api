<template>
  <BaseDialog
    :show="show"
    :title="t('admin.team.consumers.offboardTitle')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-5">
      <!-- Warning Banner -->
      <div
        class="flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-3 dark:border-red-900/50 dark:bg-red-900/20"
      >
        <Icon name="exclamationTriangle" size="md" class="mt-0.5 flex-shrink-0 text-red-500 dark:text-red-400" />
        <div>
          <p class="text-sm font-medium text-red-800 dark:text-red-300">
            {{ t('admin.team.consumers.offboardWarning') }}
          </p>
          <p class="mt-0.5 text-xs text-red-600 dark:text-red-400">
            {{ t('admin.team.consumers.offboardWarningDetail', { name: consumerName }) }}
          </p>
        </div>
      </div>

      <!-- Affected API Keys -->
      <div v-if="apiKeys.length > 0">
        <p class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.team.consumers.affectedKeys') }}
        </p>
        <div
          class="max-h-40 overflow-y-auto rounded-lg border border-gray-200 dark:border-dark-600"
        >
          <div
            v-for="key in apiKeys"
            :key="key.id"
            class="flex items-center gap-2 border-b border-gray-100 px-3 py-2 text-sm last:border-0 dark:border-dark-700"
          >
            <Icon name="key" size="sm" class="text-gray-400" />
            <span class="text-gray-700 dark:text-gray-300">{{ key.name }}</span>
            <span class="ml-auto text-xs text-gray-400 dark:text-dark-400">{{ key.id }}</span>
          </div>
        </div>
        <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
          {{ t('admin.team.consumers.affectedKeysCount', { count: apiKeys.length }) }}
        </p>
      </div>
      <div v-else class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-700">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.team.consumers.noKeysFound') }}
        </p>
      </div>

      <!-- Consequences -->
      <div class="rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-900/50 dark:bg-amber-900/20">
        <p class="mb-2 text-sm font-medium text-amber-800 dark:text-amber-300">
          {{ t('admin.team.consumers.consequences') }}
        </p>
        <ul class="space-y-1.5">
          <li class="flex items-start gap-2 text-sm text-amber-700 dark:text-amber-400">
            <Icon name="ban" size="sm" class="mt-0.5 flex-shrink-0" />
            {{ t('admin.team.consumers.consequenceKeysDisabled') }}
          </li>
          <li class="flex items-start gap-2 text-sm text-amber-700 dark:text-amber-400">
            <Icon name="xCircle" size="sm" class="mt-0.5 flex-shrink-0" />
            {{ t('admin.team.consumers.consequenceConsumerInactive') }}
          </li>
          <li class="flex items-start gap-2 text-sm text-amber-700 dark:text-amber-400">
            <Icon name="exclamationCircle" size="sm" class="mt-0.5 flex-shrink-0" />
            {{ t('admin.team.consumers.consequenceIrreversible') }}
          </li>
        </ul>
      </div>

      <!-- Confirmation Input -->
      <div>
        <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.team.consumers.confirmLabel', { name: consumerName }) }}
        </label>
        <input
          v-model="confirmText"
          type="text"
          :placeholder="consumerName"
          :class="[
            'input w-full',
            confirmError && 'border-red-500 focus:border-red-500 focus:ring-red-500/30'
          ]"
          @input="clearError"
        />
        <p v-if="confirmError" class="mt-1 text-xs text-red-500">
          {{ confirmError }}
        </p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button
          @click="handleClose"
          type="button"
          class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 dark:focus:ring-offset-dark-800"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          :disabled="loading || !confirmText.trim()"
          :class="[
            'rounded-md px-4 py-2 text-sm font-medium text-white focus:outline-none focus:ring-2 focus:ring-offset-2 dark:focus:ring-offset-dark-800',
            'bg-red-600 hover:bg-red-700 focus:ring-red-500',
            (loading || !confirmText.trim()) && 'cursor-not-allowed opacity-50'
          ]"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <LoadingSpinner size="sm" color="white" />
            {{ t('admin.team.consumers.offboarding') }}
          </span>
          <span v-else>{{ t('admin.team.consumers.confirmOffboard') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const { t } = useI18n()

export interface ApiKey {
  id: string
  name: string
}

interface Props {
  show: boolean
  consumerName: string
  apiKeys: ApiKey[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

interface Emits {
  (e: 'close'): void
  (e: 'confirm'): void
}

const emit = defineEmits<Emits>()

const confirmText = ref('')
const confirmError = ref('')

const clearError = () => {
  confirmError.value = ''
}

const validateConfirm = (): boolean => {
  const trimmed = confirmText.value.trim()
  if (!trimmed) {
    confirmError.value = t('admin.team.consumers.confirmRequired')
    return false
  }
  if (trimmed !== props.consumerName) {
    confirmError.value = t('admin.team.consumers.confirmMismatch')
    return false
  }
  return true
}

const handleConfirm = () => {
  if (props.loading) return
  if (!validateConfirm()) return
  emit('confirm')
}

const handleClose = () => {
  confirmText.value = ''
  confirmError.value = ''
  emit('close')
}
</script>
