<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900">{{ t('admin.models.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500">{{ t('admin.models.description') }}</p>
        </div>
      </div>

      <div class="card">
        <div class="p-6">
          <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="model in models"
              :key="model.id"
              class="group rounded-xl border border-gray-200 bg-white p-4 shadow-sm transition-all hover:border-primary-200 hover:shadow-md"
            >
              <template v-if="model.category === 'hint'">
                <div class="flex flex-col items-center justify-center gap-2 py-5">
                  <div :class="['flex h-10 w-10 shrink-0 items-center justify-center rounded-lg', getProviderStyle(model.vendor).gradient]">
                    <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" :d="getProviderStyle(model.vendor).icon" />
                    </svg>
                  </div>
                  <p class="text-sm text-gray-600 text-left leading-relaxed max-w-[180px]">{{ t('admin.models.hint') }}</p>
                </div>
              </template>
              <template v-else>
                <div class="flex items-start gap-3">
                  <div :class="['flex h-10 w-10 shrink-0 items-center justify-center rounded-lg', getProviderStyle(model.vendor).gradient]">
                    <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" :d="getProviderStyle(model.vendor).icon" />
                    </svg>
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center gap-2">
                      <h3 class="truncate font-semibold text-gray-900">{{ model.name }}</h3>
                      <button
                        @click="copyModelName(model.name)"
                        class="p-1 hover:bg-gray-100 rounded transition-colors"
                        :title="t('admin.models.copy')"
                      >
                        <svg v-if="copiedModel !== model.name" class="h-4 w-4 text-gray-400 hover:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                        <svg v-else class="h-4 w-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                        </svg>
                      </button>
                    </div>
                    <p class="mt-1 text-sm text-gray-500">{{ getProviderDescription(model.provider) }}</p>
                  </div>
                </div>
                <div class="mt-4 flex items-center gap-2 text-xs text-gray-400">
                  <span class="rounded-full bg-gray-100 px-2 py-1">{{ getCategoryLabel(model.category) }}</span>
                  <span class="rounded-full bg-green-100 px-2 py-1 text-green-600">{{ t('admin.models.status.available') }}</span>
                </div>
                <div v-if="model.category !== 'hint'" class="mt-3 flex flex-wrap gap-3 text-xs">
                  <template v-if="model.category === 'image'">
                    <span class="text-gray-500">{{ t('admin.models.pricing.approx') }}:</span>
                    <span class="font-medium text-gray-700">0.3$ per image</span>
                  </template>
                  <template v-else-if="model.category === 'multimodal'">
                    <span class="text-gray-500">{{ t('admin.models.pricing.approx') }}:</span>
                    <span class="font-medium text-gray-700">0.9$ per second</span>
                  </template>
                  <template v-else>
                    <!-- Dual pricing: 1.5折 and 9折 -->
                    <template v-if="getDualPricing(model.name)">
                      <div class="w-full space-y-1.5">
                        <div class="flex items-center justify-between">
                          <span class="rounded bg-red-100 px-1.5 py-0.5 text-[10px] font-medium text-red-600">1.5折</span>
                          <span class="text-gray-600 text-xs">
                            <span class="text-gray-500">{{ t('admin.models.pricing.input') }}:</span>
                            <span class="font-medium">{{ formatPrice(getDualPricing(model.name)!.tier15.input) }}</span>
                            <span class="mx-1 text-gray-400"> </span>
                            <span class="text-gray-500">{{ t('admin.models.pricing.output') }}:</span>
                            <span class="font-medium">{{ formatPrice(getDualPricing(model.name)!.tier15.output) }}</span>
                          </span>
                        </div>
                        <div class="flex items-center justify-between">
                          <span class="rounded bg-blue-100 px-1.5 py-0.5 text-[10px] font-medium text-blue-600">9折</span>
                          <span class="text-gray-600 text-xs">
                            <span class="text-gray-500">{{ t('admin.models.pricing.input') }}:</span>
                            <span class="font-medium">{{ formatPrice(getDualPricing(model.name)!.tier90.input) }}</span>
                            <span class="mx-1 text-gray-400"> </span>
                            <span class="text-gray-500">{{ t('admin.models.pricing.output') }}:</span>
                            <span class="font-medium">{{ formatPrice(getDualPricing(model.name)!.tier90.output) }}</span>
                          </span>
                        </div>
                      </div>
                    </template>
                    <!-- Standard pricing -->
                    <template v-else>
                      <div class="flex items-center gap-1">
                        <span class="text-gray-500">{{ t('admin.models.pricing.input') }}:</span>
                        <span class="font-medium text-gray-700">{{ formatPrice(modelPricing[model.name]?.input_price ?? fallbackPriceFromRate(modelUsdTokenRates[model.name])) }}</span>
                      </div>
                      <div class="flex items-center gap-1">
                        <span class="text-gray-500">{{ t('admin.models.pricing.output') }}:</span>
                        <span class="font-medium text-gray-700">{{ formatPrice(modelPricing[model.name]?.output_price ?? fallbackPriceFromRate(modelUsdTokenRates[model.name])) }}</span>
                      </div>
                      <div class="flex items-center gap-1">
                        <span class="text-gray-500">{{ t('admin.models.pricing.approx') }}:</span>
                        <span class="font-medium text-gray-700">1$≈{{ modelUsdTokenRates[model.name] || '-' }}Tokens</span>
                      </div>
                    </template>
                  </template>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import { perTokenToMTok } from '@/components/admin/channel/types'
import { useAuthStore } from '@/stores/auth'
import { apiClient } from '@/api/client'

const { t, locale } = useI18n()
const authStore = useAuthStore()

const copiedModel = ref<string | null>(null)

interface Model {
  id: string
  name: string
  provider: string
  vendor: string
  category: string
  icon: string
}

interface ModelPricing {
  input_price?: number | null
  output_price?: number | null
  cache_write_price?: number | null
  cache_read_price?: number | null
}

const modelPricing = ref<Record<string, ModelPricing>>({})

// Claude/OpenAI 新模型仅在非 threerouter 域名时显示
const isThreerouterDomain = computed(() => {
  return typeof window !== 'undefined' && window.location.hostname.includes('threerouter')
})

const fetchModelPricing = async (modelName: string) => {
  try {
    // Use public endpoint when not authenticated, admin endpoint when authenticated
    const endpoint = authStore.isAuthenticated ? '/admin/channels/model-pricing' : '/models/pricing'
    const { data: result } = await apiClient.get(endpoint, { params: { model: modelName } })
    if (result.found) {
      modelPricing.value[modelName] = {
        input_price: perTokenToMTok(result.input_price),
        output_price: perTokenToMTok(result.output_price),
        cache_write_price: perTokenToMTok(result.cache_write_price),
        cache_read_price: perTokenToMTok(result.cache_read_price),
      }
    }
  } catch (error) {
    // Silently ignore pricing fetch errors
  }
}

onMounted(() => {
  models.value.forEach(model => {
    if (model.name && model.category !== 'hint') {
      fetchModelPricing(model.name)
    }
  })
})

const formatPrice = (price: number | null | undefined): string => {
  if (price === null || price === undefined) return '-'
  return `$${price.toFixed(2)}/MTokens`
}

const fallbackPriceFromRate = (rate: string | undefined): number | undefined => {
  if (!rate || rate === '-') return undefined
  const match = rate.match(/^([\d.]+)M$/)
  if (!match) return undefined
  const tokensPerDollar = parseFloat(match[1])
  if (tokensPerDollar <= 0) return undefined
  return 1 / tokensPerDollar
}

const modelUsdTokenRates: Record<string, string> = {
  'deepseek-v4-pro': '29.49M',
  'kimi-k3': '6.18M',
  'minimax-m3': '3.40M',
  'qwen3.8-max': '3.06M',
  'glm-5.3': '3.35M',
  'seedance-2.0': '-',
  'gpt-image-2': '-',
}

const providerDescriptions: Record<string, { en: string; zh: string }> = {
  'deepseek-v4-pro': {
    en: 'DeepSeek V4 is a cutting-edge MoE-based flagship model, excelling in coding, reasoning, and long-context tasks with robust tool-use capabilities for complex workflows.',
    zh: 'DeepSeek V4 是基于 MoE 架构的前沿旗舰模型，在编码、推理和长上下文任务中表现出色，具备强大的工具调用能力。'
  },
  'minimax-m3': {
    en: 'MiniMax‑M3 is a frontier open‑weight model with 1M context, native multimodality, and top coding/agent abilities, built on the MSA sparse attention architecture.',
    zh: 'MiniMax-M3 是前沿开源权重模型，拥有 100 万上下文、原生多模态和顶级编码/智能体能力，基于 MSA 稀疏注意力架构。'
  },
  'kimi-k3': {
    en: 'Kimi-K3 is Moonshot\'s open MoE flagship with 256K context, excelling in long-horizon coding and agent swarm (300 sub-agents) for complex, multi-step tasks.',
    zh: 'Kimi-K3 是月之暗面的开源 MoE 旗舰模型，拥有 256K 上下文，擅长长期编码和智能体集群（300 个子智能体）处理复杂多步骤任务。'
  },
  'qwen3.8-max': {
    en: 'Qwen3.8-Max is Alibaba\'s agent‑centric flagship with 1M context, top-tier coding, excelling in complex workflows and multi-framework generalization.',
    zh: 'Qwen3.8-Max 是阿里的智能体旗舰模型，拥有 100 万上下文、顶级编码能力和擅长复杂工作流和多框架泛化。'
  },
  'glm-5.3': {
    en: 'GLM-5.3 is Zhipu AI\'s open MoE flagship with 200K context, excelling in 8‑hour autonomous agentic coding and topping SWE‑Bench Pro for complex software engineering tasks.',
    zh: 'GLM-5.3 是智谱 AI 的开源 MoE 旗舰模型，拥有 200K 上下文，擅长 8 小时自主智能体编码，在 SWE-Bench Pro 复杂软件工程任务中排名第一。'
  },
  'seedance-2.0': {
    en: 'Contact support via ticket after recharge. Premium video models require dedicated service.',
    zh: '充值后通过工单联系使用，好的视频模型就要专人服务。'
  },
  'gpt-image-2': {
    en: 'GPT-Image-2 (ChatGPT Images 2.0), launched by OpenAI in April 2026, is a flagship image model with reasoning, accurate Chinese rendering, high-res output and batch generation.',
    zh: 'GPT-Image-2（ChatGPT 图像 2.0）是 OpenAI 于 2026 年 4 月发布的旗舰图像模型，具备推理能力、精准的中文渲染、高分辨率输出和批量生成功能。'
  },
  'claude-haiku-4-5-20251001': {
    en: 'Claude Haiku 4.5 is Anthropic\'s fast and affordable model, ideal for high-volume tasks with near-instant response times and strong tool-use capabilities.',
    zh: 'Claude Haiku 4.5 是 Anthropic 的快速高性价比模型，适合高频率任务，响应接近即时，具备强大的工具调用能力。'
  },
  'claude-opus-4-5-20251101': {
    en: 'Claude Opus 4.5 is Anthropic\'s flagship model for complex reasoning, deep analysis, and long-context understanding, excelling in coding and multi-step agentic workflows.',
    zh: 'Claude Opus 4.5 是 Anthropic 的旗舰模型，擅长复杂推理、深度分析和长上下文理解，在编码和多步骤智能体工作流中表现出色。'
  },
  'claude-opus-4-6': {
    en: 'Claude Opus 4.6 builds on 4.5 with improved reasoning, faster inference, and enhanced code generation for enterprise-grade agentic tasks.',
    zh: 'Claude Opus 4.6 在 4.5 基础上提升推理能力、推理速度和代码生成能力，适用于企业级智能体任务。'
  },
  'claude-opus-4-7': {
    en: 'Claude Opus 4.7 further refines reasoning depth and tool-use accuracy, setting new benchmarks in coding and complex problem-solving.',
    zh: 'Claude Opus 4.7 进一步提升推理深度和工具调用精度，在编码和复杂问题解决方面树立新标杆。'
  },
  'claude-opus-4-8': {
    en: 'Claude Opus 4.8 is the latest Opus iteration with state-of-the-art reasoning, expanded context handling, and superior agentic coding performance.',
    zh: 'Claude Opus 4.8 是最新的 Opus 迭代版本，具备最先进的推理能力、扩展的上下文处理和卓越的智能体编码性能。'
  },
  'claude-opus-5': {
    en: 'Claude Opus 5 is Anthropic\'s next-generation flagship, redefining the frontier of reasoning, creativity, and autonomous task completion.',
    zh: 'Claude Opus 5 是 Anthropic 的下一代旗舰模型，重新定义推理、创造力和自主任务完成的前沿。'
  },
  'claude-sonnet-4-6': {
    en: 'Claude Sonnet 4.6 balances performance and cost with strong coding, analysis, and vision capabilities, ideal for production-scale applications.',
    zh: 'Claude Sonnet 4.6 在性能和成本之间取得平衡，具备强大的编码、分析和视觉能力，适合生产级应用。'
  },
  'claude-sonnet-5': {
    en: 'Claude Sonnet 5 delivers flagship-level reasoning at mid-tier pricing, with enhanced multi-modal understanding and agentic workflow support.',
    zh: 'Claude Sonnet 5 以中端价格提供旗舰级推理能力，增强多模态理解和智能体工作流支持。'
  },
  'claude-fable-5': {
    en: 'Claude Fable 5 is Anthropic\'s premium creative model, specialized in long-form writing, storytelling, and complex creative tasks with exceptional quality.',
    zh: 'Claude Fable 5 是 Anthropic 的旗舰创意模型，专注于长篇写作、叙事和复杂创意任务，质量卓越。'
  },
  'gpt-5.4': {
    en: 'GPT-5.4 is OpenAI\'s production workhorse with 272K context, strong reasoning and function calling, ideal for general-purpose AI applications.',
    zh: 'GPT-5.4 是 OpenAI 的生产级模型，拥有 272K 上下文，推理和函数调用能力强，适合通用 AI 应用。'
  },
  'gpt-5.5': {
    en: 'GPT-5.5 enhances 5.4 with improved multi-modal understanding, faster inference, and better tool orchestration for complex workflows.',
    zh: 'GPT-5.5 在 5.4 基础上增强多模态理解、推理速度和工具编排能力，适用于复杂工作流。'
  },
  'gpt-5.6-luna': {
    en: 'GPT-5.6 Luna is OpenAI\'s fastest and most affordable model, recently price-cut by 80%, perfect for high-volume, high-frequency tasks.',
    zh: 'GPT-5.6 Luna 是 OpenAI 最快、最具性价比的模型，近期降价 80%，适合大规模、高频率任务。'
  },
  'gpt-5.6-terra': {
    en: 'GPT-5.6 Terra is the balanced mid-tier model with strong reasoning at reduced cost, ideal for everyday enterprise workloads.',
    zh: 'GPT-5.6 Terra 是均衡型中端模型，推理强且成本优化，适合日常企业工作负载。'
  },
  'gpt-5.6-sol': {
    en: 'GPT-5.6 Sol is OpenAI\'s flagship model with the deepest reasoning and best quality, featuring a new Fast mode with 2.5x speed improvement.',
    zh: 'GPT-5.6 Sol 是 OpenAI 的旗舰模型，推理最深、质量最佳，新增 Fast 模式速度提升 2.5 倍。'
  }
}

// Dual pricing: 1.5折 (15%) and 9折 (90%) of official price
// Official prices sourced from https://api.huanxing.ai/pricing
interface DualPricing {
  officialInput: number  // USD per MTokens
  officialOutput: number // USD per MTokens
}
const dualPricingModels: Record<string, DualPricing> = {
  // Claude models (CNY → USD at ~7 CNY/USD)
  'claude-haiku-4-5-20251001': { officialInput: 1.00, officialOutput: 5.00 },
  'claude-opus-4-5-20251101': { officialInput: 5.00, officialOutput: 25.00 },
  'claude-opus-4-6': { officialInput: 5.00, officialOutput: 25.00 },
  'claude-opus-4-7': { officialInput: 5.00, officialOutput: 25.00 },
  'claude-opus-4-8': { officialInput: 5.00, officialOutput: 25.00 },
  'claude-opus-5': { officialInput: 5.00, officialOutput: 25.00 },
  'claude-sonnet-4-6': { officialInput: 3.00, officialOutput: 15.00 },
  'claude-sonnet-5': { officialInput: 2.00, officialOutput: 10.00 },
  'claude-fable-5': { officialInput: 10.00, officialOutput: 50.00 },
  // OpenAI GPT >5.4 models (official USD pricing)
  'gpt-5.4': { officialInput: 2.50, officialOutput: 10.00 },
  'gpt-5.5': { officialInput: 2.50, officialOutput: 15.00 },
  'gpt-5.6-luna': { officialInput: 1.00, officialOutput: 6.00 },
  'gpt-5.6-terra': { officialInput: 2.50, officialOutput: 15.00 },
  'gpt-5.6-sol': { officialInput: 5.00, officialOutput: 30.00 },
}

const getDualPricing = (modelName: string): { tier15: { input: number; output: number }; tier90: { input: number; output: number } } | null => {
  const pricing = dualPricingModels[modelName]
  if (!pricing) return null
  return {
    tier15: { input: pricing.officialInput * 0.15, output: pricing.officialOutput * 0.15 },
    tier90: { input: pricing.officialInput * 0.90, output: pricing.officialOutput * 0.90 },
  }
}

const models = computed<Model[]>(() => {
  const base: Model[] = [
    { id: '1', name: 'deepseek-v4-pro', provider: 'deepseek-v4-pro', vendor: 'deepseek', category: 'text', icon: '' },
    { id: '2', name: 'minimax-m3', provider: 'minimax-m3', vendor: 'minimax', category: 'text', icon: '' },
    { id: '3', name: 'kimi-k3', provider: 'kimi-k3', vendor: 'moonshot', category: 'text', icon: '' },
    { id: '4', name: 'qwen3.8-max', provider: 'qwen3.8-max', vendor: 'alibaba', category: 'text', icon: '' },
    { id: '5', name: 'glm-5.3', provider: 'glm-5.3', vendor: 'zhipu', category: 'text', icon: '' },
    { id: '6', name: 'seedance-2.0', provider: 'seedance-2.0', vendor: 'bytedance', category: 'multimodal', icon: '' },
    { id: '8', name: 'gpt-image-2', provider: 'gpt-image-2', vendor: 'openai', category: 'image', icon: '' },
  ]

  // Claude >4.5 和 OpenAI GPT >5.4 仅在非 threerouter 域名时显示
  if (!isThreerouterDomain.value) {
    base.push(
      // Claude >4.5 (高版本在前)
      { id: '18', name: 'claude-fable-5', provider: 'claude-fable-5', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '15', name: 'claude-opus-5', provider: 'claude-opus-5', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '17', name: 'claude-sonnet-5', provider: 'claude-sonnet-5', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '14', name: 'claude-opus-4-8', provider: 'claude-opus-4-8', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '13', name: 'claude-opus-4-7', provider: 'claude-opus-4-7', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '12', name: 'claude-opus-4-6', provider: 'claude-opus-4-6', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '16', name: 'claude-sonnet-4-6', provider: 'claude-sonnet-4-6', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '11', name: 'claude-opus-4-5-20251101', provider: 'claude-opus-4-5-20251101', vendor: 'anthropic', category: 'text', icon: '' },
      { id: '10', name: 'claude-haiku-4-5-20251001', provider: 'claude-haiku-4-5-20251001', vendor: 'anthropic', category: 'text', icon: '' },
      // OpenAI GPT >5.4 (高版本在前)
      { id: '23', name: 'gpt-5.6-sol', provider: 'gpt-5.6-sol', vendor: 'openai', category: 'text', icon: '' },
      { id: '22', name: 'gpt-5.6-terra', provider: 'gpt-5.6-terra', vendor: 'openai', category: 'text', icon: '' },
      { id: '21', name: 'gpt-5.6-luna', provider: 'gpt-5.6-luna', vendor: 'openai', category: 'text', icon: '' },
      { id: '20', name: 'gpt-5.5', provider: 'gpt-5.5', vendor: 'openai', category: 'text', icon: '' },
      { id: '19', name: 'gpt-5.4', provider: 'gpt-5.4', vendor: 'openai', category: 'text', icon: '' },
    )
  }

  base.push({ id: '9', name: '', provider: '', vendor: 'hint', category: 'hint', icon: '' })
  return base
})

const getProviderDescription = (providerKey: string) => {
  const desc = providerDescriptions[providerKey]
  if (!desc) return ''
  return locale.value === 'zh' ? desc.zh : desc.en
}

const providerStyles: Record<string, { gradient: string; icon: string }> = {
  bytedance: {
    gradient: 'bg-gradient-to-br from-rose-500 to-orange-500',
    icon: 'M8 5v14l11-7z'
  },
  zhipu: {
    gradient: 'bg-gradient-to-br from-blue-500 to-indigo-500',
    icon: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z'
  },
  deepseek: {
    gradient: 'bg-gradient-to-br from-cyan-500 to-teal-500',
    icon: 'M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0zM10 7v3m0 0v3m0-3h3m-3 0H7'
  },
  moonshot: {
    gradient: 'bg-gradient-to-br from-indigo-500 to-purple-600',
    icon: 'M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z'
  },
  openai: {
    gradient: 'bg-gradient-to-br from-emerald-500 to-green-500',
    icon: 'M13 10V3L4 14h7v7l9-11h-7z'
  },
  anthropic: {
    gradient: 'bg-gradient-to-br from-orange-500 to-amber-600',
    icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z'
  },
  minimax: {
    gradient: 'bg-gradient-to-br from-amber-500 to-yellow-500',
    icon: 'M13 10V3L4 14h7v7l9-11h-7z'
  },
  alibaba: {
    gradient: 'bg-gradient-to-br from-orange-500 to-red-500',
    icon: 'M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
  },
  default: {
    gradient: 'bg-gradient-to-br from-purple-500 to-blue-500',
    icon: 'M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5'
  },
  hint: {
    gradient: 'bg-gradient-to-br from-blue-400 to-indigo-500',
    icon: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z'
  }
}

const getProviderStyle = (vendor: string) => {
  return providerStyles[vendor] || providerStyles.default
}

const categoryLabels: Record<string, string> = {
  text: 'admin.models.categories.text',
  image: 'admin.models.categories.image',
  audio: 'admin.models.categories.audio',
  multimodal: 'admin.models.categories.multimodal'
}

const getCategoryLabel = (category: string) => {
  return t(categoryLabels[category] || category)
}

const copyModelName = async (name: string) => {
  try {
    await navigator.clipboard.writeText(name)
    copiedModel.value = name
    setTimeout(() => {
      copiedModel.value = null
    }, 2000)
  } catch (err) {
    console.error('Failed to copy:', err)
  }
}
</script>