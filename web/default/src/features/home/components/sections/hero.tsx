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
import { CherryStudio, LobeHub, Dify } from '@lobehub/icons'
import { ArrowRight, BookOpen, Handshake } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useStatus } from '@/hooks/use-status'
import { Button } from '@/components/ui/button'
import { HeroTerminalDemo } from '../hero-terminal-demo'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

// Stylized three-dots indicator representing "More"
const MoreIcon = () => (
  <svg
    className='text-muted-foreground/60 group-hover:text-foreground size-6 shrink-0 transition-colors'
    viewBox='0 0 24 24'
    fill='none'
    xmlns='http://www.w3.org/2000/svg'
  >
    <circle cx='6' cy='12' r='2' fill='currentColor' />
    <circle cx='12' cy='12' r='2' fill='currentColor' />
    <circle cx='18' cy='12' r='2' fill='currentColor' />
  </svg>
)

// Lettered avatar fallback for apps without a brand icon (e.g. our own tools).
const AppInitials = (props: { text: string }) => (
  <span className='flex size-6 shrink-0 items-center justify-center rounded-md bg-blue-500/10 text-[10px] font-bold text-blue-600 dark:bg-blue-400/10 dark:text-blue-400'>
    {props.text}
  </span>
)

interface SupportedApp {
  name: string
  url?: string
  icon?: React.ReactNode
  initials?: string
}

// Compatible client apps. lobe icons where available, lettered avatars for the
// rest (incl. our own OpenClaw / Hermes) so every chip stays on-brand.
const SUPPORTED_APPS: readonly SupportedApp[] = [
  {
    name: 'Cherry Studio',
    url: 'https://cherry-ai.com',
    icon: <CherryStudio.Color size={24} className='shrink-0' />,
  },
  { name: 'CC Switch', url: 'https://ccswitch.io', initials: 'CC' },
  {
    name: 'LobeChat',
    url: 'https://lobehub.com',
    icon: <LobeHub.Color size={24} className='shrink-0' />,
  },
  {
    name: 'Dify',
    url: 'https://dify.ai',
    icon: <Dify.Color size={24} className='shrink-0' />,
  },
  { name: 'OpenClaw', initials: 'OC' },
  { name: 'Hermes', initials: 'HM' },
  { name: 'Chatbox', url: 'https://chatboxai.app', initials: 'CB' },
  { name: 'NextChat', url: 'https://nextchat.dev', initials: 'NC' },
  { name: 'Open WebUI', url: 'https://openwebui.com', initials: 'OW' },
  { name: 'Cline', url: 'https://cline.bot', initials: 'CL' },
]

const APP_CHIP_CLASS =
  'group border-border/40 bg-muted/15 text-foreground/80 hover:border-border hover:bg-muted/30 hover:text-foreground flex items-center gap-2.5 rounded-full border px-4 py-2 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'

const AppChip = (props: { app: SupportedApp }) => {
  const inner = (
    <>
      {props.app.icon ?? <AppInitials text={props.app.initials ?? '?'} />}
      <span>{props.app.name}</span>
    </>
  )
  if (props.app.url) {
    return (
      <a
        href={props.app.url}
        target='_blank'
        rel='noopener noreferrer'
        className={APP_CHIP_CLASS}
      >
        {inner}
      </a>
    )
  }
  return <div className={`${APP_CHIP_CLASS} cursor-default`}>{inner}</div>
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const { status } = useStatus()
  const docsUrl =
    (status?.docs_link as string | undefined) || 'https://docs.goswitch.online'

  const renderDocsButton = () => {
    const isExternal = docsUrl.startsWith('http')
    if (isExternal) {
      return (
        <Button
          variant='outline'
          className='group border-border/50 hover:border-border hover:bg-muted/50 inline-flex h-12 items-center gap-1.5 rounded-xl px-6 text-base font-medium'
          render={
            <a href={docsUrl} target='_blank' rel='noopener noreferrer' />
          }
        >
          <BookOpen className='text-muted-foreground/80 group-hover:text-foreground size-4 transition-colors duration-200' />
          <span>{t('Docs')}</span>
        </Button>
      )
    }
    return (
      <Button
        variant='outline'
        className='group border-border/50 hover:border-border hover:bg-muted/50 inline-flex h-12 items-center gap-1.5 rounded-xl px-6 text-base font-medium'
        render={<Link to={docsUrl} />}
      >
        <BookOpen className='text-muted-foreground/80 group-hover:text-foreground size-4 transition-colors duration-200' />
        <span>{t('Docs')}</span>
      </Button>
    )
  }

  const renderCooperationButton = () => (
    <Button
      variant='outline'
      className='group border-border/50 hover:border-border hover:bg-muted/50 inline-flex h-12 items-center gap-1.5 rounded-xl px-6 text-base font-medium'
      render={<Link to='/cooperation' />}
    >
      <Handshake className='text-muted-foreground/80 group-hover:text-foreground size-4 transition-colors duration-200' />
      <span>{t('Business Cooperation')}</span>
    </Button>
  )

  return (
    <section className='relative z-10 overflow-hidden px-6 pt-24 pb-16 md:pt-32 md:pb-24 lg:pt-36 lg:pb-28'>
      {/* Radial gradient background */}
      <div
        aria-hidden
        className='pointer-events-none absolute inset-0 -z-10 opacity-25 dark:opacity-[0.12]'
        style={{
          background: [
            'radial-gradient(ellipse 60% 50% at 20% 20%, color-mix(in oklch, var(--glow-primary) 80%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 50% 40% at 80% 15%, color-mix(in oklch, var(--glow-accent) 60%, transparent) 0%, transparent 70%)',
            'radial-gradient(ellipse 40% 35% at 40% 80%, color-mix(in oklch, var(--glow-tertiary) 40%, transparent) 0%, transparent 70%)',
          ].join(', '),
        }}
      />
      {/* Grid pattern */}
      <div
        aria-hidden
        className='absolute inset-0 -z-10 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_30%,black_20%,transparent_100%)] bg-[size:4rem_4rem] opacity-[0.08]'
      />

      <div className='mx-auto grid max-w-6xl grid-cols-1 items-start gap-12 lg:grid-cols-12 lg:gap-8'>
        {/* Left Column: Title, description, action buttons and application support */}
        <div className='flex flex-col items-start text-left lg:col-span-7'>
          {/* Top Pill Badge */}
          <div
            className='landing-animate-fade-up mb-5 inline-flex items-center gap-1.5 rounded-full border border-blue-500/20 bg-blue-500/5 px-3.5 py-1.5 text-xs font-medium text-blue-600 opacity-0 shadow-xs dark:border-blue-400/20 dark:bg-blue-400/5 dark:text-blue-400'
            style={{ animationDelay: '0ms' }}
          >
            <span className='relative flex size-1.5'>
              <span className='absolute inline-flex h-full w-full animate-ping rounded-full bg-blue-400 opacity-75' />
              <span className='relative inline-flex size-1.5 rounded-full bg-blue-500 dark:bg-blue-400' />
            </span>
            <span>{t('Always-on AI gateway')}</span>
          </div>

          <h1
            className='landing-animate-fade-up text-[clamp(1.875rem,4.2vw,3rem)] leading-[1.18] font-bold tracking-[-0.02em]'
            style={{ animationDelay: '60ms' }}
          >
            {t('Multimodal APIs, unified,')}{' '}
            <br className='hidden sm:block' />
            <span className='bg-gradient-to-r from-sky-500 via-cyan-500 to-emerald-500 bg-clip-text text-transparent'>
              {t('millisecond-stable at scale.')}
            </span>
          </h1>
          <p
            className='landing-animate-fade-up text-muted-foreground/85 mt-6 max-w-2xl text-lg leading-[1.65] tracking-[-0.003em] opacity-0 md:text-xl'
            style={{ animationDelay: '120ms' }}
          >
            {t(
              'One OpenAI-compatible API — automatic failover and nearest-healthy routing keep every request online and fast.'
            )}
          </p>
          <p
            className='landing-animate-fade-up text-muted-foreground/60 mt-4 text-xs font-medium opacity-0 md:text-sm'
            style={{ animationDelay: '150ms' }}
          >
            {t(
              'OpenAI-compatible · Global and domestic models · Auto-failover'
            )}
          </p>

          <div
            className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3 opacity-0'
            style={{ animationDelay: '180ms' }}
          >
            {props.isAuthenticated ? (
              <>
                <Button
                  className='group h-12 rounded-xl px-6 text-base font-medium shadow-md transition-shadow hover:shadow-lg'
                  render={<Link to='/dashboard' />}
                >
                  {t('Go to Dashboard')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
                {renderCooperationButton()}
                {renderDocsButton()}
              </>
            ) : (
              <>
                <Button
                  className='group h-12 rounded-xl px-6 text-base font-medium shadow-md transition-shadow hover:shadow-lg'
                  render={<Link to='/sign-up' />}
                >
                  {t('Get Started')}
                  <ArrowRight className='ml-1.5 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
                </Button>
                <Button
                  variant='outline'
                  className='border-border/50 hover:border-border hover:bg-muted/50 h-12 rounded-xl px-6 text-base font-medium'
                  render={<Link to='/pricing' />}
                >
                  {t('View Pricing')}
                </Button>
                {renderCooperationButton()}
                {renderDocsButton()}
              </>
            )}
          </div>

          {/* Supported Apps (参考图二样式，进行卡片化和信息扩充设计，增加视觉高度) */}
          <div
            className='landing-animate-fade-up mt-10 w-full max-w-xl opacity-0'
            style={{ animationDelay: '240ms' }}
          >
            <div className='mb-4 flex flex-col gap-1'>
              <span className='text-muted-foreground/50 text-[10px] font-bold tracking-[0.15em] uppercase'>
                {t('Supported Applications')}
              </span>
              <p className='text-muted-foreground/60 text-xs leading-relaxed'>
                {t(
                  'Supports one-click configuration and perfectly adapts to multi-protocol setups.'
                )}
              </p>
            </div>
            <div className='flex flex-wrap items-center gap-2.5'>
              {SUPPORTED_APPS.map((app) => (
                <AppChip key={app.name} app={app} />
              ))}

              {/* "更多" */}
              <div className='group border-border/40 bg-muted/15 text-foreground/55 hover:border-border hover:bg-muted/30 hover:text-foreground flex cursor-default items-center gap-2 rounded-full border px-4 py-2 text-sm font-medium shadow-[0_1px_2.5px_rgba(0,0,0,0.01)] backdrop-blur-xs transition-all duration-300 hover:scale-[1.02]'>
                <MoreIcon />
                <span>{t('More Apps')}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right Column: Hero Terminal API Demo */}
        <div
          className='landing-animate-fade-up flex w-full justify-center opacity-0 lg:col-span-5 lg:self-center'
          style={{ animationDelay: '320ms' }}
        >
          <HeroTerminalDemo className='mt-8 lg:mt-0' />
        </div>
      </div>
    </section>
  )
}
