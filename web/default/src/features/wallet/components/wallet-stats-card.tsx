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
import { Activity, BarChart3, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatQuota } from '@/lib/format'
import { Skeleton } from '@/components/ui/skeleton'
import type { UserWalletData } from '../types'

interface WalletStatsCardProps {
  user: UserWalletData | null
  loading?: boolean
}

export function WalletStatsCard(props: WalletStatsCardProps) {
  const { t } = useTranslation()
  if (props.loading) {
    return (
      <div className='glass-card overflow-hidden rounded-xl'>
        <div className='divide-border/60 grid grid-cols-3 divide-x'>
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className='px-3 py-3 sm:px-5 sm:py-4'>
              <Skeleton className='h-3.5 w-20' />
              <Skeleton className='mt-2 h-7 w-28' />
              <Skeleton className='mt-1.5 h-3.5 w-24' />
            </div>
          ))}
        </div>
      </div>
    )
  }

  // One Google-spirit accent per stat — decorative icon chips only.
  const stats = [
    {
      label: t('Current Balance'),
      value: formatQuota(props.user?.quota ?? 0),
      description: t('Remaining quota'),
      icon: WalletCards,
      chipClass: 'bg-accent-blue/12 text-accent-blue dark:bg-accent-blue/20',
    },
    {
      label: t('Total Usage'),
      value: formatQuota(props.user?.used_quota ?? 0),
      description: t('Total consumed quota'),
      icon: BarChart3,
      chipClass:
        'bg-accent-amber/15 text-accent-amber dark:bg-accent-amber/20',
    },
    {
      label: t('API Requests'),
      value: (props.user?.request_count ?? 0).toLocaleString(),
      description: t('Total requests made'),
      icon: Activity,
      chipClass:
        'bg-accent-green/12 text-accent-green dark:bg-accent-green/20',
    },
  ]

  return (
    <div className='glass-card overflow-hidden rounded-xl'>
      <div className='divide-border/60 grid grid-cols-3 divide-x'>
        {stats.map((item) => (
          <div key={item.label} className='px-3 py-3 sm:px-5 sm:py-4'>
            <div className='flex items-center gap-2'>
              <span
                className={`flex size-6 shrink-0 items-center justify-center rounded-md ${item.chipClass}`}
                aria-hidden='true'
              >
                <item.icon className='size-3.5' />
              </span>
              <div className='text-muted-foreground truncate text-xs font-medium tracking-wider uppercase'>
                {item.label}
              </div>
            </div>

            <div className='text-foreground mt-1.5 font-mono text-base font-bold tracking-tight break-all tabular-nums sm:mt-2 sm:text-2xl'>
              {item.value}
            </div>
            <div className='text-muted-foreground/60 mt-1 hidden text-xs md:block'>
              {item.description}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
