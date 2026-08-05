/**
 * Team API endpoints
 * Handles team management, departments, consumers, and analytics
 */

import { apiClient } from './client'
import type { BasePaginationResponse, FetchOptions } from '@/types'

// ==================== Team Types ====================

export interface Team {
  id: number
  name: string
  slug: string
  description: string
  timezone: string
  language: string
  owner_user_id: number
  billing_email?: string | null
  settings?: Record<string, unknown>
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
  deleted_at?: string | null
}

export interface CreateTeamRequest {
  name: string
  description?: string | null
}

export interface UpdateTeamRequest {
  name?: string
  description?: string | null
  timezone?: string
  language?: string
  status?: 'active' | 'inactive'
}

// ==================== Department Types ====================

export interface Department {
  id: number
  team_id: number
  name: string
  cost_center_code: string | null
  parent_id: number | null
  description: string | null
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface DepartmentTreeNode extends Department {
  children: DepartmentTreeNode[]
  code?: string
  memberCount?: number
}

export interface CreateDepartmentRequest {
  name: string
  cost_center_code?: string | null
  parent_id?: number | null
  description?: string | null
}

export interface UpdateDepartmentRequest {
  name?: string
  cost_center_code?: string | null
  parent_id?: number | null
  description?: string | null
  status?: 'active' | 'inactive'
}

// ==================== Consumer Types ====================

export interface Consumer {
  id: number
  team_id: number
  department_id: number | null
  type: string
  name: string
  email: string | null
  phone: string | null
  title: string | null
  app_description: string | null
  description: string
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface CreateConsumerRequest {
  name: string
  email?: string | null
  phone?: string | null
  title?: string | null
  type?: string
  description?: string | null
  dept_id?: number
}

export interface UpdateConsumerRequest {
  name?: string
  email?: string
  phone?: string | null
  title?: string | null
  type?: string
  description?: string | null
  dept_id?: number
  status?: 'active' | 'inactive'
}

// ==================== Analytics Types ====================

export interface TeamAnalyticsOverview {
  team_id: number
  total_requests: number
  input_tokens: number
  output_tokens: number
  total_cost: number
  actual_cost: number
}

export interface DepartmentRankingItem {
  department_id: number
  department_name: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  total_cost: number
  actual_cost: number
}

export interface ConsumerRankingItem {
  consumer_id: number
  consumer_name: string
  consumer_type: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  total_cost: number
  actual_cost: number
}

// ==================== Member Types ====================

export interface TeamMember {
  id: number
  team_id: number
  user_id: number
  role: string
  status: string
  department_id: number | null
  consumer_id: number | null
  joined_at: string
  created_at: string
  updated_at: string
  user?: {
    id: number
    name: string
    email: string
    avatar_url?: string
  }
}

export interface InviteMemberRequest {
  email: string
  role: string
  display_name?: string
}

export interface DailyTrendPoint {
  date: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  total_cost: number
  actual_cost: number
}

export interface ModelDistributionItem {
  model_name: string
  total_requests: number
  input_tokens: number
  output_tokens: number
  total_cost: number
  actual_cost: number
}

// ==================== Team APIs ====================

/**
 * List all teams
 * @param options - Fetch options (e.g. signal for cancellation)
 * @returns Paginated list of teams
 */
export async function listTeams(
  options?: FetchOptions
): Promise<BasePaginationResponse<Team>> {
  const { data } = await apiClient.get<BasePaginationResponse<Team>>('/teams', {
    signal: options?.signal,
  })
  return data
}

/**
 * Create a new team
 * @param payload - Team creation data
 * @returns Created team
 */
export async function createTeam(payload: CreateTeamRequest): Promise<Team> {
  const { data } = await apiClient.post<Team>('/teams', payload)
  return data
}

/**
 * Get team details by ID
 * @param id - Team ID
 * @returns Team details
 */
export async function getTeam(id: number): Promise<Team> {
  const { data } = await apiClient.get<Team>(`/teams/${id}`)
  return data
}

/**
 * Update a team
 * @param id - Team ID
 * @param payload - Team update data
 * @returns Updated team
 */
export async function updateTeam(
  id: number,
  payload: UpdateTeamRequest
): Promise<Team> {
  const { data } = await apiClient.put<Team>(`/teams/${id}`, payload)
  return data
}

/**
 * Delete a team
 * @param id - Team ID
 */
export async function deleteTeam(id: number): Promise<void> {
  await apiClient.delete(`/teams/${id}`)
}

// ==================== Department APIs ====================

/**
 * List departments for a team
 * @param teamId - Team ID
 * @param options - Fetch options
 * @returns Paginated list of departments
 */
export async function listDepartments(
  teamId: number,
  options?: FetchOptions
): Promise<BasePaginationResponse<Department>> {
  const { data } = await apiClient.get<BasePaginationResponse<Department>>(
    `/teams/${teamId}/departments`,
    { signal: options?.signal }
  )
  return data
}

/**
 * Create a department under a team
 * @param teamId - Team ID
 * @param payload - Department creation data
 * @returns Created department
 */
export async function createDepartment(
  teamId: number,
  payload: CreateDepartmentRequest
): Promise<Department> {
  const { data } = await apiClient.post<Department>(
    `/teams/${teamId}/departments`,
    payload
  )
  return data
}

/**
 * Get department tree for a team
 * @param teamId - Team ID
 * @returns Department tree structure
 */
export async function getDepartmentTree(
  teamId: number
): Promise<DepartmentTreeNode[]> {
  const { data } = await apiClient.get<DepartmentTreeNode[]>(
    `/teams/${teamId}/departments/tree`
  )
  return data
}

/**
 * Get a specific department
 * @param teamId - Team ID
 * @param deptId - Department ID
 * @returns Department details
 */
export async function getDepartment(
  teamId: number,
  deptId: number
): Promise<Department> {
  const { data } = await apiClient.get<Department>(
    `/teams/${teamId}/departments/${deptId}`
  )
  return data
}

/**
 * Update a department
 * @param teamId - Team ID
 * @param deptId - Department ID
 * @param payload - Department update data
 * @returns Updated department
 */
export async function updateDepartment(
  teamId: number,
  deptId: number,
  payload: UpdateDepartmentRequest
): Promise<Department> {
  const { data } = await apiClient.put<Department>(
    `/teams/${teamId}/departments/${deptId}`,
    payload
  )
  return data
}

/**
 * Delete a department
 * @param teamId - Team ID
 * @param deptId - Department ID
 */
export async function deleteDepartment(
  teamId: number,
  deptId: number
): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/departments/${deptId}`)
}

// ==================== Consumer APIs ====================

/**
 * List consumers for a team
 * @param teamId - Team ID
 * @param options - Fetch options
 * @param params - Optional params (page, page_size, dept_id to filter by department)
 * @returns Paginated list of consumers
 */
export async function listConsumers(
  teamId: number,
  options?: FetchOptions,
  params?: { page?: number; page_size?: number; dept_id?: number }
): Promise<BasePaginationResponse<Consumer>> {
  const { data } = await apiClient.get<BasePaginationResponse<Consumer>>(
    `/teams/${teamId}/consumers`,
    { signal: options?.signal, params }
  )
  return data
}

/**
 * Create a consumer under a team
 * @param teamId - Team ID
 * @param payload - Consumer creation data
 * @returns Created consumer
 */
export async function createConsumer(
  teamId: number,
  payload: CreateConsumerRequest
): Promise<Consumer> {
  const { data } = await apiClient.post<Consumer>(
    `/teams/${teamId}/consumers`,
    payload
  )
  return data
}

/**
 * Get a specific consumer
 * @param teamId - Team ID
 * @param consumerId - Consumer ID
 * @returns Consumer details
 */
export async function getConsumer(
  teamId: number,
  consumerId: number
): Promise<Consumer> {
  const { data } = await apiClient.get<Consumer>(
    `/teams/${teamId}/consumers/${consumerId}`
  )
  return data
}

/**
 * Update a consumer
 * @param teamId - Team ID
 * @param consumerId - Consumer ID
 * @param payload - Consumer update data
 * @returns Updated consumer
 */
export async function updateConsumer(
  teamId: number,
  consumerId: number,
  payload: UpdateConsumerRequest
): Promise<Consumer> {
  const { data } = await apiClient.put<Consumer>(
    `/teams/${teamId}/consumers/${consumerId}`,
    payload
  )
  return data
}

/**
 * Delete a consumer
 * @param teamId - Team ID
 * @param consumerId - Consumer ID
 */
export async function deleteConsumer(
  teamId: number,
  consumerId: number
): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/consumers/${consumerId}`)
}

// ==================== Member APIs ====================

/**
 * List members for a team
 * @param teamId - Team ID
 * @param params - Pagination params (page, page_size)
 * @param options - Fetch options
 * @returns Paginated list of team members
 */
export async function listMembers(
  teamId: number,
  params?: { page?: number; page_size?: number },
  options?: FetchOptions
): Promise<BasePaginationResponse<TeamMember>> {
  const { data } = await apiClient.get<BasePaginationResponse<TeamMember>>(
    `/teams/${teamId}/members`,
    { params, signal: options?.signal }
  )
  return data
}

/**
 * Invite a member to the team
 * @param teamId - Team ID
 * @param payload - Invite member request
 * @returns Created team member
 */
export async function inviteMember(
  teamId: number,
  payload: InviteMemberRequest
): Promise<TeamMember> {
  const { data } = await apiClient.post<TeamMember>(
    `/teams/${teamId}/members/invite`,
    payload
  )
  return data
}

/**
 * Update a member's role
 * @param teamId - Team ID
 * @param memberId - Member ID
 * @param role - New role
 */
export async function updateMemberRole(
  teamId: number,
  memberId: number,
  role: string
): Promise<void> {
  await apiClient.put(`/teams/${teamId}/members/${memberId}`, { role })
}

/**
 * Update a member's status (active/inactive)
 * @param teamId - Team ID
 * @param memberId - Member ID
 * @param status - New status ('active' | 'inactive')
 */
export async function updateMemberStatus(
  teamId: number,
  memberId: number,
  status: string
): Promise<void> {
  await apiClient.put(`/teams/${teamId}/members/${memberId}`, { status })
}

/**
 * Remove a member from the team
 * @param teamId - Team ID
 * @param memberId - Member ID
 */
export async function removeMember(
  teamId: number,
  memberId: number
): Promise<void> {
  await apiClient.delete(`/teams/${teamId}/members/${memberId}`)
}

// ==================== Analytics APIs ====================

/** Optional date range query params for analytics endpoints */
export interface AnalyticsDateRange {
  start_date?: string
  end_date?: string
}

/**
 * Get team analytics overview
 * @param teamId - Team ID
 * @param dateRange - Optional date range (start_date / end_date in YYYY-MM-DD)
 * @returns Analytics overview data
 */
export async function getAnalyticsOverview(
  teamId: number,
  dateRange?: AnalyticsDateRange
): Promise<TeamAnalyticsOverview> {
  const { data } = await apiClient.get<TeamAnalyticsOverview>(
    `/teams/${teamId}/analytics/overview`,
    { params: dateRange }
  )
  return data
}

/**
 * Get department ranking
 * @param teamId - Team ID
 * @param dateRange - Optional date range (start_date / end_date in YYYY-MM-DD)
 * @returns Department ranking list
 */
export async function getDepartmentRanking(
  teamId: number,
  dateRange?: AnalyticsDateRange
): Promise<DepartmentRankingItem[]> {
  const { data } = await apiClient.get<DepartmentRankingItem[]>(
    `/teams/${teamId}/analytics/departments/ranking`,
    { params: dateRange }
  )
  return data
}

/**
 * Get consumer ranking
 * @param teamId - Team ID
 * @param dateRange - Optional date range (start_date / end_date in YYYY-MM-DD)
 * @returns Consumer ranking list
 */
export async function getConsumerRanking(
  teamId: number,
  dateRange?: AnalyticsDateRange
): Promise<ConsumerRankingItem[]> {
  const { data } = await apiClient.get<ConsumerRankingItem[]>(
    `/teams/${teamId}/analytics/consumers/ranking`,
    { params: dateRange }
  )
  return data
}

/**
 * Get daily trend data
 * @param teamId - Team ID
 * @param dateRange - Optional date range (start_date / end_date in YYYY-MM-DD)
 * @param granularity - Granularity of trend data ('day' or 'hour'), defaults to 'day'
 * @returns Daily trend points
 */
export async function getDailyTrend(
  teamId: number,
  dateRange?: AnalyticsDateRange,
  granularity: 'day' | 'hour' = 'day'
): Promise<DailyTrendPoint[]> {
  const params: Record<string, string> = {}
  if (dateRange?.start_date) params.start_date = dateRange.start_date
  if (dateRange?.end_date) params.end_date = dateRange.end_date
  params.granularity = granularity
  const { data } = await apiClient.get<DailyTrendPoint[]>(
    `/teams/${teamId}/analytics/trend`,
    { params }
  )
  return data
}

/**
 * Get model distribution
 * @param teamId - Team ID
 * @param dateRange - Optional date range (start_date / end_date in YYYY-MM-DD)
 * @returns Model distribution list
 */
export async function getModelDistribution(
  teamId: number,
  dateRange?: AnalyticsDateRange
): Promise<ModelDistributionItem[]> {
  const { data } = await apiClient.get<ModelDistributionItem[]>(
    `/teams/${teamId}/analytics/models/distribution`,
    { params: dateRange }
  )
  return data
}

// ==================== API Namespace ====================

export const teamAPI = {
  // Team
  listTeams,
  createTeam,
  getTeam,
  updateTeam,
  deleteTeam,

  // Department
  listDepartments,
  createDepartment,
  getDepartmentTree,
  getDepartment,
  updateDepartment,
  deleteDepartment,

  // Consumer
  listConsumers,
  createConsumer,
  getConsumer,
  updateConsumer,
  deleteConsumer,

  // Member
  listMembers,
  inviteMember,
  updateMemberRole,
  updateMemberStatus,
  removeMember,

  // Analytics
  getAnalyticsOverview,
  getDepartmentRanking,
  getConsumerRanking,
  getDailyTrend,
  getModelDistribution,
}

export default teamAPI
