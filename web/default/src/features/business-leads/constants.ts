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
import { type TFunction } from 'i18next'
import { type StatusBadgeProps } from '@/components/status-badge'
import { COOPERATION_TYPE_OPTIONS } from '@/features/cooperation/constants'

// Shared React Query key root; status mutations and delete invalidate it.
export const BUSINESS_LEAD_QUERY_KEY = 'business-leads'

// ============================================================================
// Business Lead Status Configuration (matches model/business_lead.go)
// ============================================================================

export const BUSINESS_LEAD_STATUS = {
  PENDING: 'pending',
  CONTACTED: 'contacted',
  ARCHIVED: 'archived',
} as const

export type BusinessLeadStatus =
  (typeof BUSINESS_LEAD_STATUS)[keyof typeof BUSINESS_LEAD_STATUS]

export const BUSINESS_LEAD_STATUS_VALUES = Object.values(
  BUSINESS_LEAD_STATUS
) as [BusinessLeadStatus, ...BusinessLeadStatus[]]

// labelKey values are i18n keys; use t(config.labelKey) in components.
export const BUSINESS_LEAD_STATUSES: Record<
  string,
  Pick<StatusBadgeProps, 'variant'> & {
    labelKey: string
    value: BusinessLeadStatus
  }
> = {
  [BUSINESS_LEAD_STATUS.PENDING]: {
    labelKey: 'Pending',
    variant: 'warning',
    value: BUSINESS_LEAD_STATUS.PENDING,
  },
  [BUSINESS_LEAD_STATUS.CONTACTED]: {
    labelKey: 'Contacted',
    variant: 'success',
    value: BUSINESS_LEAD_STATUS.CONTACTED,
  },
  [BUSINESS_LEAD_STATUS.ARCHIVED]: {
    labelKey: 'Archived',
    variant: 'neutral',
    value: BUSINESS_LEAD_STATUS.ARCHIVED,
  },
}

export function getBusinessLeadStatusOptions(t: TFunction) {
  return Object.values(BUSINESS_LEAD_STATUSES).map((config) => ({
    label: t(config.labelKey),
    value: config.value,
  }))
}

// ============================================================================
// Cooperation Type label lookup — reuses the public form's option list so the
// admin board and the public page stay in sync. labelKey goes through t().
// ============================================================================

export const COOPERATION_TYPE_LABEL_KEYS: Record<string, string> =
  Object.fromEntries(
    COOPERATION_TYPE_OPTIONS.map((option) => [option.value, option.labelKey])
  )

export function getCooperationTypeLabel(t: TFunction, value: string): string {
  const labelKey = COOPERATION_TYPE_LABEL_KEYS[value]
  return labelKey ? t(labelKey) : value
}

// ============================================================================
// Messages (i18n keys; use t(...) when displaying)
// ============================================================================

export const SUCCESS_MESSAGES = {
  STATUS_UPDATED: 'Lead status updated successfully',
  LEAD_DELETED: 'Business lead deleted successfully',
} as const
