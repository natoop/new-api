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
import { useEffect, useMemo, useState } from 'react'
import {
  ArrowLeft,
  BadgePercent,
  CalendarClock,
  Crown,
  ExternalLink,
  Loader2,
  Package,
  QrCode,
  X,
} from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { DEFAULT_CURRENCY_CONFIG } from '@/stores/system-config-store'
import { formatQuota } from '@/lib/format'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Dialog } from '@/components/dialog'
import { GroupBadge } from '@/components/group-badge'
import {
  getSelfSubscriptionFull,
  paySubscriptionBalance,
  paySubscriptionCreem,
  paySubscriptionEpay,
  paySubscriptionStripe,
  paySubscriptionWaffoPancake,
  paySubscriptionXunhu,
  validateSubscriptionPromo,
} from '@/features/subscriptions/api'
import { formatDuration, formatResetPeriod } from '@/features/subscriptions/lib'
import type {
  PlanRecord,
  SubscriptionPromoValidation,
} from '@/features/subscriptions/types'
import { formatPlanAmount, submitPaymentForm } from '../../lib'
import type { TopupInfo } from '../../types'

interface XunhuOrder {
  payType: 'wechat' | 'alipay'
  payUrl?: string
  qrUrl?: string
  orderNo?: string
}

interface PlanPurchaseDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  planRecord: PlanRecord | null
  topupInfo: TopupInfo | null
  userQuota?: number
  purchaseLimit?: number
  purchaseCount?: number
  onPurchaseSuccess?: () => void | Promise<void>
}

function formatDiscountBps(discountBps: number): string {
  const percent = discountBps / 100
  return `-${Number.isInteger(percent) ? percent : percent.toFixed(1)}%`
}

export function PlanPurchaseDialog(props: PlanPurchaseDialogProps) {
  const { t } = useTranslation()
  const { currency } = useSystemConfig()

  const [paying, setPaying] = useState<string | null>(null)
  const [promoInput, setPromoInput] = useState('')
  const [promoChecking, setPromoChecking] = useState(false)
  const [promo, setPromo] = useState<SubscriptionPromoValidation | null>(null)
  const [xunhuOrder, setXunhuOrder] = useState<XunhuOrder | null>(null)
  const [selectedEpayMethod, setSelectedEpayMethod] = useState('')

  const epayMethods = useMemo(
    () =>
      (props.topupInfo?.pay_methods ?? []).filter(
        (m) =>
          m?.type &&
          m.type !== 'stripe' &&
          m.type !== 'creem' &&
          m.type !== 'waffo' &&
          m.type !== 'waffo_pancake'
      ),
    [props.topupInfo?.pay_methods]
  )

  // Reset transient state whenever the dialog closes / plan changes
  useEffect(() => {
    if (!props.open) {
      setPaying(null)
      setPromoInput('')
      setPromo(null)
      setXunhuOrder(null)
      setSelectedEpayMethod('')
    } else if (epayMethods.length > 0) {
      setSelectedEpayMethod(epayMethods[0].type)
    }
  }, [props.open, epayMethods])

  // Lightweight polling while a Xunhu order is pending: when a new
  // subscription shows up, treat the payment as completed.
  useEffect(() => {
    if (!xunhuOrder || !props.open) return

    let stopped = false
    let baseline: number | null = null

    const tick = async () => {
      try {
        const res = await getSelfSubscriptionFull()
        if (stopped || !res.success || !res.data) return
        const count = (res.data.all_subscriptions || []).length
        if (baseline === null) {
          baseline = count
          return
        }
        if (count > baseline) {
          stopped = true
          toast.success(t('Subscription purchased successfully'))
          void props.onPurchaseSuccess?.()
          props.onOpenChange(false)
        }
      } catch {
        // ignore polling errors
      }
    }

    void tick()
    const timer = window.setInterval(tick, 5000)
    return () => {
      stopped = true
      window.clearInterval(timer)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [xunhuOrder, props.open])

  const plan = props.planRecord?.plan
  if (!plan) return null

  const enableXunhu = !!props.topupInfo?.enable_xunhu_topup
  const hasStripe = !!props.topupInfo?.enable_stripe_topup && !!plan.stripe_price_id
  const hasCreem = !!props.topupInfo?.enable_creem_topup && !!plan.creem_product_id
  const hasWaffoPancake =
    !!props.topupInfo?.enable_waffo_pancake_topup &&
    !!plan.waffo_pancake_product_id
  const hasEpay = !!props.topupInfo?.enable_online_topup && epayMethods.length > 0

  const originalAmount = Number(plan.price_amount || 0)
  const finalAmount = promo ? Number(promo.final_amount) : originalAmount
  const promoCode = promo?.code

  const quotaPerUnit =
    currency?.quotaPerUnit && currency.quotaPerUnit > 0
      ? currency.quotaPerUnit
      : DEFAULT_CURRENCY_CONFIG.quotaPerUnit
  const balanceCost = Math.max(0, Math.ceil(finalAmount * quotaPerUnit))
  const userQuota = Math.max(0, Number(props.userQuota || 0))
  const allowBalancePay = plan.allow_balance_pay !== false
  const insufficientBalance = userQuota < balanceCost
  const limitReached =
    (props.purchaseLimit || 0) > 0 &&
    (props.purchaseCount || 0) >= (props.purchaseLimit || 0)
  const totalAmount = Number(plan.total_amount || 0)

  const handleApplyPromo = async () => {
    const code = promoInput.trim()
    if (!code || promoChecking) return
    setPromoChecking(true)
    try {
      const res = await validateSubscriptionPromo({ code, plan_id: plan.id })
      if (res.success && res.data) {
        setPromo(res.data)
        toast.success(t('Promo code applied'))
      } else {
        toast.error(res.message || t('Invalid promo code'))
      }
    } catch {
      toast.error(t('Request failed'))
    } finally {
      setPromoChecking(false)
    }
  }

  const handleRemovePromo = () => {
    setPromo(null)
    setPromoInput('')
  }

  const handlePayBalance = async () => {
    setPaying('balance')
    try {
      const res = await paySubscriptionBalance({
        plan_id: plan.id,
        promo_code: promoCode,
      })
      if (res.success) {
        toast.success(t('Subscription purchased successfully'))
        void props.onPurchaseSuccess?.()
        props.onOpenChange(false)
      }
    } catch {
      toast.error(t('Payment request failed'))
    } finally {
      setPaying(null)
    }
  }

  const handlePayXunhu = async (payType: 'wechat' | 'alipay') => {
    setPaying(payType)
    try {
      const res = await paySubscriptionXunhu({
        plan_id: plan.id,
        pay_type: payType,
        promo_code: promoCode,
      })
      if (res.success && res.data && (res.data.pay_url || res.data.qr_url)) {
        setXunhuOrder({
          payType,
          payUrl: res.data.pay_url,
          qrUrl: res.data.qr_url,
          orderNo: res.data.order_no,
        })
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
      setPaying(null)
    }
  }

  const handlePayStripe = async () => {
    setPaying('stripe')
    try {
      const res = await paySubscriptionStripe({
        plan_id: plan.id,
        promo_code: promoCode,
      })
      if (res.message === 'success' && res.data?.pay_link) {
        window.open(res.data.pay_link, '_blank')
        toast.success(t('Payment page opened'))
        props.onOpenChange(false)
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
      setPaying(null)
    }
  }

  const handlePayCreem = async () => {
    setPaying('creem')
    try {
      const res = await paySubscriptionCreem({ plan_id: plan.id })
      if (res.message === 'success' && res.data?.checkout_url) {
        window.open(res.data.checkout_url, '_blank')
        toast.success(t('Payment page opened'))
        props.onOpenChange(false)
      }
    } catch {
      toast.error(t('Payment request failed'))
    } finally {
      setPaying(null)
    }
  }

  // In-tab redirect — user-gesture context is lost across the await, so a
  // popup would be blocked (same rationale as the wallet topup hook).
  const handlePayWaffoPancake = async () => {
    setPaying('waffo_pancake')
    try {
      const res = await paySubscriptionWaffoPancake({ plan_id: plan.id })
      if (res.message === 'success' && res.data?.checkout_url) {
        toast.success(t('Redirecting to payment page...'))
        window.location.href = res.data.checkout_url
      }
    } catch {
      toast.error(t('Payment request failed'))
    } finally {
      setPaying(null)
    }
  }

  const handlePayEpay = async () => {
    if (!selectedEpayMethod) {
      toast.error(t('Please select a payment method'))
      return
    }
    setPaying('epay')
    try {
      const res = await paySubscriptionEpay({
        plan_id: plan.id,
        payment_method: selectedEpayMethod,
        promo_code: promoCode,
      })
      if (res.message === 'success' && res.url) {
        submitPaymentForm(res.url, (res.data as Record<string, unknown>) || {})
        toast.success(t('Payment initiated'))
        props.onOpenChange(false)
      }
    } catch {
      toast.error(t('Payment request failed'))
    } finally {
      setPaying(null)
    }
  }

  const selectedEpayMethodLabel =
    epayMethods.find((m) => m.type === selectedEpayMethod)?.name ||
    selectedEpayMethod ||
    t('Select payment method')

  const payDisabled = !!paying || limitReached

  return (
    <Dialog
      open={props.open}
      onOpenChange={props.onOpenChange}
      title={
        <>
          <Crown className='h-5 w-5' />
          {t('Purchase Subscription')}
        </>
      }
      contentClassName='max-sm:w-[calc(100vw-1.5rem)] sm:max-w-md'
      titleClassName='flex items-center gap-2'
      contentHeight='auto'
      bodyClassName='space-y-4'
    >
      {xunhuOrder ? (
        // ------------------------------------------------------------------
        // Xunhu payment view: QR code + cashier link + completion hint
        // ------------------------------------------------------------------
        <div className='space-y-4'>
          <div className='bg-muted/50 flex flex-col items-center gap-3 rounded-xl border p-4 sm:p-5'>
            <div className='text-sm font-medium'>
              {xunhuOrder.payType === 'wechat' ? t('WeChat Pay') : t('Alipay')}{' '}
              ·{' '}
              <span className='text-primary font-bold'>
                {formatPlanAmount(finalAmount, plan.currency)}
              </span>
            </div>
            {xunhuOrder.qrUrl ? (
              <>
                <div className='rounded-lg bg-white p-3 shadow-sm'>
                  <QRCodeSVG value={xunhuOrder.qrUrl} size={176} />
                </div>
                <p className='text-muted-foreground flex items-center gap-1.5 text-xs'>
                  <QrCode className='h-3.5 w-3.5' />
                  {t('Scan the QR code to pay')}
                </p>
              </>
            ) : (
              <p className='text-muted-foreground text-xs'>
                {t('Complete the payment in the checkout page.')}
              </p>
            )}
            {xunhuOrder.payUrl && (
              <Button
                variant='outline'
                className='w-full gap-2'
                onClick={() => window.open(xunhuOrder.payUrl, '_blank')}
              >
                <ExternalLink className='h-4 w-4' />
                {t('Open Checkout')}
              </Button>
            )}
            {xunhuOrder.orderNo && (
              <p className='text-muted-foreground/70 text-[11px]'>
                {t('Order Number')}: {xunhuOrder.orderNo}
              </p>
            )}
          </div>

          <Alert>
            <Loader2 className='h-4 w-4 animate-spin' />
            <AlertDescription>
              {t(
                'Waiting for payment confirmation... Your subscription will be activated automatically once the payment completes.'
              )}
            </AlertDescription>
          </Alert>

          <div className='flex gap-2'>
            <Button
              variant='ghost'
              className='gap-1.5'
              onClick={() => setXunhuOrder(null)}
            >
              <ArrowLeft className='h-4 w-4' />
              {t('Back')}
            </Button>
            <Button
              className='flex-1'
              onClick={() => {
                void props.onPurchaseSuccess?.()
                props.onOpenChange(false)
              }}
            >
              {t('I have completed payment')}
            </Button>
          </div>
        </div>
      ) : (
        // ------------------------------------------------------------------
        // Order form view: summary + promo code + payment methods
        // ------------------------------------------------------------------
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
                {promo && (
                  <>
                    <span className='text-muted-foreground text-sm line-through'>
                      {formatPlanAmount(originalAmount, plan.currency)}
                    </span>
                    <Badge variant='secondary' className='gap-0.5'>
                      <BadgePercent className='h-3 w-3' />
                      {formatDiscountBps(promo.discount_bps)}
                    </Badge>
                  </>
                )}
                <span className='text-primary text-lg font-bold'>
                  {formatPlanAmount(finalAmount, plan.currency)}
                </span>
              </span>
            </div>
          </div>

          {/* Promo code */}
          <div className='space-y-2 rounded-xl border p-3'>
            <Label className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
              {t('Promo Code')}
            </Label>
            {promo ? (
              <div className='border-primary/20 bg-primary/5 flex items-center justify-between gap-2 rounded-lg border px-3 py-2'>
                <span className='flex min-w-0 items-center gap-2 text-sm'>
                  <BadgePercent className='text-primary h-4 w-4 shrink-0' />
                  <span className='truncate font-mono font-medium'>
                    {promo.code}
                  </span>
                  <span className='text-primary text-xs font-semibold'>
                    {formatDiscountBps(promo.discount_bps)}
                  </span>
                </span>
                <Button
                  variant='ghost'
                  size='icon'
                  className='h-7 w-7 shrink-0'
                  onClick={handleRemovePromo}
                  aria-label={t('Remove')}
                >
                  <X className='h-3.5 w-3.5' />
                </Button>
              </div>
            ) : (
              <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
                <Input
                  value={promoInput}
                  onChange={(e) => setPromoInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') void handleApplyPromo()
                  }}
                  placeholder={t('Enter promo code (optional)')}
                  className='h-9 min-w-0'
                />
                <Button
                  variant='outline'
                  className='h-9'
                  onClick={handleApplyPromo}
                  disabled={promoChecking || !promoInput.trim()}
                >
                  {promoChecking && (
                    <Loader2 className='mr-1.5 h-3.5 w-3.5 animate-spin' />
                  )}
                  {t('Apply')}
                </Button>
              </div>
            )}
          </div>

          {limitReached && (
            <Alert variant='destructive'>
              <AlertDescription>
                {t('Purchase limit reached')} ({props.purchaseCount}/
                {props.purchaseLimit})
              </AlertDescription>
            </Alert>
          )}

          {/* Balance payment */}
          <div className='flex flex-col gap-2 rounded-xl border p-3'>
            <div className='flex items-center justify-between gap-2 text-xs'>
              <span className='text-muted-foreground'>{t('Required')}</span>
              <span>{formatQuota(balanceCost)}</span>
            </div>
            <div className='flex items-center justify-between gap-2 text-xs'>
              <span className='text-muted-foreground'>{t('Available')}</span>
              <span>{formatQuota(userQuota)}</span>
            </div>
            {!allowBalancePay ? (
              <Alert variant='destructive'>
                <AlertDescription>
                  {t('This plan does not allow balance redemption')}
                </AlertDescription>
              </Alert>
            ) : (
              insufficientBalance && (
                <Alert variant='destructive'>
                  <AlertDescription>
                    {t('Insufficient balance')}
                  </AlertDescription>
                </Alert>
              )
            )}
            <Button
              onClick={handlePayBalance}
              disabled={payDisabled || !allowBalancePay || insufficientBalance}
            >
              {paying === 'balance' && (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              )}
              {t('Pay with Balance')}
            </Button>
          </div>

          {/* WeChat / Alipay (Xunhu) */}
          {enableXunhu && (
            <div className='grid grid-cols-2 gap-2'>
              <Button
                variant='outline'
                onClick={() => handlePayXunhu('wechat')}
                disabled={payDisabled}
              >
                {paying === 'wechat' && (
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                )}
                {t('WeChat Pay')}
              </Button>
              <Button
                variant='outline'
                onClick={() => handlePayXunhu('alipay')}
                disabled={payDisabled}
              >
                {paying === 'alipay' && (
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                )}
                {t('Alipay')}
              </Button>
            </div>
          )}

          {/* Other gateways (kept conditional on topup info) */}
          {(hasStripe || hasCreem || hasWaffoPancake || hasEpay) && (
            <div className='space-y-2'>
              <p className='text-muted-foreground text-xs'>
                {t('Other payment methods')}
              </p>
              {(hasStripe || hasCreem || hasWaffoPancake) && (
                <div className='grid grid-cols-2 gap-2 sm:flex'>
                  {hasStripe && (
                    <Button
                      variant='outline'
                      className='flex-1'
                      onClick={handlePayStripe}
                      disabled={payDisabled}
                    >
                      Stripe
                    </Button>
                  )}
                  {hasCreem && (
                    <Button
                      variant='outline'
                      className='flex-1'
                      onClick={handlePayCreem}
                      disabled={payDisabled || !!promo}
                    >
                      Creem
                    </Button>
                  )}
                  {hasWaffoPancake && (
                    <Button
                      variant='outline'
                      className='flex-1'
                      onClick={handlePayWaffoPancake}
                      disabled={payDisabled || !!promo}
                    >
                      Waffo Pancake
                    </Button>
                  )}
                </div>
              )}
              {promo && (hasCreem || hasWaffoPancake) && (
                <p className='text-muted-foreground text-[11px]'>
                  {t('Promo codes are not supported by Creem / Waffo Pancake.')}
                </p>
              )}
              {hasEpay && (
                <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
                  <Select
                    items={epayMethods.map((m) => ({
                      value: m.type,
                      label: m.name || m.type,
                    }))}
                    value={selectedEpayMethod}
                    onValueChange={(v) =>
                      v !== null && setSelectedEpayMethod(v)
                    }
                    disabled={limitReached}
                  >
                    <SelectTrigger className='flex-1'>
                      <SelectValue>{selectedEpayMethodLabel}</SelectValue>
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {epayMethods.map((m) => (
                          <SelectItem key={m.type} value={m.type}>
                            {m.name || m.type}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <Button
                    variant='outline'
                    onClick={handlePayEpay}
                    disabled={payDisabled || !selectedEpayMethod}
                  >
                    {paying === 'epay' && (
                      <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                    )}
                    {t('Pay')}
                  </Button>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </Dialog>
  )
}
