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

// ============================================================================
// Redemption Schema & Types
// ============================================================================

export const REDEMPTION_TYPE_VALUES = ['balance', 'plan', 'promo'] as const

export type RedemptionType = (typeof REDEMPTION_TYPE_VALUES)[number]

export const redemptionSchema = z.object({
  id: z.number(),
  user_id: z.number(),
  name: z.string(),
  key: z.string(),
  status: z.number(), // 1: enabled, 2: disabled, 3: used
  quota: z.number(),
  created_time: z.number(),
  redeemed_time: z.number(),
  expired_time: z.number(), // 0 for never expires
  used_user_id: z.number(),
  type: z.enum(REDEMPTION_TYPE_VALUES).catch('balance'), // balance | plan | promo
  plan_id: z.number().catch(0), // plan type: subscription plan to activate
  discount_bps: z.number().catch(0), // promo type: discount in basis points (2000 = 20% off)
  max_uses: z.number().catch(0), // promo type: max uses, 0 = unlimited
  used_count: z.number().catch(0), // promo type: times used
})

export type Redemption = z.infer<typeof redemptionSchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface GetRedemptionsParams {
  p?: number
  page_size?: number
}

export interface GetRedemptionsResponse {
  success: boolean
  message?: string
  data?: {
    items: Redemption[]
    total: number
    page: number
    page_size: number
  }
}

export interface SearchRedemptionsParams {
  keyword?: string
  p?: number
  page_size?: number
}

export interface RedemptionFormData {
  id?: number
  name: string
  type: RedemptionType
  quota: number
  expired_time: number
  count?: number // Only for create
  status?: number // Only for status update
  plan_id?: number // Only for plan type
  discount_bps?: number // Only for promo type
  max_uses?: number // Only for promo type
}

// Minimal shape of a subscription plan used for the plan dropdown / title mapping
export interface SubscriptionPlanOption {
  id: number
  title: string
  enabled: boolean
}

// ============================================================================
// Dialog Types
// ============================================================================

export type RedemptionsDialogType = 'create' | 'update' | 'delete' | 'view'
