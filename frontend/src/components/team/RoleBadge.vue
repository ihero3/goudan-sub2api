<template>
  <span
    :class="[
      'inline-flex items-center gap-1.5 font-medium',
      sizeClasses,
      colorClasses
    ]"
  >
    <Icon :name="iconName" :size="iconSize" />
    <span>{{ displayRole }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

type Role = 'owner' | 'admin' | 'member' | 'viewer'
type Size = 'sm' | 'md' | 'lg'

interface Props {
  role: Role
  size?: Size
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md'
})

const displayRole = computed(() => t(`admin.team.members.roles.${props.role}`))

const iconName = computed(() => {
      switch (props.role) {
        case 'owner':
          return 'user' as const
        case 'admin':
          return 'user' as const
        case 'member':
          return 'user' as const
        case 'viewer':
          return 'eye' as const
        default:
          return 'user' as const
      }
    })

const iconSize = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'xs'
    case 'lg':
      return 'md'
    default:
      return 'sm'
  }
})

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'rounded px-2 py-0.5 text-xs'
    case 'lg':
      return 'rounded-lg px-3 py-1.5 text-sm'
    default:
      return 'rounded-md px-2.5 py-1 text-xs'
  }
})

const colorClasses = computed(() => {
  switch (props.role) {
    case 'owner':
      return 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300'
    case 'admin':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
    case 'member':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
    case 'viewer':
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
  }
})
</script>
