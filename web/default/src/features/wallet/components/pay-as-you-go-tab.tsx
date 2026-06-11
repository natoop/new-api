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
import { useEffect, useMemo, useRef, useState } from 'react'
import { Loader2, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { usePayment } from '../hooks/use-payment'
import {
  formatCurrency,
  getDefaultPaymentType,
  getMinTopupAmount,
  mergePresetAmounts,
} from '../lib'
import type { PaymentMethod, PresetAmount, TopupInfo } from '../types'

const MONEY_PREFIX = '¥'

function money(n: number): string {
  return `${MONEY_PREFIX}${formatCurrency(n)}`
}

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

  const [amount, setAmount] = useState<number>(min)
  const [method, setMethod] = useState<string>(getDefaultPaymentType(topupInfo))
  // Until the user touches the amount, keep it pinned to the real minimum
  // that arrives asynchronously with topupInfo.
  const touchedRef = useRef(false)
  const {
    amount: payPrice,
    calculating,
    processing,
    calculatePaymentAmount,
    processPayment,
  } = usePayment()

  useEffect(() => {
    setMethod(getDefaultPaymentType(topupInfo))
    if (!touchedRef.current) {
      setAmount(getMinTopupAmount(topupInfo))
    }
  }, [topupInfo])

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

  const valid = amount > 0 && amount >= min && Boolean(method)

  const handlePay = async () => {
    if (!valid || processing) return
    const ok = await processPayment(amount, method)
    if (ok) await onPaid?.()
  }

  return (
    <Card className='glass-card rounded-xl py-0'>
      <CardContent className='space-y-5 p-4 sm:p-5'>
        {/* Preset amounts */}
        <div className='space-y-2'>
          <p className='text-sm font-medium'>{t('Choose an amount')}</p>
          <div className='grid grid-cols-3 gap-2 sm:grid-cols-6'>
            {presets.map((p) => {
              const active = amount === p.value
              return (
                <button
                  key={p.value}
                  type='button'
                  onClick={() => {
                    touchedRef.current = true
                    setAmount(p.value)
                  }}
                  className={cn(
                    'flex h-16 flex-col items-center justify-center rounded-xl border text-sm font-semibold transition-colors',
                    active
                      ? 'border-primary bg-primary/10 text-primary'
                      : 'hover:border-primary/40 hover:bg-muted/50'
                  )}
                >
                  <span>{money(p.value)}</span>
                  {p.discount && p.discount < 1 ? (
                    <span className='text-accent-coral mt-0.5 text-[11px] font-medium'>
                      {Math.round((1 - p.discount) * 100)}% OFF
                    </span>
                  ) : null}
                </button>
              )
            })}
          </div>
        </div>

        {/* Custom amount */}
        <div className='space-y-2'>
          <p className='text-sm font-medium'>{t('Or enter a custom amount')}</p>
          <div className='relative max-w-xs'>
            <span className='text-muted-foreground absolute top-1/2 left-3 -translate-y-1/2 text-sm'>
              {MONEY_PREFIX}
            </span>
            <Input
              type='number'
              min={min}
              step={1}
              value={amount || ''}
              onChange={(e) => {
                touchedRef.current = true
                // Whole units only — the backend floors the order amount, so
                // the quoted price must match what actually gets charged.
                setAmount(Math.max(0, Math.floor(Number(e.target.value) || 0)))
              }}
              className='h-10 pl-7'
              placeholder={String(min)}
            />
          </div>
          <p className='text-muted-foreground text-xs'>
            {t('Minimum {{amount}}', { amount: money(min) })}
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
                    onClick={() => setMethod(m.type)}
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
          <div>
            <p className='text-muted-foreground text-xs'>{t('Amount due')}</p>
            <p className='text-2xl font-semibold tabular-nums'>
              {calculating ? (
                <Loader2 className='size-5 animate-spin' />
              ) : (
                money(payPrice || amount)
              )}
            </p>
          </div>
          <Button
            size='lg'
            disabled={!valid || processing}
            onClick={handlePay}
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
  )
}
