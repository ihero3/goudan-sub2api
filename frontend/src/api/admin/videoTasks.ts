/**
 * Admin Video Tasks API endpoints
 * 管理后台视频任务管理（对应后端 /admin/video-tasks）
 */

import { apiClient } from '../client'

/**
 * 视频任务状态。
 * - processing: 等待上游结果
 * - succeeded: 成功（包含 video_url）
 * - failed: 失败
 * - cancelled: 已取消
 */
export type VideoTaskStatus = 'processing' | 'succeeded' | 'failed' | 'cancelled'

/**
 * 视频任务记录。
 */
export interface VideoTask {
  id: number
  local_id: string
  user_id: number
  api_key_id: number
  public_model: string
  upstream_model: string
  account_id: number
  upstream_task_id: string
  status: VideoTaskStatus
  resolution: string
  duration_sec: number
  video_url: string
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

export interface ListVideoTasksParams {
  user_id: number
  page?: number
  page_size?: number
  status?: VideoTaskStatus
  signal?: AbortSignal
}

/**
 * 列出视频任务（按用户筛选）。
 * 后端当前仅支持 user_id 过滤，status 在前端内存过滤。
 */
export async function list(params: ListVideoTasksParams): Promise<PaginatedResponse<VideoTask>> {
  const queryParams: Record<string, any> = {
    user_id: params.user_id,
    page: params.page ?? 1,
    page_size: params.page_size ?? 20
  }
  if (params.status) {
    queryParams.status = params.status
  }
  const { data } = await apiClient.get<PaginatedResponse<VideoTask>>('/admin/video-tasks', {
    params: queryParams,
    signal: params.signal
  })
  return data
}

/**
 * 按 ID 或 local_id 获取任务详情。
 */
export async function get(idOrLocalID: string | number): Promise<VideoTask> {
  const { data } = await apiClient.get<VideoTask>(`/admin/video-tasks/${idOrLocalID}`)
  return data
}

/**
 * 取消任务。
 */
export async function cancel(id: number): Promise<{ id: number; status: string }> {
  const { data } = await apiClient.post<{ id: number; status: string }>(`/admin/video-tasks/${id}/cancel`)
  return data
}

export const videoTasksAPI = {
  list,
  get,
  cancel
}

export default videoTasksAPI
