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
import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  ChevronLeft,
  ChevronRight,
  GitFork,
  Pencil,
  Plus,
  RefreshCcw,
  Search,
  Trash2,
  Wallet,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { formatCurrencyUSD } from '@/lib/format'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { SectionPageLayout } from '@/components/layout'
import { AgentCombobox, formatAgentLabel } from './agent-combobox'
import {
  adminAdjustAgentBalance,
  adminGetAttribution,
  adminGetCoupons,
  adminGetGiftRules,
  adminGetOpsAuthorizations,
  adminGetPackages,
  adminGetProfit,
  adminGetSubscriptionPlans,
  adminGetUserOptions,
  adminGrantOpsAuthorization,
  adminIssueCoupons,
  adminSaveAgent,
  adminSaveGiftRule,
  adminSavePackage,
  adminRevokeOpsAuthorization,
  adminSearchAgents,
  adminUpdateAgentStatus,
  adminUpdateGiftRuleStatus,
  adminUpdatePackageStatus,
  fetchAdminDistributionOrders,
} from './api'
import { DateRangeField } from './date-fields'
import {
  distributionCouponSourceLabel,
  distributionOrderStatusLabel,
  distributionOrderStatuses,
  distributionOrderTypeLabel,
  distributionOrderTypes,
  distributionPriceTargetLabel,
  distributionPriceTypeLabel,
  distributionSourceTypeLabel,
  distributionStatusLabel,
} from './labels'
import type {
  DistributionAgent,
  DistributionAttributionLog,
  DistributionCoupon,
  DistributionGiftRule,
  DistributionOpsAuthorization,
  DistributionOrderRecord,
  DistributionPackage,
  DistributionProfit,
  DistributionSubscriptionPlanRecord,
  DistributionUserOption,
} from './types'

type AdminTab =
  | 'agents'
  | 'packages'
  | 'orders'
  | 'coupons'
  | 'gift'
  | 'ops'
  | 'profit'
  | 'attribution'

type DialogKind =
  | 'agent'
  | 'balance'
  | 'package'
  | 'price'
  | 'coupon'
  | 'gift'
  | 'ops'
  | null

const pageSize = 10

function formatTime(value?: number) {
  if (!value) return '-'
  return new Date(value * 1000).toLocaleString()
}

function StatusBadge({ status }: { status?: string }) {
  const { t } = useTranslation()
  return <Badge variant='secondary'>{distributionStatusLabel(status, t)}</Badge>
}

function formatMoney(value?: number) {
  return formatCurrencyUSD(value)
}

function packageTitle(pkg: DistributionPackage) {
  return pkg.subscription_title || pkg.name || '-'
}

function packageSubtitle(pkg: DistributionPackage) {
  return pkg.subscription_subtitle || pkg.description || ''
}

function planLabel(record?: DistributionSubscriptionPlanRecord) {
  if (!record) return '-'
  const subtitle = record.plan.subtitle?.trim()
  return subtitle ? `${record.plan.title} · ${subtitle}` : record.plan.title
}

function formatUserLabel(user?: DistributionUserOption) {
  if (!user) return '-'
  const displayName = user.display_name?.trim()
  const username = user.username?.trim()
  const email = user.email?.trim()
  if (displayName && username) return `${displayName} (${username})`
  return displayName || username || email || '-'
}

function formatFallbackId(id?: number) {
  return id ? `#${id}` : '-'
}

function SelectDisplay({
  label,
  placeholder,
}: {
  label?: string
  placeholder: string
}) {
  return (
    <span
      data-slot='select-value'
      className='flex min-w-0 flex-1 items-center truncate text-left'
    >
      {label || placeholder}
    </span>
  )
}

function TablePager({
  page,
  total,
  onPageChange,
}: {
  page: number
  total: number
  onPageChange: (page: number) => void
}) {
  const { t } = useTranslation()
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  return (
    <div className='flex items-center justify-between gap-3 border-t pt-3 text-sm'>
      <div className='text-muted-foreground'>
        {t('Page')} {page} / {totalPages}
      </div>
      <div className='flex items-center gap-2'>
        <Button
          variant='outline'
          size='sm'
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeft className='mr-1 h-4 w-4' />
          {t('Previous page')}
        </Button>
        <Button
          variant='outline'
          size='sm'
          disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}
        >
          {t('Next page')}
          <ChevronRight className='ml-1 h-4 w-4' />
        </Button>
      </div>
    </div>
  )
}

function EmptyTableRow({
  colSpan,
  loading,
}: {
  colSpan: number
  loading?: boolean
}) {
  const { t } = useTranslation()
  return (
    <tr>
      <td
        colSpan={colSpan}
        className='text-muted-foreground py-8 text-center text-sm'
      >
        {loading ? (
          <span className='animate-pulse'>{t('Loading')}</span>
        ) : (
          t('No data')
        )}
      </td>
    </tr>
  )
}

function apiActionError(message: string | undefined, fallback: string) {
  return message?.trim() || fallback
}

function packageActionError(message?: string) {
  // Result is passed through t() at the call site, so the fallback stays an i18n key.
  const fallback = apiActionError(message, 'Action failed')
  const normalized = fallback.toLowerCase()
  if (
    normalized.includes('subscription plan already has') ||
    normalized.includes('distribution package subscription plan already exists')
  ) {
    return 'This subscription plan already has a package.'
  }
  if (normalized.includes('tier 1 agent price must be less')) {
    return 'Tier 1 agent price must be less than or equal to tier 2 agent price.'
  }
  if (normalized.includes('agent prices must be less')) {
    return 'Agent prices must be less than or equal to the subscription plan price.'
  }
  return fallback
}

function subscriptionPlanPriceAmount(
  record?: DistributionSubscriptionPlanRecord
) {
  if (!record) return 0
  return Number(Number(record.plan.price_amount || 0).toFixed(2))
}

function nextDistributionStatus(status?: string) {
  return status === 'enabled' ? 'disabled' : 'enabled'
}

function TabCard({
  title,
  action,
  children,
}: {
  title: string
  action?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <Card>
      <CardHeader className='flex flex-row items-center justify-between gap-3'>
        <CardTitle>{title}</CardTitle>
        {action}
      </CardHeader>
      <CardContent className='space-y-3 overflow-x-auto'>
        {children}
      </CardContent>
    </Card>
  )
}

const emptyAgentForm = {
  id: '',
  user_id: '',
  name: '',
  balance: '',
  parent_agent_id: '0',
  level: '2',
  contact: '',
  remark: '',
  status: 'enabled',
}

const emptyPackageForm = {
  id: '',
  subscription_plan_id: '',
  status: 'enabled',
  agent_price: '',
  secondary_agent_price: '',
  sort_order: '',
}

const emptyPriceForm = {
  target_type: 'level',
  package_id: '',
  customer_user_id: '',
  agent_level: '1',
  price_type: 'fixed',
  price_value: '',
  status: 'enabled',
  remark: '',
}

const emptyOrderFilters = {
  keyword: '',
  order_type: '',
  plan_id: '',
  status: '',
  start_time: '',
  end_time: '',
}

const emptyGiftForm = {
  name: '',
  package_id: '',
  gift_package_id: '',
  trigger_quantity: '',
  gift_quantity: '',
  starts_at: '',
  expires_at: '',
  status: 'enabled',
}

type CouponItemDraft = {
  count: string
  amount: string
  validity_days: string
}

const emptyCouponItem: CouponItemDraft = {
  count: '1',
  amount: '',
  validity_days: '7',
}

const MAX_COUPONS_PER_ISSUE = 100

function couponItemsTotalCount(items: CouponItemDraft[]) {
  return items.reduce((sum, item) => sum + (Number(item.count) || 0), 0)
}

export function AgentAdmin() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<AdminTab>('agents')
  const [dialogKind, setDialogKind] = useState<DialogKind>(null)
  const [loading, setLoading] = useState(false)
  const [agents, setAgents] = useState<DistributionAgent[]>([])
  const [agentOptions, setAgentOptions] = useState<DistributionAgent[]>([])
  const [packages, setPackages] = useState<DistributionPackage[]>([])
  const [packageOptions, setPackageOptions] = useState<DistributionPackage[]>(
    []
  )
  const [subscriptionPlans, setSubscriptionPlans] = useState<
    DistributionSubscriptionPlanRecord[]
  >([])
  const [users, setUsers] = useState<DistributionUserOption[]>([])
  const [profit, setProfit] = useState<DistributionProfit[]>([])
  const [attribution, setAttribution] = useState<DistributionAttributionLog[]>(
    []
  )
  const [giftRules, setGiftRules] = useState<DistributionGiftRule[]>([])
  const [opsAuth, setOpsAuth] = useState<DistributionOpsAuthorization[]>([])
  const [agentPage, setAgentPage] = useState(1)
  const [agentTotal, setAgentTotal] = useState(0)
  const [packagePage, setPackagePage] = useState(1)
  const [packageTotal, setPackageTotal] = useState(0)
  const [profitPage, setProfitPage] = useState(1)
  const [profitTotal, setProfitTotal] = useState(0)
  const [attributionPage, setAttributionPage] = useState(1)
  const [attributionTotal, setAttributionTotal] = useState(0)
  const [giftRulePage, setGiftRulePage] = useState(1)
  const [giftRuleTotal, setGiftRuleTotal] = useState(0)
  const [orders, setOrders] = useState<DistributionOrderRecord[]>([])
  const [orderPage, setOrderPage] = useState(1)
  const [orderTotal, setOrderTotal] = useState(0)
  const [orderFilterDraft, setOrderFilterDraft] = useState(emptyOrderFilters)
  const [orderFilters, setOrderFilters] = useState(emptyOrderFilters)
  const [opsPage, setOpsPage] = useState(1)
  const [opsTotal, setOpsTotal] = useState(0)
  const [coupons, setCoupons] = useState<DistributionCoupon[]>([])
  const [couponPage, setCouponPage] = useState(1)
  const [couponTotal, setCouponTotal] = useState(0)
  const [couponFilterAgentId, setCouponFilterAgentId] = useState('0')
  const [couponFilterAgent, setCouponFilterAgent] =
    useState<DistributionAgent | null>(null)
  const [couponIssueAgentId, setCouponIssueAgentId] = useState('')
  const [couponIssueAgent, setCouponIssueAgent] =
    useState<DistributionAgent | null>(null)
  const [couponItems, setCouponItems] = useState<CouponItemDraft[]>([
    { ...emptyCouponItem },
  ])
  const [couponRemark, setCouponRemark] = useState('')
  const [balanceAgentId, setBalanceAgentId] = useState('')
  const [balanceAgent, setBalanceAgent] = useState<DistributionAgent | null>(
    null
  )
  const [balanceDelta, setBalanceDelta] = useState('')
  const [balanceRemark, setBalanceRemark] = useState('')
  const [agentParentAgent, setAgentParentAgent] =
    useState<DistributionAgent | null>(null)
  const [agentDialogMode, setAgentDialogMode] = useState<'create' | 'edit'>(
    'create'
  )
  const [packageDialogMode, setPackageDialogMode] = useState<'create' | 'edit'>(
    'create'
  )
  const [agentForm, setAgentForm] = useState(emptyAgentForm)
  const [packageForm, setPackageForm] = useState(emptyPackageForm)
  const [priceForm, setPriceForm] = useState(emptyPriceForm)
  const [giftForm, setGiftForm] = useState(emptyGiftForm)
  const [opsUserId, setOpsUserId] = useState('')
  const [opsRemark, setOpsRemark] = useState('')

  const agentLabelMap = useMemo(
    () =>
      new Map(agentOptions.map((agent) => [agent.id, formatAgentLabel(agent)])),
    [agentOptions]
  )
  const packageLabelMap = useMemo(
    () => new Map(packageOptions.map((pkg) => [pkg.id, packageTitle(pkg)])),
    [packageOptions]
  )
  const subscriptionPlanLabelMap = useMemo(
    () =>
      new Map(
        subscriptionPlans.map((record) => [record.plan.id, planLabel(record)])
      ),
    [subscriptionPlans]
  )
  const userLabelMap = useMemo(
    () => new Map(users.map((user) => [user.id, formatUserLabel(user)])),
    [users]
  )
  const selectedPackageSubscriptionPlan = useMemo(
    () =>
      subscriptionPlans.find(
        (record) => record.plan.id === Number(packageForm.subscription_plan_id)
      ),
    [packageForm.subscription_plan_id, subscriptionPlans]
  )
  const selectedPackageSubscriptionPriceAmount = subscriptionPlanPriceAmount(
    selectedPackageSubscriptionPlan
  )

  const resetAgentForm = useCallback(() => {
    setAgentForm(emptyAgentForm)
    setAgentParentAgent(null)
    setAgentDialogMode('create')
  }, [])

  const resetPackageForm = useCallback(() => {
    setPackageForm(emptyPackageForm)
    setPackageDialogMode('create')
  }, [])

  const resetGiftForm = useCallback(() => {
    setGiftForm(emptyGiftForm)
  }, [])

  const resetBalanceForm = useCallback(() => {
    setBalanceAgentId('')
    setBalanceAgent(null)
    setBalanceDelta('')
    setBalanceRemark('')
  }, [])

  const resetOpsForm = useCallback(() => {
    setOpsUserId('')
    setOpsRemark('')
  }, [])

  const resetCouponForm = useCallback(() => {
    setCouponIssueAgentId('')
    setCouponIssueAgent(null)
    setCouponItems([{ ...emptyCouponItem }])
    setCouponRemark('')
  }, [])

  const closeDialog = useCallback(() => {
    setDialogKind(null)
  }, [])

  const loadAgentOptions = useCallback(async () => {
    const res = await adminSearchAgents({ p: 1, page_size: 100 }).catch(
      () => null
    )
    if (res?.success) setAgentOptions(res.data?.items || [])
  }, [])

  const loadPackageOptions = useCallback(async () => {
    const res = await adminGetPackages({ p: 1, page_size: 100 }).catch(
      () => null
    )
    if (res?.success) setPackageOptions(res.data?.items || [])
  }, [])

  const loadSubscriptionPlans = useCallback(async () => {
    const res = await adminGetSubscriptionPlans().catch(() => null)
    if (res?.success) {
      setSubscriptionPlans(
        (res.data || []).filter((record) => record.plan.enabled)
      )
    }
  }, [])

  const loadUsers = useCallback(async () => {
    const res = await adminGetUserOptions().catch(() => null)
    if (res?.success) setUsers(res.data?.items || [])
  }, [])

  const refreshTab = useCallback(
    async (tab: AdminTab) => {
      setLoading(true)
      try {
        if (tab === 'agents') {
          const res = await adminSearchAgents({
            p: agentPage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setAgents(res.data?.items || [])
            setAgentTotal(res.data?.total || 0)
            setAgentOptions((current) => {
              const map = new Map(current.map((agent) => [agent.id, agent]))
              for (const agent of res.data?.items || [])
                map.set(agent.id, agent)
              return [...map.values()]
            })
          }
        }
        if (tab === 'packages') {
          await loadSubscriptionPlans()
          const res = await adminGetPackages({
            p: packagePage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setPackages(res.data?.items || [])
            setPackageTotal(res.data?.total || 0)
            setPackageOptions((current) => {
              const map = new Map(current.map((pkg) => [pkg.id, pkg]))
              for (const pkg of res.data?.items || []) map.set(pkg.id, pkg)
              return [...map.values()]
            })
          }
        }
        if (tab === 'orders') {
          await loadSubscriptionPlans()
          const res = await fetchAdminDistributionOrders({
            p: orderPage,
            page_size: pageSize,
            keyword: orderFilters.keyword,
            order_type: orderFilters.order_type || undefined,
            plan_id: orderFilters.plan_id
              ? Number(orderFilters.plan_id)
              : undefined,
            status: orderFilters.status || undefined,
            start_time: orderFilters.start_time
              ? Number(orderFilters.start_time)
              : undefined,
            end_time: orderFilters.end_time
              ? Number(orderFilters.end_time)
              : undefined,
          }).catch(() => null)
          if (res?.success) {
            setOrders(res.data?.items || [])
            setOrderTotal(res.data?.total || 0)
          }
        }
        if (tab === 'coupons') {
          await loadAgentOptions()
          const res = await adminGetCoupons({
            agent_id: Number(couponFilterAgentId) || 0,
            p: couponPage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setCoupons(res.data?.items || [])
            setCouponTotal(res.data?.total || 0)
          }
        }
        if (tab === 'gift') {
          await loadPackageOptions()
          const res = await adminGetGiftRules({
            p: giftRulePage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setGiftRules(res.data?.items || [])
            setGiftRuleTotal(res.data?.total || 0)
          }
        }
        if (tab === 'ops') {
          await loadUsers()
          const res = await adminGetOpsAuthorizations({
            p: opsPage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setOpsAuth(res.data?.items || [])
            setOpsTotal(res.data?.total || 0)
          }
        }
        if (tab === 'profit') {
          await loadAgentOptions()
          const res = await adminGetProfit({
            p: profitPage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setProfit(res.data?.items || [])
            setProfitTotal(res.data?.total || 0)
          }
        }
        if (tab === 'attribution') {
          await loadUsers()
          const res = await adminGetAttribution({
            p: attributionPage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setAttribution(res.data?.items || [])
            setAttributionTotal(res.data?.total || 0)
          }
        }
      } finally {
        setLoading(false)
      }
    },
    [
      agentPage,
      attributionPage,
      couponFilterAgentId,
      couponPage,
      giftRulePage,
      loadAgentOptions,
      loadPackageOptions,
      loadSubscriptionPlans,
      loadUsers,
      opsPage,
      orderFilters,
      orderPage,
      packagePage,
      profitPage,
    ]
  )

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void refreshTab(activeTab)
  }, [activeTab, refreshTab])

  const openDialog = useCallback(
    async (kind: DialogKind) => {
      setDialogKind(kind)
      if (kind === 'agent') {
        resetAgentForm()
        setAgentDialogMode('create')
      }
      if (kind === 'coupon') resetCouponForm()
      if (kind === 'agent' || kind === 'ops') await loadUsers()
      if (kind === 'balance') await loadAgentOptions()
      if (kind === 'gift') await loadPackageOptions()
      if (kind === 'package') await loadSubscriptionPlans()
    },
    [
      loadAgentOptions,
      loadPackageOptions,
      loadSubscriptionPlans,
      loadUsers,
      resetAgentForm,
      resetCouponForm,
    ]
  )

  const openEditAgentDialog = useCallback(
    async (row: DistributionAgent) => {
      setAgentDialogMode('edit')
      setAgentForm({
        id: String(row.id),
        user_id: String(row.user_id),
        name: row.name,
        balance: String(row.balance),
        parent_agent_id: String(row.parent_agent_id || 0),
        level: String(row.level || 2),
        contact: row.contact,
        remark: row.remark,
        status: row.status,
      })
      setAgentParentAgent(
        agentOptions.find((agent) => agent.id === row.parent_agent_id) || null
      )
      setDialogKind('agent')
      await loadUsers()
    },
    [agentOptions, loadUsers]
  )

  const openEditPackageDialog = useCallback(
    async (row: DistributionPackage) => {
      setPackageDialogMode('edit')
      setPackageForm({
        id: String(row.id),
        subscription_plan_id: String(row.subscription_plan_id || ''),
        status: row.status,
        agent_price: String(row.agent_price),
        secondary_agent_price: String(row.secondary_agent_price),
        sort_order: String(row.sort_order),
      })
      setDialogKind('package')
      await loadSubscriptionPlans()
    },
    [loadSubscriptionPlans]
  )

  async function handleSaveAgent() {
    const payload =
      agentDialogMode === 'edit'
        ? {
            id: Number(agentForm.id),
            parent_agent_id: Number(agentForm.parent_agent_id || 0),
            level: Number(agentForm.level || 2),
          }
        : {
            user_id: Number(agentForm.user_id),
            name: agentForm.name,
            balance: Number(agentForm.balance),
            parent_agent_id: Number(agentForm.parent_agent_id),
            level: Number(agentForm.level || 2),
            contact: agentForm.contact,
            remark: agentForm.remark,
            status: agentForm.status,
          }
    const res = await adminSaveAgent(payload)
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    resetAgentForm()
    closeDialog()
    setAgentPage(1)
    await refreshTab('agents')
  }

  async function handleAdjustBalance() {
    const res = await adminAdjustAgentBalance(
      Number(balanceAgentId),
      Number(balanceDelta),
      balanceRemark
    )
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    resetBalanceForm()
    closeDialog()
    await refreshTab('agents')
  }

  async function handleSavePackage() {
    const subscriptionPlanId = Number(packageForm.subscription_plan_id)
    const agentPrice = Number(packageForm.agent_price)
    const secondaryAgentPrice = Number(packageForm.secondary_agent_price)
    if (!subscriptionPlanId || !selectedPackageSubscriptionPlan) {
      toast.error(t('Please select a subscription plan.'))
      return
    }
    if (agentPrice > secondaryAgentPrice) {
      toast.error(
        t(
          'Tier 1 agent price must be less than or equal to tier 2 agent price.'
        )
      )
      return
    }
    if (
      agentPrice > selectedPackageSubscriptionPriceAmount ||
      secondaryAgentPrice > selectedPackageSubscriptionPriceAmount
    ) {
      toast.error(
        t(
          'Agent prices must be less than or equal to the subscription plan price.'
        )
      )
      return
    }
    const res = await adminSavePackage({
      id: packageDialogMode === 'edit' ? Number(packageForm.id) : undefined,
      subscription_plan_id: subscriptionPlanId,
      status: packageForm.status,
      agent_price: agentPrice,
      secondary_agent_price: secondaryAgentPrice,
      sort_order: Number(packageForm.sort_order),
    })
    if (!res.success) {
      toast.error(t(packageActionError(res.message)))
      return
    }
    toast.success(t('Saved'))
    resetPackageForm()
    closeDialog()
    setPackagePage(1)
    await refreshTab('packages')
  }

  async function handleSaveGiftRule() {
    const res = await adminSaveGiftRule({
      name: giftForm.name,
      package_id: Number(giftForm.package_id),
      gift_package_id: Number(giftForm.gift_package_id),
      trigger_quantity: Number(giftForm.trigger_quantity),
      gift_quantity: Number(giftForm.gift_quantity),
      starts_at: Number(giftForm.starts_at),
      expires_at: Number(giftForm.expires_at),
      status: giftForm.status,
    })
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    resetGiftForm()
    closeDialog()
    setGiftRulePage(1)
    await refreshTab('gift')
  }

  async function handleIssueCoupons() {
    if (!couponIssueAgentId || !Number(couponIssueAgentId)) {
      toast.error(t('Select agent'))
      return
    }
    const items = couponItems.map((item) => ({
      count: Number(item.count),
      amount: Number(item.amount),
      validity_days: Number(item.validity_days),
    }))
    if (
      items.some(
        (item) =>
          !Number.isInteger(item.count) ||
          item.count <= 0 ||
          Number.isNaN(item.amount) ||
          item.amount <= 0 ||
          !Number.isInteger(item.validity_days) ||
          item.validity_days <= 0
      )
    ) {
      toast.error(
        t('Each item requires count, amount and validity days greater than 0.')
      )
      return
    }
    const totalCount = items.reduce((sum, item) => sum + item.count, 0)
    if (totalCount > MAX_COUPONS_PER_ISSUE) {
      toast.error(t('A maximum of 100 coupons can be issued at once.'))
      return
    }
    const res = await adminIssueCoupons({
      agent_id: Number(couponIssueAgentId),
      items,
      remark: couponRemark.trim() || undefined,
    })
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Coupons issued'))
    resetCouponForm()
    closeDialog()
    setCouponPage(1)
    await refreshTab('coupons')
  }

  async function handleGrantOps() {
    const res = await adminGrantOpsAuthorization(Number(opsUserId), opsRemark)
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Granted'))
    resetOpsForm()
    closeDialog()
    setOpsPage(1)
    await refreshTab('ops')
  }

  async function handleToggleAgentStatus(row: DistributionAgent) {
    const res = await adminUpdateAgentStatus(
      row.id,
      nextDistributionStatus(row.status)
    )
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    await refreshTab('agents')
  }

  async function handleTogglePackageStatus(row: DistributionPackage) {
    const res = await adminUpdatePackageStatus(
      row.id,
      nextDistributionStatus(row.status)
    )
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    await refreshTab('packages')
  }

  async function handleToggleGiftRuleStatus(row: DistributionGiftRule) {
    const res = await adminUpdateGiftRuleStatus(
      row.id,
      nextDistributionStatus(row.status)
    )
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    await refreshTab('gift')
  }

  function handleSearchOrders() {
    setOrderPage(1)
    setOrderFilters({ ...orderFilterDraft })
  }

  function handleResetOrderFilters() {
    setOrderFilterDraft(emptyOrderFilters)
    setOrderFilters({ ...emptyOrderFilters })
    setOrderPage(1)
  }

  async function handleRevokeOps(row: DistributionOpsAuthorization) {
    const res = await adminRevokeOpsAuthorization(row.user_id)
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Saved'))
    await refreshTab('ops')
  }

  async function handleReGrantOps(row: DistributionOpsAuthorization) {
    const res = await adminGrantOpsAuthorization(row.user_id, '')
    if (!res.success) {
      toast.error(apiActionError(res.message, t('Action failed')))
      return
    }
    toast.success(t('Granted'))
    await refreshTab('ops')
  }

  const dialogTitle =
    dialogKind === 'agent'
      ? agentDialogMode === 'edit'
        ? t('Edit Agent')
        : t('Add Agent')
      : dialogKind === 'balance'
        ? t('Adjust Balance')
        : dialogKind === 'package'
          ? packageDialogMode === 'edit'
            ? t('Edit Prices')
            : t('Add Package')
          : dialogKind === 'coupon'
            ? t('Issue Coupons')
            : dialogKind === 'gift'
              ? t('Add Gift Rule')
              : dialogKind === 'ops'
                ? t('Grant Access')
                : ''

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Agent Management')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Button variant='outline' onClick={() => void refreshTab(activeTab)}>
          <RefreshCcw className='mr-2 h-4 w-4' />
          {loading ? t('Loading') : t('Refresh')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <Tabs
          value={activeTab}
          onValueChange={(value) => setActiveTab(value as AdminTab)}
          className='space-y-4'
        >
          <TabsList className='max-w-full flex-wrap justify-start group-data-horizontal/tabs:h-auto'>
            <TabsTrigger value='agents'>{t('Agents')}</TabsTrigger>
            <TabsTrigger value='packages'>{t('Packages')}</TabsTrigger>
            <TabsTrigger value='orders'>{t('Order Lookup')}</TabsTrigger>
            <TabsTrigger value='coupons'>{t('Coupons')}</TabsTrigger>
            <TabsTrigger value='gift'>{t('Gift Rules')}</TabsTrigger>
            <TabsTrigger value='ops'>{t('Operations Access')}</TabsTrigger>
            <TabsTrigger value='profit'>{t('Profit')}</TabsTrigger>
            <TabsTrigger value='attribution'>
              {t('Attribution Logs')}
            </TabsTrigger>
          </TabsList>

          <TabsContent value='agents' className='space-y-4'>
            <TabCard
              title={t('Agents')}
              action={
                <div className='flex flex-wrap gap-2'>
                  <Button
                    variant='outline'
                    onClick={() => void openDialog('balance')}
                  >
                    <Wallet className='mr-2 h-4 w-4' />
                    {t('Adjust Balance')}
                  </Button>
                  <Button onClick={() => void openDialog('agent')}>
                    <Plus className='mr-2 h-4 w-4' />
                    {t('Add Agent')}
                  </Button>
                </div>
              }
            >
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Name')}</th>
                    <th>{t('User')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Balance')}</th>
                    <th>{t('Agent Level')}</th>
                    <th>{t('Parent')}</th>
                    <th>{t('Actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {agents.length === 0 && (
                    <EmptyTableRow colSpan={7} loading={loading} />
                  )}
                  {agents.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>{row.name}</td>
                      <td>{formatAgentLabel(row)}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatMoney(row.balance)}</td>
                      <td>{`${t('Level')} ${row.level || 2}`}</td>
                      <td>
                        {agentLabelMap.get(row.parent_agent_id) ||
                          formatFallbackId(row.parent_agent_id)}
                      </td>
                      <td className='space-x-2 py-2'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void openEditAgentDialog(row)}
                        >
                          <GitFork className='mr-2 h-4 w-4' />
                          {t('Adjust Hierarchy')}
                        </Button>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void handleToggleAgentStatus(row)}
                        >
                          {t(row.status === 'enabled' ? 'Disable' : 'Enable')}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={agentPage}
                total={agentTotal}
                onPageChange={setAgentPage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='packages' className='space-y-4'>
            <TabCard
              title={t('Packages')}
              action={
                <Button onClick={() => void openDialog('package')}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Add Package')}
                </Button>
              }
            >
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Subscription Plan')}</th>
                    <th>{t('Tier 1 Agent Price')}</th>
                    <th>{t('Tier 2 Agent Price')}</th>
                    <th>{t('Sort Order')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {packages.length === 0 && (
                    <EmptyTableRow colSpan={6} loading={loading} />
                  )}
                  {packages.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>
                        <div className='font-medium'>{packageTitle(row)}</div>
                        {packageSubtitle(row) && (
                          <div className='text-muted-foreground text-xs'>
                            {packageSubtitle(row)}
                          </div>
                        )}
                      </td>
                      <td>{formatMoney(row.agent_price)}</td>
                      <td>{formatMoney(row.secondary_agent_price)}</td>
                      <td>{row.sort_order}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td className='space-x-2 py-2'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void openEditPackageDialog(row)}
                        >
                          <Pencil className='mr-2 h-4 w-4' />
                          {t('Edit Prices')}
                        </Button>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void handleTogglePackageStatus(row)}
                        >
                          {t(row.status === 'enabled' ? 'Disable' : 'Enable')}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={packagePage}
                total={packageTotal}
                onPageChange={setPackagePage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='orders' className='space-y-4'>
            <TabCard title={t('Order Lookup')}>
              <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-5'>
                <div className='space-y-2'>
                  <Label>{t('Buyer')}</Label>
                  <Input
                    value={orderFilterDraft.keyword}
                    placeholder={t(
                      'Search buyer username, email or display name'
                    )}
                    onChange={(e) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        keyword: e.target.value,
                      }))
                    }
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Order Type')}</Label>
                  <Select
                    value={orderFilterDraft.order_type || 'all'}
                    onValueChange={(value) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        order_type: !value || value === 'all' ? '' : value,
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={
                          orderFilterDraft.order_type
                            ? distributionOrderTypeLabel(
                                orderFilterDraft.order_type,
                                t
                              )
                            : t('All Types')
                        }
                        placeholder={t('All Types')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        <SelectItem value='all'>{t('All Types')}</SelectItem>
                        {distributionOrderTypes.map((type) => (
                          <SelectItem key={type} value={type}>
                            {distributionOrderTypeLabel(type, t)}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Subscription Plan')}</Label>
                  <Select
                    value={orderFilterDraft.plan_id || 'all'}
                    onValueChange={(value) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        plan_id: !value || value === 'all' ? '' : value,
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={
                          orderFilterDraft.plan_id
                            ? subscriptionPlanLabelMap.get(
                                Number(orderFilterDraft.plan_id)
                              )
                            : t('All Plans')
                        }
                        placeholder={t('All Plans')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        <SelectItem value='all'>{t('All Plans')}</SelectItem>
                        {subscriptionPlans.map((record) => (
                          <SelectItem
                            key={record.plan.id}
                            value={String(record.plan.id)}
                          >
                            {planLabel(record)}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Status')}</Label>
                  <Select
                    value={orderFilterDraft.status || 'all'}
                    onValueChange={(value) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        status: !value || value === 'all' ? '' : value,
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={
                          orderFilterDraft.status
                            ? distributionOrderStatusLabel(
                                orderFilterDraft.status,
                                t
                              )
                            : t('All Status')
                        }
                        placeholder={t('All Status')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        <SelectItem value='all'>{t('All Status')}</SelectItem>
                        {distributionOrderStatuses.map((status) => (
                          <SelectItem key={status} value={status}>
                            {distributionOrderStatusLabel(status, t)}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='md:col-span-2'>
                  <DateRangeField
                    startLabel='Start Time'
                    endLabel='End Time'
                    startValue={orderFilterDraft.start_time}
                    endValue={orderFilterDraft.end_time}
                    onStartChange={(value) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        start_time: value,
                      }))
                    }
                    onEndChange={(value) =>
                      setOrderFilterDraft((state) => ({
                        ...state,
                        end_time: value,
                      }))
                    }
                    rangePicker
                  />
                </div>
                <div className='flex items-end gap-2'>
                  <Button onClick={handleSearchOrders}>
                    <Search className='mr-2 h-4 w-4' />
                    {t('Search')}
                  </Button>
                  <Button variant='outline' onClick={handleResetOrderFilters}>
                    {t('Reset')}
                  </Button>
                </div>
              </div>
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Order No')}</th>
                    <th>{t('Order Type')}</th>
                    <th>{t('Buyer')}</th>
                    <th>{t('Subscription Plan')}</th>
                    <th>{t('Activation Code')}</th>
                    <th>{t('Original Amount')}</th>
                    <th>{t('Paid Amount')}</th>
                    <th>{t('Commission')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Created At')}</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.length === 0 && (
                    <EmptyTableRow colSpan={10} loading={loading} />
                  )}
                  {orders.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2 font-mono text-xs'>{row.order_no}</td>
                      <td>{distributionOrderTypeLabel(row.order_type, t)}</td>
                      <td>
                        <div className='font-medium'>
                          {row.buy_user_name?.trim() ||
                            row.buyer_display_name?.trim() ||
                            row.buyer_username?.trim() ||
                            '-'}
                        </div>
                        {row.buyer_email && (
                          <div className='text-muted-foreground text-xs'>
                            {row.buyer_email}
                          </div>
                        )}
                      </td>
                      <td>
                        {row.subscription_title || row.package_name || '-'}
                      </td>
                      <td className='font-mono text-xs'>
                        {row.agent_active_code || '-'}
                      </td>
                      <td>{formatMoney(row.original_amount)}</td>
                      <td>{formatMoney(row.paid_amount)}</td>
                      <td>{formatMoney(row.commission_amount)}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatTime(row.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={orderPage}
                total={orderTotal}
                onPageChange={setOrderPage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='coupons' className='space-y-4'>
            <TabCard
              title={t('Coupons')}
              action={
                <Button onClick={() => void openDialog('coupon')}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Issue Coupons')}
                </Button>
              }
            >
              <div className='flex flex-col gap-2 md:flex-row md:items-end'>
                <div className='space-y-2 md:w-80'>
                  <Label>{t('Agent')}</Label>
                  <AgentCombobox
                    value={couponFilterAgentId}
                    selectedAgent={couponFilterAgent || undefined}
                    onValueChange={(value) => {
                      setCouponPage(1)
                      setCouponFilterAgentId(value || '0')
                    }}
                    onAgentSelected={setCouponFilterAgent}
                    placeholder={t('All Agents')}
                    includeEmpty
                    emptyLabel={t('All Agents')}
                  />
                </div>
              </div>
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Code')}</th>
                    <th>{t('Agent')}</th>
                    <th>{t('Amount')}</th>
                    <th>{t('Source')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Expires At')}</th>
                    <th>{t('Used At')}</th>
                    <th>{t('Remark')}</th>
                  </tr>
                </thead>
                <tbody>
                  {coupons.length === 0 && (
                    <EmptyTableRow colSpan={8} loading={loading} />
                  )}
                  {coupons.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2 font-mono text-xs'>{row.code}</td>
                      <td>
                        {row.agent_name ||
                          agentLabelMap.get(row.agent_id) ||
                          formatFallbackId(row.agent_id)}
                      </td>
                      <td>{formatMoney(row.amount)}</td>
                      <td>{distributionCouponSourceLabel(row.source, t)}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatTime(row.expires_at)}</td>
                      <td>{formatTime(row.used_at)}</td>
                      <td>{row.remark || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={couponPage}
                total={couponTotal}
                onPageChange={setCouponPage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='gift' className='space-y-4'>
            <TabCard
              title={t('Gift Rules')}
              action={
                <Button onClick={() => void openDialog('gift')}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Add Gift Rule')}
                </Button>
              }
            >
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Name')}</th>
                    <th>{t('Trigger')}</th>
                    <th>{t('Gift')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {giftRules.length === 0 && (
                    <EmptyTableRow colSpan={5} loading={loading} />
                  )}
                  {giftRules.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>{row.name}</td>
                      <td>
                        {packageLabelMap.get(row.package_id) ||
                          formatFallbackId(row.package_id)}{' '}
                        × {row.trigger_quantity}
                      </td>
                      <td>
                        {row.gift_quantity} /{' '}
                        {packageLabelMap.get(row.gift_package_id) ||
                          formatFallbackId(row.gift_package_id)}
                      </td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td className='py-2'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void handleToggleGiftRuleStatus(row)}
                        >
                          {t(row.status === 'enabled' ? 'Disable' : 'Enable')}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={giftRulePage}
                total={giftRuleTotal}
                onPageChange={setGiftRulePage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='ops' className='space-y-4'>
            <TabCard
              title={t('Operations Access')}
              action={
                <Button onClick={() => void openDialog('ops')}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Grant Access')}
                </Button>
              }
            >
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('User')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Granted')}</th>
                    <th>{t('Actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {opsAuth.length === 0 && (
                    <EmptyTableRow colSpan={4} loading={loading} />
                  )}
                  {opsAuth.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>
                        {userLabelMap.get(row.user_id) ||
                          formatFallbackId(row.user_id)}
                      </td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatTime(row.granted_at)}</td>
                      <td className='py-2'>
                        {row.status === 'granted' ? (
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() => void handleRevokeOps(row)}
                          >
                            {t('Revoke')}
                          </Button>
                        ) : (
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={() => void handleReGrantOps(row)}
                          >
                            {t('Grant')}
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={opsPage}
                total={opsTotal}
                onPageChange={setOpsPage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='profit' className='space-y-4'>
            <TabCard title={t('Profit')}>
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    {/* <th className='py-2'>{t('No')}</th> */}
                    <th>{t('Agent')}</th>
                    <th>{t('Child')}</th>
                    <th>{t('Amount')}</th>
                  </tr>
                </thead>
                <tbody>
                  {profit.length === 0 && (
                    <EmptyTableRow colSpan={3} loading={loading} />
                  )}
                  {profit.map((row) => (
                    <tr key={row.id} className='border-b'>
                      {/* <td className='py-2 font-mono text-xs'>{row.profit_no}</td> */}
                      <td>
                        {agentLabelMap.get(row.agent_id) ||
                          formatFallbackId(row.agent_id)}
                      </td>
                      <td>
                        {agentLabelMap.get(row.child_agent_id) ||
                          formatFallbackId(row.child_agent_id)}
                      </td>
                      <td>{formatMoney(row.amount)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={profitPage}
                total={profitTotal}
                onPageChange={setProfitPage}
              />
            </TabCard>
          </TabsContent>

          <TabsContent value='attribution' className='space-y-4'>
            <TabCard title={t('Attribution Logs')}>
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Customer')}</th>
                    <th>{t('Event')}</th>
                    <th>{t('Source')}</th>
                    <th>{t('Time')}</th>
                  </tr>
                </thead>
                <tbody>
                  {attribution.length === 0 && (
                    <EmptyTableRow colSpan={4} loading={loading} />
                  )}
                  {attribution.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>
                        {userLabelMap.get(row.customer_user_id) ||
                          formatFallbackId(row.customer_user_id)}
                      </td>
                      <td>{distributionSourceTypeLabel(row.event_type, t)}</td>
                      <td>{distributionSourceTypeLabel(row.source_type, t)}</td>
                      <td>{formatTime(row.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={attributionPage}
                total={attributionTotal}
                onPageChange={setAttributionPage}
              />
            </TabCard>
          </TabsContent>
        </Tabs>

        <Dialog
          open={dialogKind !== null}
          onOpenChange={(open) => !open && closeDialog()}
        >
          <DialogContent className='max-h-[85vh] overflow-y-auto sm:max-w-2xl'>
            <DialogHeader>
              <DialogTitle>{dialogTitle}</DialogTitle>
            </DialogHeader>

            {dialogKind === 'balance' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2 md:col-span-2'>
                  <Label>{t('Agent')}</Label>
                  <AgentCombobox
                    value={balanceAgentId}
                    selectedAgent={balanceAgent || undefined}
                    onValueChange={setBalanceAgentId}
                    onAgentSelected={setBalanceAgent}
                    placeholder={t('Select agent')}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Delta')}</Label>
                  <Input
                    value={balanceDelta}
                    onChange={(e) => setBalanceDelta(e.target.value)}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={balanceRemark}
                    onChange={(e) => setBalanceRemark(e.target.value)}
                  />
                </div>
              </div>
            )}

            {dialogKind === 'agent' && (
              <div className='grid gap-4 md:grid-cols-2'>
                {agentDialogMode === 'create' && (
                  <>
                    <div className='space-y-2'>
                      <Label>{t('User')}</Label>
                      <Select
                        value={agentForm.user_id}
                        onValueChange={(value) =>
                          setAgentForm((state) => ({
                            ...state,
                            user_id: value ?? '',
                          }))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectDisplay
                            label={userLabelMap.get(Number(agentForm.user_id))}
                            placeholder={t('Select user')}
                          />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            {users.map((user) => (
                              <SelectItem key={user.id} value={String(user.id)}>
                                {formatUserLabel(user)}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className='space-y-2'>
                      <Label>{t('Name')}</Label>
                      <Input
                        value={agentForm.name}
                        onChange={(e) =>
                          setAgentForm((state) => ({
                            ...state,
                            name: e.target.value,
                          }))
                        }
                      />
                    </div>
                    <div className='space-y-2'>
                      <Label>{t('Balance')}</Label>
                      <Input
                        value={agentForm.balance}
                        onChange={(e) =>
                          setAgentForm((state) => ({
                            ...state,
                            balance: e.target.value,
                          }))
                        }
                      />
                    </div>
                    <div className='space-y-2'>
                      <Label>{t('Status')}</Label>
                      <Select
                        value={agentForm.status}
                        onValueChange={(value) =>
                          setAgentForm((state) => ({
                            ...state,
                            status: value ?? '',
                          }))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectDisplay
                            label={distributionStatusLabel(agentForm.status, t)}
                            placeholder={t('Status')}
                          />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            {['enabled', 'disabled'].map((status) => (
                              <SelectItem key={status} value={status}>
                                {t(
                                  status === 'enabled' ? 'Enabled' : 'Disabled'
                                )}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className='space-y-2 md:col-span-2'>
                      <Label>{t('Contact')}</Label>
                      <Input
                        value={agentForm.contact}
                        onChange={(e) =>
                          setAgentForm((state) => ({
                            ...state,
                            contact: e.target.value,
                          }))
                        }
                      />
                    </div>
                    <div className='space-y-2 md:col-span-2'>
                      <Label>{t('Remark')}</Label>
                      <Textarea
                        value={agentForm.remark}
                        onChange={(e) =>
                          setAgentForm((state) => ({
                            ...state,
                            remark: e.target.value,
                          }))
                        }
                      />
                    </div>
                  </>
                )}
                {agentDialogMode === 'edit' && (
                  <>
                    <div className='space-y-2 md:col-span-2'>
                      <Label>{t('Parent Agent')}</Label>
                      <AgentCombobox
                        value={agentForm.parent_agent_id || '0'}
                        selectedAgent={agentParentAgent || undefined}
                        onValueChange={(value) =>
                          setAgentForm((state) => ({
                            ...state,
                            parent_agent_id: value,
                          }))
                        }
                        onAgentSelected={setAgentParentAgent}
                        placeholder={t('Select parent agent')}
                        includeEmpty
                        emptyLabel={t('No parent agent')}
                      />
                    </div>
                    <div className='space-y-2 md:col-span-2'>
                      <Label>{t('Agent Level')}</Label>
                      <Select
                        value={agentForm.level}
                        onValueChange={(value) =>
                          setAgentForm((state) => ({
                            ...state,
                            level: value ?? '2',
                          }))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectDisplay
                            label={`${t('Level')} ${agentForm.level || '2'}`}
                            placeholder={t('Select level')}
                          />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            {['1', '2'].map((level) => (
                              <SelectItem key={level} value={level}>
                                {t('Level')} {level}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                  </>
                )}
              </div>
            )}

            {dialogKind === 'package' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2 md:col-span-2'>
                  <Label>{t('Subscription Plan')}</Label>
                  <Select
                    disabled={packageDialogMode === 'edit'}
                    value={packageForm.subscription_plan_id}
                    onValueChange={(value) =>
                      setPackageForm((state) => ({
                        ...state,
                        subscription_plan_id: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={subscriptionPlanLabelMap.get(
                          Number(packageForm.subscription_plan_id)
                        )}
                        placeholder={t('Select subscription plan')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {subscriptionPlans.map((record) => (
                          <SelectItem
                            key={record.plan.id}
                            value={String(record.plan.id)}
                          >
                            {planLabel(record)}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  {packageDialogMode === 'edit' && (
                    <p className='text-muted-foreground text-xs'>
                      {t('Package cannot be changed when editing.')}
                    </p>
                  )}
                  {selectedPackageSubscriptionPlan && (
                    <p className='text-muted-foreground text-xs'>
                      {t('Package guide price')}:{' '}
                      {formatMoney(selectedPackageSubscriptionPriceAmount)}
                    </p>
                  )}
                  <p className='rounded-md border border-red-200 bg-red-50 px-3 py-2 text-xs text-red-700 dark:border-red-900/50 dark:bg-red-950/20 dark:text-red-300'>
                    {t(
                      'Tier 1 price must be less than or equal to tier 2 price and both must not exceed the subscription plan price.'
                    )}
                  </p>
                </div>
                {(
                  [
                    ['agent_price', 'Tier 1 Agent Price'],
                    ['secondary_agent_price', 'Tier 2 Agent Price'],
                    ['sort_order', 'Sort Order'],
                  ] as const
                ).map(([key, label]) => (
                  <div key={key} className='space-y-2'>
                    <Label>{t(label)}</Label>
                    <Input
                      type='number'
                      min={0}
                      max={
                        ['agent_price', 'secondary_agent_price'].includes(key)
                          ? selectedPackageSubscriptionPriceAmount
                          : undefined
                      }
                      step={
                        ['agent_price', 'secondary_agent_price'].includes(key)
                          ? '0.01'
                          : '1'
                      }
                      value={packageForm[key]}
                      onChange={(e) =>
                        setPackageForm((state) => ({
                          ...state,
                          [key]: e.target.value,
                        }))
                      }
                    />
                    {['agent_price', 'secondary_agent_price'].includes(key) && (
                      <p className='text-muted-foreground text-xs'>
                        {t('Enter an amount with up to 2 decimal places.')}
                      </p>
                    )}
                  </div>
                ))}
                <div className='space-y-2'>
                  <Label>{t('Status')}</Label>
                  <Select
                    value={packageForm.status}
                    onValueChange={(value) =>
                      setPackageForm((state) => ({
                        ...state,
                        status: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={distributionStatusLabel(packageForm.status, t)}
                        placeholder={t('Status')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {['enabled', 'disabled'].map((status) => (
                          <SelectItem key={status} value={status}>
                            {t(status === 'enabled' ? 'Enabled' : 'Disabled')}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}

            {dialogKind === 'price' && (
              <div className='space-y-6'>
                <div className='space-y-2'>
                  <Label>{t('Package')}</Label>
                  <Select
                    value={priceForm.package_id}
                    onValueChange={(value) =>
                      setPriceForm((state) => ({
                        ...state,
                        package_id: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={packageLabelMap.get(
                          Number(priceForm.package_id)
                        )}
                        placeholder={t('Select package')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {packageOptions.map((pkg) => (
                          <SelectItem key={pkg.id} value={String(pkg.id)}>
                            {pkg.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>

                <div className='grid gap-4 rounded-md border p-4 md:grid-cols-2'>
                  <div className='space-y-2'>
                    <Label>{t('Target Type')}</Label>
                    <Select
                      value={priceForm.target_type}
                      onValueChange={(value) => {
                        const targetType = value ?? 'level'
                        setPriceForm((state) => ({
                          ...state,
                          target_type: targetType,
                          customer_user_id:
                            targetType === 'customer'
                              ? state.customer_user_id
                              : '',
                          agent_level:
                            targetType === 'level'
                              ? state.agent_level || '1'
                              : '',
                        }))
                      }}
                    >
                      <SelectTrigger className='w-full'>
                        <SelectDisplay
                          label={distributionPriceTargetLabel(
                            priceForm.target_type,
                            t
                          )}
                          placeholder={t('Target Type')}
                        />
                      </SelectTrigger>
                      <SelectContent alignItemWithTrigger={false}>
                        <SelectGroup>
                          {['customer', 'level'].map((target) => (
                            <SelectItem key={target} value={target}>
                              {distributionPriceTargetLabel(target, t)}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>

                  {priceForm.target_type === 'customer' ? (
                    <div className='space-y-2'>
                      <Label>{t('Customer')}</Label>
                      <Select
                        value={priceForm.customer_user_id}
                        onValueChange={(value) =>
                          setPriceForm((state) => ({
                            ...state,
                            customer_user_id: value ?? '',
                          }))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectDisplay
                            label={userLabelMap.get(
                              Number(priceForm.customer_user_id)
                            )}
                            placeholder={t('Select user')}
                          />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            {users.map((user) => (
                              <SelectItem key={user.id} value={String(user.id)}>
                                {formatUserLabel(user)}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                  ) : (
                    <div className='space-y-2'>
                      <Label>{t('Agent Level')}</Label>
                      <Select
                        value={priceForm.agent_level}
                        onValueChange={(value) =>
                          setPriceForm((state) => ({
                            ...state,
                            agent_level: value ?? '',
                          }))
                        }
                      >
                        <SelectTrigger className='w-full'>
                          <SelectDisplay
                            label={
                              priceForm.agent_level
                                ? `${t('Agent Level')} ${priceForm.agent_level}`
                                : ''
                            }
                            placeholder={t('Select level')}
                          />
                        </SelectTrigger>
                        <SelectContent alignItemWithTrigger={false}>
                          <SelectGroup>
                            {['1', '2'].map((level) => (
                              <SelectItem key={level} value={level}>
                                {t('Agent Level')} {level}
                              </SelectItem>
                            ))}
                          </SelectGroup>
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                </div>

                <div className='grid gap-4 rounded-md border p-4 md:grid-cols-2'>
                  <div className='space-y-2'>
                    <Label>{t('Price Type')}</Label>
                    <Select
                      value={priceForm.price_type}
                      onValueChange={(value) =>
                        setPriceForm((state) => ({
                          ...state,
                          price_type: value ?? 'fixed',
                          price_value: '',
                        }))
                      }
                    >
                      <SelectTrigger className='w-full'>
                        <SelectDisplay
                          label={distributionPriceTypeLabel(
                            priceForm.price_type,
                            t
                          )}
                          placeholder={t('Price Type')}
                        />
                      </SelectTrigger>
                      <SelectContent alignItemWithTrigger={false}>
                        <SelectGroup>
                          {['fixed', 'discount'].map((type) => (
                            <SelectItem key={type} value={type}>
                              {distributionPriceTypeLabel(type, t)}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>
                  <div className='space-y-2'>
                    <Label>
                      {t(
                        priceForm.price_type === 'discount'
                          ? 'Discount'
                          : 'Price'
                      )}
                    </Label>
                    <Input
                      type='number'
                      min={priceForm.price_type === 'discount' ? 1 : 0}
                      max={priceForm.price_type === 'discount' ? 10 : undefined}
                      step='1'
                      value={priceForm.price_value}
                      placeholder={
                        priceForm.price_type === 'discount'
                          ? t('Enter a discount from 1 to 10')
                          : t('Enter an amount')
                      }
                      onChange={(e) =>
                        setPriceForm((state) => ({
                          ...state,
                          price_value: e.target.value,
                        }))
                      }
                    />
                    <p className='text-muted-foreground text-xs'>
                      {priceForm.price_type === 'discount'
                        ? t(
                            'Enter 1-10: 1 means 10% of the original price, 2 means 20%, and 10 means the original price.'
                          )
                        : t('Enter an amount with up to 2 decimal places.')}
                    </p>
                  </div>
                </div>

                <div className='grid gap-4 md:grid-cols-2'>
                  <div className='space-y-2'>
                    <Label>{t('Status')}</Label>
                    <Select
                      value={priceForm.status}
                      onValueChange={(value) =>
                        setPriceForm((state) => ({
                          ...state,
                          status: value ?? '',
                        }))
                      }
                    >
                      <SelectTrigger className='w-full'>
                        <SelectDisplay
                          label={distributionStatusLabel(priceForm.status, t)}
                          placeholder={t('Status')}
                        />
                      </SelectTrigger>
                      <SelectContent alignItemWithTrigger={false}>
                        <SelectGroup>
                          {['enabled', 'disabled'].map((status) => (
                            <SelectItem key={status} value={status}>
                              {t(status === 'enabled' ? 'Enabled' : 'Disabled')}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={priceForm.remark}
                    onChange={(e) =>
                      setPriceForm((state) => ({
                        ...state,
                        remark: e.target.value,
                      }))
                    }
                  />
                </div>
              </div>
            )}

            {dialogKind === 'coupon' && (
              <div className='space-y-4'>
                <div className='space-y-2'>
                  <Label>{t('Agent')}</Label>
                  <AgentCombobox
                    value={couponIssueAgentId}
                    selectedAgent={couponIssueAgent || undefined}
                    onValueChange={setCouponIssueAgentId}
                    onAgentSelected={setCouponIssueAgent}
                    placeholder={t('Select agent')}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Coupon Items')}</Label>
                  {couponItems.map((item, index) => (
                    <div
                      key={index}
                      className='grid grid-cols-[1fr_1fr_1fr_auto] items-end gap-2'
                    >
                      <div className='space-y-1'>
                        <Label className='text-muted-foreground text-xs'>
                          {t('Quantity')}
                        </Label>
                        <Input
                          type='number'
                          min={1}
                          step='1'
                          value={item.count}
                          onChange={(e) =>
                            setCouponItems((state) =>
                              state.map((row, i) =>
                                i === index
                                  ? { ...row, count: e.target.value }
                                  : row
                              )
                            )
                          }
                        />
                      </div>
                      <div className='space-y-1'>
                        <Label className='text-muted-foreground text-xs'>
                          {t('Amount')}
                        </Label>
                        <Input
                          type='number'
                          min={0}
                          step='0.01'
                          placeholder='0.00'
                          value={item.amount}
                          onChange={(e) =>
                            setCouponItems((state) =>
                              state.map((row, i) =>
                                i === index
                                  ? { ...row, amount: e.target.value }
                                  : row
                              )
                            )
                          }
                        />
                      </div>
                      <div className='space-y-1'>
                        <Label className='text-muted-foreground text-xs'>
                          {t('Validity (Days)')}
                        </Label>
                        <Input
                          type='number'
                          min={1}
                          step='1'
                          value={item.validity_days}
                          onChange={(e) =>
                            setCouponItems((state) =>
                              state.map((row, i) =>
                                i === index
                                  ? { ...row, validity_days: e.target.value }
                                  : row
                              )
                            )
                          }
                        />
                      </div>
                      <Button
                        variant='outline'
                        size='sm'
                        disabled={couponItems.length <= 1}
                        onClick={() =>
                          setCouponItems((state) =>
                            state.filter((_, i) => i !== index)
                          )
                        }
                      >
                        <Trash2 className='h-4 w-4' />
                      </Button>
                    </div>
                  ))}
                  <div className='flex items-center justify-between'>
                    <Button
                      variant='outline'
                      size='sm'
                      onClick={() =>
                        setCouponItems((state) => [
                          ...state,
                          { ...emptyCouponItem },
                        ])
                      }
                    >
                      <Plus className='mr-2 h-4 w-4' />
                      {t('Add Item')}
                    </Button>
                    <span className='text-muted-foreground text-sm'>
                      {t('Total coupons')}: {couponItemsTotalCount(couponItems)}
                      /{MAX_COUPONS_PER_ISSUE}
                    </span>
                  </div>
                  <p className='text-muted-foreground text-xs'>
                    {t(
                      'Issued coupons are not refundable and will be destroyed when they expire.'
                    )}
                  </p>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={couponRemark}
                    onChange={(e) => setCouponRemark(e.target.value)}
                  />
                </div>
              </div>
            )}

            {dialogKind === 'gift' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2'>
                  <Label>{t('Package')}</Label>
                  <Select
                    value={giftForm.package_id}
                    onValueChange={(value) =>
                      setGiftForm((state) => ({
                        ...state,
                        package_id: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={packageLabelMap.get(Number(giftForm.package_id))}
                        placeholder={t('Select package')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {packageOptions.map((pkg) => (
                          <SelectItem key={pkg.id} value={String(pkg.id)}>
                            {pkg.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Gift Package')}</Label>
                  <Select
                    value={giftForm.gift_package_id}
                    onValueChange={(value) =>
                      setGiftForm((state) => ({
                        ...state,
                        gift_package_id: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={packageLabelMap.get(
                          Number(giftForm.gift_package_id)
                        )}
                        placeholder={t('Select gift package')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {packageOptions.map((pkg) => (
                          <SelectItem key={pkg.id} value={String(pkg.id)}>
                            {pkg.name}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                {(
                  [
                    ['name', 'Name'],
                    ['trigger_quantity', 'Trigger Quantity'],
                    ['gift_quantity', 'Gift Quantity'],
                  ] as const
                ).map(([key, label]) => (
                  <div key={key} className='space-y-2'>
                    <Label>{t(label)}</Label>
                    <Input
                      value={giftForm[key]}
                      onChange={(e) =>
                        setGiftForm((state) => ({
                          ...state,
                          [key]: e.target.value,
                        }))
                      }
                    />
                  </div>
                ))}
                <div className='md:col-span-2'>
                  <DateRangeField
                    startValue={giftForm.starts_at}
                    endValue={giftForm.expires_at}
                    onStartChange={(value) =>
                      setGiftForm((state) => ({ ...state, starts_at: value }))
                    }
                    onEndChange={(value) =>
                      setGiftForm((state) => ({ ...state, expires_at: value }))
                    }
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Status')}</Label>
                  <Select
                    value={giftForm.status}
                    onValueChange={(value) =>
                      setGiftForm((state) => ({
                        ...state,
                        status: value ?? '',
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={distributionStatusLabel(giftForm.status, t)}
                        placeholder={t('Status')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {['enabled', 'disabled'].map((status) => (
                          <SelectItem key={status} value={status}>
                            {t(status === 'enabled' ? 'Enabled' : 'Disabled')}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}

            {dialogKind === 'ops' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2'>
                  <Label>{t('User')}</Label>
                  <Select
                    value={opsUserId}
                    onValueChange={(value) => setOpsUserId(value ?? '')}
                  >
                    <SelectTrigger className='w-full'>
                      <SelectDisplay
                        label={userLabelMap.get(Number(opsUserId))}
                        placeholder={t('Select user')}
                      />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {users.map((user) => (
                          <SelectItem key={user.id} value={String(user.id)}>
                            {formatUserLabel(user)}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={opsRemark}
                    onChange={(e) => setOpsRemark(e.target.value)}
                  />
                </div>
              </div>
            )}

            <DialogFooter>
              <Button variant='outline' onClick={closeDialog}>
                {t('Cancel')}
              </Button>
              {dialogKind === 'balance' && (
                <Button onClick={() => void handleAdjustBalance()}>
                  {t('Save')}
                </Button>
              )}
              {dialogKind === 'agent' && (
                <Button onClick={() => void handleSaveAgent()}>
                  {t('Save')}
                </Button>
              )}
              {dialogKind === 'package' && (
                <Button onClick={() => void handleSavePackage()}>
                  {t('Save')}
                </Button>
              )}
              {dialogKind === 'coupon' && (
                <Button onClick={() => void handleIssueCoupons()}>
                  {t('Issue')}
                </Button>
              )}
              {dialogKind === 'gift' && (
                <Button onClick={() => void handleSaveGiftRule()}>
                  {t('Save')}
                </Button>
              )}
              {dialogKind === 'ops' && (
                <Button onClick={() => void handleGrantOps()}>
                  {t('Grant')}
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
