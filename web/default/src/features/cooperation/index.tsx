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
  ExternalLink,
  Handshake,
  Boxes,
  Network,
  Headset,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { PublicLayout } from '@/components/layout'

const COOPERATION_FORM_URL =
  'https://packymellc.feishu.cn/share/base/form/shrcn0MQ2rxSWdr2N9WKT5cyQ3n'

const PARTNERSHIPS = [
  {
    icon: Boxes,
    title: 'API wholesale & volume pricing',
    desc: 'Tiered volume discounts that scale with usage, covering global frontier and domestic models alike.',
  },
  {
    icon: Network,
    title: 'Reseller & channel partnership',
    desc: 'Open your own agent console with custom pricing and promo codes; orders are auto-attributed and commission settled.',
  },
  {
    icon: Headset,
    title: 'Dedicated technical integration',
    desc: 'Multi-protocol onboarding and white-label options, with a dedicated contact channel and SLA to go live fast.',
  },
] as const

export function Cooperation() {
  const { t } = useTranslation()

  return (
    <PublicLayout showMainContainer={false}>
      <section className='relative overflow-hidden px-6 py-12 md:py-16'>
        {/* Brand-tinted radial glow, consistent with the landing hero */}
        <div
          aria-hidden
          className='pointer-events-none absolute inset-0 -z-10 opacity-25 dark:opacity-[0.12]'
          style={{
            background: [
              'radial-gradient(ellipse 55% 45% at 25% 15%, oklch(0.72 0.18 250 / 75%) 0%, transparent 70%)',
              'radial-gradient(ellipse 45% 40% at 80% 20%, oklch(0.72 0.16 155 / 45%) 0%, transparent 70%)',
            ].join(', '),
          }}
        />

        <div className='mx-auto max-w-5xl'>
          {/* Header */}
          <div className='max-w-2xl'>
            <div className='border-primary/20 bg-primary/5 text-primary mb-5 inline-flex items-center gap-1.5 rounded-full border px-3.5 py-1.5 text-xs font-medium'>
              <Handshake className='size-3.5' />
              <span>{t('Partner with us')}</span>
            </div>
            <h1 className='text-3xl leading-tight font-bold tracking-tight md:text-4xl'>
              {t('Business Cooperation')}
            </h1>
            <p className='text-muted-foreground/85 mt-4 text-base leading-relaxed'>
              {t(
                "Whether you're an API reseller, distribution agent, SaaS integrator, or enterprise buyer, we offer competitive volume pricing, a stable multi-protocol gateway, and dedicated support to help you onboard fast and scale."
              )}
            </p>
          </div>

          {/* Partnership tracks */}
          <div className='mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
            {PARTNERSHIPS.map((item) => (
              <div
                key={item.title}
                className='group border-border/60 bg-card hover:border-primary/30 relative overflow-hidden rounded-2xl border p-5 shadow-sm transition-all duration-300 hover:shadow-md'
              >
                <div className='bg-primary/10 text-primary ring-primary/15 mb-4 flex size-10 items-center justify-center rounded-xl ring-1 ring-inset'>
                  <item.icon className='size-5' />
                </div>
                <h3 className='text-base font-semibold'>{t(item.title)}</h3>
                <p className='text-muted-foreground/80 mt-2 text-sm leading-relaxed'>
                  {t(item.desc)}
                </p>
              </div>
            ))}
          </div>

          {/* Lead form */}
          <div className='mt-12 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
            <div>
              <h2 className='text-xl font-semibold tracking-tight'>
                {t('Tell us about your needs')}
              </h2>
              <p className='text-muted-foreground/80 mt-1.5 text-sm'>
                {t(
                  'Leave your details and our team will reach out to discuss the partnership that fits you.'
                )}
              </p>
            </div>
            <Button
              variant='outline'
              className='shrink-0'
              render={
                <a
                  href={COOPERATION_FORM_URL}
                  target='_blank'
                  rel='noopener noreferrer'
                />
              }
            >
              <ExternalLink className='size-4' />
              {t('Open cooperation form in a new tab')}
            </Button>
          </div>

          <div className='border-border/60 bg-card mt-4 overflow-hidden rounded-2xl border shadow-sm'>
            <iframe
              src={COOPERATION_FORM_URL}
              className='h-[64vh] w-full border-0'
              title={t('Business Cooperation')}
            />
          </div>
          <p className='text-muted-foreground/60 mt-3 text-center text-xs'>
            {t(
              "If the form doesn't load here, use the button above to open it in a new tab."
            )}
          </p>
        </div>
      </section>
    </PublicLayout>
  )
}
