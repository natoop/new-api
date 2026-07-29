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
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { AnimateInView } from '@/components/animate-in-view'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section className='relative z-10 overflow-hidden px-6 py-24 md:py-28'>
      {/* Gradient mesh background */}
      <div
        aria-hidden
        className='absolute inset-0 -z-10 opacity-20 dark:opacity-[0.08]'
        style={{
          background: [
            'radial-gradient(ellipse 50% 50% at 30% 50%, color-mix(in oklch, var(--glow-primary) 70%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 70% 40%, color-mix(in oklch, var(--glow-accent) 50%, transparent) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <AnimateInView
        className='mx-auto max-w-2xl text-center'
        animation='scale-in'
      >
        <h2 className='text-3xl leading-[1.08] font-bold tracking-[-0.02em] md:text-5xl'>
          {t('Run on a gateway')}
          <br />
          <span className='bg-gradient-to-r from-sky-500 via-cyan-500 to-emerald-500 bg-clip-text text-transparent'>
            {t('that stays up and stays fast')}
          </span>
        </h2>
        <p className='text-muted-foreground/80 mx-auto mt-5 max-w-md text-[15px] leading-[1.7] tracking-[-0.002em] md:text-base'>
          {t(
            'Start free — multi-channel auto-failover, continuous health checks, and one unified OpenAI-compatible API, included.'
          )}
        </p>
        <div className='mt-8 flex items-center justify-center gap-3'>
          <Button className='group rounded-lg' render={<Link to='/sign-up' />}>
            {t('Get Started')}
            <ArrowRight className='ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5' />
          </Button>
          <Button
            variant='outline'
            className='border-border/50 hover:border-border hover:bg-muted/50 rounded-lg'
            render={<Link to='/pricing' />}
          >
            {t('View Pricing')}
          </Button>
        </div>
      </AnimateInView>
    </section>
  )
}
