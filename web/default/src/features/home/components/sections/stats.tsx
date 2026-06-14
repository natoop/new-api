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
import { useRef, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { useStatus } from '@/hooks/use-status'

interface CounterProps {
  end: number
  suffix?: string
  prefix?: string
  duration?: number
  decimals?: number
}

function Counter(props: CounterProps) {
  const { end, suffix = '', prefix = '', duration = 1600, decimals = 0 } = props
  const ref = useRef<HTMLSpanElement>(null)
  const startedRef = useRef(false)

  const formatValue = useCallback(
    (v: number) =>
      decimals > 0 ? v.toFixed(decimals) : Math.round(v).toLocaleString(),
    [decimals]
  )

  const animate = useCallback(() => {
    const el = ref.current
    if (!el) return
    const start = performance.now()
    const step = (now: number) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      el.textContent = `${prefix}${formatValue(eased * end)}${suffix}`
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [end, duration, prefix, suffix, formatValue])

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mq.matches) {
      el.textContent = `${prefix}${formatValue(end)}${suffix}`
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !startedRef.current) {
          startedRef.current = true
          animate()
          observer.unobserve(el)
        }
      },
      { threshold: 0.5 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [animate, end, prefix, suffix, formatValue])

  return (
    <span ref={ref} className='tabular-nums'>
      {prefix}0{suffix}
    </span>
  )
}

interface StatsProps {
  className?: string
}

interface StatItem {
  end: number
  suffix: string
  label: string
  decimals?: number
}

export function Stats(_props: StatsProps) {
  const { t } = useTranslation()
  const { status } = useStatus()

  const uptime =
    typeof status?.measured_uptime_pct === 'number'
      ? (status.measured_uptime_pct as number)
      : 99.9
  const latencyMs =
    typeof status?.typical_latency_ms === 'number'
      ? (status.typical_latency_ms as number)
      : 400
  const channels =
    typeof status?.channel_count === 'number'
      ? (status.channel_count as number)
      : 50
  const models =
    typeof status?.model_count === 'number'
      ? (status.model_count as number)
      : 100

  const stats: StatItem[] = [
    { end: latencyMs, suffix: 'ms', label: t('typical first-token latency') },
    { end: uptime, suffix: '%', decimals: 1, label: t('measured uptime') },
    { end: channels, suffix: '+', label: t('channels with auto-failover') },
    { end: models, suffix: '+', label: t('models behind one endpoint') },
  ]

  return (
    <div className='border-border/40 bg-muted/10 relative z-10 border-y'>
      <div className='mx-auto max-w-6xl px-6 py-10 md:py-12'>
        <div className='grid grid-cols-2 gap-8 md:grid-cols-4 md:gap-12'>
          {stats.map((s) => (
            <div
              key={s.label}
              className='flex flex-col items-center text-center'
            >
              <span className='text-3xl font-bold tracking-[-0.02em] md:text-4xl'>
                <Counter end={s.end} suffix={s.suffix} decimals={s.decimals} />
              </span>
              <span className='text-muted-foreground/75 mt-2 text-xs tracking-[0.01em]'>
                {s.label}
              </span>
            </div>
          ))}
        </div>
        <p className='text-muted-foreground/55 mt-8 text-center text-[11px] tracking-[0.01em]'>
          {t('Measured / typical figures, not guarantees.')}
        </p>
      </div>
    </div>
  )
}
