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
import {
  Activity,
  Gauge,
  Network,
  ShieldCheck,
  Layers,
  Globe,
  Wallet,
  Plug,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      id: 'failover',
      num: '01',
      title: t('Automatic failover'),
      desc: t(
        'When a channel slows or errors, traffic reroutes to the next healthy upstream within seconds — no manual switching, no downtime for your users.'
      ),
      span: 'md:col-span-2',
      icon: <Activity className='size-4 text-sky-400' />,
      visual: (
        <div className='mt-4 grid grid-cols-3 gap-2'>
          {['OpenAI', 'Claude', 'Gemini', 'DeepSeek', 'Qwen', 'Llama'].map(
            (name) => (
              <div
                key={name}
                className='border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-sky-500/30 hover:bg-sky-500/5'
              >
                {name}
              </div>
            )
          )}
        </div>
      ),
    },
    {
      id: 'monitoring',
      num: '02',
      title: t('Continuous health monitoring'),
      desc: t(
        'Live health, latency, and availability for every channel, checked on a fixed interval.'
      ),
      span: 'md:col-span-1',
      icon: <Gauge className='size-4 text-emerald-400' />,
      visual: (
        <div className='mt-4 flex items-center justify-center'>
          <div className='flex size-16 items-center justify-center rounded-2xl border border-emerald-500/20 bg-emerald-500/5'>
            <Gauge className='size-7 text-emerald-500/70' strokeWidth={1.5} />
          </div>
        </div>
      ),
    },
    {
      id: 'routing',
      num: '03',
      title: t('Nearest-healthy routing'),
      desc: t(
        'Requests are sent to the lowest-latency healthy channel automatically.'
      ),
      span: 'md:col-span-1',
      icon: <Network className='size-4 text-violet-400' />,
      visual: (
        <div className='mt-4 flex items-center justify-center'>
          <div className='flex size-16 items-center justify-center rounded-2xl border border-violet-500/20 bg-violet-500/5'>
            <Network className='size-7 text-violet-500/70' strokeWidth={1.5} />
          </div>
        </div>
      ),
    },
    {
      id: 'keys',
      num: '04',
      title: t('Key isolation'),
      desc: t(
        'Issue scoped keys per project, each with its own quota and rate limit — revoke any time, with zero data retention by default.'
      ),
      span: 'md:col-span-2',
      icon: <ShieldCheck className='size-4 text-amber-400' />,
      visual: (
        <div className='mt-4'>
          <div className='grid grid-cols-2 gap-2 sm:grid-cols-3'>
            {['key · prod', 'key · dev', 'key · ci', 'key · staging', 'key · app', 'key · n'].map(
              (name) => (
                <div
                  key={name}
                  className='border-border/30 bg-muted/20 text-muted-foreground flex items-center justify-center gap-1.5 rounded-lg border px-3 py-2 text-xs transition-colors duration-300 hover:border-amber-500/30 hover:bg-amber-500/5'
                >
                  <ShieldCheck className='size-3.5 shrink-0 text-amber-500/70' />
                  <span className='truncate'>{name}</span>
                </div>
              )
            )}
          </div>
          <div className='text-muted-foreground mt-3 flex items-center justify-center gap-1.5 text-xs'>
            <ShieldCheck className='size-3.5 text-emerald-500' />
            {t('0 data retained')}
          </div>
        </div>
      ),
    },
  ]

  const additionalFeatures = [
    {
      icon: <Layers className='size-5' strokeWidth={1.5} />,
      title: t('High concurrency'),
      desc: t('Absorbs traffic spikes without degradation.'),
    },
    {
      icon: <Globe className='size-5' strokeWidth={1.5} />,
      title: t('Multi-region'),
      desc: t('Stable access to global models across regions.'),
    },
    {
      icon: <Wallet className='size-5' strokeWidth={1.5} />,
      title: t('Unified billing'),
      desc: t('One balance across every model and route.'),
    },
    {
      icon: <Plug className='size-5' strokeWidth={1.5} />,
      title: t('Drop-in migration'),
      desc: t(
        'Point your base URL at GoSwitch — your SDK and code stay unchanged.'
      ),
    },
  ]

  return (
    <section className='relative z-10 px-6 py-24 md:py-28'>
      <div className='mx-auto max-w-6xl'>
        <AnimateInView className='mb-16 max-w-lg' animation='fade-up'>
          <p className='text-muted-foreground/55 mb-3 text-[11px] font-semibold tracking-[0.16em] uppercase'>
            {t('Core Capabilities')}
          </p>
          <h2 className='text-3xl leading-[1.12] font-bold tracking-[-0.02em] md:text-4xl'>
            {t('Engineered to stay up under real traffic')}
          </h2>
        </AnimateInView>

        {/* Bento grid */}
        <div className='border-border/40 bg-border/40 grid gap-px overflow-hidden rounded-xl border md:grid-cols-3'>
          {features.map((f, i) => (
            <AnimateInView
              key={f.id}
              delay={i * 100}
              animation='scale-in'
              className={`bg-background group hover:bg-muted/20 p-7 transition-colors duration-300 md:p-8 ${f.span}`}
            >
              <div className='mb-3 flex items-center gap-3'>
                <span className='border-border/40 bg-muted text-muted-foreground flex size-7 items-center justify-center rounded-md border text-[10px] font-semibold tabular-nums'>
                  {f.num}
                </span>
                <h3 className='text-sm font-semibold'>{f.title}</h3>
              </div>
              <p className='text-muted-foreground text-sm leading-[1.65] tracking-[-0.002em]'>
                {f.desc}
              </p>
              {f.visual}
            </AnimateInView>
          ))}
        </div>

        {/* Additional features row */}
        <div className='mt-12 grid grid-cols-2 gap-8 md:grid-cols-4 md:gap-12'>
          {additionalFeatures.map((f, i) => (
            <AnimateInView
              key={f.title}
              delay={i * 100}
              animation='fade-up'
              className='flex flex-col items-center text-center'
            >
              <div className='text-muted-foreground border-border/50 bg-muted/30 group-hover:text-foreground mb-3 flex size-12 items-center justify-center rounded-xl border transition-colors'>
                {f.icon}
              </div>
              <h3 className='mb-1.5 text-sm font-semibold'>{f.title}</h3>
              <p className='text-muted-foreground max-w-[200px] text-xs leading-relaxed'>
                {f.desc}
              </p>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
