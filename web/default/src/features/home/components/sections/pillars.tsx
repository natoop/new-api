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
import { ShieldCheck, Activity, Zap, type LucideIcon } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

interface Pillar {
  id: string
  icon: LucideIcon
  title: string
  desc: string
  badge: string
  accent: string
  surface: string
  dot: string
  card: string
  scale: string
  delay: number
}

// Stability sits center and slightly larger (the spec's reliability axis);
// Security and Speed flank it. Colors reuse the welcome-dialog token set.
const PILLARS: readonly Pillar[] = [
  {
    id: 'security',
    icon: ShieldCheck,
    title: 'Security',
    desc: 'Per-key isolation with zero data retention by default. Prompts and responses are not stored or reused — your keys stay yours.',
    badge: 'Zero data retention by default',
    accent: 'text-emerald-500',
    surface: 'border-emerald-500/20 bg-emerald-500/10',
    dot: 'bg-emerald-500',
    card: '',
    scale: '',
    delay: 100,
  },
  {
    id: 'stability',
    icon: Activity,
    title: 'Stability',
    desc: 'When an upstream slows or errors, multi-channel auto-failover reroutes traffic within seconds. High availability is the default and is monitored continuously.',
    badge: 'Auto-failover · continuous health checks',
    accent: 'text-sky-500',
    surface: 'border-sky-500/20 bg-sky-500/10',
    dot: 'bg-sky-500',
    card: 'border-sky-500/30',
    scale: 'md:scale-[1.04]',
    delay: 0,
  },
  {
    id: 'speed',
    icon: Zap,
    title: 'Speed',
    desc: 'Each request is routed to the nearest healthy channel over a single OpenAI-compatible protocol — lower latency, less integration overhead.',
    badge: 'Nearest-healthy routing',
    accent: 'text-amber-500',
    surface: 'border-amber-500/20 bg-amber-500/10',
    dot: 'bg-amber-500',
    card: '',
    scale: '',
    delay: 100,
  },
] as const

export function Pillars() {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 overflow-hidden px-6 py-24 md:py-28'>
      {/* Ambient brand glow */}
      <div
        aria-hidden
        className='absolute inset-0 -z-10 opacity-20 dark:opacity-[0.08]'
        style={{
          background: [
            'radial-gradient(ellipse 50% 45% at 50% 20%, oklch(0.72 0.18 250 / 70%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 20% 75%, oklch(0.72 0.16 155 / 45%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 80% 70%, oklch(0.78 0.14 75 / 40%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-14 text-center' animation='fade-up'>
          <h2 className='text-3xl leading-[1.12] font-bold tracking-[-0.02em] md:text-4xl'>
            {t('Built to stay up. Tuned to stay fast.')}
          </h2>
          <p className='text-muted-foreground/80 mx-auto mt-4 max-w-xl text-[15px] leading-[1.7] tracking-[-0.002em] md:text-base'>
            {t(
              'Reliability and speed you can operate on — security underpins both.'
            )}
          </p>
        </AnimateInView>

        <div className='grid items-stretch gap-6 md:grid-cols-3'>
          {PILLARS.map((pillar) => (
            <AnimateInView
              key={pillar.id}
              animation='scale-in'
              delay={pillar.delay}
              className={`glass-card flex flex-col rounded-2xl border p-6 ${pillar.card} ${pillar.scale}`}
            >
              <div
                className={`flex size-9 items-center justify-center rounded-lg border ${pillar.surface}`}
              >
                <pillar.icon className={`size-4.5 ${pillar.accent}`} />
              </div>
              <h3 className='mt-4 text-lg font-semibold'>{t(pillar.title)}</h3>
              <p className='text-muted-foreground mt-2 text-sm leading-[1.65] tracking-[-0.002em]'>
                {t(pillar.desc)}
              </p>
              <div className='text-muted-foreground mt-auto flex items-center gap-2 pt-5 text-xs'>
                <span className={`size-1.5 rounded-full ${pillar.dot}`} />
                {t(pillar.badge)}
              </div>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
