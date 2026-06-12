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
import { ArrowRight, Crown, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { ROLE } from '@/lib/roles'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

// Agent entry point on the wallet page:
// - Agents (role === ROLE.AGENT): prominent gradient banner linking to /agent
// - Everyone else (incl. admin/root): a single low-key marketing line linking to /agent/guide
export function AgentBanner() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const isAgent = (user?.role ?? 0) === ROLE.AGENT

  if (isAgent) {
    return (
      <Card className='from-primary/10 via-primary/5 to-card rounded-xl bg-gradient-to-r py-0'>
        <CardContent className='space-y-3 p-4 sm:p-5'>
          <div className='flex items-start justify-between gap-3'>
            <div className='flex min-w-0 items-center gap-3'>
              <div className='bg-primary/15 text-primary flex h-9 w-9 shrink-0 items-center justify-center rounded-lg'>
                <Crown className='h-4 w-4' />
              </div>
              <div className='min-w-0'>
                <div className='text-sm font-semibold'>
                  {t('Agent program upgraded')}
                </div>
                <p className='text-muted-foreground line-clamp-2 text-xs'>
                  {t('Track your referrals, commissions and withdrawals')}
                </p>
              </div>
            </div>
            <Badge variant='secondary'>{t('Agent')}</Badge>
          </div>

          <Button className='w-full gap-2' render={<Link to='/agent' />}>
            {t('Open Agent Center')}
            <ArrowRight className='h-4 w-4' />
          </Button>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card className='rounded-xl py-0'>
      <CardContent className='space-y-3 p-4 sm:p-5'>
        <div className='flex items-start justify-between gap-3'>
          <div className='flex min-w-0 items-center gap-3'>
            <div className='bg-accent-amber/15 text-accent-amber flex h-9 w-9 shrink-0 items-center justify-center rounded-lg'>
              <Sparkles className='h-4 w-4' />
            </div>
            <div className='min-w-0'>
              <div className='text-sm font-semibold'>
                {t('Not an agent yet')}
              </div>
              <p className='text-muted-foreground line-clamp-2 text-xs'>
                {t(
                  'Upgrade to agent to unlock reseller pricing and commission workflows.'
                )}
              </p>
            </div>
          </div>
          <Badge variant='outline'>{t('Referral Program')}</Badge>
        </div>

        <Button
          variant='outline'
          className='w-full gap-2'
          render={<Link to='/agent/guide' />}
        >
          {t('View upgrade guide')}
          <ArrowRight className='h-4 w-4' />
        </Button>
      </CardContent>
    </Card>
  )
}
