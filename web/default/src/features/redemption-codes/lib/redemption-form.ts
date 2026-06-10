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
import { z } from 'zod'
import type { TFunction } from 'i18next'
import { parseQuotaFromDollars, quotaUnitsToDollars } from '@/lib/format'
import {
  REDEMPTION_VALIDATION,
  getRedemptionFormErrorMessages,
} from '../constants'
import {
  REDEMPTION_TYPE_VALUES,
  type RedemptionFormData,
  type Redemption,
  type RedemptionType,
} from '../types'

// ============================================================================
// Form Schema (use getRedemptionFormSchema(t) in components for i18n messages)
// ============================================================================

export function getRedemptionFormSchema(t: TFunction) {
  const msg = getRedemptionFormErrorMessages(t)
  return z
    .object({
      name: z
        .string()
        .min(REDEMPTION_VALIDATION.NAME_MIN_LENGTH, msg.NAME_LENGTH_INVALID)
        .max(REDEMPTION_VALIDATION.NAME_MAX_LENGTH, msg.NAME_LENGTH_INVALID),
      type: z.enum(REDEMPTION_TYPE_VALUES),
      quota_dollars: z.number().min(0, t('Quota must be a positive number')),
      plan_id: z.number().optional(),
      discount_percent: z.number().optional(),
      max_uses: z.number().optional(),
      expired_time: z.date().optional(),
      count: z
        .number()
        .min(REDEMPTION_VALIDATION.COUNT_MIN, msg.COUNT_INVALID)
        .max(REDEMPTION_VALIDATION.COUNT_MAX, msg.COUNT_INVALID)
        .optional(),
    })
    .superRefine((data, ctx) => {
      if (data.type === 'plan' && (!data.plan_id || data.plan_id <= 0)) {
        ctx.addIssue({
          code: 'custom',
          path: ['plan_id'],
          message: t('Please select a plan'),
        })
      }
      if (data.type === 'promo') {
        const percent = data.discount_percent ?? 0
        // 0 < bps < 10000 on backend, i.e. 0 < percent < 100
        if (percent <= 0 || percent >= 100) {
          ctx.addIssue({
            code: 'custom',
            path: ['discount_percent'],
            message: t('Discount must be greater than 0 and less than 100'),
          })
        }
        if ((data.max_uses ?? 0) < 0) {
          ctx.addIssue({
            code: 'custom',
            path: ['max_uses'],
            message: t('Max uses cannot be negative'),
          })
        }
      }
    })
}

export type RedemptionFormValues = {
  name: string
  type: RedemptionType
  quota_dollars: number
  plan_id?: number
  discount_percent?: number // UI percent: 20 = 20% off (stored as 2000 bps)
  max_uses?: number // 0 = unlimited
  expired_time?: Date
  count?: number
}

// ============================================================================
// Form Defaults
// ============================================================================

export const REDEMPTION_FORM_DEFAULT_VALUES: RedemptionFormValues = {
  name: '',
  type: 'balance',
  quota_dollars: 10,
  plan_id: undefined,
  discount_percent: 20,
  max_uses: 0,
  expired_time: undefined,
  count: 1,
}

// ============================================================================
// Form Data Transformation
// ============================================================================

/**
 * Transform form data to API payload.
 * Only fields relevant to the selected type are included.
 */
export function transformFormDataToPayload(
  data: RedemptionFormValues
): RedemptionFormData {
  const base: RedemptionFormData = {
    name: data.name,
    type: data.type,
    quota: 0,
    expired_time: data.expired_time
      ? Math.floor(data.expired_time.getTime() / 1000)
      : 0,
    count: data.count || 1,
  }

  switch (data.type) {
    case 'plan':
      return { ...base, plan_id: data.plan_id }
    case 'promo':
      return {
        ...base,
        // UI percent -> basis points (20% -> 2000 bps)
        discount_bps: Math.round((data.discount_percent ?? 0) * 100),
        max_uses: data.max_uses ?? 0,
      }
    default:
      return { ...base, quota: parseQuotaFromDollars(data.quota_dollars) }
  }
}

/**
 * Transform redemption data to form defaults
 */
export function transformRedemptionToFormDefaults(
  redemption: Redemption
): RedemptionFormValues {
  return {
    name: redemption.name,
    type: redemption.type || 'balance',
    quota_dollars: quotaUnitsToDollars(redemption.quota),
    plan_id: redemption.plan_id > 0 ? redemption.plan_id : undefined,
    // basis points -> UI percent (2000 bps -> 20%)
    discount_percent:
      redemption.discount_bps > 0 ? redemption.discount_bps / 100 : undefined,
    max_uses: redemption.max_uses ?? 0,
    expired_time:
      redemption.expired_time > 0
        ? new Date(redemption.expired_time * 1000)
        : undefined,
    count: 1,
  }
}
