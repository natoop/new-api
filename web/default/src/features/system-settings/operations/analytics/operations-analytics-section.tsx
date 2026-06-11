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
  Package,
  RefreshCw,
  Server,
  ShoppingCart,
  TrendingUp,
  Users,
  Wifi,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard } from '@/features/dashboard/components/ui/stat-card'
import { cn } from '@/lib/utils'
import { formatNumber, formatPercent, formatQuota } from '@/lib/format'
import {
  getOpsOverview,
  getOpsPaymentProviders,
  getOpsPlanSales,
  getOpsRevenueTrend,
  getOpsUserGrowth,
  type OpsOverview,
  type PaymentProviderStat,
  type PlanSales,
  type RevenueTrendPoint,
  type UserGrowthPoint,
} from './api'
import { OpsTimeSeriesChart } from './ops-charts'

// Top-up / subscription `money` is settled in the deployment's payment-gateway
// currency (CNY for the WeChat/Alipay/Xunhu rails this build ships with).
const MONEY_PREFIX = '¥'
const RANGE_OPTIONS = [7, 30, 90] as const

function formatMoney(n: number): string {
  // Proper retail-style RMB: thousands grouping + always two decimals.
  return `${MONEY_PREFIX}${Intl.NumberFormat(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n)}`
}

function dayLabel(unixSeconds: number): string {
  // Backend buckets days at UTC midnight — label with UTC date parts so the
  // axis never drifts a day in non-UTC timezones.
  const d = new Date(unixSeconds * 1000)
  const mm = String(d.getUTCMonth() + 1).padStart(2, '0')
  const dd = String(d.getUTCDate()).padStart(2, '0')
  return `${mm}/${dd}`
}

interface OpsData {
  overview: OpsOverview | null
  revenue: RevenueTrendPoint[]
  growth: UserGrowthPoint[]
  providers: PaymentProviderStat[]
  plans: PlanSales[]
}

function useOpsAnalytics(days: number) {
  const [data, setData] = useState<OpsData>({
    overview: null,
    revenue: [],
    growth: [],
    providers: [],
    plans: [],
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
      getOpsPlanSales(days),
    ])
      .then(([overview, revenue, growth, providers, plans]) => {
        if (cancelled) return
        setData({
          overview,
          revenue: revenue.series,
          growth: growth.series,
          providers: providers.providers,
          plans: plans.plans,
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
  className?: string
}) {
  const Icon = props.icon
  return (
    <div className={cn('bg-card overflow-hidden rounded-xl border', props.className)}>
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
            {t(`${opt} days`)}
          </button>
        )
      })}
    </div>
  )
}

const RANK_BADGES = ['bg-accent-amber', 'bg-muted-foreground', 'bg-accent-coral']

function PlanSalesPanel(props: { plans: PlanSales[] }) {
  const { t } = useTranslation()
  const max = Math.max(1, ...props.plans.map((p) => p.sales_count))
  if (!props.plans.length) {
    return (
      <p className='text-muted-foreground py-6 text-center text-sm'>
        {t('No plan sales in this period')}
      </p>
    )
  }
  return (
    <div className='flex flex-col gap-3'>
      {props.plans.map((p, i) => (
        <div key={p.plan_id} className='flex flex-col gap-1.5'>
          <div className='flex items-center justify-between gap-2 text-sm'>
            <span className='flex min-w-0 items-center gap-2'>
              <span
                className={cn(
                  'flex size-5 shrink-0 items-center justify-center rounded-full text-[11px] font-bold text-white',
                  i < 3 ? RANK_BADGES[i] : 'bg-muted-foreground/50'
                )}
              >
                {i + 1}
              </span>
              <span className='truncate font-medium'>{p.title}</span>
            </span>
            <span className='shrink-0 tabular-nums'>
              {formatNumber(p.sales_count)} {t('sold')}
            </span>
          </div>
          <div className='bg-muted h-2 overflow-hidden rounded-full'>
            <div
              className='bg-accent-blue h-full rounded-full'
              style={{ width: `${(p.sales_count / max) * 100}%` }}
            />
          </div>
          <div className='text-muted-foreground text-right text-xs'>
            {formatMoney(p.revenue)}
          </div>
        </div>
      ))}
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
  const { overview, revenue, growth, providers, plans } = data

  const revenueSeries = useMemo(
    () => revenue.map((p) => ({ label: dayLabel(p.date), value: p.revenue })),
    [revenue]
  )
  const growthSeries = useMemo(
    () => growth.map((p) => ({ label: dayLabel(p.date), value: p.cumulative })),
    [growth]
  )
  const revenueSpark = useMemo(() => revenue.map((p) => p.revenue), [revenue])
  const newUsersSpark = useMemo(() => growth.map((p) => p.new_users), [growth])

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
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
        <StatCard
          title={t('Platform revenue')}
          value={overview ? formatMoney(overview.total_revenue) : '-'}
          description={t('Top-up + subscription, period')}
          icon={TrendingUp}
          tone='green'
          loading={loading}
          sparkline={revenueSpark}
          sparklineVariant='line'
          details={[
            {
              label: t('Top-up'),
              value: overview ? formatMoney(overview.revenue) : '-',
            },
            {
              label: t('Subscription'),
              value: overview ? formatMoney(overview.subscription_revenue) : '-',
            },
          ]}
        />
        <StatCard
          title={t('Online users')}
          value={overview ? formatNumber(overview.online_users) : '-'}
          description={t('Active in last 15 min')}
          icon={Wifi}
          tone='coral'
          loading={loading}
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
          title={t('Subscription sales')}
          value={overview ? formatNumber(overview.subscription_orders) : '-'}
          description={t('Plan orders in period')}
          icon={Package}
          tone='amber'
          loading={loading}
          details={[
            {
              label: t('Success'),
              value: overview ? String(overview.subscription_success) : '-',
              tone: 'success',
            },
            {
              label: t('Revenue'),
              value: overview ? formatMoney(overview.subscription_revenue) : '-',
            },
          ]}
        />
        <StatCard
          title={t('Top-up orders')}
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
        <Panel title={t('Platform revenue trend')} icon={TrendingUp}>
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

      {/* Plan ranking + providers + channel health */}
      <div className='grid grid-cols-1 gap-4 xl:grid-cols-3'>
        <Panel title={t('Plan sales ranking')} icon={Package}>
          {loading ? (
            <Skeleton className='h-40 w-full rounded-lg' />
          ) : (
            <PlanSalesPanel plans={plans} />
          )}
        </Panel>
        <Panel title={t('Payment providers')} icon={CreditCard}>
          {loading ? (
            <Skeleton className='h-40 w-full rounded-lg' />
          ) : (
            <ProvidersPanel providers={providers} />
          )}
        </Panel>
        <Panel title={t('Channel health')} icon={Server}>
          {loading ? (
            <Skeleton className='h-40 w-full rounded-lg' />
          ) : (
            <ChannelHealthPanel overview={overview} />
          )}
        </Panel>
      </div>
    </div>
  )
}
