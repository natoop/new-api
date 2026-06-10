/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { api } from '@/lib/api'
import type {
  ApiResponse,
  DistributionAgent,
  DistributionAttributionLog,
  DistributionCustomerOwnership,
  DistributionGiftRule,
  DistributionInventory,
  DistributionInvitation,
  DistributionLedger,
  DistributionOpsAuthorization,
  DistributionPackage,
  DistributionInventoryPackageOption,
  DistributionProfile,
  DistributionProfit,
  DistributionPromoCode,
  DistributionSubscriptionPlanRecord,
  DistributionUserOption,
  PaginatedData,
} from './types'

const silentReadConfig = {
  skipBusinessError: true,
  skipErrorHandler: true,
}

function createIdempotencyKey() {
  const cryptoApi = globalThis.crypto
  if (cryptoApi && typeof cryptoApi.randomUUID === 'function') {
    return cryptoApi.randomUUID()
  }
  if (cryptoApi && typeof cryptoApi.getRandomValues === 'function') {
    const values = new Uint32Array(4)
    cryptoApi.getRandomValues(values)
    return `idem:${Date.now().toString(36)}:${Array.from(values)
      .map((value) => value.toString(36))
      .join(':')}`
  }
  return `idem:${Date.now().toString(36)}:${Math.random()
    .toString(36)
    .slice(2)}:${Math.random().toString(36).slice(2)}`
}

function isRecordNotFoundMessage(message?: string) {
  return (message || '').toLowerCase().includes('record not found')
}

async function getSilent<T>(
  path: string,
  emptyData: T
): Promise<ApiResponse<T>> {
  try {
    const res = await api.get<ApiResponse<T>>(path, silentReadConfig)
    if (!res.data.success && isRecordNotFoundMessage(res.data.message)) {
      return { success: true, message: '', data: emptyData }
    }
    return res.data
  } catch (error) {
    const message =
      (error as { response?: { data?: { message?: string } } })?.response?.data
        ?.message || ''
    if (isRecordNotFoundMessage(message)) {
      return { success: true, message: '', data: emptyData }
    }
    return { success: false, message, data: emptyData }
  }
}

export async function getAgentProfile() {
  return getSilent<DistributionProfile | null>('/api/agent/profile', null)
}

export async function getAgentPackages(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  return getSilent<PaginatedData<DistributionPackage>>(
    `/api/agent/packages?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function purchaseAgentPackage(packageId: number) {
  try {
    const res = await api.post(
      '/api/agent/purchase',
      {
        package_id: packageId,
        idempotency_key: createIdempotencyKey(),
      },
      {
        skipBusinessError: true,
        skipErrorHandler: true,
      }
    )
    return res.data as ApiResponse
  } catch (error) {
    const message =
      (error as { response?: { data?: { message?: string } } })?.response?.data
        ?.message || ''
    return { success: false, message } as ApiResponse
  }
}

export async function getAgentInventory(
  params: {
    p?: number
    page_size?: number
    keyword?: string
  } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  if (params.keyword?.trim()) query.set('keyword', params.keyword.trim())
  return getSilent<PaginatedData<DistributionInventory>>(
    `/api/agent/inventory?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function getAgentInventoryPackageOptions() {
  return getSilent<DistributionInventoryPackageOption[]>(
    '/api/agent/inventory/packages',
    []
  )
}

export async function assignAgentInventory(
  inventoryId: number,
  customerUserId: number
) {
  const res = await api.post('/api/agent/inventory/assign', {
    inventory_id: inventoryId,
    customer_user_id: customerUserId,
  })
  return res.data as ApiResponse
}

export async function refundAgentInventory(inventoryId: number) {
  const res = await api.post('/api/agent/inventory/refund', {
    inventory_id: inventoryId,
  })
  return res.data as ApiResponse
}

export async function getAgentLedger(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  return getSilent<PaginatedData<DistributionLedger>>(
    `/api/agent/ledger?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function getAgentProfit(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  return getSilent<PaginatedData<DistributionProfit>>(
    `/api/agent/profit?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function getAgentCustomers(
  params: {
    p?: number
    page_size?: number
    keyword?: string
  } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  if (params.keyword?.trim()) query.set('keyword', params.keyword.trim())
  return getSilent<PaginatedData<DistributionCustomerOwnership>>(
    `/api/agent/customers?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function getAgentInvitations() {
  return getSilent<DistributionInvitation[]>('/api/agent/invitations', [])
}

export async function createAgentInvitation(payload: {
  invitee_email: string
  level: number
  expires_at: number
}) {
  const res = await api.post('/api/agent/invitations', {
    ...payload,
    idempotency_key: createIdempotencyKey(),
  })
  return res.data as ApiResponse
}

export async function acceptAgentInvitation(invitationNo: string) {
  const res = await api.post('/api/agent/invitations/accept', {
    invitation_no: invitationNo,
  })
  return res.data as ApiResponse
}

export async function getAgentPromoCodes(
  params: {
    p?: number
    page_size?: number
    time_filter?: string
    usage_filter?: string
  } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  if (params.time_filter) query.set('time_filter', params.time_filter)
  if (params.usage_filter) query.set('usage_filter', params.usage_filter)
  return getSilent<PaginatedData<DistributionPromoCode>>(
    `/api/agent/promo-codes?${query.toString()}`,
    {
      items: [],
      total: 0,
      page: params.p || 1,
      page_size: params.page_size || 10,
    }
  )
}

export async function saveAgentPromoCode(
  payload: Partial<DistributionPromoCode>
) {
  const path = payload.id
    ? `/api/agent/promo-codes/${payload.id}`
    : '/api/agent/promo-codes'
  const res = payload.id
    ? await api.put(path, payload)
    : await api.post(path, payload)
  return res.data as ApiResponse
}

export async function adminGetAgents() {
  const res = await api.get<ApiResponse<PaginatedData<DistributionAgent>>>(
    '/api/agent-admin/agents'
  )
  return res.data
}

export async function adminSearchAgents(
  params: {
    p?: number
    page_size?: number
    keyword?: string
  } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  if (params.keyword?.trim()) query.set('keyword', params.keyword.trim())
  const res = await api.get<ApiResponse<PaginatedData<DistributionAgent>>>(
    `/api/agent-admin/agents?${query.toString()}`
  )
  return res.data
}

export async function adminSaveAgent(payload: Partial<DistributionAgent>) {
  const res = await api.post('/api/agent-admin/agents', payload)
  return res.data as ApiResponse
}

export async function adminUpdateAgentStatus(agentId: number, status: string) {
  const res = await api.put(`/api/agent-admin/agents/${agentId}/status`, {
    status,
  })
  return res.data as ApiResponse
}

export async function adminAdjustAgentBalance(
  agentId: number,
  delta: number,
  description: string
) {
  const res = await api.post(`/api/agent-admin/agents/${agentId}/balance`, {
    delta,
    description,
    idempotency_key: createIdempotencyKey(),
  })
  return res.data as ApiResponse
}

export async function adminGetPackages(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  const res = await api.get<ApiResponse<PaginatedData<DistributionPackage>>>(
    `/api/agent-admin/packages?${query.toString()}`
  )
  return res.data
}

export async function adminGetSubscriptionPlans() {
  const res = await api.get<ApiResponse<DistributionSubscriptionPlanRecord[]>>(
    '/api/subscription/admin/plans'
  )
  return res.data
}

export async function adminSavePackage(payload: Partial<DistributionPackage>) {
  const path = payload.id
    ? `/api/agent-admin/packages/${payload.id}`
    : '/api/agent-admin/packages'
  const res = payload.id
    ? await api.put(path, payload)
    : await api.post(path, payload)
  return res.data as ApiResponse
}

export async function adminUpdatePackageStatus(
  packageId: number,
  status: string
) {
  const res = await api.patch(`/api/agent-admin/packages/${packageId}/status`, {
    status,
  })
  return res.data as ApiResponse
}

export async function adminGetProfit(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  const res = await api.get<ApiResponse<PaginatedData<DistributionProfit>>>(
    `/api/agent-admin/profit?${query.toString()}`
  )
  return res.data
}

export async function adminGetAttribution(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  const res = await api.get<
    ApiResponse<PaginatedData<DistributionAttributionLog>>
  >(`/api/agent-admin/attribution?${query.toString()}`)
  return res.data
}

export async function adminGetGiftRules(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  const res = await api.get<ApiResponse<PaginatedData<DistributionGiftRule>>>(
    `/api/agent-admin/gift-rules?${query.toString()}`
  )
  return res.data
}

export async function adminSaveGiftRule(
  payload: Partial<DistributionGiftRule>
) {
  const path = payload.id
    ? `/api/agent-admin/gift-rules/${payload.id}`
    : '/api/agent-admin/gift-rules'
  const res = payload.id
    ? await api.put(path, payload)
    : await api.post(path, payload)
  return res.data as ApiResponse
}

export async function adminUpdateGiftRuleStatus(
  ruleId: number,
  status: string
) {
  const res = await api.patch(`/api/agent-admin/gift-rules/${ruleId}/status`, {
    status,
  })
  return res.data as ApiResponse
}

export async function adminGetOpsAuthorizations(
  params: { p?: number; page_size?: number } = {}
) {
  const query = new URLSearchParams()
  query.set('p', String(params.p || 1))
  query.set('page_size', String(params.page_size || 10))
  const res = await api.get<
    ApiResponse<PaginatedData<DistributionOpsAuthorization>>
  >(`/api/agent-admin/ops-auth?${query.toString()}`)
  return res.data
}

export async function adminGrantOpsAuthorization(
  userId: number,
  remark: string
) {
  const res = await api.post('/api/agent-admin/ops-auth', {
    user_id: userId,
    remark,
  })
  return res.data as ApiResponse
}

export async function adminRevokeOpsAuthorization(userId: number) {
  const res = await api.delete(`/api/agent-admin/ops-auth/${userId}`)
  return res.data as ApiResponse
}

export async function adminGetUserOptions() {
  const res = await api.get<
    ApiResponse<{
      items: DistributionUserOption[]
      total: number
      page: number
      page_size: number
    }>
  >('/api/user/?p=1&page_size=1000')
  return res.data
}
