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
import { useState, useEffect, useMemo, useCallback } from 'react'
import i18next from 'i18next'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { DEFAULT_CURRENCY_CONFIG } from '@/stores/system-config-store'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Skeleton } from '@/components/ui/skeleton'
import { getRechargePlans } from '../api'
import { getDiscountLabel } from '../lib'
import type { RechargePlan, RechargePlanRecord } from '../types'
import { PlanPurchaseDialog } from './dialogs/plan-purchase-dialog'
import { RechargeTierCard } from './recharge-tier-card'

interface PurchasePlansTabProps {
  userQuota?: number
  onPurchaseSuccess?: () => void | Promise<void>
  onGoTopup?: () => void
}

/**
 * Quota a tier credits versus the quota its price buys at list rate; below 1
 * means the tier is a discount.
 */
function getTierDiscount(plan: RechargePlan, quotaPerUnit: number): number {
  const credited = Number(plan.total_amount || 0)
  if (credited <= 0) return 1
  const listQuota = Number(plan.price_amount || 0) * quotaPerUnit
  return listQuota > 0 ? listQuota / credited : 1
}

export function PurchasePlansTab({
  userQuota,
  onPurchaseSuccess,
  onGoTopup,
}: PurchasePlansTabProps) {
  const { t } = useTranslation()
  const { currency } = useSystemConfig()

  const [plans, setPlans] = useState<RechargePlanRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [purchaseOpen, setPurchaseOpen] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState<RechargePlanRecord | null>(
    null
  )

  const quotaPerUnit =
    currency?.quotaPerUnit && currency.quotaPerUnit > 0
      ? currency.quotaPerUnit
      : DEFAULT_CURRENCY_CONFIG.quotaPerUnit

  // A failed request must not look like "no tiers configured": surface it the
  // same way PlanPurchaseDialog surfaces its own failures. Translating through
  // i18next directly (as useRedemption does) keeps this callback's identity
  // stable, so the mount effect below still runs exactly once.
  const fetchPlans = useCallback(async () => {
    try {
      const res = await getRechargePlans()
      setPlans(res.success ? res.data || [] : [])
    } catch {
      setPlans([])
      toast.error(i18next.t('Failed to load recharge tiers'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    // fetchPlans only sets state after awaiting the request; the compiler can
    // not prove that through the catch branch, so it reports a sync setState.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void fetchPlans()
  }, [fetchPlans])

  // Recommended tier: middle price among the listed tiers (no backend marker
  // field exists on the plan schema).
  const recommendedPlanId = useMemo(() => {
    const valid = plans.filter((p) => p?.plan)
    if (valid.length < 2) return null
    const sorted = [...valid].sort(
      (a, b) =>
        Number(a.plan.price_amount || 0) - Number(b.plan.price_amount || 0)
    )
    return sorted[Math.floor((sorted.length - 1) / 2)].plan.id
  }, [plans])

  if (loading) {
    return (
      <div className='space-y-4'>
        <Skeleton className='h-16 w-full rounded-xl' />
        <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className='h-56 w-full rounded-xl' />
          ))}
        </div>
      </div>
    )
  }

  return (
    <>
      <div className='space-y-4 sm:space-y-5'>
        {plans.length > 0 ? (
          <div className='grid grid-cols-1 gap-3 pt-2 sm:grid-cols-2 sm:gap-4 xl:grid-cols-3'>
            {plans.map((record) => {
              const plan = record?.plan
              if (!plan) return null
              return (
                <RechargeTierCard
                  key={plan.id}
                  plan={plan}
                  creditedQuota={Number(plan.total_amount || 0)}
                  discountLabel={getDiscountLabel(
                    getTierDiscount(plan, quotaPerUnit)
                  )}
                  recommended={plan.id === recommendedPlanId}
                  onSelect={() => {
                    setSelectedPlan(record)
                    setPurchaseOpen(true)
                  }}
                />
              )
            })}
          </div>
        ) : (
          <p className='text-muted-foreground py-8 text-center text-sm'>
            {t('No recharge tiers available')}
          </p>
        )}
      </div>

      <PlanPurchaseDialog
        open={purchaseOpen}
        onOpenChange={setPurchaseOpen}
        planRecord={selectedPlan}
        userQuota={userQuota}
        onPurchaseSuccess={onPurchaseSuccess}
        onGoTopup={onGoTopup}
      />
    </>
  )
}
