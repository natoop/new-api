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
import { useCallback, useEffect, useMemo, useState } from 'react'
import { ArrowRight, Crown, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Skeleton } from '@/components/ui/skeleton'
import { StatusBadge } from '@/components/status-badge'
import {
  getPublicPlans,
  getSelfSubscriptionFull,
} from '@/features/subscriptions/api'
import type {
  PlanRecord,
  UserSubscriptionRecord,
} from '@/features/subscriptions/types'

interface MySubscriptionCardProps {
  refreshKey?: number
  onGoPlans: () => void
}

function getRemainingDays(sub: UserSubscriptionRecord): number {
  const endTime = sub?.subscription?.end_time || 0
  if (!endTime) return 0
  return Math.max(0, Math.ceil((endTime - Date.now() / 1000) / 86400))
}

export function MySubscriptionCard({
  refreshKey = 0,
  onGoPlans,
}: MySubscriptionCardProps) {
  const { t } = useTranslation()
  const [plans, setPlans] = useState<PlanRecord[]>([])
  const [activeSubscriptions, setActiveSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const fetchData = useCallback(async () => {
    const [plansRes, subscriptionRes] = await Promise.allSettled([
      getPublicPlans(),
      getSelfSubscriptionFull(),
    ])

    if (plansRes.status === 'fulfilled' && plansRes.value.success) {
      setPlans(plansRes.value.data || [])
    }
    if (
      subscriptionRes.status === 'fulfilled' &&
      subscriptionRes.value.success &&
      subscriptionRes.value.data
    ) {
      setActiveSubscriptions(subscriptionRes.value.data.subscriptions || [])
    }
  }, [])

  useEffect(() => {
    const init = async () => {
      setLoading(true)
      try {
        await fetchData()
      } finally {
        setLoading(false)
      }
    }
    void init()
  }, [fetchData, refreshKey])

  const planTitleMap = useMemo(() => {
    const map = new Map<number, string>()
    for (const record of plans) {
      const plan = record?.plan
      if (plan?.id) map.set(plan.id, plan.title || '')
    }
    return map
  }, [plans])

  const current = activeSubscriptions[0]
  const subscription = current?.subscription
  const total = Number(subscription?.amount_total || 0)
  const used = Number(subscription?.amount_used || 0)
  const remaining = total > 0 ? Math.max(0, total - used) : 0
  const usagePercent =
    total > 0 ? Math.min(100, Math.round((used / total) * 100)) : 0
  const planTitle =
    (subscription?.plan_id && planTitleMap.get(subscription.plan_id)) ||
    t('Subscription')

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await fetchData()
    } finally {
      setRefreshing(false)
    }
  }

  if (loading) {
    return (
      <Card className='py-0'>
        <CardContent className='space-y-4 p-4 sm:p-5'>
          <div className='flex items-center gap-3'>
            <Skeleton className='size-9 rounded-lg' />
            <div className='flex-1 space-y-2'>
              <Skeleton className='h-4 w-32' />
              <Skeleton className='h-3 w-44' />
            </div>
          </div>
          <Skeleton className='h-20 w-full rounded-xl' />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className='py-0'>
      <CardContent className='space-y-4 p-4 sm:p-5'>
        <div className='flex items-start justify-between gap-3'>
          <div className='flex min-w-0 items-center gap-3'>
            <div
              className={cn(
                'flex size-9 shrink-0 items-center justify-center rounded-lg',
                current
                  ? 'bg-primary/10 text-primary'
                  : 'bg-muted text-muted-foreground'
              )}
            >
              <Crown className='size-4' />
            </div>
            <div className='min-w-0'>
              <h3 className='truncate text-sm font-semibold'>
                {t('My Subscription')}
              </h3>
              <p className='text-muted-foreground truncate text-xs'>
                {current
                  ? t('Active subscription')
                  : t('Choose a plan to start using subscription quota.')}
              </p>
            </div>
          </div>
          <Button
            variant='ghost'
            size='icon'
            className='size-8 shrink-0'
            onClick={handleRefresh}
            disabled={refreshing}
          >
            <RefreshCw
              className={cn('size-3.5', refreshing && 'animate-spin')}
            />
          </Button>
        </div>

        {current && subscription ? (
          <div className='bg-muted/20 space-y-3 rounded-xl border p-3'>
            <div className='flex items-start justify-between gap-3'>
              <div className='min-w-0'>
                <p className='truncate text-sm font-semibold'>{planTitle}</p>
                <p className='text-muted-foreground mt-0.5 text-xs'>
                  {t('{{count}} days remaining', {
                    count: getRemainingDays(current),
                  })}
                </p>
              </div>
              <StatusBadge
                label={t('Active')}
                variant='success'
                copyable={false}
              />
            </div>

            <div className='space-y-2'>
              <div className='flex items-center justify-between text-xs'>
                <span className='text-muted-foreground'>
                  {t('Subscription usage')}
                </span>
                <span className='font-medium tabular-nums'>
                  {total > 0 ? `${usagePercent}%` : t('Unlimited')}
                </span>
              </div>
              {total > 0 ? (
                <>
                  <Progress value={usagePercent} className='h-1.5' />
                  <div className='text-muted-foreground flex items-center justify-between gap-2 text-xs'>
                    <span>{formatQuota(used)}</span>
                    <span>
                      {t('Remaining')} {formatQuota(remaining)}
                    </span>
                  </div>
                </>
              ) : (
                <div className='text-muted-foreground text-xs'>
                  {t('Total Quota')}: {t('Unlimited')}
                </div>
              )}
            </div>
          </div>
        ) : (
          <div className='bg-muted/20 rounded-xl border border-dashed p-3'>
            <p className='text-sm font-medium'>{t('No active subscription')}</p>
            <p className='text-muted-foreground mt-1 text-xs'>
              {t('Subscribe to a plan for model access')}
            </p>
            <Button
              variant='outline'
              size='sm'
              className='mt-3 w-full gap-2'
              onClick={onGoPlans}
            >
              {t('Browse plans')}
              <ArrowRight className='size-3.5' />
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
