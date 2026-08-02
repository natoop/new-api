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
import { Check, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { formatPlanAmount } from '../lib'
import type { RechargePlan } from '../types'

interface RechargeTierCardProps {
  plan: RechargePlan
  /** Quota credited to the wallet balance on purchase. */
  creditedQuota: number
  /** Pre-computed discount label, empty when the tier has no discount. */
  discountLabel: string
  recommended: boolean
  onSelect: () => void
}

export function RechargeTierCard({
  plan,
  creditedQuota,
  discountLabel,
  recommended,
  onSelect,
}: RechargeTierCardProps) {
  const { t } = useTranslation()

  const benefits = [
    creditedQuota > 0
      ? `${t('Credited Quota')}: ${formatQuota(creditedQuota)}`
      : null,
    t('Never expires'),
    plan.upgrade_group ? `${t('Upgrade Group')}: ${plan.upgrade_group}` : null,
  ].filter(Boolean) as string[]

  return (
    <Card
      className={cn(
        'relative rounded-xl py-0 transition-shadow hover:shadow-sm',
        recommended &&
          'border-primary/60 from-primary/[0.05] bg-gradient-to-b to-transparent shadow-sm'
      )}
    >
      {recommended && (
        <Badge className='absolute -top-2.5 left-1/2 -translate-x-1/2 gap-1 shadow-sm'>
          <Sparkles className='h-3 w-3' />
          {t('Recommended')}
        </Badge>
      )}
      <CardContent className='flex h-full flex-col p-4 sm:p-5'>
        <div className='flex min-w-0 items-start justify-between gap-2'>
          <div className='min-w-0'>
            <h4 className='truncate text-base font-semibold'>
              {plan.title || t('Recharge Tiers')}
            </h4>
            <p className='text-muted-foreground mt-0.5 line-clamp-2 min-h-4 text-xs'>
              {plan.subtitle || ''}
            </p>
          </div>
          {discountLabel && (
            <Badge variant='secondary' className='shrink-0'>
              {discountLabel}
            </Badge>
          )}
        </div>

        <div className='py-3'>
          <span
            className={cn(
              'text-3xl font-bold tracking-tight',
              recommended && 'text-primary'
            )}
          >
            {formatPlanAmount(Number(plan.price_amount || 0), plan.currency)}
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

        <Button
          variant={recommended ? 'default' : 'outline'}
          className='w-full'
          onClick={onSelect}
        >
          {t('Recharge Now')}
        </Button>
      </CardContent>
    </Card>
  )
}
