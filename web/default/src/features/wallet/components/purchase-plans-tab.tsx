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
import { Check, RefreshCw, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { StatusBadge, dotColorMap, textColorMap } from '@/components/status-badge'
import {
  getPublicPlans,
  getSelfSubscriptionFull,
  updateBillingPreference,
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
}

function getBillingPreferenceLabel(
  preference: string,
  t: (key: string) => string
): string {
  switch (preference) {
    case 'subscription_first':
      return t('Subscription First')
    case 'wallet_first':
      return t('Wallet First')
    case 'subscription_only':
      return t('Subscription Only')
    case 'wallet_only':
      return t('Wallet Only')
    default:
      return preference
  }
}

export function PurchasePlansTab({
  topupInfo,
  userQuota,
  onPurchaseSuccess,
}: PurchasePlansTabProps) {
  const { t } = useTranslation()

  const [plans, setPlans] = useState<PlanRecord[]>([])
  const [activeSubscriptions, setActiveSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [allSubscriptions, setAllSubscriptions] = useState<
    UserSubscriptionRecord[]
  >([])
  const [billingPreference, setBillingPreference] =
    useState('subscription_first')
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

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
        setBillingPreference(
          res.data.billing_preference || 'subscription_first'
        )
        setActiveSubscriptions(res.data.subscriptions || [])
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

  const handleRefresh = async () => {
    setRefreshing(true)
    try {
      await fetchSelfSubscription()
    } finally {
      setRefreshing(false)
    }
  }

  const handlePreferenceChange = async (pref: string) => {
    const previous = billingPreference
    setBillingPreference(pref)
    try {
      const res = await updateBillingPreference(pref)
      if (res.success) {
        toast.success(t('Updated successfully'))
        setBillingPreference(res.data?.billing_preference || pref)
      } else {
        setBillingPreference(previous)
      }
    } catch {
      toast.error(t('Request failed'))
      setBillingPreference(previous)
    }
  }

  const hasActive = activeSubscriptions.length > 0
  const hasAny = allSubscriptions.length > 0
  const disablePref = !hasActive
  const isSubPref =
    billingPreference === 'subscription_first' ||
    billingPreference === 'subscription_only'
  const displayPref =
    disablePref && isSubPref ? 'wallet_first' : billingPreference

  const planPurchaseCountMap = useMemo(() => {
    const map = new Map<number, number>()
    for (const sub of allSubscriptions) {
      const planId = sub?.subscription?.plan_id
      if (!planId) continue
      map.set(planId, (map.get(planId) || 0) + 1)
    }
    return map
  }, [allSubscriptions])

  const planTitleMap = useMemo(() => {
    const map = new Map<number, string>()
    for (const p of plans) {
      if (p?.plan?.id) {
        map.set(p.plan.id, p.plan.title || '')
      }
    }
    return map
  }, [plans])

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

  const getRemainingDays = (sub: UserSubscriptionRecord) => {
    const endTime = sub?.subscription?.end_time || 0
    if (!endTime) return 0
    const now = Date.now() / 1000
    return Math.max(0, Math.ceil((endTime - now) / 86400))
  }

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
        {/* My subscriptions & billing preference */}
        {hasAny && (
          <div className='rounded-xl border p-3 sm:p-4'>
            <div className='flex flex-wrap items-center justify-between gap-2.5 sm:gap-3'>
              <div className='flex min-w-0 flex-wrap items-center gap-2'>
                <span className='text-sm font-medium'>
                  {t('My Subscriptions')}
                </span>
                <span className='flex items-center gap-1.5 text-xs font-medium'>
                  <span
                    className={cn(
                      'size-1.5 shrink-0 rounded-full',
                      hasActive ? dotColorMap.success : dotColorMap.neutral
                    )}
                    aria-hidden='true'
                  />
                  {hasActive ? (
                    <span className={cn(textColorMap.success)}>
                      {activeSubscriptions.length} {t('active')}
                    </span>
                  ) : (
                    <span className='text-muted-foreground'>
                      {t('No Active')}
                    </span>
                  )}
                  {allSubscriptions.length > activeSubscriptions.length && (
                    <>
                      <span className='text-muted-foreground/30'>·</span>
                      <span className='text-muted-foreground'>
                        {allSubscriptions.length - activeSubscriptions.length}{' '}
                        {t('expired')}
                      </span>
                    </>
                  )}
                </span>
              </div>
              <div className='flex w-full items-center gap-2 sm:w-auto'>
                <Select
                  items={[
                    'subscription_first',
                    'wallet_first',
                    'subscription_only',
                    'wallet_only',
                  ].map((value) => ({
                    value,
                    label: getBillingPreferenceLabel(value, t),
                  }))}
                  value={displayPref}
                  onValueChange={(v) => v !== null && handlePreferenceChange(v)}
                >
                  <SelectTrigger className='h-8 flex-1 text-xs sm:w-[140px] sm:flex-none'>
                    <SelectValue>
                      {getBillingPreferenceLabel(displayPref, t)}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent alignItemWithTrigger={false}>
                    <SelectGroup>
                      <SelectItem
                        value='subscription_first'
                        disabled={disablePref}
                      >
                        {getBillingPreferenceLabel('subscription_first', t)}
                        {disablePref ? ` (${t('No Active')})` : ''}
                      </SelectItem>
                      <SelectItem value='wallet_first'>
                        {getBillingPreferenceLabel('wallet_first', t)}
                      </SelectItem>
                      <SelectItem
                        value='subscription_only'
                        disabled={disablePref}
                      >
                        {getBillingPreferenceLabel('subscription_only', t)}
                        {disablePref ? ` (${t('No Active')})` : ''}
                      </SelectItem>
                      <SelectItem value='wallet_only'>
                        {getBillingPreferenceLabel('wallet_only', t)}
                      </SelectItem>
                    </SelectGroup>
                  </SelectContent>
                </Select>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-8 w-8'
                  onClick={handleRefresh}
                  disabled={refreshing}
                >
                  <RefreshCw
                    className={`h-3.5 w-3.5 ${refreshing ? 'animate-spin' : ''}`}
                  />
                </Button>
              </div>
            </div>

            <Separator className='my-3' />
            <div className='max-h-64 space-y-3 overflow-y-auto pr-1'>
              {allSubscriptions.map((sub) => {
                const subscription = sub.subscription
                const subTotal = Number(subscription?.amount_total || 0)
                const subUsed = Number(subscription?.amount_used || 0)
                const subRemain =
                  subTotal > 0 ? Math.max(0, subTotal - subUsed) : 0
                const usagePercent =
                  subTotal > 0 ? Math.round((subUsed / subTotal) * 100) : 0
                const planTitle = planTitleMap.get(subscription?.plan_id) || ''
                const now = Date.now() / 1000
                const isExpired = (subscription?.end_time || 0) < now
                const isCancelled = subscription?.status === 'cancelled'
                const isActive = subscription?.status === 'active' && !isExpired

                return (
                  <div
                    key={subscription?.id}
                    className='bg-background rounded-md border p-3 text-xs'
                  >
                    <div className='flex items-center justify-between'>
                      <div className='flex items-center gap-2'>
                        <span className='font-medium'>
                          {planTitle
                            ? `${planTitle} · ${t('Subscription')} #${subscription?.id}`
                            : `${t('Subscription')} #${subscription?.id}`}
                        </span>
                        {isActive ? (
                          <StatusBadge
                            label={t('Active')}
                            variant='success'
                            copyable={false}
                          />
                        ) : isCancelled ? (
                          <StatusBadge
                            label={t('Cancelled')}
                            variant='neutral'
                            copyable={false}
                          />
                        ) : (
                          <StatusBadge
                            label={t('Expired')}
                            variant='neutral'
                            copyable={false}
                          />
                        )}
                      </div>
                      {isActive && (
                        <span className='text-muted-foreground'>
                          {t('{{count}} days remaining', {
                            count: getRemainingDays(sub),
                          })}
                        </span>
                      )}
                    </div>
                    <div className='text-muted-foreground mt-1.5'>
                      {isActive
                        ? t('Until')
                        : isCancelled
                          ? t('Cancelled at')
                          : t('Expired at')}{' '}
                      {new Date(
                        (subscription?.end_time || 0) * 1000
                      ).toLocaleString()}
                    </div>
                    {isActive && (subscription?.next_reset_time ?? 0) > 0 && (
                      <div className='text-muted-foreground mt-1'>
                        {t('Next reset')}:{' '}
                        {new Date(
                          subscription!.next_reset_time! * 1000
                        ).toLocaleString()}
                      </div>
                    )}
                    <div className='text-muted-foreground mt-1'>
                      {t('Total Quota')}:{' '}
                      {subTotal > 0 ? (
                        <>
                          {formatQuota(subUsed)}/{formatQuota(subTotal)} ·{' '}
                          {t('Remaining')} {formatQuota(subRemain)}
                          <span className='ml-2'>
                            {t('Used')} {usagePercent}%
                          </span>
                        </>
                      ) : (
                        t('Unlimited')
                      )}
                    </div>
                    {subTotal > 0 && isActive && (
                      <Progress value={usagePercent} className='mt-2 h-1.5' />
                    )}
                  </div>
                )
              })}
            </div>

            {disablePref && isSubPref && (
              <p className='text-muted-foreground mt-2 text-xs'>
                {t(
                  'Preference saved as {{pref}}, but no active subscription. Wallet will be used automatically.',
                  {
                    pref:
                      billingPreference === 'subscription_only'
                        ? t('Subscription Only')
                        : t('Subscription First'),
                  }
                )}
              </p>
            )}
          </div>
        )}

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
                          <Check className='text-primary h-3.5 w-3.5 shrink-0' />
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
