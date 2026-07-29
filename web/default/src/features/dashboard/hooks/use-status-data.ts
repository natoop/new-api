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
import { useStatus } from '@/hooks/use-status'
import type { AnnouncementItem, ApiInfoItem, FAQItem } from '../types'

/**
 * Get specific list from status data
 */
export function useStatusData<T = unknown>(
  enabledKey: string,
  dataKey: string
): { items: T[]; loading: boolean } {
  const { status, loading } = useStatus()
  const enabled = status ? status[enabledKey] !== false : false
  const items = (enabled ? status?.[dataKey] || [] : []) as T[]

  return { items, loading }
}

/**
 * Get API info list
 */
export function useApiInfo() {
  return useStatusData<ApiInfoItem>('api_info_enabled', 'api_info')
}

/**
 * Get announcements list
 */
export function useAnnouncements() {
  return useStatusData<AnnouncementItem>(
    'announcements_enabled',
    'announcements'
  )
}

/**
 * Get FAQ list
 */
export function useFAQ() {
  return useStatusData<FAQItem>('faq_enabled', 'faq')
}

/** 后端把这些内容字段当空串/空数组下发，取一个安全的长度。 */
function hasEntries(value: unknown): boolean {
  return Array.isArray(value) && value.length > 0
}

/**
 * Get dashboard content panel visibility
 *
 * 开关是"缺省为开"（`!== false`），而内容字段在后端默认为空——
 * 于是全新部署的概览页会恒定挂着几张空面板。这里要求开关开**且**确实有内容，
 * 已经配置过内容的部署不受影响。
 *
 * uptimeKuma 是例外：它在 status 里没有对应的内容字段，
 * 数据来自系统设置的 uptime_kuma_groups，只能由面板自己判断。
 */
export function useDashboardContentVisibility() {
  const { status } = useStatus()
  const hasStatus = Boolean(status)

  return {
    apiInfo:
      hasStatus &&
      status?.api_info_enabled !== false &&
      hasEntries(status?.api_info),
    announcements:
      hasStatus &&
      status?.announcements_enabled !== false &&
      hasEntries(status?.announcements),
    faq: hasStatus && status?.faq_enabled !== false && hasEntries(status?.faq),
    uptimeKuma: hasStatus && status?.uptime_kuma_enabled !== false,
  }
}
