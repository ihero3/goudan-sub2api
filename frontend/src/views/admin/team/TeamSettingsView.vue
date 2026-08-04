<template>
  <AppLayout>
    <div class="mx-auto max-w-5xl space-y-6">
      <!-- Page Header -->
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
          {{ t('admin.team.settings.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
          {{ t('admin.team.settings.subtitle') }}
        </p>
      </div>

      <!-- Tab Navigation -->
      <div class="border-b border-gray-200 dark:border-dark-600">
        <nav class="-mb-px flex space-x-6" aria-label="Tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            type="button"
            :class="[
              'inline-flex items-center gap-2 border-b-2 px-1 py-3 text-sm font-medium transition-colors',
              activeTab === tab.key
                ? 'border-primary-500 text-primary-600 dark:border-primary-400 dark:text-primary-400'
                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-dark-400 dark:hover:border-dark-500 dark:hover:text-gray-300',
            ]"
            @click="activeTab = tab.key"
          >
            <Icon :name="tab.icon" size="sm" />
            {{ tab.label }}
          </button>
        </nav>
      </div>

      <!-- Tab 1: General Settings -->
      <div v-show="activeTab === 'general'" class="space-y-6">
        <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-600">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.team.settings.general.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.team.settings.general.description') }}
            </p>
          </div>
          <div class="space-y-6 p-6">
            <!-- Team Name -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.general.teamName') }}
              </label>
              <input
                v-model="generalForm.name"
                type="text"
                class="input w-full"
                :placeholder="t('admin.team.settings.general.teamNamePlaceholder')"
              />
            </div>

            <!-- Team Description -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.general.descriptionLabel') }}
              </label>
              <textarea
                v-model="generalForm.description"
                rows="3"
                class="input w-full"
                :placeholder="t('admin.team.settings.general.descriptionPlaceholder')"
              />
            </div>

            <!-- Timezone -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.general.timezone') }}
              </label>
              <select v-model="generalForm.timezone" class="input w-full">
                <option v-for="tz in timezones" :key="tz.value" :value="tz.value">
                  {{ tz.label }}
                </option>
              </select>
            </div>

            <!-- Language -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.general.language') }}
              </label>
              <select v-model="generalForm.language" class="input w-full">
                <option value="zh">{{ t('admin.team.settings.general.lang.zh') }}</option>
                <option value="en">{{ t('admin.team.settings.general.lang.en') }}</option>
              </select>
            </div>

            <!-- Save Button -->
            <div class="flex justify-end pt-2">
              <button type="button" class="btn btn-primary" @click="saveGeneral">
                <Icon name="check" size="sm" class="mr-1.5" />
                {{ t('common.save') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab 2: Members Management -->
      <div v-show="activeTab === 'members'" class="space-y-6">
        <!-- Invite Member -->
        <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-600">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.team.settings.members.inviteTitle') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.team.settings.members.inviteDescription') }}
            </p>
          </div>
          <div class="p-6">
            <div class="flex flex-col gap-3 sm:flex-row">
              <input
                v-model="inviteForm.name"
                type="text"
                class="input w-full sm:w-48"
                :placeholder="t('admin.team.settings.members.namePlaceholder')"
                maxlength="100"
              />
              <input
                v-model="inviteForm.email"
                type="email"
                class="input flex-1"
                :placeholder="t('admin.team.settings.members.emailPlaceholder')"
              />
              <select v-model="inviteForm.role" class="input w-full sm:w-40">
                <option value="member">{{ t('admin.team.settings.members.roles.member') }}</option>
                <option value="manager">{{ t('admin.team.settings.members.roles.manager') }}</option>
                <option value="admin">{{ t('admin.team.settings.members.roles.admin') }}</option>
              </select>
              <button type="button" class="btn btn-primary whitespace-nowrap" @click="inviteMember">
                <Icon name="plus" size="sm" class="mr-1.5" />
                {{ t('admin.team.settings.members.inviteButton') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Members List -->
        <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-600">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.team.settings.members.listTitle') }}
            </h2>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-left text-sm">
              <thead class="bg-gray-50 text-xs uppercase text-gray-500 dark:bg-dark-700 dark:text-dark-400">
                <tr>
                  <th class="px-6 py-3">{{ t('admin.team.settings.members.columns.name') }}</th>
                  <th class="px-6 py-3">{{ t('admin.team.settings.members.columns.email') }}</th>
                  <th class="px-6 py-3">{{ t('admin.team.settings.members.columns.role') }}</th>
                  <th class="px-6 py-3">{{ t('admin.team.settings.members.columns.status') }}</th>
                  <th class="px-6 py-3">{{ t('admin.team.settings.members.columns.joinedAt') }}</th>
                  <th class="px-6 py-3 text-right">{{ t('admin.team.settings.members.columns.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-dark-600">
                <tr
                  v-for="member in members"
                  :key="member.id"
                  class="transition-colors hover:bg-gray-50 dark:hover:bg-dark-700/50"
                >
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-3">
                      <div
                        class="flex h-9 w-9 items-center justify-center rounded-full bg-primary-100 text-sm font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300"
                      >
                        {{ getInitials(member.user?.name || member.user?.email || '?') }}
                      </div>
                      <div class="flex flex-col">
                        <span class="font-medium text-gray-900 dark:text-white">{{ member.user?.name || '-' }}</span>
                        <span class="text-xs text-gray-500 dark:text-dark-400">{{ member.user?.email || '-' }}</span>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-4 text-gray-600 dark:text-gray-400">{{ member.user?.email || '-' }}</td>
                  <td class="px-6 py-4">
                    <select
                      v-model="member.role"
                      class="rounded-md border border-gray-300 bg-white px-2.5 py-1.5 text-sm text-gray-700 focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500 dark:border-dark-500 dark:bg-dark-700 dark:text-gray-200"
                      @change="updateRole(member)"
                    >
                      <option value="member">{{ t('admin.team.settings.members.roles.member') }}</option>
                      <option value="manager">{{ t('admin.team.settings.members.roles.manager') }}</option>
                      <option value="admin">{{ t('admin.team.settings.members.roles.admin') }}</option>
                    </select>
                  </td>
                  <td class="px-6 py-4">
                    <div class="flex items-center gap-2">
                      <span
                        :class="[
                          'inline-block h-2 w-2 rounded-full',
                          member.status === 'active'
                            ? 'bg-green-500'
                            : member.status === 'pending'
                              ? 'bg-amber-500'
                              : 'bg-gray-400',
                        ]"
                      ></span>
                      <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
                        {{ t('admin.team.members.status.' + member.status) }}
                      </span>
                    </div>
                  </td>
                  <td class="px-6 py-4 text-gray-500 dark:text-dark-400">{{ formatDate(member.joined_at) }}</td>
                  <td class="px-6 py-4 text-right">
                    <button
                      class="rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
                      :title="t('common.remove')"
                      @click="removeMember(member)"
                    >
                      <Icon name="trash" size="sm" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Tab 3: Billing Settings -->
      <div v-show="activeTab === 'billing'" class="space-y-6">
        <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-600 dark:bg-dark-800">
          <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-600">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.team.settings.billing.title') }}
            </h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('admin.team.settings.billing.description') }}
            </p>
          </div>
          <div class="space-y-6 p-6">
            <!-- Payment Method -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.billing.paymentMethod') }}
              </label>
              <select v-model="billingForm.paymentMethod" class="input w-full">
                <option value="credit_card">{{ t('admin.team.settings.billing.methods.creditCard') }}</option>
                <option value="alipay">{{ t('admin.team.settings.billing.methods.alipay') }}</option>
                <option value="wechat_pay">{{ t('admin.team.settings.billing.methods.wechatPay') }}</option>
                <option value="bank_transfer">{{ t('admin.team.settings.billing.methods.bankTransfer') }}</option>
              </select>
            </div>

            <!-- Billing Address -->
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ t('admin.team.settings.billing.billingAddress') }}
              </label>
              <textarea
                v-model="billingForm.billingAddress"
                rows="3"
                class="input w-full"
                :placeholder="t('admin.team.settings.billing.addressPlaceholder')"
              />
            </div>

            <!-- Invoice Settings -->
            <div class="space-y-4">
              <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                {{ t('admin.team.settings.billing.invoiceSettings') }}
              </h3>
              <div class="flex items-center gap-3">
                <Toggle v-model="billingForm.autoInvoice" />
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  {{ t('admin.team.settings.billing.autoInvoice') }}
                </span>
              </div>
              <div class="flex items-center gap-3">
                <Toggle v-model="billingForm.invoiceEmailEnabled" />
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  {{ t('admin.team.settings.billing.invoiceEmail') }}
                </span>
              </div>
              <div v-if="billingForm.invoiceEmailEnabled">
                <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.team.settings.billing.invoiceEmailAddress') }}
                </label>
                <input
                  v-model="billingForm.invoiceEmail"
                  type="email"
                  class="input w-full"
                  :placeholder="t('admin.team.settings.billing.invoiceEmailPlaceholder')"
                />
              </div>
            </div>

            <!-- Save Button -->
            <div class="flex justify-end pt-2">
              <button type="button" class="btn btn-primary" @click="saveBilling">
                <Icon name="check" size="sm" class="mr-1.5" />
                {{ t('common.save') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab 4: Danger Zone -->
      <div v-show="activeTab === 'danger'" class="space-y-6">
        <div class="rounded-xl border border-red-200 bg-white dark:border-red-800 dark:bg-dark-800">
          <div class="border-b border-red-100 px-6 py-4 dark:border-red-800/50">
            <h2 class="text-lg font-semibold text-red-700 dark:text-red-400">
              {{ t('admin.team.settings.danger.title') }}
            </h2>
            <p class="mt-1 text-sm text-red-600/80 dark:text-red-400/70">
              {{ t('admin.team.settings.danger.description') }}
            </p>
          </div>
          <div class="space-y-6 p-6">
            <!-- Transfer Ownership -->
            <div class="flex items-center justify-between rounded-lg border border-gray-200 p-4 dark:border-dark-600">
              <div>
                <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                  {{ t('admin.team.settings.danger.transferOwnership') }}
                </h3>
                <p class="mt-0.5 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('admin.team.settings.danger.transferOwnershipDesc') }}
                </p>
              </div>
              <button type="button" class="btn btn-secondary" @click="showTransferDialog = true">
                <Icon name="arrowRight" size="sm" class="mr-1.5" />
                {{ t('admin.team.settings.danger.transferButton') }}
              </button>
            </div>

            <!-- Delete Team -->
            <div class="flex items-center justify-between rounded-lg border border-red-200 bg-red-50/50 p-4 dark:border-red-800 dark:bg-red-900/10">
              <div>
                <h3 class="text-sm font-medium text-red-700 dark:text-red-400">
                  {{ t('admin.team.settings.danger.deleteTeam') }}
                </h3>
                <p class="mt-0.5 text-sm text-red-600/80 dark:text-red-400/70">
                  {{ t('admin.team.settings.danger.deleteTeamDesc') }}
                </p>
              </div>
              <button type="button" class="btn bg-red-600 text-white hover:bg-red-700" @click="showDeleteDialog = true">
                <Icon name="trash" size="sm" class="mr-1.5" />
                {{ t('admin.team.settings.danger.deleteButton') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Transfer Ownership Dialog -->
    <ConfirmDialog
      :show="showTransferDialog"
      :title="t('admin.team.settings.danger.transferDialogTitle')"
      :message="t('admin.team.settings.danger.transferDialogMessage')"
      @confirm="confirmTransfer"
      @cancel="showTransferDialog = false"
    />

    <!-- Delete Team Dialog -->
    <ConfirmDialog
      :show="showDeleteDialog"
      :title="t('admin.team.settings.danger.deleteDialogTitle')"
      :message="t('admin.team.settings.danger.deleteDialogMessage')"
      danger
      @confirm="confirmDeleteTeam"
      @cancel="showDeleteDialog = false"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useTeamContext } from '@/composables/useTeamContext'
import { teamAPI } from '@/api/team'
import type { TeamMember } from '@/api/team'
import { formatDateTime } from '@/utils/format'
import Icon from '@/components/icons/Icon.vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Toggle from '@/components/common/Toggle.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

type IconName = 'cog' | 'users' | 'creditCard' | 'exclamationTriangle'

const { t } = useI18n()
const appStore = useAppStore()
const { teamId, fetchCurrentTeam } = useTeamContext()

// Tabs
const tabs = [
  { key: 'general', label: t('admin.team.settings.tabs.general'), icon: 'cog' as IconName },
]
const activeTab = ref('general')

// Timezones
const timezones = [
  { value: 'Asia/Shanghai', label: 'Asia/Shanghai (UTC+8)' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo (UTC+9)' },
  { value: 'Asia/Singapore', label: 'Asia/Singapore (UTC+8)' },
  { value: 'America/New_York', label: 'America/New_York (UTC-5/UTC-4)' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles (UTC-8/UTC-7)' },
  { value: 'Europe/London', label: 'Europe/London (UTC+0/UTC+1)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC+1/UTC+2)' },
  { value: 'Australia/Sydney', label: 'Australia/Sydney (UTC+10/UTC+11)' },
  { value: 'UTC', label: 'UTC' },
]

// General Settings Form
interface GeneralForm {
  name: string
  description: string
  timezone: string
  language: string
}

const generalForm = reactive<GeneralForm>({
  name: '',
  description: '',
  timezone: 'Asia/Shanghai',
  language: 'zh-CN',
})

const savingGeneral = ref(false)

const saveGeneral = async () => {
  if (!teamId.value) {
    appStore.showError(t('common.error'))
    return
  }
  savingGeneral.value = true
  try {
    await teamAPI.updateTeam(teamId.value, {
      name: generalForm.name,
      description: generalForm.description,
      timezone: generalForm.timezone,
      language: generalForm.language,
    })
    appStore.showSuccess(t('admin.team.settings.general.saveSuccess'))
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  } finally {
    savingGeneral.value = false
  }
}

// Members - TeamMember type imported from @/api/team
const members = ref<TeamMember[]>([])
const membersLoading = ref(false)

const inviteForm = reactive({
  email: '',
  name: '',
  role: 'member' as 'admin' | 'manager' | 'member',
})
const inviting = ref(false)

const getInitials = (name: string): string => {
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)
}

const formatDate = (date: string): string => {
  return formatDateTime(date)
}

const fetchMembers = async () => {
  if (!teamId.value) return
  membersLoading.value = true
  try {
    const response = await teamAPI.listMembers(teamId.value)
    members.value = response.items || []
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  } finally {
    membersLoading.value = false
  }
}

const inviteMember = async () => {
  if (!inviteForm.email) {
    appStore.showError(t('admin.team.settings.members.emailRequired'))
    return
  }
  if (!teamId.value) {
    appStore.showError(t('common.error'))
    return
  }
  inviting.value = true
  try {
    await teamAPI.inviteMember(teamId.value, {
      email: inviteForm.email,
      role: inviteForm.role,
      display_name: inviteForm.name.trim(),
    })
    appStore.showSuccess(t('admin.team.settings.members.inviteSuccess'))
    inviteForm.email = ''
    inviteForm.name = ''
    inviteForm.role = 'member'
    await fetchMembers()
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  } finally {
    inviting.value = false
  }
}

const updateRole = async (member: TeamMember) => {
  if (!teamId.value) {
    appStore.showError(t('common.error'))
    return
  }
  const previousRole = member.role
  try {
    await teamAPI.updateMemberRole(teamId.value, member.id, member.role)
    appStore.showSuccess(
      t('admin.team.settings.members.roleUpdated', { name: member.user?.name || '' }),
    )
  } catch (err: any) {
    // Restore the previous role in the UI so the select reflects the real state
    member.role = previousRole
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  }
}

const removeMember = async (member: TeamMember) => {
  if (!teamId.value) {
    appStore.showError(t('common.error'))
    return
  }
  try {
    await teamAPI.removeMember(teamId.value, member.id)
    appStore.showSuccess(
      t('admin.team.settings.members.memberRemoved', { name: member.user?.name || '' }),
    )
    await fetchMembers()
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  }
}

// Billing Form
interface BillingForm {
  paymentMethod: string
  billingAddress: string
  autoInvoice: boolean
  invoiceEmailEnabled: boolean
  invoiceEmail: string
}

const billingForm = reactive<BillingForm>({
  paymentMethod: 'credit_card',
  billingAddress: '',
  autoInvoice: true,
  invoiceEmailEnabled: true,
  invoiceEmail: '',
})

const saveBilling = () => {
  // 后端暂未提供账单 API
  appStore.showInfo(t('common.notAvailable'))
}

// Danger Zone
const showTransferDialog = ref(false)
const showDeleteDialog = ref(false)

const confirmTransfer = () => {
  // 后端暂未提供转移所有权 API
  appStore.showInfo(t('common.notAvailable'))
  showTransferDialog.value = false
}

const confirmDeleteTeam = () => {
  // 后端暂未提供删除团队 API（避免误操作）
  appStore.showInfo(t('common.notAvailable'))
  showDeleteDialog.value = false
}

onMounted(async () => {
  await fetchCurrentTeam()
  if (!teamId.value) return
  // 加载团队信息填充通用设置表单
  try {
    const team = await teamAPI.getTeam(teamId.value)
    generalForm.name = team.name
    generalForm.description = team.description ?? ''
    generalForm.timezone = team.timezone || 'Asia/Shanghai'
    generalForm.language = team.language || 'zh-CN'
  } catch (err: any) {
    const msg = err?.response?.data?.message || err?.message || t('common.error')
    appStore.showError(msg)
  }
  // 加载成员列表
  await fetchMembers()
})
</script>
