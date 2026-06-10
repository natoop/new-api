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
  Plus,
  RefreshCcw,
  Wallet,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
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
  adminGetGiftRules,
  adminGetOpsAuthorizations,
  adminGetPackages,
  adminGetProfit,
  adminGetSubscriptionPlans,
  adminGetUserOptions,
  adminGrantOpsAuthorization,
  adminSaveAgent,
  adminSaveGiftRule,
  adminSavePackage,
  adminRevokeOpsAuthorization,
  adminSearchAgents,
  adminUpdateAgentStatus,
  adminUpdateGiftRuleStatus,
  adminUpdatePackageStatus,
} from './api'
import { DateRangeField } from './date-fields'
import {
  distributionPriceTargetLabel,
  distributionPriceTypeLabel,
  distributionSourceTypeLabel,
  distributionStatusLabel,
} from './labels'
import type {
  DistributionAgent,
  DistributionAttributionLog,
  DistributionGiftRule,
  DistributionOpsAuthorization,
  DistributionPackage,
  DistributionProfit,
  DistributionSubscriptionPlanRecord,
  DistributionUserOption,
} from './types'

type AdminTab =
  | 'agents'
  | 'packages'
  | 'gift'
  | 'ops'
  | 'profit'
  | 'attribution'

type DialogKind =
  | 'agent'
  | 'balance'
  | 'package'
  | 'price'
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
  return ((value ?? 0) / 100).toFixed(2)
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

function EmptyTableRow({ colSpan }: { colSpan: number }) {
  const { t } = useTranslation()
  return (
    <tr>
      <td
        colSpan={colSpan}
        className='text-muted-foreground py-8 text-center text-sm'
      >
        {t('No data')}
      </td>
    </tr>
  )
}

function apiActionError(message?: string) {
  return message?.trim() || 'Action failed'
}

function packageActionError(message?: string) {
  const fallback = apiActionError(message)
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

function subscriptionPlanPriceCents(
  record?: DistributionSubscriptionPlanRecord
) {
  if (!record) return 0
  return Math.round(Number(record.plan.price_amount || 0) * 100)
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
  const [opsPage, setOpsPage] = useState(1)
  const [opsTotal, setOpsTotal] = useState(0)
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
  const selectedPackageSubscriptionPriceCents = subscriptionPlanPriceCents(
    selectedPackageSubscriptionPlan
  )

  const resetAgentForm = useCallback(() => {
    setAgentForm(emptyAgentForm)
    setAgentParentAgent(null)
    setAgentDialogMode('create')
  }, [])

  const resetPackageForm = useCallback(() => {
    setPackageForm(emptyPackageForm)
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
      giftRulePage,
      loadAgentOptions,
      loadPackageOptions,
      loadSubscriptionPlans,
      loadUsers,
      opsPage,
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
      toast.error(apiActionError(res.message))
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
      toast.error(apiActionError(res.message))
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
      agentPrice > selectedPackageSubscriptionPriceCents ||
      secondaryAgentPrice > selectedPackageSubscriptionPriceCents
    ) {
      toast.error(
        t(
          'Agent prices must be less than or equal to the subscription plan price.'
        )
      )
      return
    }
    const res = await adminSavePackage({
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
      toast.error(apiActionError(res.message))
      return
    }
    toast.success(t('Saved'))
    resetGiftForm()
    closeDialog()
    setGiftRulePage(1)
    await refreshTab('gift')
  }

  async function handleGrantOps() {
    const res = await adminGrantOpsAuthorization(Number(opsUserId), opsRemark)
    if (!res.success) {
      toast.error(apiActionError(res.message))
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
      toast.error(apiActionError(res.message))
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
      toast.error(apiActionError(res.message))
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
      toast.error(apiActionError(res.message))
      return
    }
    toast.success(t('Saved'))
    await refreshTab('gift')
  }

  async function handleRevokeOps(row: DistributionOpsAuthorization) {
    const res = await adminRevokeOpsAuthorization(row.user_id)
    if (!res.success) {
      toast.error(apiActionError(res.message))
      return
    }
    toast.success(t('Saved'))
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
          ? t('Add Package')
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
                  {agents.length === 0 && <EmptyTableRow colSpan={7} />}
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
                    <th>{t('Tier 1 Agent Price (USD)')}</th>
                    <th>{t('Tier 2 Agent Price (USD)')}</th>
                    <th>{t('Sort Order')}</th>
                    <th>{t('Status')}</th>
                    <th>{t('Actions')}</th>
                  </tr>
                </thead>
                <tbody>
                  {packages.length === 0 && <EmptyTableRow colSpan={6} />}
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
                      <td className='py-2'>
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
                  {giftRules.length === 0 && <EmptyTableRow colSpan={5} />}
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
                  {opsAuth.length === 0 && <EmptyTableRow colSpan={4} />}
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
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => void handleRevokeOps(row)}
                        >
                          {t('Revoke')}
                        </Button>
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
                  {profit.length === 0 && <EmptyTableRow colSpan={3} />}
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
                  {attribution.length === 0 && <EmptyTableRow colSpan={4} />}
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
                  <p className='text-muted-foreground text-xs'>
                    {t(
                      'Tier 1 price must be less than or equal to tier 2 price and both must not exceed the subscription plan price.'
                    )}
                  </p>
                </div>
                {(
                  [
                    ['agent_price', 'Tier 1 Agent Price (USD)'],
                    ['secondary_agent_price', 'Tier 2 Agent Price (USD)'],
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
                          ? selectedPackageSubscriptionPriceCents
                          : undefined
                      }
                      step='1'
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
                        {t(
                          'Enter a USD amount in cents, for example 1000 means 10.00 USD.'
                        )}
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
                          : t('Enter the price in cents')
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
                        : t(
                            'The price is stored in cents; for example, enter 1000 for 10.00.'
                          )}
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
