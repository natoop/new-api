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
import { useEffect, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { ArrowRight, ShieldCheck, Activity, Zap } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { getAnnouncementColorClass } from '@/lib/colors'
import { formatDateTimeObject } from '@/lib/time'
import { cn } from '@/lib/utils'
import { useNotifications } from '@/hooks/use-notifications'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Markdown } from '@/components/ui/markdown'
import { Dialog } from '@/components/dialog'

const STORAGE_PREFIX = 'goswitch_home_welcome_dismissed_'
const BRAND_FINGERPRINT = 'brand-v1'
const OPEN_DELAY_MS = 600

const PILLARS = [
  {
    id: 'stability',
    icon: Activity,
    title: 'Stability',
    desc: 'Auto-failover keeps you online, monitored continuously.',
    accent: 'text-sky-500',
    surface: 'border-sky-500/20 bg-sky-500/10',
  },
  {
    id: 'speed',
    icon: Zap,
    title: 'Speed',
    desc: 'Nearest-healthy routing over one OpenAI-compatible protocol — low latency, minimal code.',
    accent: 'text-amber-500',
    surface: 'border-amber-500/20 bg-amber-500/10',
  },
  {
    id: 'security',
    icon: ShieldCheck,
    title: 'Security',
    desc: 'Per-key isolation with zero data retention keeps every request private.',
    accent: 'text-emerald-500',
    surface: 'border-emerald-500/20 bg-emerald-500/10',
  },
] as const

function shortHash(input: string): string {
  let h = 0
  for (let i = 0; i < input.length; i += 1) {
    h = (h << 5) - h + input.charCodeAt(i)
    h |= 0
  }
  return (h >>> 0).toString(36)
}

function buildFingerprint(items: Record<string, unknown>[]): string {
  if (items.length === 0) return BRAND_FINGERPRINT
  return items
    .map((it) => {
      const a = it as { type?: string; content?: string; publishDate?: string }
      return `${a.type || ''}|${(a.content || '').trim()}|${a.publishDate || ''}`
    })
    .join('::')
}

export function WelcomeAnnouncement() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user
  const notifications = useNotifications()
  const visible = notifications.announcementsEnabled
    ? notifications.announcements.slice(0, 3)
    : []
  const dismissKey = STORAGE_PREFIX + shortHash(buildFingerprint(visible))
  const [open, setOpen] = useState(false)
  const [dontShowAgain, setDontShowAgain] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') return
    if (notifications.loading) return
    if (window.localStorage.getItem(dismissKey)) return

    const timeoutId = window.setTimeout(() => setOpen(true), OPEN_DELAY_MS)
    return () => window.clearTimeout(timeoutId)
  }, [notifications.loading, dismissKey])

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next && dontShowAgain && typeof window !== 'undefined') {
      window.localStorage.setItem(dismissKey, '1')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      contentClassName='glass-card sm:max-w-lg'
      title={
        <div className='flex items-center gap-2'>
          <span className='text-sm font-bold tracking-[-0.01em]'>GoSwitch</span>
          <span className='bg-gradient-to-r from-sky-500 via-cyan-500 to-emerald-500 bg-clip-text text-transparent'>
            {t('Welcome to GoSwitch')}
          </span>
        </div>
      }
      description={t(
        'The stable, fast gateway for AI models — auto-failover keeps you online, nearest-healthy routing keeps you quick.'
      )}
      footer={
        <div className='flex w-full flex-col gap-3 sm:flex-row sm:items-center sm:justify-between'>
          <label className='text-muted-foreground flex cursor-pointer items-center gap-2 text-sm select-none'>
            <Checkbox
              checked={dontShowAgain}
              onCheckedChange={(checked) => setDontShowAgain(checked === true)}
            />
            {t('Do not show again')}
          </label>
          <div className='flex items-center gap-2'>
            <Button variant='outline' onClick={() => handleOpenChange(false)}>
              {t('Got it')}
            </Button>
            <Button
              className='group'
              onClick={() => handleOpenChange(false)}
              render={<Link to={isAuthenticated ? '/dashboard' : '/sign-up'} />}
            >
              {isAuthenticated ? t('Go to Dashboard') : t('Get Started')}
              <ArrowRight className='ml-1 size-4 transition-transform duration-200 group-hover:translate-x-0.5' />
            </Button>
          </div>
        </div>
      }
    >
      {visible.length > 0 ? (
        <div className='max-h-[min(48vh,22rem)] space-y-3 overflow-y-auto pr-1'>
          {visible.map((item, idx) => {
            const announcement = item as {
              type?: string
              content?: string
              extra?: string
              publishDate?: string
            }
            const date = announcement.publishDate
              ? formatDateTimeObject(new Date(announcement.publishDate))
              : ''
            return (
              <div
                key={idx}
                className='border-border/40 bg-muted/20 flex items-start gap-3 rounded-xl border p-3.5'
              >
                <span
                  className={cn(
                    'mt-1.5 inline-block size-2 shrink-0 rounded-full',
                    getAnnouncementColorClass(announcement.type)
                  )}
                />
                <div className='min-w-0 flex-1'>
                  <div className='text-sm leading-[1.65]'>
                    <Markdown>{announcement.content || ''}</Markdown>
                  </div>
                  {announcement.extra ? (
                    <div className='text-muted-foreground mt-1 text-xs'>
                      <Markdown>{announcement.extra}</Markdown>
                    </div>
                  ) : null}
                  {date ? (
                    <div className='text-muted-foreground/70 mt-1.5 text-[11px] tracking-[0.01em]'>
                      {date}
                    </div>
                  ) : null}
                </div>
              </div>
            )
          })}
        </div>
      ) : (
        <div className='space-y-3'>
          {PILLARS.map((pillar) => (
            <div
              key={pillar.id}
              className='border-border/40 bg-muted/20 flex items-start gap-3 rounded-xl border p-3.5'
            >
              <div
                className={`flex size-9 shrink-0 items-center justify-center rounded-lg border ${pillar.surface}`}
              >
                <pillar.icon className={`size-4.5 ${pillar.accent}`} />
              </div>
              <div className='min-w-0'>
                <h4 className='text-sm font-semibold'>{t(pillar.title)}</h4>
                <p className='text-muted-foreground mt-0.5 text-sm leading-relaxed'>
                  {t(pillar.desc)}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </Dialog>
  )
}
