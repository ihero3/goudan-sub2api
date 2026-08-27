/**
 * 深合并两个普通对象（用于 locale 字典合并）。
 *
 * 规则：
 * - 两侧同为普通对象时递归合并；
 * - 其余情况 override 优先（override 为 undefined 时保留 base）；
 * - 数组视为原子值，整体替换。
 */
type PlainObject = Record<string, unknown>

function isPlainObject(value: unknown): value is PlainObject {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

export function deepMerge<T>(base: T, override: unknown): T {
  if (isPlainObject(base) && isPlainObject(override)) {
    const out: PlainObject = { ...base }
    for (const key of Object.keys(override)) {
      out[key] = deepMerge((base as PlainObject)[key], override[key])
    }
    return out as T
  }
  return (override === undefined ? base : override) as T
}
