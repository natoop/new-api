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
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Dialog } from '@/components/dialog'

const DISMISS_STORAGE_KEY = 'goswitch_home_announcement_dismissed_v1'
const OPEN_DELAY_MS = 600

const PILLARS = [
  {
    id: 'stability',
    icon: Activity,
    title: 'Stability',
    desc: 'Automatic multi-channel failover keeps you online, monitored continuously.',
    accent: 'text-sky-500',
    surface: 'border-sky-500/20 bg-sky-500/10',
  },
  {
    id: 'security',
    icon: ShieldCheck,
    title: 'Security',
    desc: 'Per-key isolation with zero data retention keeps every request private.',
    accent: 'text-emerald-500',
    surface: 'border-emerald-500/20 bg-emerald-500/10',
  },
  {
    id: 'efficiency',
    icon: Zap,
    title: 'Efficiency',
    desc: 'One unified protocol gives low-latency access to global AI models.',
    accent: 'text-amber-500',
    surface: 'border-amber-500/20 bg-amber-500/10',
  },
] as const

export function WelcomeAnnouncement() {
  const { t } = useTranslation()
  const { auth } = useAuthStore()
  const isAuthenticated = !!auth.user
  const [open, setOpen] = useState(false)
  const [dontShowAgain, setDontShowAgain] = useState(false)

  useEffect(() => {
    if (typeof window === 'undefined') return
    if (window.localStorage.getItem(DISMISS_STORAGE_KEY)) return

    const timeoutId = window.setTimeout(() => setOpen(true), OPEN_DELAY_MS)
    return () => window.clearTimeout(timeoutId)
  }, [])

  const handleOpenChange = (next: boolean) => {
    setOpen(next)
    if (!next && dontShowAgain && typeof window !== 'undefined') {
      window.localStorage.setItem(DISMISS_STORAGE_KEY, '1')
    }
  }

  return (
    <Dialog
      open={open}
      onOpenChange={handleOpenChange}
      contentClassName='glass-card sm:max-w-lg'
      title={
        <span className='bg-gradient-to-r from-sky-500 via-cyan-500 to-emerald-500 bg-clip-text text-transparent'>
          {t('Welcome to GoSwitch')}
        </span>
      }
      description={t(
        'The reliable gateway for AI models — stable by default, secure by design, efficient at scale.'
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
    </Dialog>
  )
}
