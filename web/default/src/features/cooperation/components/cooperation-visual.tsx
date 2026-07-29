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

interface Brand {
  key: string
  name: string
}

// Domestic vendors lead the wall; a few global names are sprinkled in.
const COLUMN_A: readonly Brand[] = [
  { key: 'DeepSeek.Color', name: 'DeepSeek' },
  { key: 'Qwen.Color', name: 'Qwen' },
  { key: 'Kimi.Color', name: 'Kimi' },
  { key: 'Zhipu.Color', name: 'ChatGLM' },
  { key: 'Doubao.Color', name: 'Doubao' },
  { key: 'Hunyuan.Color', name: 'Hunyuan' },
  { key: 'OpenAI.Color', name: 'OpenAI' },
]

const COLUMN_B: readonly Brand[] = [
  { key: 'Wenxin.Color', name: 'ERNIE' },
  { key: 'Spark.Color', name: 'Spark' },
  { key: 'Minimax.Color', name: 'MiniMax' },
  { key: 'Baichuan.Color', name: 'Baichuan' },
  { key: 'Yi.Color', name: 'Yi' },
  { key: 'Claude.Color', name: 'Claude' },
  { key: 'Gemini.Color', name: 'Gemini' },
]

function BrandChip(props: { brand: Brand }) {
  return (
    <div className='border-border/50 bg-background/70 flex items-center gap-2.5 rounded-xl border px-3.5 py-2.5 shadow-sm backdrop-blur-sm'>
      <span className='flex size-6 shrink-0 items-center justify-center'>
        {getLobeIcon(props.brand.key, 22)}
      </span>
      <span className='text-foreground/85 truncate text-sm font-medium'>
        {props.brand.name}
      </span>
    </div>
  )
}

function ScrollColumn(props: {
  brands: readonly Brand[]
  direction: 'up' | 'down'
}) {
  const animation =
    props.direction === 'up' ? 'animate-scroll-up' : 'animate-scroll-down'
  return (
    <div className={`flex flex-col gap-3 ${animation}`}>
      {[...props.brands, ...props.brands].map((brand, i) => (
        <BrandChip key={`${brand.key}-${i}`} brand={brand} />
      ))}
    </div>
  )
}

export function CooperationVisual() {
  const { t } = useTranslation()

  return (
    <div className='glass-card relative flex h-full min-h-80 flex-col overflow-hidden rounded-2xl'>
      {/* Brand-tinted gradient wash */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 opacity-70 dark:opacity-50'
        style={{
          background: [
            'radial-gradient(ellipse 60% 50% at 20% 15%, color-mix(in oklch, var(--glow-primary) 35%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 55% 50% at 85% 85%, color-mix(in oklch, var(--glow-tertiary) 28%, transparent) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      {/* Header */}
      <div className='relative z-10 px-6 pt-6'>
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

      {/* Scrolling brand wall — domestic-led, global sprinkled in; runs to the
          card's bottom edge so its height tracks the form column beside it. */}
      <div
        aria-hidden
        className='relative z-10 mt-6 min-h-44 flex-1 overflow-hidden px-6 [mask-image:linear-gradient(to_bottom,transparent,#000_10%,#000_92%,transparent)]'
      >
        <div className='grid grid-cols-2 gap-3'>
          <ScrollColumn brands={COLUMN_A} direction='up' />
          <ScrollColumn brands={COLUMN_B} direction='down' />
        </div>
      </div>
    </div>
  )
}
