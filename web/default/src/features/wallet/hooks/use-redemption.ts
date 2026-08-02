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
import { useState, useCallback } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import { redeemTopupCode } from '../api'
import type { RedeemResultData } from '../types'

// ============================================================================
// Redemption Hook
// ============================================================================

/**
 * Normalized redemption outcome. The backend returns
 * { quota, type: "balance" | "plan", plan_id, plan_title }; legacy plain-int
 * responses are normalized to a balance outcome for safety.
 *
 * The `plan` type only comes from agent inventory codes, which now credit the
 * recharge tier's quota straight to the wallet instead of activating a
 * subscription. That branch carries both the tier title and the quota that
 * actually landed in the wallet.
 */
export interface RedeemOutcome {
  type: 'balance' | 'plan'
  quota: number
  planId?: number
  planTitle?: string
}

function normalizeRedeemResult(data: RedeemResultData | number): RedeemOutcome {
  if (typeof data === 'number') {
    return { type: 'balance', quota: data }
  }
  return {
    type: data.type === 'plan' ? 'plan' : 'balance',
    quota: Number(data.quota) || 0,
    planId: data.plan_id,
    planTitle: data.plan_title,
  }
}

export function useRedemption() {
  const [redeeming, setRedeeming] = useState(false)

  const redeemCode = useCallback(
    async (code: string): Promise<RedeemOutcome | null> => {
      if (!code || code.trim() === '') {
        toast.error(i18next.t('Please enter a redemption code'))
        return null
      }

      try {
        setRedeeming(true)
        const response = await redeemTopupCode({ key: code.trim() })

        if (response.success && response.data !== undefined) {
          return normalizeRedeemResult(response.data)
        }

        // Promo codes are rejected here with a dedicated backend message
        // ("促销码请在购买套餐时使用") — surface it as-is.
        toast.error(response.message || i18next.t('Redemption failed'))
        return null
      } catch (_error) {
        toast.error(i18next.t('Redemption failed'))
        return null
      } finally {
        setRedeeming(false)
      }
    },
    []
  )

  return {
    redeeming,
    redeemCode,
  }
}
