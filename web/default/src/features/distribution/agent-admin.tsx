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
import { ChevronLeft, ChevronRight, Plus, RefreshCcw, Wallet } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
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
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { SectionPageLayout } from '@/components/layout'
import {
  adminAdjustAgentBalance,
  adminGetAttribution,
  adminGetGiftRules,
  adminGetOpsAuthorizations,
  adminGetPackages,
  adminGetPriceConfigs,
  adminGetProfit,
  adminGetUserOptions,
  adminGrantOpsAuthorization,
  adminSaveAgent,
  adminSaveGiftRule,
  adminSavePackage,
  adminSavePriceConfig,
  adminSearchAgents,
} from './api'
import { AgentCombobox, formatAgentLabel } from './agent-combobox'
import { DateRangeField } from './date-fields'
import type {
  DistributionAgent,
  DistributionAttributionLog,
  DistributionGiftRule,
  DistributionOpsAuthorization,
  DistributionPackage,
  DistributionPriceConfig,
  DistributionProfit,
  DistributionUserOption,
} from './types'

type AdminTab =
  | 'agents'
  | 'packages'
  | 'price'
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
  return <Badge variant='secondary'>{status || '-'}</Badge>
}

function formatMoney(value?: number) {
  return ((value ?? 0) / 100).toFixed(2)
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
        className='py-8 text-center text-sm text-muted-foreground'
      >
        {t('No data')}
      </td>
    </tr>
  )
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
      <CardContent className='space-y-3 overflow-x-auto'>{children}</CardContent>
    </Card>
  )
}

const emptyAgentForm = {
  user_id: '',
  name: '',
  balance: '',
  commission_bps: '',
  parent_agent_id: '0',
  contact: '',
  remark: '',
  status: 'enabled',
}

const emptyPackageForm = {
  name: '',
  sku: '',
  description: '',
  status: 'enabled',
  agent_price: '',
  retail_price: '',
  credit_amount: '',
  sort_order: '',
}

const emptyPriceForm = {
  scope_type: 'global',
  package_id: '',
  level: '',
  parent_agent_id: '0',
  agent_id: '0',
  unit_price: '',
  tier1_cost_price: '',
  tier2_cost_price: '',
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
  const [packageOptions, setPackageOptions] = useState<DistributionPackage[]>([])
  const [users, setUsers] = useState<DistributionUserOption[]>([])
  const [priceConfigs, setPriceConfigs] = useState<DistributionPriceConfig[]>([])
  const [profit, setProfit] = useState<DistributionProfit[]>([])
  const [attribution, setAttribution] = useState<DistributionAttributionLog[]>([])
  const [giftRules, setGiftRules] = useState<DistributionGiftRule[]>([])
  const [opsAuth, setOpsAuth] = useState<DistributionOpsAuthorization[]>([])
  const [agentPage, setAgentPage] = useState(1)
  const [agentTotal, setAgentTotal] = useState(0)
  const [packagePage, setPackagePage] = useState(1)
  const [packageTotal, setPackageTotal] = useState(0)
  const [pricePage, setPricePage] = useState(1)
  const [priceTotal, setPriceTotal] = useState(0)
  const [profitPage, setProfitPage] = useState(1)
  const [profitTotal, setProfitTotal] = useState(0)
  const [attributionPage, setAttributionPage] = useState(1)
  const [attributionTotal, setAttributionTotal] = useState(0)
  const [giftRulePage, setGiftRulePage] = useState(1)
  const [giftRuleTotal, setGiftRuleTotal] = useState(0)
  const [opsPage, setOpsPage] = useState(1)
  const [opsTotal, setOpsTotal] = useState(0)
  const [balanceAgentId, setBalanceAgentId] = useState('')
  const [balanceAgent, setBalanceAgent] = useState<DistributionAgent | null>(null)
  const [balanceDelta, setBalanceDelta] = useState('')
  const [balanceRemark, setBalanceRemark] = useState('')
  const [agentParentAgent, setAgentParentAgent] = useState<DistributionAgent | null>(null)
  const [priceParentAgent, setPriceParentAgent] = useState<DistributionAgent | null>(null)
  const [priceAgent, setPriceAgent] = useState<DistributionAgent | null>(null)
  const [agentForm, setAgentForm] = useState(emptyAgentForm)
  const [packageForm, setPackageForm] = useState(emptyPackageForm)
  const [priceForm, setPriceForm] = useState(emptyPriceForm)
  const [giftForm, setGiftForm] = useState(emptyGiftForm)
  const [opsUserId, setOpsUserId] = useState('')
  const [opsRemark, setOpsRemark] = useState('')

  const agentLabelMap = useMemo(
    () => new Map(agentOptions.map((agent) => [agent.id, formatAgentLabel(agent)])),
    [agentOptions]
  )
  const packageLabelMap = useMemo(
    () => new Map(packageOptions.map((pkg) => [pkg.id, pkg.name])),
    [packageOptions]
  )
  const userLabelMap = useMemo(
    () => new Map(users.map((user) => [user.id, formatUserLabel(user)])),
    [users]
  )

  const resetAgentForm = useCallback(() => {
    setAgentForm(emptyAgentForm)
    setAgentParentAgent(null)
  }, [])

  const resetPackageForm = useCallback(() => {
    setPackageForm(emptyPackageForm)
  }, [])

  const resetPriceForm = useCallback(() => {
    setPriceForm(emptyPriceForm)
    setPriceParentAgent(null)
    setPriceAgent(null)
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
    const res = await adminSearchAgents({ p: 1, page_size: 100 }).catch(() => null)
    if (res?.success) setAgentOptions(res.data?.items || [])
  }, [])

  const loadPackageOptions = useCallback(async () => {
    const res = await adminGetPackages({ p: 1, page_size: 100 }).catch(() => null)
    if (res?.success) setPackageOptions(res.data?.items || [])
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
              for (const agent of res.data?.items || []) map.set(agent.id, agent)
              return [...map.values()]
            })
          }
        }
        if (tab === 'packages') {
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
        if (tab === 'price') {
          await Promise.all([loadPackageOptions(), loadAgentOptions()])
          const res = await adminGetPriceConfigs({
            p: pricePage,
            page_size: pageSize,
          }).catch(() => null)
          if (res?.success) {
            setPriceConfigs(res.data?.items || [])
            setPriceTotal(res.data?.total || 0)
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
      loadUsers,
      opsPage,
      packagePage,
      pricePage,
      profitPage,
    ]
  )

  useEffect(() => {
    void refreshTab(activeTab)
  }, [activeTab, refreshTab])

  const openDialog = useCallback(
    async (kind: DialogKind) => {
      setDialogKind(kind)
      if (kind === 'agent' || kind === 'ops') await loadUsers()
      if (kind === 'balance') await loadAgentOptions()
      if (kind === 'price') await Promise.all([loadPackageOptions(), loadAgentOptions()])
      if (kind === 'gift') await loadPackageOptions()
    },
    [loadAgentOptions, loadPackageOptions, loadUsers]
  )

  async function handleSaveAgent() {
    await adminSaveAgent({
      user_id: Number(agentForm.user_id),
      name: agentForm.name,
      balance: Number(agentForm.balance),
      commission_bps: Number(agentForm.commission_bps),
      parent_agent_id: Number(agentForm.parent_agent_id),
      contact: agentForm.contact,
      remark: agentForm.remark,
      status: agentForm.status,
    })
    toast.success(t('Saved'))
    resetAgentForm()
    closeDialog()
    setAgentPage(1)
    await refreshTab('agents')
  }

  async function handleAdjustBalance() {
    await adminAdjustAgentBalance(
      Number(balanceAgentId),
      Number(balanceDelta),
      balanceRemark
    )
    toast.success(t('Saved'))
    resetBalanceForm()
    closeDialog()
    await refreshTab('agents')
  }

  async function handleSavePackage() {
    await adminSavePackage({
      name: packageForm.name,
      sku: packageForm.sku,
      description: packageForm.description,
      status: packageForm.status,
      agent_price: Number(packageForm.agent_price),
      retail_price: Number(packageForm.retail_price),
      credit_amount: Number(packageForm.credit_amount),
      sort_order: Number(packageForm.sort_order),
    })
    toast.success(t('Saved'))
    resetPackageForm()
    closeDialog()
    setPackagePage(1)
    await refreshTab('packages')
  }

  async function handleSavePriceConfig() {
    await adminSavePriceConfig({
      scope_type: priceForm.scope_type,
      package_id: Number(priceForm.package_id),
      level: Number(priceForm.level),
      parent_agent_id: Number(priceForm.parent_agent_id),
      agent_id: Number(priceForm.agent_id),
      unit_price: Number(priceForm.unit_price),
      tier1_cost_price: Number(priceForm.tier1_cost_price),
      tier2_cost_price: Number(priceForm.tier2_cost_price),
      status: priceForm.status,
      remark: priceForm.remark,
    })
    toast.success(t('Saved'))
    resetPriceForm()
    closeDialog()
    setPricePage(1)
    await refreshTab('price')
  }

  async function handleSaveGiftRule() {
    await adminSaveGiftRule({
      name: giftForm.name,
      package_id: Number(giftForm.package_id),
      gift_package_id: Number(giftForm.gift_package_id),
      trigger_quantity: Number(giftForm.trigger_quantity),
      gift_quantity: Number(giftForm.gift_quantity),
      starts_at: Number(giftForm.starts_at),
      expires_at: Number(giftForm.expires_at),
      status: giftForm.status,
    })
    toast.success(t('Saved'))
    resetGiftForm()
    closeDialog()
    setGiftRulePage(1)
    await refreshTab('gift')
  }

  async function handleGrantOps() {
    await adminGrantOpsAuthorization(Number(opsUserId), opsRemark)
    toast.success(t('Granted'))
    resetOpsForm()
    closeDialog()
    setOpsPage(1)
    await refreshTab('ops')
  }

  const dialogTitle =
    dialogKind === 'agent'
      ? t('Add Agent')
      : dialogKind === 'balance'
        ? t('Adjust Balance')
        : dialogKind === 'package'
          ? t('Add Package')
          : dialogKind === 'price'
            ? t('Add Price Config')
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
            <TabsTrigger value='price'>{t('Price Configs')}</TabsTrigger>
            <TabsTrigger value='gift'>{t('Gift Rules')}</TabsTrigger>
            <TabsTrigger value='ops'>{t('Operations Access')}</TabsTrigger>
            <TabsTrigger value='profit'>{t('Profit')}</TabsTrigger>
            <TabsTrigger value='attribution'>{t('Attribution Logs')}</TabsTrigger>
          </TabsList>

          <TabsContent value='agents' className='space-y-4'>
            <TabCard
              title={t('Agents')}
              action={
                <div className='flex flex-wrap gap-2'>
                  <Button variant='outline' onClick={() => void openDialog('balance')}>
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
                    <th>{t('Parent')}</th>
                  </tr>
                </thead>
                <tbody>
                  {agents.length === 0 && <EmptyTableRow colSpan={5} />}
                  {agents.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>{row.name}</td>
                      <td>{formatAgentLabel(row)}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatMoney(row.balance)}</td>
                      <td>
                        {agentLabelMap.get(row.parent_agent_id) ||
                          formatFallbackId(row.parent_agent_id)}
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
                    <th className='py-2'>{t('Name')}</th>
                    <th>{t('SKU')}</th>
                    <th>{t('Agent Price')}</th>
                    <th>{t('Retail Price')}</th>
                    <th>{t('Credit Amount')}</th>
                    <th>{t('Status')}</th>
                  </tr>
                </thead>
                <tbody>
                  {packages.length === 0 && <EmptyTableRow colSpan={6} />}
                  {packages.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>{row.name}</td>
                      <td>{row.sku}</td>
                      <td>{formatMoney(row.agent_price)}</td>
                      <td>{formatMoney(row.retail_price)}</td>
                      <td>{row.credit_amount}</td>
                      <td>
                        <StatusBadge status={row.status} />
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

          <TabsContent value='price' className='space-y-4'>
            <TabCard
              title={t('Price Configs')}
              action={
                <Button onClick={() => void openDialog('price')}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Add Price Config')}
                </Button>
              }
            >
              <table className='w-full text-sm'>
                <thead>
                  <tr className='border-b text-left'>
                    <th className='py-2'>{t('Scope')}</th>
                    <th>{t('Package')}</th>
                    <th>{t('Unit Price')}</th>
                    <th>{t('Status')}</th>
                  </tr>
                </thead>
                <tbody>
                  {priceConfigs.length === 0 && <EmptyTableRow colSpan={4} />}
                  {priceConfigs.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>
                        {t(
                          row.scope_type === 'global'
                            ? 'Global'
                            : row.scope_type === 'level'
                              ? 'Level'
                              : 'Agent'
                        )}
                      </td>
                      <td>
                        {packageLabelMap.get(row.package_id) ||
                          formatFallbackId(row.package_id)}
                      </td>
                      <td>{formatMoney(row.unit_price)}</td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager
                page={pricePage}
                total={priceTotal}
                onPageChange={setPricePage}
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
                  </tr>
                </thead>
                <tbody>
                  {giftRules.length === 0 && <EmptyTableRow colSpan={4} />}
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
                  </tr>
                </thead>
                <tbody>
                  {opsAuth.length === 0 && <EmptyTableRow colSpan={3} />}
                  {opsAuth.map((row) => (
                    <tr key={row.id} className='border-b'>
                      <td className='py-2'>
                        {userLabelMap.get(row.user_id) || formatFallbackId(row.user_id)}
                      </td>
                      <td>
                        <StatusBadge status={row.status} />
                      </td>
                      <td>{formatTime(row.granted_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <TablePager page={opsPage} total={opsTotal} onPageChange={setOpsPage} />
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
                  {profit.length === 0 && <EmptyTableRow colSpan={4} />}
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
                      <td>{row.event_type}</td>
                      <td>{row.source_type}</td>
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

        <Dialog open={dialogKind !== null} onOpenChange={(open) => !open && closeDialog()}>
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
                <div className='space-y-2'>
                  <Label>{t('User')}</Label>
                  <Select
                    value={agentForm.user_id}
                    onValueChange={(value) =>
                      setAgentForm((state) => ({ ...state, user_id: value }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Select user')} />
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
                {(
                  [
                    ['name', 'Name'],
                    ['balance', 'Balance'],
                    ['commission_bps', 'Commission BPS'],
                    ['contact', 'Contact'],
                  ] as const
                ).map(([key, label]) => (
                  <div key={key} className='space-y-2'>
                    <Label>{t(label)}</Label>
                    <Input
                      value={agentForm[key]}
                      onChange={(e) =>
                        setAgentForm((state) => ({
                          ...state,
                          [key]: e.target.value,
                        }))
                      }
                    />
                  </div>
                ))}
                <div className='space-y-2 md:col-span-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={agentForm.remark}
                    onChange={(e) =>
                      setAgentForm((state) => ({ ...state, remark: e.target.value }))
                    }
                  />
                </div>
              </div>
            )}

            {dialogKind === 'package' && (
              <div className='grid gap-4 md:grid-cols-2'>
                {(
                  [
                    ['name', 'Name'],
                    ['sku', 'SKU'],
                    ['agent_price', 'Agent Price'],
                    ['retail_price', 'Retail Price'],
                    ['credit_amount', 'Credit Amount'],
                    ['sort_order', 'Sort Order'],
                  ] as const
                ).map(([key, label]) => (
                  <div key={key} className='space-y-2'>
                    <Label>{t(label)}</Label>
                    <Input
                      value={packageForm[key]}
                      onChange={(e) =>
                        setPackageForm((state) => ({
                          ...state,
                          [key]: e.target.value,
                        }))
                      }
                    />
                  </div>
                ))}
                <div className='space-y-2 md:col-span-2'>
                  <Label>{t('Description')}</Label>
                  <Textarea
                    value={packageForm.description}
                    onChange={(e) =>
                      setPackageForm((state) => ({
                        ...state,
                        description: e.target.value,
                      }))
                    }
                  />
                </div>
              </div>
            )}

            {dialogKind === 'price' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2'>
                  <Label>{t('Scope Type')}</Label>
                  <Select
                    value={priceForm.scope_type}
                    onValueChange={(value) =>
                      setPriceForm((state) => ({ ...state, scope_type: value }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Scope Type')} />
                    </SelectTrigger>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        {['global', 'level', 'agent'].map((scope) => (
                          <SelectItem key={scope} value={scope}>
                            {t(
                              scope === 'global'
                                ? 'Global'
                                : scope === 'level'
                                  ? 'Level'
                                  : 'Agent'
                            )}
                          </SelectItem>
                        ))}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </div>
                <div className='space-y-2'>
                  <Label>{t('Package')}</Label>
                  <Select
                    value={priceForm.package_id}
                    onValueChange={(value) =>
                      setPriceForm((state) => ({ ...state, package_id: value }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Select package')} />
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
                  <Label>{t('Parent Agent')}</Label>
                  <AgentCombobox
                    value={priceForm.parent_agent_id || '0'}
                    selectedAgent={priceParentAgent || undefined}
                    onValueChange={(value) =>
                      setPriceForm((state) => ({ ...state, parent_agent_id: value }))
                    }
                    onAgentSelected={setPriceParentAgent}
                    placeholder={t('Select parent agent')}
                    includeEmpty
                    emptyLabel={t('No parent agent')}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Agent')}</Label>
                  <AgentCombobox
                    value={priceForm.agent_id || '0'}
                    selectedAgent={priceAgent || undefined}
                    onValueChange={(value) =>
                      setPriceForm((state) => ({ ...state, agent_id: value }))
                    }
                    onAgentSelected={setPriceAgent}
                    placeholder={t('Select agent')}
                    includeEmpty
                    emptyLabel={t('No agent')}
                  />
                </div>
                {(
                  [
                    ['level', 'Level'],
                    ['unit_price', 'Unit Price'],
                    ['tier1_cost_price', 'Tier1 Cost'],
                    ['tier2_cost_price', 'Tier2 Cost'],
                  ] as const
                ).map(([key, label]) => (
                  <div key={key} className='space-y-2'>
                    <Label>{t(label)}</Label>
                    <Input
                      value={priceForm[key]}
                      onChange={(e) =>
                        setPriceForm((state) => ({
                          ...state,
                          [key]: e.target.value,
                        }))
                      }
                    />
                  </div>
                ))}
                <div className='space-y-2 md:col-span-2'>
                  <Label>{t('Remark')}</Label>
                  <Textarea
                    value={priceForm.remark}
                    onChange={(e) =>
                      setPriceForm((state) => ({ ...state, remark: e.target.value }))
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
                      setGiftForm((state) => ({ ...state, package_id: value }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Select package')} />
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
                        gift_package_id: value,
                      }))
                    }
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Select gift package')} />
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
              </div>
            )}

            {dialogKind === 'ops' && (
              <div className='grid gap-4 md:grid-cols-2'>
                <div className='space-y-2'>
                  <Label>{t('User')}</Label>
                  <Select
                    value={opsUserId}
                    onValueChange={(value) => setOpsUserId(value)}
                  >
                    <SelectTrigger className='w-full'>
                      <SelectValue placeholder={t('Select user')} />
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
                <Button onClick={() => void handleAdjustBalance()}>{t('Save')}</Button>
              )}
              {dialogKind === 'agent' && (
                <Button onClick={() => void handleSaveAgent()}>{t('Save')}</Button>
              )}
              {dialogKind === 'package' && (
                <Button onClick={() => void handleSavePackage()}>{t('Save')}</Button>
              )}
              {dialogKind === 'price' && (
                <Button onClick={() => void handleSavePriceConfig()}>{t('Save')}</Button>
              )}
              {dialogKind === 'gift' && (
                <Button onClick={() => void handleSaveGiftRule()}>{t('Save')}</Button>
              )}
              {dialogKind === 'ops' && (
                <Button onClick={() => void handleGrantOps()}>{t('Grant')}</Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
