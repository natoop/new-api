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
import {
  CheckCircle2,
  Crown,
  ExternalLink,
  Gift,
  Loader2,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useRedemption, type RedeemOutcome } from '../hooks/use-redemption'

// ============================================================================
// Shared redemption section (used by the top quick-redeem card and the
// "Redemption Code" tab inside WalletTabs)
// ============================================================================

interface RedemptionSectionProps {
  /** Whether redemption is enabled (topupInfo.enable_redemption !== false) */
  enabled?: boolean
  /** Optional external link to obtain redemption codes */
  topupLink?: string
  /** Called after a successful redemption (refresh user/subscriptions) */
  onRedeemed?: (outcome: RedeemOutcome) => void | Promise<void>
}

export function RedemptionSection({
  enabled = true,
  topupLink,
  onRedeemed,
}: RedemptionSectionProps) {
  const { t } = useTranslation()
  const [code, setCode] = useState('')
  const [outcome, setOutcome] = useState<RedeemOutcome | null>(null)
  const { redeeming, redeemCode } = useRedemption()

  const handleRedeem = async () => {
    if (redeeming) return
    const result = await redeemCode(code)
    if (result) {
      setOutcome(result)
      setCode('')
      await onRedeemed?.(result)
    }
  }

  if (!enabled) {
    return (
      <Alert>
        <AlertDescription>
          {t(
            'Redemption codes are disabled until the administrator confirms compliance terms.'
          )}
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <div className='space-y-2.5'>
      <div className='grid grid-cols-[minmax(0,1fr)_auto] gap-2'>
        <Input
          value={code}
          onChange={(e) => {
            setCode(e.target.value)
            if (outcome) setOutcome(null)
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter') void handleRedeem()
          }}
          placeholder={t('Enter your redemption code')}
          className='h-9 min-w-0'
        />
        <Button
          onClick={handleRedeem}
          disabled={redeeming || !code.trim()}
          className='h-9 px-5'
        >
          {redeeming && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
          {t('Redeem')}
        </Button>
      </div>

      {outcome && (
        <div className='border-primary/20 bg-primary/5 flex items-center gap-2.5 rounded-xl border px-3.5 py-2.5 text-sm'>
          {outcome.type === 'plan' ? (
            <Crown className='text-primary h-4 w-4 shrink-0' />
          ) : (
            <CheckCircle2 className='text-primary h-4 w-4 shrink-0' />
          )}
          <span className='min-w-0 font-medium'>
            {outcome.type === 'plan'
              ? t('Plan “{{title}}” activated', {
                  title: outcome.planTitle || '',
                })
              : t('Balance credited +{{amount}}', {
                  amount: formatQuota(outcome.quota),
                })}
          </span>
        </div>
      )}

      {topupLink && (
        <p className='text-muted-foreground text-xs'>
          {t('Need a redemption code?')}{' '}
          <a
            href={topupLink}
            target='_blank'
            rel='noopener noreferrer'
            className='inline-flex items-center gap-1 underline-offset-4 hover:underline'
          >
            {t('Get one here')}
            <ExternalLink className='h-3 w-3' />
          </a>
        </p>
      )}
    </div>
  )
}

// ============================================================================
// Quick-redeem card (pinned below wallet stats)
// ============================================================================

export function RedemptionCard(props: RedemptionSectionProps) {
  const { t } = useTranslation()

  return (
    <Card className='from-primary/[0.05] via-card to-card rounded-xl bg-gradient-to-r py-0'>
      <CardContent className='flex flex-col gap-3 p-4 sm:flex-row sm:items-start sm:gap-6 sm:p-5'>
        <div className='flex items-center gap-3 sm:w-60 sm:shrink-0 sm:pt-0.5'>
          <div className='bg-primary/10 text-primary flex h-9 w-9 shrink-0 items-center justify-center rounded-lg'>
            <Gift className='h-4 w-4' />
          </div>
          <div className='min-w-0'>
            <h3 className='text-sm font-semibold'>{t('Have a Code?')}</h3>
            <p className='text-muted-foreground text-xs'>
              {t('Redeem instantly to your account')}
            </p>
          </div>
        </div>
        <div className='min-w-0 flex-1'>
          <RedemptionSection {...props} />
        </div>
      </CardContent>
    </Card>
  )
}
