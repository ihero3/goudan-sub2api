<template>
  <BaseDialog
    :show="show"
    :title="t('admin.team.members.inviteMember')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-5">
      <!-- Email Input -->
      <div>
        <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.team.members.emailLabel') }}
        </label>
        <input
          v-model="email"
          type="email"
          :placeholder="t('admin.team.members.emailPlaceholder')"
          :class="[
            'input w-full',
            emailError && 'border-red-500 focus:border-red-500 focus:ring-red-500/30'
          ]"
          @input="clearError"
          @keydown.enter="handleSubmit"
        />
        <p v-if="emailError" class="mt-1 text-xs text-red-500">
          {{ emailError }}
        </p>
      </div>

      <!-- Role Select -->
      <div>
        <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.team.members.roleLabel') }}
        </label>
        <Select
          v-model="selectedRole"
          :options="roleOptions"
          class="w-full"
        />
        <p class="mt-1.5 text-xs text-gray-500 dark:text-dark-400">
          {{ roleDescription }}
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
          @click="handleSubmit"
          type="button"
          :disabled="loading || !email.trim()"
          :class="[
            'rounded-md px-4 py-2 text-sm font-medium text-white focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:focus:ring-offset-dark-800',
            'bg-primary-600 hover:bg-primary-700',
            (loading || !email.trim()) && 'cursor-not-allowed opacity-50'
          ]"
        >
          <span v-if="loading" class="flex items-center gap-2">
            <LoadingSpinner size="sm" />
            {{ t('admin.team.members.inviting') }}
          </span>
          <span v-else>{{ t('admin.team.members.sendInvite') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const { t } = useI18n()

type Role = 'owner' | 'admin' | 'member' | 'viewer'

interface Props {
  show: boolean
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false
})

interface Emits {
  (e: 'close'): void
  (e: 'submit', payload: { email: string; role: Role }): void
}

const emit = defineEmits<Emits>()

const email = ref('')
const selectedRole = ref<Role>('member')
const emailError = ref('')

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

const roleOptions = computed(() => [
  { value: 'owner', label: t('admin.team.members.roles.owner') },
  { value: 'admin', label: t('admin.team.members.roles.admin') },
  { value: 'member', label: t('admin.team.members.roles.member') },
  { value: 'viewer', label: t('admin.team.members.roles.viewer') }
])

const roleDescription = computed(() => {
  switch (selectedRole.value) {
    case 'owner':
      return t('admin.team.members.roleDescriptions.owner')
    case 'admin':
      return t('admin.team.members.roleDescriptions.admin')
    case 'member':
      return t('admin.team.members.roleDescriptions.member')
    case 'viewer':
      return t('admin.team.members.roleDescriptions.viewer')
    default:
      return ''
  }
})

const clearError = () => {
  emailError.value = ''
}

const validateEmail = (): boolean => {
  const trimmed = email.value.trim()
  if (!trimmed) {
    emailError.value = t('admin.team.members.emailRequired')
    return false
  }
  if (!emailRegex.test(trimmed)) {
    emailError.value = t('admin.team.members.emailInvalid')
    return false
  }
  return true
}

const handleSubmit = () => {
  if (props.loading) return
  if (!validateEmail()) return

  emit('submit', {
    email: email.value.trim(),
    role: selectedRole.value
  })
}

const handleClose = () => {
  email.value = ''
  selectedRole.value = 'member'
  emailError.value = ''
  emit('close')
}
</script>
