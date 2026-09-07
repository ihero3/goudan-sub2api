<template>
  <div class="w-full">
    <template v-if="!verified">
      <button
        type="button"
        class="btn btn-secondary w-full"
        :disabled="loading || acquiring"
        @click="startChallenge"
      >
        {{ loading ? t('auth.clickCaptcha.verifying') : captchaToken ? t('auth.clickCaptcha.restart') : t('auth.clickCaptcha.start') }}
      </button>

      <div v-if="challenge" class="mt-4">
        <p class="mb-2 text-center text-sm text-gray-600 dark:text-dark-300">
          {{ t('auth.clickCaptcha.promptPrefix') }}<span class="font-medium text-primary-600">{{ challenge.prompt.join(promptSeparator) }}</span>
        </p>
        <div class="relative h-56 w-full overflow-hidden rounded-xl border border-dashed border-gray-300 bg-gray-50 dark:border-dark-600 dark:bg-dark-800/40">
          <button
            v-for="cell in challenge.grid"
            :key="cell.cell_id"
            type="button"
            class="absolute flex h-12 w-12 items-center justify-center rounded-lg text-3xl leading-none transition-transform"
            :style="cellStyle(cell.cell_id)"
            :class="[
              selected.has(cell.cell_id) ? 'bg-primary-100/80 ring-2 ring-primary-500 dark:bg-primary-900/40' : 'hover:scale-110',
              error ? 'ring-1 ring-red-400' : '',
            ]"
            :disabled="submitting || selected.has(cell.cell_id)"
            @click="handleClick(cell.cell_id)"
          >
            {{ cell.content }}
          </button>
          <p v-if="error" class="absolute inset-x-0 bottom-1 text-center text-sm text-red-500">{{ error }}</p>
        </div>
      </div>
    </template>
    <div v-else class="rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700 dark:bg-green-900/20 dark:text-green-400">
      ✓ {{ t('auth.clickCaptcha.passed') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { createClickCaptchaChallenge, verifyClickCaptcha } from '@/api/registrationRisk'
import type { ClickCaptchaChallenge } from '@/types'

const emit = defineEmits<{ (e: 'verified', token: string): void; (e: 'expired'): void }>()

const { t } = useI18n()

const promptSeparator = computed(() => t('auth.clickCaptcha.promptSeparator'))

interface CellPosition {
  left: string
  top: string
  rotate: number
  scale: number
}

const positions = ref<Record<string, CellPosition>>({})

// 在 5x5 的松格子里随机取位并在格内抖动，生成"散乱画布"布局。
// 固定网格 + 文字极易被 OCR / 文本识别机器人破解；乱布局 + emoji 提升图像识别难度。
function buildLayout(count: number): CellPosition[] {
  const cols = 5
  const rows = 5
  const pool = Array.from({ length: cols * rows }, (_, i) => i)
  for (let i = pool.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[pool[i], pool[j]] = [pool[j], pool[i]]
  }
  const chosen = pool.slice(0, count)
  const cellW = 100 / cols
  const cellH = 100 / rows
  return chosen.map((gi) => {
    const r = Math.floor(gi / cols)
    const c = gi % cols
    const jitterX = (Math.random() - 0.5) * cellW * 0.7
    const jitterY = (Math.random() - 0.5) * cellH * 0.7
    return {
      left: `${(c * cellW + cellW / 2 + jitterX).toFixed(1)}%`,
      top: `${(r * cellH + cellH / 2 + jitterY).toFixed(1)}%`,
      rotate: Math.floor(Math.random() * 40 - 20),
      scale: +(0.9 + Math.random() * 0.25).toFixed(2),
    }
  })
}

function cellStyle(cellId: string) {
  const pos = positions.value[cellId]
  if (!pos) return {}
  return {
    left: pos.left,
    top: pos.top,
    transform: `translate(-50%, -50%) rotate(${pos.rotate}deg) scale(${pos.scale})`,
  }
}

const loading = ref(false)
const acquiring = ref(false)
const submitting = ref(false)
const challenge = ref<ClickCaptchaChallenge | null>(null)
const captchaToken = ref('')
const error = ref('')
const selected = ref(new Set<string>())
const order = ref<string[]>([])
const verified = ref(false)

async function startChallenge() {
  if (acquiring.value) return
  acquiring.value = true
  error.value = ''
  loading.value = true
  try {
    // 重新验证时需要旧 token 失效（后端 token 一次性，重复使用会失败）
    captchaToken.value = ''
    challenge.value = await createClickCaptchaChallenge()
    const layout = buildLayout(challenge.value.grid.length)
    const posMap: Record<string, CellPosition> = {}
    challenge.value.grid.forEach((cell, i) => {
      posMap[cell.cell_id] = layout[i]
    })
    positions.value = posMap
    selected.value = new Set()
    order.value = []
    verified.value = false
    loading.value = false
  } catch {
    error.value = t('auth.clickCaptcha.initFailed')
    loading.value = false
  } finally {
    acquiring.value = false
  }
}

function handleClick(cellId: string) {
  if (submitting.value || !challenge.value || selected.value.has(cellId)) return
  selected.value.add(cellId)
  order.value.push(cellId)
  if (order.value.length >= challenge.value.prompt.length) {
    submit()
  }
}

async function submit() {
  if (!challenge.value || submitting.value) return
  submitting.value = true
  error.value = ''
  try {
    const result = await verifyClickCaptcha(challenge.value.challenge_id, order.value)
    captchaToken.value = result.captcha_token
    verified.value = true
    emit('verified', captchaToken.value)
  } catch {
    error.value = t('auth.clickCaptcha.wrongOrder')
    await startChallenge()
  } finally {
    submitting.value = false
  }
}
</script>
