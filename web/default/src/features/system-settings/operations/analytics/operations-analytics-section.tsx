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
import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import {
  Activity,
  CreditCard,
  RefreshCw,
  Server,
  ShoppingCart,
  TrendingUp,
  Users,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { StatCard } from '@/features/dashboard/components/ui/stat-card'
import { cn } from '@/lib/utils'
import { formatNumber, formatPercent, formatQuota } from '@/lib/format'
import {
  getOpsOverview,
  getOpsPaymentProviders,
  getOpsRevenueTrend,
  getOpsUserGrowth,
  type OpsOverview,
  type PaymentProviderStat,
  type RevenueTrendPoint,
  type UserGrowthPoint,
} from './api'
import { OpsTimeSeriesChart } from './ops-charts'

// Top-up `money` is settled in the deployment's payment-gateway currency
// (CNY for the WeChat/Alipay/Xunhu rails this build ships with).
const MONEY_PREFIX = '¥'
const RANGE_OPTIONS = [7, 30, 90] as const

function formatMoney(n: number): string {
  return `${MONEY_PREFIX}${formatNumber(n)}`
}

function dayLabel(unixSeconds: number): string {
  const d = new Date(unixSeconds * 1000)
  const mm = String(d.getMonth() + 1).padStart(2, '0')
  const dd = String(d.getDate()).padStart(2, '0')
  return `${mm}/${dd}`
}

interface OpsData {
  overview: OpsOverview | null
  revenue: RevenueTrendPoint[]
  growth: UserGrowthPoint[]
  providers: PaymentProviderStat[]
}

function useOpsAnalytics(days: number) {
  const [data, setData] = useState<OpsData>({
    overview: null,
    revenue: [],
    growth: [],
    providers: [],
  })
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)

  const refresh = useCallback(() => setReloadKey((k) => k + 1), [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(false)
    Promise.all([
      getOpsOverview(days),
      getOpsRevenueTrend(days),
      getOpsUserGrowth(days),
      getOpsPaymentProviders(days),
    ])
      .then(([overview, revenue, growth, providers]) => {
        if (cancelled) return
        setData({
          overview,
          revenue: revenue.series,
          growth: growth.series,
          providers: providers.providers,
        })
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [days, reloadKey])

  return { data, loading, error, refresh }
}

function Panel(props: {
  title: string
  icon: typeof TrendingUp
  extra?: ReactNode
  children: ReactNode
}) {
  const Icon = props.icon
  return (
    <div className='bg-card overflow-hidden rounded-xl border'>
      <div className='flex items-center justify-between gap-2 border-b px-4 py-3'>
        <div className='flex items-center gap-2'>
          <Icon className='text-muted-foreground/70 size-4' />
          <span className='text-sm font-semibold'>{props.title}</span>
        </div>
        {props.extra}
      </div>
      <div className='p-4'>{props.children}</div>
    </div>
  )
}

function RangeSelector(props: {
  value: number
  onChange: (days: number) => void
}) {
  const { t } = useTranslation()
  return (
    <div className='bg-muted/60 inline-flex h-8 rounded-lg border p-0.5'>
      {RANGE_OPTIONS.map((opt) => {
        const active = props.value === opt
        return (
          <button
            key={opt}
            type='button'
            onClick={() => props.onChange(opt)}
            className={cn(
              'rounded-md px-3 text-xs font-medium transition-colors',
              active
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {t('{{count}}d', { count: opt })}
          </button>
        )
      })}
    </div>
  )
}

function ProvidersPanel(props: { providers: PaymentProviderStat[] }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...props.providers.map((p) => p.revenue))
  if (!props.providers.length) {
    return (
      <p className='text-muted-foreground py-6 text-center text-sm'>
        {t('No payment data in this period')}
      </p>
    )
  }
  return (
    <div className='flex flex-col gap-3'>
      {props.providers.map((p) => (
        <div key={p.provider} className='flex flex-col gap-1.5'>
          <div className='flex items-center justify-between text-sm'>
            <span className='font-medium capitalize'>{p.provider}</span>
            <span className='tabular-nums'>{formatMoney(p.revenue)}</span>
          </div>
          <div className='bg-muted h-2 overflow-hidden rounded-full'>
            <div
              className='bg-accent-green h-full rounded-full'
              style={{ width: `${(p.revenue / max) * 100}%` }}
            />
          </div>
          <div className='text-muted-foreground flex items-center justify-between text-xs'>
            <span>
              {p.order_success}/{p.order_total} {t('orders')}
            </span>
            <span>
              {t('Success')} {formatPercent(p.success_rate * 100)}
            </span>
          </div>
        </div>
      ))}
    </div>
  )
}

function ChannelHealthPanel(props: { overview: OpsOverview | null }) {
  const { t } = useTranslation()
  const ch = props.overview?.channel
  const rows: { label: string; value: number; dot: string }[] = [
    { label: t('Enabled'), value: ch?.enabled ?? 0, dot: 'bg-accent-green' },
    {
      label: t('Manually disabled'),
      value: ch?.manually_disabled ?? 0,
      dot: 'bg-muted-foreground',
    },
    {
      label: t('Auto disabled'),
      value: ch?.auto_disabled ?? 0,
      dot: 'bg-accent-coral',
    },
  ]
  const total = ch?.total ?? 0
  return (
    <div className='flex flex-col gap-4'>
      <div className='flex items-baseline gap-2'>
        <span className='text-2xl font-semibold tabular-nums'>{total}</span>
        <span className='text-muted-foreground text-sm'>
          {t('total channels')}
        </span>
      </div>
      <div className='bg-muted flex h-2.5 overflow-hidden rounded-full'>
        {total > 0 &&
          rows.map((r) => (
            <div
              key={r.label}
              className={cn('h-full', r.dot)}
              style={{ width: `${(r.value / total) * 100}%` }}
            />
          ))}
      </div>
      <div className='grid grid-cols-3 gap-3'>
        {rows.map((r) => (
          <div key={r.label} className='flex flex-col gap-1'>
            <span className='flex items-center gap-1.5 text-xs'>
              <span className={cn('size-2 rounded-full', r.dot)} />
              <span className='text-muted-foreground'>{r.label}</span>
            </span>
            <span className='text-lg font-semibold tabular-nums'>
              {r.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

export function OperationsAnalyticsSection() {
  const { t } = useTranslation()
  const [days, setDays] = useState(30)
  const { data, loading, error, refresh } = useOpsAnalytics(days)
  const { overview, revenue, growth, providers } = data

  const revenueSeries = useMemo(
    () => revenue.map((p) => ({ label: dayLabel(p.date), value: p.revenue })),
    [revenue]
  )
  const growthSeries = useMemo(
    () =>
      growth.map((p) => ({ label: dayLabel(p.date), value: p.cumulative })),
    [growth]
  )
  const newUserSeries = useMemo(
    () => growth.map((p) => ({ label: dayLabel(p.date), value: p.new_users })),
    [growth]
  )
  const revenueSpark = useMemo(() => revenue.map((p) => p.revenue), [revenue])
  const newUsersSpark = useMemo(
    () => growth.map((p) => p.new_users),
    [growth]
  )

  return (
    <div className='flex flex-col gap-6'>
      {/* Header: intent hint + range + refresh */}
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <p className='text-muted-foreground text-sm'>
          {t('Real-time operations overview for the selected period.')}
        </p>
        <div className='flex items-center gap-2'>
          <RangeSelector value={days} onChange={setDays} />
          <button
            type='button'
            onClick={refresh}
            className='text-muted-foreground hover:text-foreground hover:bg-muted flex size-8 items-center justify-center rounded-lg border transition-colors'
            aria-label={t('Refresh')}
          >
            <RefreshCw className={cn('size-4', loading && 'animate-spin')} />
          </button>
        </div>
      </div>

      {error ? (
        <div className='border-destructive/30 bg-destructive/5 text-destructive rounded-xl border p-6 text-center text-sm'>
          {t('Failed to load operations data. Please retry.')}
        </div>
      ) : null}

      {/* KPI cards */}
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4'>
        <StatCard
          title={t('Revenue')}
          value={overview ? formatMoney(overview.revenue) : '-'}
          description={t('Paid orders in period')}
          icon={TrendingUp}
          tone='green'
          loading={loading}
          sparkline={revenueSpark}
          sparklineVariant='line'
        />
        <StatCard
          title={t('Orders')}
          value={overview ? formatNumber(overview.order_total) : '-'}
          description={t('Total / success rate')}
          icon={ShoppingCart}
          tone='blue'
          loading={loading}
          details={[
            {
              label: t('Success'),
              value: overview ? String(overview.order_success) : '-',
              tone: 'success',
            },
            {
              label: t('Rate'),
              value: overview
                ? formatPercent(overview.order_success_rate * 100)
                : '-',
            },
          ]}
        />
        <StatCard
          title={t('Users')}
          value={overview ? formatNumber(overview.new_users) : '-'}
          description={t('New in period')}
          icon={Users}
          tone='blue'
          loading={loading}
          sparkline={newUsersSpark}
          sparklineVariant='bars'
          details={[
            {
              label: t('Active'),
              value: overview ? formatNumber(overview.active_users) : '-',
            },
            {
              label: t('Total'),
              value: overview ? formatNumber(overview.total_users) : '-',
            },
          ]}
        />
        <StatCard
          title={t('Consumption')}
          value={overview ? formatQuota(overview.consumption_quota) : '-'}
          description={t('Quota used in period')}
          icon={Activity}
          tone='amber'
          loading={loading}
          details={[
            {
              label: t('Requests'),
              value: overview ? formatNumber(overview.request_count) : '-',
            },
            {
              label: t('Tokens'),
              value: overview ? formatNumber(overview.token_count) : '-',
            },
          ]}
        />
      </div>

      {/* Trend charts */}
      <div className='grid grid-cols-1 gap-4 xl:grid-cols-2'>
        <Panel title={t('Revenue trend')} icon={TrendingUp}>
          <OpsTimeSeriesChart
            data={revenueSeries}
            kind='area'
            tone='green'
            loading={loading}
            valueFormatter={formatMoney}
          />
        </Panel>
        <Panel title={t('User growth')} icon={Users}>
          <OpsTimeSeriesChart
            data={growthSeries}
            kind='area'
            tone='blue'
            loading={loading}
            valueFormatter={(v) => formatNumber(v)}
          />
        </Panel>
      </div>

      {/* New users + providers + channel health */}
      <div className='grid grid-cols-1 gap-4 xl:grid-cols-3'>
        <Panel title={t('New users per day')} icon={Users}>
          <OpsTimeSeriesChart
            data={newUserSeries}
            kind='bar'
            tone='blue'
            loading={loading}
            valueFormatter={(v) => formatNumber(v)}
          />
        </Panel>
        <Panel title={t('Payment providers')} icon={CreditCard}>
          <ProvidersPanel providers={providers} />
        </Panel>
        <Panel title={t('Channel health')} icon={Server}>
          <ChannelHealthPanel overview={overview} />
        </Panel>
      </div>
    </div>
  )
}
