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
import { Loader2, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  formatBillingCurrencyFromUSD,
  formatLocalCurrencyAmount,
} from '@/lib/currency'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { usePayment } from '../hooks/use-payment'
import {
  getDefaultPaymentType,
  getDiscountLabel,
  getMinTopupAmount,
  mergePresetAmounts,
} from '../lib'
import type { PaymentMethod, PresetAmount, TopupInfo } from '../types'
import { PaymentConfirmDialog } from './dialogs/payment-confirm-dialog'

const XUNHU_PAYMENT_TYPES = ['xunhu-wechat', 'xunhu-alipay']

interface PayAsYouGoTabProps {
  topupInfo: TopupInfo | null
  onPaid?: () => void | Promise<void>
}

export function PayAsYouGoTab({ topupInfo, onPaid }: PayAsYouGoTabProps) {
  const { t } = useTranslation()

  const min = getMinTopupAmount(topupInfo)
  const presets = useMemo<PresetAmount[]>(
    () =>
      mergePresetAmounts(
        topupInfo?.amount_options ?? [],
        topupInfo?.discount ?? {}
      ),
    [topupInfo]
  )
  const methods: PaymentMethod[] = useMemo(
    () => topupInfo?.pay_methods ?? [],
    [topupInfo]
  )

  const [manualAmount, setManualAmount] = useState<number>(min)
  const [selectedPresetAmount, setSelectedPresetAmount] = useState<
    number | null
  >(null)
  const [preferredMethod, setPreferredMethod] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const {
    amount: payPrice,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
  } = usePayment()

  const amount = selectedPresetAmount ?? (manualAmount > 0 ? manualAmount : min)
  const defaultMethod = useMemo(
    () => getDefaultPaymentType(topupInfo),
    [topupInfo]
  )
  const method =
    preferredMethod &&
    (methods.length === 0 || methods.some((m) => m.type === preferredMethod))
      ? preferredMethod
      : defaultMethod

  // Debounced server-side quote — rapid typing must not spray requests.
  useEffect(() => {
    if (!(amount > 0 && amount >= min && method)) return
    const timer = setTimeout(() => {
      void calculatePaymentAmount(amount, method)
    }, 300)
    return () => clearTimeout(timer)
  }, [amount, method, min, calculatePaymentAmount])

  const hasChannel =
    methods.length > 0 ||
    Boolean(topupInfo?.enable_online_topup) ||
    Boolean(topupInfo?.enable_xunhu_topup)

  const discountRate = useMemo(() => {
    const presetDiscount = presets.find((p) => p.value === amount)?.discount
    const configuredDiscount = topupInfo?.discount?.[amount]
    const discount = presetDiscount ?? configuredDiscount ?? 1
    return discount > 0 ? discount : 1
  }, [amount, presets, topupInfo?.discount])
  const hasDiscount = discountRate > 0 && discountRate < 1
  const selectedMethod =
    methods.find((m) => m.type === method) ||
    ({
      name: method || t('Payment method'),
      type: method,
    } satisfies PaymentMethod)
  const quotedPaymentAmount = payPrice > 0 ? payPrice : null
  const originalPaymentAmount =
    hasDiscount && quotedPaymentAmount !== null
      ? quotedPaymentAmount / discountRate
      : quotedPaymentAmount
  const discountAmount =
    originalPaymentAmount !== null && quotedPaymentAmount !== null
      ? Math.max(0, originalPaymentAmount - quotedPaymentAmount)
      : null

  const valid =
    amount > 0 &&
    amount >= min &&
    Boolean(method) &&
    quotedPaymentAmount !== null &&
    !calculating

  // 原始(42e786c8): 数字在前 xunhu_fund_symbol在后
  // 修改: xunhu_fund_symbol在前 数字在后
  const formatPaymentAmount = (value: number) => {
    if (XUNHU_PAYMENT_TYPES.includes(method) && topupInfo?.xunhu_fund_symbol) {
      return `${topupInfo.xunhu_fund_symbol}${value.toFixed(2)}`
    }
    return formatLocalCurrencyAmount(value, {
      digitsLarge: 2,
      digitsSmall: 2,
      abbreviate: false,
    })
  }

  if (!hasChannel) {
    return (
      <Card className='glass-card rounded-xl py-0'>
        <CardContent className='p-6'>
          <p className='text-muted-foreground text-center text-sm'>
            {t('Pay-as-you-go top-up is not enabled yet.')}
          </p>
        </CardContent>
      </Card>
    )
  }

  const handlePay = async () => {
    if (!valid || processing) return
    const ok = await processPayment(amount, method)
    if (ok) {
      setConfirmOpen(false)
      await onPaid?.()
    }
  }

  return (
    <>
      <Card className='glass-card rounded-xl py-0'>
        <CardContent className='space-y-5 p-4 sm:p-5'>
          {/* Preset amounts */}
          <div className='space-y-2'>
            <div className='flex flex-wrap items-end justify-between gap-2'>
              <div>
                <p className='text-sm font-medium'>{t('Top-up tiers')}</p>
                <p className='text-muted-foreground text-xs'>
                  {t('Pick a higher tier to unlock a better rate.')}
                </p>
              </div>
            </div>
            <div className='grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4'>
              {presets.map((p) => {
                const active = selectedPresetAmount === p.value
                const discount = p.discount && p.discount > 0 ? p.discount : 1
                const discountedValue = p.value * discount
                const cardHasDiscount = discount < 1
                return (
                  <button
                    key={p.value}
                    type='button'
                    onClick={() => {
                      setSelectedPresetAmount(p.value)
                    }}
                    className={cn(
                      'relative flex min-h-24 flex-col rounded-xl border p-3 text-sm transition-colors',
                      active
                        ? 'border-primary bg-primary/10 text-primary shadow-sm'
                        : 'hover:border-primary/40 hover:bg-muted/50',
                      cardHasDiscount
                        ? 'items-start justify-between text-left'
                        : 'items-center justify-center text-center'
                    )}
                  >
                    {cardHasDiscount ? (
                      <>
                        <div className='flex w-full items-start justify-between gap-2'>
                          <span className='text-muted-foreground text-xs line-through'>
                            {formatBillingCurrencyFromUSD(p.value, {
                              digitsLarge: 2,
                              digitsSmall: 2,
                              abbreviate: false,
                            })}
                          </span>
                          <span className='text-accent-coral bg-accent-coral/10 rounded-full px-2 py-0.5 text-[11px] font-bold'>
                            {getDiscountLabel(discount)}
                          </span>
                        </div>
                        <div>
                          <div className='text-xl font-bold tabular-nums'>
                            {formatBillingCurrencyFromUSD(discountedValue, {
                              digitsLarge: 2,
                              digitsSmall: 2,
                              abbreviate: false,
                            })}
                          </div>
                          <div className='text-muted-foreground mt-1 text-xs'>
                            {t('Credit')}{' '}
                            {formatBillingCurrencyFromUSD(p.value, {
                              digitsLarge: 2,
                              digitsSmall: 2,
                              abbreviate: false,
                            })}
                          </div>
                        </div>
                      </>
                    ) : (
                      <span className='text-lg font-semibold tabular-nums'>
                        {formatBillingCurrencyFromUSD(p.value, {
                          digitsLarge: 2,
                          digitsSmall: 2,
                          abbreviate: false,
                        })}
                      </span>
                    )}
                  </button>
                )
              })}
            </div>
          </div>

          {/* Custom amount */}
          <div className='space-y-2'>
            <p className='text-sm font-medium'>
              {t('Or enter a custom amount')}
            </p>
            <div className='relative max-w-xs'>
              <span className='text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2 text-sm'>
                {XUNHU_PAYMENT_TYPES.includes(method) && topupInfo?.xunhu_fund_symbol
                  ? topupInfo.xunhu_fund_symbol
                  : '$'}
              </span>
              <Input
                type='number'
                min={min}
                step={1}
                value={manualAmount || ''}
                onChange={(e) => {
                  setSelectedPresetAmount(null)
                  // Whole units only — the backend floors the order amount, so
                  // the quoted price must match what actually gets charged.
                  setManualAmount(
                    Math.max(0, Math.floor(Number(e.target.value) || 0))
                  )
                }}
                className='h-10 pl-7'
                placeholder={String(min)}
              />
            </div>
            <p className='text-muted-foreground text-xs'>
              {t('Minimum {{amount}}', {
                amount: formatBillingCurrencyFromUSD(min, {
                  digitsLarge: 2,
                  digitsSmall: 2,
                  abbreviate: false,
                }),
              })}
            </p>
          </div>

          {/* Payment methods */}
          {methods.length > 0 ? (
            <div className='space-y-2'>
              <p className='text-sm font-medium'>{t('Payment method')}</p>
              <div className='flex flex-wrap gap-2'>
                {methods.map((m) => {
                  const active = method === m.type
                  return (
                    <button
                      key={m.type}
                      type='button'
                      onClick={() => setPreferredMethod(m.type)}
                      className={cn(
                        'flex items-center gap-2 rounded-lg border px-3 py-2 text-sm font-medium transition-colors',
                        active
                          ? 'border-primary bg-primary/10 text-primary'
                          : 'hover:border-primary/40 hover:bg-muted/50'
                      )}
                    >
                      {m.icon ? (
                        <img src={m.icon} alt='' className='size-4' />
                      ) : null}
                      {m.name}
                    </button>
                  )
                })}
              </div>
            </div>
          ) : null}

          {/* Summary + pay */}
          <div className='flex flex-wrap items-center justify-between gap-3 border-t pt-4'>
            <div className='grid gap-1 text-sm'>
              <div className='flex items-center gap-2'>
                <span className='text-muted-foreground'>
                  {t('Topup Amount')}
                </span>
                <span className='font-medium'>
                  {formatBillingCurrencyFromUSD(amount, {
                    digitsLarge: 2,
                    digitsSmall: 2,
                    abbreviate: false,
                  })}
                </span>
              </div>
              <div className='flex items-center gap-2'>
                <span className='text-muted-foreground'>
                  {t('Original price')}
                </span>
                {calculating ? (
                  <Loader2 className='size-4 animate-spin' />
                ) : hasDiscount && originalPaymentAmount !== null ? (
                  <span className='text-muted-foreground line-through'>
                    {formatPaymentAmount(originalPaymentAmount)}
                  </span>
                ) : quotedPaymentAmount !== null ? (
                  <span className='font-medium'>
                    {formatPaymentAmount(quotedPaymentAmount)}
                  </span>
                ) : (
                  <span className='text-muted-foreground'>--</span>
                )}
              </div>
              {hasDiscount ? (
                <div className='flex items-center gap-2'>
                  <span className='text-muted-foreground'>{t('Discount')}</span>
                  {calculating ? (
                    <Loader2 className='size-4 animate-spin' />
                  ) : discountAmount !== null ? (
                    <span className='text-accent-coral font-semibold'>
                      -
                      {formatPaymentAmount(discountAmount)}
                    </span>
                  ) : (
                    <span className='text-muted-foreground'>--</span>
                  )}
                </div>
              ) : null}
              <div className='flex items-baseline gap-2'>
                <span className='text-muted-foreground text-xs'>
                  {t('Amount due')}
                </span>
                <span className='text-2xl font-semibold tabular-nums'>
                  {calculating ? (
                    <Loader2 className='size-5 animate-spin' />
                  ) : quotedPaymentAmount !== null ? (
                    formatPaymentAmount(quotedPaymentAmount)
                  ) : (
                    '--'
                  )}
                </span>
              </div>
            </div>
            <Button
              size='lg'
              disabled={!valid || processing}
              onClick={() => setConfirmOpen(true)}
              className='gap-2'
            >
              {processing ? (
                <Loader2 className='size-4 animate-spin' />
              ) : (
                <WalletCards className='size-4' />
              )}
              {t('Pay now')}
            </Button>
          </div>
        </CardContent>
      </Card>

      <PaymentConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        onConfirm={handlePay}
        topupAmount={amount}
        paymentAmount={quotedPaymentAmount ?? 0}
        paymentMethod={selectedMethod}
        calculating={calculating}
        processing={processing}
        discountRate={discountRate}
        xunhuFundSymbol={topupInfo?.xunhu_fund_symbol}
      />
    </>
  )
}
