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
import { useTranslation } from 'react-i18next'
import { getLobeIcon } from '@/lib/lobe-icon'
import { AnimateInView } from '@/components/animate-in-view'

interface ModelBrand {
  key: string
  label: string
}

// Curated brands. Each key resolves to a colored icon via @lobehub/icons (.Color);
// unknown keys degrade to a lettered avatar rather than a broken image.
const BRANDS: readonly ModelBrand[] = [
  // Global frontier providers
  { key: 'OpenAI.Color', label: 'OpenAI' },
  { key: 'Claude.Color', label: 'Claude' },
  { key: 'Gemini.Color', label: 'Gemini' },
  { key: 'Mistral.Color', label: 'Mistral' },
  { key: 'Meta.Color', label: 'Llama' },
  { key: 'Cohere.Color', label: 'Cohere' },
  { key: 'Perplexity.Color', label: 'Perplexity' },
  { key: 'Nvidia.Color', label: 'Nemotron' },
  // Domestic providers
  { key: 'DeepSeek.Color', label: 'DeepSeek' },
  { key: 'Qwen.Color', label: 'Qwen' },
  { key: 'Kimi.Color', label: 'Kimi' },
  { key: 'Zhipu.Color', label: 'ChatGLM' },
  { key: 'Doubao.Color', label: 'Doubao' },
  { key: 'Hunyuan.Color', label: 'Hunyuan' },
  { key: 'Wenxin.Color', label: 'ERNIE' },
  { key: 'Spark.Color', label: 'Spark' },
  { key: 'Minimax.Color', label: 'MiniMax' },
  { key: 'Stepfun.Color', label: 'Stepfun' },
  { key: 'SenseNova.Color', label: 'SenseNova' },
  { key: 'InternLM.Color', label: 'InternLM' },
  { key: 'Baichuan.Color', label: 'Baichuan' },
  { key: 'Yi.Color', label: 'Yi' },
]

export function ModelBrands() {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 overflow-hidden px-6 py-24 md:py-28'>
      {/* Soft brand-tinted glow */}
      <div
        aria-hidden
        className='absolute inset-0 -z-10 opacity-20 dark:opacity-[0.08]'
        style={{
          background: [
            'radial-gradient(ellipse 45% 45% at 25% 30%, oklch(0.7 0.15 250 / 60%) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 40% at 75% 70%, oklch(0.65 0.12 200 / 45%) 0%, transparent 70%)',
          ].join(', '),
        }}
      />

      <div className='mx-auto max-w-5xl'>
        <AnimateInView className='mb-12 text-center' animation='fade-up'>
          <span className='text-muted-foreground/50 text-[10px] font-bold tracking-[0.15em] uppercase'>
            {t('Model Providers')}
          </span>
          <h2 className='mt-3 text-3xl leading-tight font-bold tracking-tight md:text-4xl'>
            {t('Every major provider, one stable endpoint')}
          </h2>
          <p className='text-muted-foreground/80 mx-auto mt-4 max-w-xl text-sm leading-relaxed md:text-base'>
            {t(
              'Global and domestic models, unified behind a single gateway with automatic failover between them.'
            )}
          </p>
        </AnimateInView>

        <AnimateInView
          className='grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5'
          animation='fade-up'
          delay={120}
        >
          {BRANDS.map((brand) => (
            <div
              key={brand.key}
              className='group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-3 rounded-xl border px-4 py-3.5 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'
            >
              <span className='flex size-7 shrink-0 items-center justify-center'>
                {getLobeIcon(brand.key, 26)}
              </span>
              <span className='truncate'>{brand.label}</span>
            </div>
          ))}
        </AnimateInView>
      </div>
    </section>
  )
}
