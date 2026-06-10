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

// Agent entry point on the wallet page:
// - Agents (role === ROLE.AGENT): prominent gradient banner linking to /agent
// - Everyone else (incl. admin/root): a single low-key marketing line linking to /agent/guide
export function AgentBanner() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const isAgent = (user?.role ?? 0) === ROLE.AGENT

  if (isAgent) {
    return (
      <Link
        to='/agent'
        className='group from-primary/10 via-primary/5 to-card hover:border-primary/40 flex items-center justify-between gap-3 rounded-xl border bg-gradient-to-r px-4 py-3.5 transition-colors sm:px-5'
      >
        <div className='flex min-w-0 items-center gap-3'>
          <div className='bg-primary/15 text-primary flex h-9 w-9 shrink-0 items-center justify-center rounded-lg'>
            <Crown className='h-4 w-4' />
          </div>
          <div className='min-w-0'>
            <div className='text-sm font-semibold'>{t('Agent Center')}</div>
            <p className='text-muted-foreground truncate text-xs'>
              {t('Track your referrals, commissions and withdrawals')}
            </p>
          </div>
        </div>
        <ArrowRight className='text-muted-foreground group-hover:text-foreground h-4 w-4 shrink-0 transition-transform group-hover:translate-x-0.5' />
      </Link>
    )
  }

  return (
    <Link
      to='/agent/guide'
      className='text-muted-foreground hover:text-foreground group inline-flex items-center gap-1.5 self-start text-xs transition-colors'
    >
      <Sparkles className='h-3.5 w-3.5' />
      {t('Learn about the agent program — share to earn')}
      <ArrowRight className='h-3 w-3 transition-transform group-hover:translate-x-0.5' />
    </Link>
  )
}
