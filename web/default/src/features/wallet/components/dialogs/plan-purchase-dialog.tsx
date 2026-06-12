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
import { useState } from 'react'
import { CalendarClock, Crown, Loader2, Package } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { DEFAULT_CURRENCY_CONFIG } from '@/stores/system-config-store'
import { formatQuota } from '@/lib/format'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Dialog } from '@/components/dialog'
import { GroupBadge } from '@/components/group-badge'
import { paySubscriptionBalance } from '@/features/subscriptions/api'
import { formatDuration, formatResetPeriod } from '@/features/subscriptions/lib'
import type { PlanRecord } from '@/features/subscriptions/types'
import { formatPlanAmount } from '../../lib'
import type { TopupInfo } from '../../types'

interface PlanPurchaseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  planRecord: PlanRecord | null
  topupInfo: TopupInfo | null
  userQuota?: number
  purchaseLimit?: number
  purchaseCount?: number
  onPurchaseSuccess?: () => void | Promise<void>
  onGoTopup?: () => void
}

export function PlanPurchaseDialog(props: PlanPurchaseDialogProps) {
  const { t } = useTranslation()
  const { currency } = useSystemConfig()

  const [paying, setPaying] = useState(false)

  const plan = props.planRecord?.plan
  if (!plan) return null

  const originalAmount = Number(plan.price_amount || 0)
  const totalAmount = Number(plan.total_amount || 0)
  const quotaPerUnit =
    currency?.quotaPerUnit && currency.quotaPerUnit > 0
      ? currency.quotaPerUnit
      : DEFAULT_CURRENCY_CONFIG.quotaPerUnit
  const balanceCost = Math.max(0, Math.ceil(originalAmount * quotaPerUnit))
  const userQuota = Math.max(0, Number(props.userQuota || 0))
  const insufficientBalance = userQuota < balanceCost
  const limitReached =
    (props.purchaseLimit || 0) > 0 &&
    (props.purchaseCount || 0) >= (props.purchaseLimit || 0)

  const resetState = () => {
    setPaying(false)
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) resetState()
    props.onOpenChange(open)
  }

  const handlePayBalance = async () => {
    if (paying || insufficientBalance || limitReached) return
    setPaying(true)
    try {
      const res = await paySubscriptionBalance({
        plan_id: plan.id,
      })
      if (res.success) {
        toast.success(t('Subscription purchased successfully'))
        void props.onPurchaseSuccess?.()
        handleOpenChange(false)
      } else {
        toast.error(
          res.message && res.message !== 'success'
            ? res.message
            : t('Payment request failed')
        )
      }
    } catch {
      toast.error(t('Payment request failed'))
    } finally {
      setPaying(false)
    }
  }

  return (
    <Dialog
      open={props.open}
      onOpenChange={handleOpenChange}
      title={
        <>
          <Crown className='h-5 w-5' />
          {t('Purchase Subscription')}
        </>
      }
      contentClassName='glass-card max-sm:w-[calc(100vw-1.5rem)] sm:max-w-md'
      titleClassName='flex items-center gap-2'
      contentHeight='auto'
      bodyClassName='space-y-4'
    >
      <div className='space-y-3 sm:space-y-4'>
        <div className='bg-muted/50 space-y-2.5 rounded-xl border p-3 sm:space-y-3 sm:p-4'>
          <div className='flex justify-between'>
            <span className='text-muted-foreground text-sm'>
              {t('Plan Name')}
            </span>
            <span className='max-w-[200px] truncate text-sm font-medium'>
              {plan.title}
            </span>
          </div>
          <div className='flex items-center justify-between'>
            <span className='text-muted-foreground text-sm'>
              {t('Validity Period')}
            </span>
            <span className='flex items-center gap-1 text-sm'>
              <CalendarClock className='h-3.5 w-3.5' />
              {formatDuration(plan, t)}
            </span>
          </div>
          {formatResetPeriod(plan, t) !== t('No Reset') && (
            <div className='flex justify-between'>
              <span className='text-muted-foreground text-sm'>
                {t('Reset Period')}
              </span>
              <span className='text-sm'>{formatResetPeriod(plan, t)}</span>
            </div>
          )}
          <div className='flex items-center justify-between'>
            <span className='text-muted-foreground text-sm'>
              {t('Total Quota')}
            </span>
            <span className='flex items-center gap-1 text-sm'>
              <Package className='h-3.5 w-3.5' />
              {totalAmount > 0 ? formatQuota(totalAmount) : t('Unlimited')}
            </span>
          </div>
          {plan.upgrade_group && (
            <div className='flex items-center justify-between'>
              <span className='text-muted-foreground text-sm'>
                {t('Upgrade Group')}
              </span>
              <GroupBadge group={plan.upgrade_group} />
            </div>
          )}
          <Separator />
          <div className='flex items-center justify-between'>
            <span className='text-sm font-medium'>{t('Amount Due')}</span>
            <span className='flex items-baseline gap-2'>
              <span className='text-primary text-lg font-bold'>
                {formatPlanAmount(originalAmount, plan.currency)}
              </span>
            </span>
          </div>
        </div>

        {limitReached && (
          <Alert variant='destructive'>
            <AlertDescription>
              {t('Purchase limit reached')} ({props.purchaseCount}/
              {props.purchaseLimit})
            </AlertDescription>
          </Alert>
        )}

        <div className='flex flex-col gap-2 rounded-xl border p-3'>
          <div className='flex items-center justify-between gap-3'>
            <span className='text-sm font-medium'>{t('Pay with Balance')}</span>
            <Badge variant='secondary'>{t('Balance only')}</Badge>
          </div>
          <div className='flex items-center justify-between gap-2 text-xs'>
            <span className='text-muted-foreground'>{t('Required')}</span>
            <span>{formatQuota(balanceCost)}</span>
          </div>
          <div className='flex items-center justify-between gap-2 text-xs'>
            <span className='text-muted-foreground'>{t('Available')}</span>
            <span>{formatQuota(userQuota)}</span>
          </div>
          {insufficientBalance && (
            <Alert variant='destructive'>
              <AlertDescription className='flex flex-col gap-2'>
                <span>
                  {t(
                    'Insufficient balance. Please recharge your balance first.'
                  )}
                </span>
                <Button
                  type='button'
                  variant='secondary'
                  size='sm'
                  className='self-start'
                  onClick={() => {
                    handleOpenChange(false)
                    props.onGoTopup?.()
                  }}
                >
                  {t('Go to top up')}
                </Button>
              </AlertDescription>
            </Alert>
          )}
          <Button
            onClick={handlePayBalance}
            disabled={paying || insufficientBalance || limitReached}
          >
            {paying && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {t('Pay with Balance')}
          </Button>
        </div>
      </div>
    </Dialog>
  )
}
