/**
 * zh 字典 shim：上游模块化字典 + 本 fork 差异，深合并。
 *
 * 背景：本 fork 曾在此文件维护完整的扁平字典，导致 `import('./locales/zh')`
 * 优先解析到本文件、遮蔽了上游的 locales/zh/ 模块化目录，上游新增的翻译
 * key 在运行时全部丢失（表现为大量 [intlify] Not found 警告）。
 *
 * 现在的维护方式：
 * - 上游文案改动 → locales/zh/ 下的模块化文件；
 * - fork 专属文案/覆盖 → locales/zh-fork.ts（其值优先）；
 * - 运行时最终字典 = deepMerge(locales/zh/, locales/zh-fork.ts)。
 */
import { deepMerge } from '../deepMerge'

import modular from './zh/index'
import fork from './zh-fork'

export default deepMerge(modular, fork)
