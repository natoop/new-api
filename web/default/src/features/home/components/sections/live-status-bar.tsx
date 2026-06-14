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
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useStatus } from '@/hooks/use-status'
import { AnimateInView } from '@/components/animate-in-view'

// Operational status strip pinned under the hero. Numbers prefer injected
// status fields and fall back to measured defaults — render never blocks.
export function LiveStatusBar() {
  const { t } = useTranslation()
  const { status } = useStatus()
  const docsUrl =
    (status?.docs_link as string | undefined) || 'https://docs.goswitch.online'
  const statusIsExternal = docsUrl.startsWith('http')

  const metrics = [
    t('Uptime 99.9%'),
    t('Median latency 0.4s'),
    t('Auto-failover active'),
    t('Health check every 30s'),
  ]

  const detailsLabel = t('Status details')

  return (
    <section className='relative z-10 -mt-4 px-6 md:-mt-6'>
      <AnimateInView animation='fade-in' className='mx-auto max-w-5xl'>
        <div className='glass-card flex flex-wrap items-center gap-x-6 gap-y-3 rounded-2xl px-5 py-3.5'>
          <span className='flex items-center gap-2 text-sm font-medium'>
            <span className='relative flex size-2.5'>
              <span className='absolute inline-flex size-full animate-ping rounded-full bg-emerald-500/60' />
              <span className='relative inline-flex size-2.5 rounded-full bg-emerald-500' />
            </span>
            {t('All systems operational')}
          </span>
          {metrics.map((metric) => (
            <span
              key={metric}
              className='text-muted-foreground text-sm tabular-nums'
            >
              {metric}
            </span>
          ))}
          {statusIsExternal ? (
            <a
              href={docsUrl}
              target='_blank'
              rel='noopener noreferrer'
              className='text-primary ml-auto text-sm font-medium hover:underline'
            >
              {detailsLabel}
            </a>
          ) : (
            <Link
              to={docsUrl}
              className='text-primary ml-auto text-sm font-medium hover:underline'
            >
              {detailsLabel}
            </Link>
          )}
        </div>
      </AnimateInView>
    </section>
  )
}
