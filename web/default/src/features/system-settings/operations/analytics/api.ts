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

// Backend payloads (controller/ops_analytics.go). All timestamps are unix
// seconds; `date` fields are day-aligned (UTC midnight).

export interface ChannelHealth {
  total: number
  enabled: number
  manually_disabled: number
  auto_disabled: number
}

export interface OpsOverview {
  days: number
  start_timestamp: number
  end_timestamp: number
  revenue: number
  order_total: number
  order_success: number
  order_success_rate: number
  new_users: number
  total_users: number
  enabled_users: number
  active_users: number
  consumption_quota: number
  consumption_usd: number
  request_count: number
  token_count: number
  channel: ChannelHealth
}

export interface RevenueTrendPoint {
  date: number
  revenue: number
  order_count: number
  success_count: number
}

export interface UserGrowthPoint {
  date: number
  new_users: number
  cumulative: number
}

export interface PaymentProviderStat {
  provider: string
  revenue: number
  order_total: number
  order_success: number
  success_rate: number
}

type ApiEnvelope<T> = { success: boolean; message: string; data: T }

async function getOps<T>(path: string, days: number): Promise<T> {
  const res = await api.get<ApiEnvelope<T>>(`/api/ops/${path}`, {
    params: { days },
  })
  return res.data.data
}

export function getOpsOverview(days: number) {
  return getOps<OpsOverview>('overview', days)
}

export function getOpsRevenueTrend(days: number) {
  return getOps<{ series: RevenueTrendPoint[] }>('revenue-trend', days)
}

export function getOpsUserGrowth(days: number) {
  return getOps<{ baseline: number; series: UserGrowthPoint[] }>(
    'user-growth',
    days
  )
}

export function getOpsPaymentProviders(days: number) {
  return getOps<{ providers: PaymentProviderStat[] }>('payment-providers', days)
}
