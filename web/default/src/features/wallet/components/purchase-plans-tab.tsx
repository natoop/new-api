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
import { Check, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  getPublicPlans,
  getSelfSubscriptionFull,
} from '@/features/subscriptions/api'
import { formatDuration, formatResetPeriod } from '@/features/subscriptions/lib'
import type {
  PlanRecord,
  UserSubscriptionRecord,
} from '@/features/subscriptions/types'
import type { TopupInfo } from '../types'
import { PlanPurchaseDialog } from './dialogs/plan-purchase-dialog'

interface PurchasePlansTabProps {
  topupInfo: TopupInfo | null
  userQuota?: number
  onPurchaseSuccess?: () => void | Promise<void>
  onGoTopup?: () => void
}

export function PurchasePlansTab({
  topupInfo,
  userQuota,
  onPurchaseSuccess,
  onGoTopup,
}: PurchasePlansTabProps) {
  const { t } = useTranslation()

  const [plans, setPlans] = useState<PlanRecord[]>([])
  const [allSubscriptions, setAllSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [loading, setLoading] = useState(true)

  const [purchaseOpen, setPurchaseOpen] = useState(false)
  const [selectedPlan, setSelectedPlan] = useState<PlanRecord | null>(null)

  const fetchPlans = useCallback(async () => {
    try {
      const res = await getPublicPlans()
      if (res.success) {
        setPlans(res.data || [])
      }
    } catch {
      setPlans([])
    }
  }, [])

  const fetchSelfSubscription = useCallback(async () => {
    try {
      const res = await getSelfSubscriptionFull()
      if (res.success && res.data) {
        setAllSubscriptions(res.data.all_subscriptions || [])
      }
    } catch {
      // ignore
    }
  }, [])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      await Promise.all([fetchPlans(), fetchSelfSubscription()])
      setLoading(false)
    }
    init()
  }, [fetchPlans, fetchSelfSubscription])

  const planPurchaseCountMap = useMemo(() => {
    const map = new Map<number, number>()
    for (const sub of allSubscriptions) {
      const planId = sub?.subscription?.plan_id
      if (!planId) continue
      map.set(planId, (map.get(planId) || 0) + 1)
    }
    return map
  }, [allSubscriptions])

  // Recommended tier: middle price among the listed plans (no backend marker
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

  const handlePurchaseSuccess = useCallback(async () => {
    await Promise.all([
      fetchSelfSubscription(),
      Promise.resolve(onPurchaseSuccess?.()),
    ])
  }, [fetchSelfSubscription, onPurchaseSuccess])

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
        {/* Marketing plan grid */}
        {plans.length > 0 ? (
          <div className='grid grid-cols-1 gap-3 pt-2 sm:grid-cols-2 sm:gap-4 xl:grid-cols-3'>
            {plans.map((p) => {
              const plan = p?.plan
              if (!plan) return null
              const totalAmount = Number(plan.total_amount || 0)
              const price = Number(plan.price_amount || 0).toFixed(2)
              const isRecommended = plan.id === recommendedPlanId
              const limit = Number(plan.max_purchase_per_user || 0)
              const count = planPurchaseCountMap.get(plan.id) || 0
              const reached = limit > 0 && count >= limit

              const benefits = [
                `${t('Validity Period')}: ${formatDuration(plan, t)}`,
                formatResetPeriod(plan, t) !== t('No Reset')
                  ? `${t('Quota Reset')}: ${formatResetPeriod(plan, t)}`
                  : null,
                totalAmount > 0
                  ? `${t('Total Quota')}: ${formatQuota(totalAmount)}`
                  : `${t('Total Quota')}: ${t('Unlimited')}`,
                limit > 0 ? `${t('Purchase Limit')}: ${limit}` : null,
                plan.upgrade_group
                  ? `${t('Upgrade Group')}: ${plan.upgrade_group}`
                  : null,
              ].filter(Boolean) as string[]

              return (
                <Card
                  key={plan.id}
                  className={cn(
                    'relative rounded-xl py-0 transition-shadow hover:shadow-sm',
                    isRecommended &&
                      'border-primary/60 from-primary/[0.05] bg-gradient-to-b to-transparent shadow-sm'
                  )}
                >
                  {isRecommended && (
                    <Badge className='absolute -top-2.5 left-1/2 -translate-x-1/2 gap-1 shadow-sm'>
                      <Sparkles className='h-3 w-3' />
                      {t('Recommended')}
                    </Badge>
                  )}
                  <CardContent className='flex h-full flex-col p-4 sm:p-5'>
                    <div className='min-w-0'>
                      <h4 className='truncate text-base font-semibold'>
                        {plan.title || t('Subscription Plans')}
                      </h4>
                      <p className='text-muted-foreground mt-0.5 line-clamp-2 min-h-4 text-xs'>
                        {plan.subtitle || ''}
                      </p>
                    </div>

                    <div className='flex items-baseline gap-1 py-3'>
                      <span
                        className={cn(
                          'text-3xl font-bold tracking-tight',
                          isRecommended && 'text-primary'
                        )}
                      >
                        ${price}
                      </span>
                      <span className='text-muted-foreground text-xs'>
                        / {formatDuration(plan, t)}
                      </span>
                    </div>

                    <div className='flex-1 space-y-2 pb-4'>
                      {benefits.map((label) => (
                        <div
                          key={label}
                          className='text-muted-foreground flex items-center gap-2 text-xs'
                        >
                          <Check className='text-accent-green h-3.5 w-3.5 shrink-0' />
                          <span>{label}</span>
                        </div>
                      ))}
                    </div>

                    {reached ? (
                      <Tooltip>
                        <TooltipTrigger render={<div />}>
                          <Button variant='outline' className='w-full' disabled>
                            {t('Limit Reached')}
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          {t('Purchase limit reached')} ({count}/{limit})
                        </TooltipContent>
                      </Tooltip>
                    ) : (
                      <Button
                        variant={isRecommended ? 'default' : 'outline'}
                        className='w-full'
                        onClick={() => {
                          setSelectedPlan(p)
                          setPurchaseOpen(true)
                        }}
                      >
                        {t('Subscribe Now')}
                      </Button>
                    )}
                  </CardContent>
                </Card>
              )
            })}
          </div>
        ) : (
          <p className='text-muted-foreground py-8 text-center text-sm'>
            {t('No plans available')}
          </p>
        )}
      </div>

      <PlanPurchaseDialog
        open={purchaseOpen}
        onOpenChange={(open) => {
          setPurchaseOpen(open)
          if (!open) {
            fetchSelfSubscription()
          }
        }}
        planRecord={selectedPlan}
        topupInfo={topupInfo}
        userQuota={userQuota}
        onPurchaseSuccess={handlePurchaseSuccess}
        onGoTopup={onGoTopup}
        purchaseLimit={
          selectedPlan?.plan?.max_purchase_per_user
            ? Number(selectedPlan.plan.max_purchase_per_user)
            : undefined
        }
        purchaseCount={
          selectedPlan?.plan?.id
            ? planPurchaseCountMap.get(selectedPlan.plan.id)
            : undefined
        }
      />
    </>
  )
}
