/**
 * Admin Media Tasks API endpoints
 * 管理后台媒体任务管理（图片 / 视频 / 音频，对应后端 /admin/media-tasks）
 */

import { apiClient } from '../client'

/**
 * 媒体任务状态。
 */
export type MediaTaskStatus = 'processing' | 'succeeded' | 'failed' | 'cancelled'

/**
 * 媒体任务记录。
 */
export interface MediaTask {
  id: number
  local_id: string
  media_kind: string
  user_id: number
  api_key_id: number
  public_model: string
  upstream_model: string
  account_id: number
  upstream_task_id: string
  status: MediaTaskStatus
  resolution: string
  duration_sec: number
  media_url: string
  thumbnail_url: string
  error_message: string
  cost_usd: number
  created_at: string
  updated_at: string
  finished_at: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface ListMediaTasksParams {
  user_id?: number
  page?: number
  page_size?: number
  status?: MediaTaskStatus
  media_kind?: string
  signal?: AbortSignal
}

/**
 * 列出媒体任务（支持 user_id / status / media_kind 过滤，SQL 层下推）。
 */
export async function list(params: ListMediaTasksParams): Promise<PaginatedResponse<MediaTask>> {
  const queryParams: Record<string, any> = {
    page: params.page ?? 1,
    page_size: params.page_size ?? 20
  }
  if (params.user_id) queryParams.user_id = params.user_id
  if (params.status) queryParams.status = params.status
  if (params.media_kind) queryParams.media_kind = params.media_kind
  const { data } = await apiClient.get<PaginatedResponse<MediaTask>>('/admin/media-tasks', {
    params: queryParams,
    signal: params.signal
  })
  return data
}

/**
 * 按 ID 或 local_id 获取任务详情。
 */
export async function get(idOrLocalID: string | number): Promise<MediaTask> {
  const { data } = await apiClient.get<MediaTask>(`/admin/media-tasks/${idOrLocalID}`)
  return data
}

/**
 * 取消任务。
 */
export async function cancel(id: number): Promise<{ id: number; status: string }> {
  const { data } = await apiClient.post<{ id: number; status: string }>(`/admin/media-tasks/${id}/cancel`)
  return data
}

export const mediaTasksAPI = {
  list,
  get,
  cancel
}

export default mediaTasksAPI
