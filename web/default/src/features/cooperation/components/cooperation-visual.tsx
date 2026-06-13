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
import { Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { getLobeIcon } from '@/lib/lobe-icon'
import { AnimateInView } from '@/components/animate-in-view'

// Floating brand chips — keys verified renderable via @lobehub/icons (.Color).
// `top`/`left` place each chip; `delay` staggers the entrance.
const FLOATING_ICONS = [
  { key: 'OpenAI.Color', top: '8%', left: '12%', size: 40, delay: 0 },
  { key: 'Claude.Color', top: '18%', left: '64%', size: 48, delay: 90 },
  { key: 'Gemini.Color', top: '40%', left: '24%', size: 44, delay: 180 },
  { key: 'DeepSeek.Color', top: '34%', left: '74%', size: 38, delay: 270 },
  { key: 'Qwen.Color', top: '62%', left: '14%', size: 42, delay: 360 },
  { key: 'Kimi.Color', top: '70%', left: '58%', size: 40, delay: 450 },
  { key: 'Mistral.Color', top: '78%', left: '36%', size: 36, delay: 540 },
  { key: 'Doubao.Color', top: '52%', left: '50%', size: 46, delay: 630 },
] as const

export function CooperationVisual() {
  const { t } = useTranslation()

  return (
    <div className='glass-card relative h-full min-h-80 overflow-hidden rounded-2xl p-6'>
      {/* Brand-tinted gradient wash */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 opacity-70 dark:opacity-50'
        style={{
          background: [
            'radial-gradient(ellipse 60% 50% at 20% 20%, oklch(0.72 0.18 250 / 35%) 0%, transparent 70%)',
            'radial-gradient(ellipse 55% 50% at 80% 75%, oklch(0.72 0.16 155 / 28%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      {/* Floating colored model icons */}
      <div aria-hidden className='absolute inset-0'>
        {FLOATING_ICONS.map((item) => (
          <span
            key={item.key}
            className='absolute'
            style={{ top: item.top, left: item.left }}
          >
            <AnimateInView animation='scale-in' delay={item.delay}>
              <span
                className='cooperation-float bg-background/70 ring-border/50 flex items-center justify-center rounded-2xl p-2.5 shadow-md ring-1 backdrop-blur-sm'
                style={{ animationDelay: `${item.delay}ms` }}
              >
                {getLobeIcon(item.key, item.size)}
              </span>
            </AnimateInView>
          </span>
        ))}
      </div>

      {/* Caption */}
      <div className='relative flex h-full flex-col justify-end'>
        <div className='bg-primary/10 text-primary ring-primary/15 mb-3 flex size-10 w-fit items-center justify-center rounded-xl ring-1 ring-inset'>
          <Sparkles className='size-5' />
        </div>
        <h3 className='text-lg font-semibold tracking-tight'>
          {t('One gateway, every leading model')}
        </h3>
        <p className='text-muted-foreground/80 mt-2 max-w-sm text-sm leading-relaxed'>
          {t(
            'Global frontier and domestic models, unified behind a single stable endpoint — ready to power your products at scale.'
          )}
        </p>
      </div>
    </div>
  )
}
