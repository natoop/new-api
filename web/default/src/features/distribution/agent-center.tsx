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
import { Link } from '@tanstack/react-router'
import {
  ChevronLeft,
  ChevronRight,
  Plus,
  RefreshCcw,
  Search,
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
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CopyButton } from '@/components/copy-button'
import { SectionPageLayout } from '@/components/layout'
import { generateAffiliateLink } from '@/features/wallet/lib'
import {
  applyAgentCoupon,
  getAgentCoupons,
  getAgentCustomers,
  getAgentInventory,
  getAgentInventoryPackageOptions,
  getAgentLedger,
  getAgentPackages,
  getAgentProfile,
  getAgentProfit,
  purchaseAgentPackage,
  refundAgentInventory,
} from './api'
import {
  distributionCouponSourceLabel,
  distributionLedgerEntryLabel,
  distributionSourceTypeLabel,
  distributionStatusLabel,
} from './labels'
import type {
  DistributionCoupon,
  DistributionCustomerOwnership,
  DistributionInventory,
  DistributionInventoryPackageOption,
  DistributionLedger,
  DistributionPackage,
  DistributionProfile,
  DistributionProfit,
} from './types'

const PAGE_SIZE = 10
type AgentTab =
  | 'packages'
  | 'inventory'
  | 'ledger'
  | 'profit'
  | 'customers'
  | 'coupons'

function formatTime(value?: number) {
  if (!value) return '-'
  return new Date(value * 1000).toLocaleString()
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

function formatUserLabel(item?: {
  customer_user_id?: number
  assigned_to?: number
  username?: string
  display_name?: string
  email?: string
}) {
  if (!item) return '-'
  const name = item.display_name || item.username || item.email
  const id = item.customer_user_id || item.assigned_to
  if (name && id) return `${name} (#${id})`
  return name || (id ? `#${id}` : '-')
}

function getPurchaseMessageKey(message?: string) {
  const normalized = (message || '').toLowerCase()
  if (normalized.includes('insufficient balance')) return 'Insufficient balance'
  if (normalized.includes('agent profile not found'))
    return 'Agent profile not found'
  if (normalized.includes('package not found')) return 'Package not found'
  if (normalized.includes('record not found')) return 'Data not found'
  return 'Purchase failed'
}

function apiActionError(message?: string) {
  return message?.trim() || 'Action failed'
}

function normalizeCouponAmountInput(value: string) {
  if (value.trim() === '') return ''
  const amount = Number(value)
  if (Number.isNaN(amount)) return ''
  return String(Number(Math.max(0, amount).toFixed(2)))
}

function canRefundInventory(row: DistributionInventory) {
  return row.status === 'available' && row.assigned_to === 0
}

function StatusBadge({ status }: { status?: string }) {
  const { t } = useTranslation()
  return <Badge variant='secondary'>{distributionStatusLabel(status, t)}</Badge>
}

function ClipboardButton({ text }: { text: string }) {
  const { t } = useTranslation()
  return (
    <CopyButton
      value={text}
      variant='outline'
      className='bg-background size-9 shrink-0'
      iconClassName='size-4'
      tooltip={t('Copy referral link')}
      aria-label={t('Copy referral link')}
    />
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
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return (
    <div className='flex items-center justify-between gap-3 pt-3 text-sm'>
      <span className='text-muted-foreground'>
        {t('Total')}: {total}
      </span>
      <div className='flex items-center gap-2'>
        <Button
          variant='outline'
          size='sm'
          disabled={page <= 1}
          onClick={() => onPageChange(Math.max(1, page - 1))}
        >
          <ChevronLeft className='h-4 w-4' />
        </Button>
        <span className='min-w-16 text-center'>
          {page}/{totalPages}
        </span>
        <Button
          variant='outline'
          size='sm'
          disabled={page >= totalPages}
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        >
          <ChevronRight className='h-4 w-4' />
        </Button>
      </div>
    </div>
  )
}

export function AgentCenter() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<AgentTab>('packages')
  const [profile, setProfile] = useState<DistributionProfile | null>(null)
  const [packages, setPackages] = useState<DistributionPackage[]>([])
  const [inventory, setInventory] = useState<DistributionInventory[]>([])
  const [ledger, setLedger] = useState<DistributionLedger[]>([])
  const [profit, setProfit] = useState<DistributionProfit[]>([])
  const [customers, setCustomers] = useState<DistributionCustomerOwnership[]>(
    []
  )
  const [coupons, setCoupons] = useState<DistributionCoupon[]>([])
  const [inventoryPackageOptions, setInventoryPackageOptions] = useState<
    DistributionInventoryPackageOption[]
  >([])

  const [packagePage, setPackagePage] = useState(1)
  const [inventoryPage, setInventoryPage] = useState(1)
  const [ledgerPage, setLedgerPage] = useState(1)
  const [profitPage, setProfitPage] = useState(1)
  const [customerPage, setCustomerPage] = useState(1)
  const [couponPage, setCouponPage] = useState(1)

  const [packageTotal, setPackageTotal] = useState(0)
  const [inventoryTotal, setInventoryTotal] = useState(0)
  const [ledgerTotal, setLedgerTotal] = useState(0)
  const [profitTotal, setProfitTotal] = useState(0)
  const [customerTotal, setCustomerTotal] = useState(0)
  const [couponTotal, setCouponTotal] = useState(0)

  const [inventoryKeyword, setInventoryKeyword] = useState('')
  const [inventoryKeywordInput, setInventoryKeywordInput] = useState('')
  const [customerKeyword, setCustomerKeyword] = useState('')
  const [customerKeywordInput, setCustomerKeywordInput] = useState('')

  const [couponDialogOpen, setCouponDialogOpen] = useState(false)
  const [couponAmount, setCouponAmount] = useState('')
  const [couponSubmitting, setCouponSubmitting] = useState(false)
  const [loading, setLoading] = useState(false)

  const refresh = useCallback(
    async (tab: AgentTab) => {
      setLoading(true)
      try {
        const profileRes = await getAgentProfile().catch(() => null)
        if (profileRes?.success) setProfile(profileRes.data)

        if (tab === 'packages') {
          const res = await getAgentPackages({
            p: packagePage,
            page_size: PAGE_SIZE,
          }).catch(() => null)
          if (res?.success) {
            setPackages(res.data.items || [])
            setPackageTotal(res.data.total || 0)
          }
        }

        if (tab === 'inventory') {
          const [res, packageOptionsRes] = await Promise.all([
            getAgentInventory({
              p: inventoryPage,
              page_size: PAGE_SIZE,
              keyword: inventoryKeyword,
            }).catch(() => null),
            getAgentInventoryPackageOptions().catch(() => null),
          ])
          if (res?.success) {
            setInventory(res.data.items || [])
            setInventoryTotal(res.data.total || 0)
          }
          if (packageOptionsRes?.success) {
            setInventoryPackageOptions(packageOptionsRes.data || [])
          }
        }

        if (tab === 'ledger') {
          const res = await getAgentLedger({
            p: ledgerPage,
            page_size: PAGE_SIZE,
          }).catch(() => null)
          if (res?.success) {
            setLedger(res.data.items || [])
            setLedgerTotal(res.data.total || 0)
          }
        }

        if (tab === 'profit') {
          const res = await getAgentProfit({
            p: profitPage,
            page_size: PAGE_SIZE,
          }).catch(() => null)
          if (res?.success) {
            setProfit(res.data.items || [])
            setProfitTotal(res.data.total || 0)
          }
        }

        if (tab === 'customers') {
          const res = await getAgentCustomers({
            p: customerPage,
            page_size: PAGE_SIZE,
            keyword: customerKeyword,
          }).catch(() => null)
          if (res?.success) {
            setCustomers(res.data.items || [])
            setCustomerTotal(res.data.total || 0)
          }
        }

        if (tab === 'coupons') {
          const res = await getAgentCoupons({
            p: couponPage,
            page_size: PAGE_SIZE,
          }).catch(() => null)
          if (res?.success) {
            setCoupons(res.data.items || [])
            setCouponTotal(res.data.total || 0)
          }
        }
      } finally {
        setLoading(false)
      }
    },
    [
      couponPage,
      customerKeyword,
      customerPage,
      inventoryKeyword,
      inventoryPage,
      ledgerPage,
      packagePage,
      profitPage,
    ]
  )

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void refresh(activeTab)
  }, [activeTab, refresh])

  const summaryCards = useMemo(
    () => [
      {
        title: t('Balance'),
        value: formatMoney(profile?.agent?.balance),
      },
      {
        title: t('Available Inventory'),
        value: String(profile?.available_inventory ?? 0),
      },
      {
        title: t('Customer Count'),
        value: String(customerTotal),
      },
      {
        title: t('Coupon Count'),
        value: String(couponTotal),
      },
    ],
    [couponTotal, customerTotal, profile, t]
  )
  const referralLink = profile?.aff_code
    ? generateAffiliateLink(profile.aff_code)
    : ''
  const packageLabelMap = useMemo(() => {
    const next = new Map<number, string>()
    packages.forEach((pkg) => next.set(pkg.id, packageTitle(pkg)))
    inventoryPackageOptions.forEach((option) =>
      next.set(option.package_id, option.package_name)
    )
    return next
  }, [inventoryPackageOptions, packages])

  async function handleApplyCoupon() {
    const amount = Number(couponAmount)
    if (!couponAmount || Number.isNaN(amount) || amount <= 0) {
      toast.error(t('Coupon amount must be greater than 0.'))
      return
    }
    setCouponSubmitting(true)
    try {
      const result = await applyAgentCoupon(amount)
      if (!result.success) {
        toast.error(apiActionError(result.message))
        return
      }
      toast.success(t('Coupon applied'))
      setCouponDialogOpen(false)
      setCouponAmount('')
      setCouponPage(1)
      await refresh('coupons')
    } finally {
      setCouponSubmitting(false)
    }
  }

  async function handleRefundInventory(row: DistributionInventory) {
    const result = await refundAgentInventory(row.id)
    if (!result.success) {
      toast.error(apiActionError(result.message))
      return
    }
    toast.success(t('Refunded'))
    setLedgerPage(1)
    await refresh('inventory')
  }

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Agent Center')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
        <Link
          to='/agent/guide'
          className='text-muted-foreground hover:text-foreground text-sm underline-offset-4 hover:underline'
        >
          {t('Agent Program Guide')}
        </Link>
        <Button variant='outline' onClick={() => void refresh(activeTab)}>
          <RefreshCcw className='mr-2 h-4 w-4' />
          {loading ? t('Loading') : t('Refresh')}
        </Button>
      </SectionPageLayout.Actions>
      <SectionPageLayout.Content>
        <div className='grid gap-4 md:grid-cols-4'>
          {summaryCards.map((item) => (
            <Card key={item.title}>
              <CardHeader className='pb-2'>
                <CardTitle className='text-sm font-medium'>
                  {item.title}
                </CardTitle>
              </CardHeader>
              <CardContent className='text-2xl font-semibold'>
                {item.value}
              </CardContent>
            </Card>
          ))}
        </div>

        <Tabs
          value={activeTab}
          onValueChange={(value) => setActiveTab(value as AgentTab)}
          className='mt-6 space-y-4'
        >
          <TabsList className='max-w-full flex-wrap justify-start group-data-horizontal/tabs:h-auto'>
            <TabsTrigger value='packages'>{t('Packages')}</TabsTrigger>
            <TabsTrigger value='inventory'>{t('Inventory')}</TabsTrigger>
            <TabsTrigger value='ledger'>{t('Ledger')}</TabsTrigger>
            <TabsTrigger value='profit'>{t('Profit')}</TabsTrigger>
            <TabsTrigger value='customers'>{t('Customers')}</TabsTrigger>
            <TabsTrigger value='coupons'>{t('Coupons')}</TabsTrigger>
          </TabsList>

          <TabsContent value='packages'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Available Packages')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4'>
                {packages.length === 0 && (
                  <div className='text-muted-foreground rounded-md border border-dashed py-8 text-center text-sm'>
                    {t('No data')}
                  </div>
                )}
                {packages.map((pkg) => (
                  <div
                    key={pkg.id}
                    className='flex flex-wrap items-center justify-between gap-3 rounded-md border p-3'
                  >
                    <div className='space-y-1'>
                      <div className='font-medium'>{packageTitle(pkg)}</div>
                      {packageSubtitle(pkg) && (
                        <div className='text-muted-foreground text-sm'>
                          {packageSubtitle(pkg)}
                        </div>
                      )}
                      <div className='text-muted-foreground text-xs'>
                        {t('Tier 1 Agent Price')}:{' '}
                        {formatMoney(pkg.agent_price)} ·{' '}
                        {t('Tier 2 Agent Price')}:{' '}
                        {formatMoney(pkg.secondary_agent_price)}
                      </div>
                    </div>
                    <Button
                      onClick={async () => {
                        const result = await purchaseAgentPackage(pkg.id)
                        if (result.success) {
                          toast.success(t('Purchased'))
                          setInventoryPage(1)
                          setLedgerPage(1)
                          await refresh('packages')
                          return
                        }
                        toast.error(t(getPurchaseMessageKey(result.message)))
                      }}
                    >
                      <Plus className='mr-2 h-4 w-4' />
                      {t('Buy')}
                    </Button>
                  </div>
                ))}
                <TablePager
                  page={packagePage}
                  total={packageTotal}
                  onPageChange={setPackagePage}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='inventory'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Inventory')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4'>
                <div className='flex flex-col gap-2 md:flex-row md:items-end'>
                  <div className='space-y-2 md:w-80'>
                    <Label>{t('Search customer')}</Label>
                    <Input
                      value={inventoryKeywordInput}
                      onChange={(event) =>
                        setInventoryKeywordInput(event.target.value)
                      }
                      onKeyDown={(event) => {
                        if (event.key === 'Enter') {
                          setInventoryPage(1)
                          setInventoryKeyword(inventoryKeywordInput)
                        }
                      }}
                    />
                  </div>
                  <Button
                    variant='outline'
                    onClick={() => {
                      setInventoryPage(1)
                      setInventoryKeyword(inventoryKeywordInput)
                    }}
                  >
                    <Search className='mr-2 h-4 w-4' />
                    {t('Search')}
                  </Button>
                </div>
                <div className='overflow-x-auto'>
                  <table className='w-full text-sm'>
                    <thead>
                      <tr className='border-b text-left'>
                        <th className='py-2'>{t('Redeem Code')}</th>
                        <th>{t('Status')}</th>
                        <th>{t('Package')}</th>
                        <th>{t('Customer')}</th>
                        <th>{t('Action')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {inventory.length === 0 && <EmptyTableRow colSpan={5} />}
                      {inventory.map((row) => (
                        <tr key={row.id} className='border-b'>
                          <td className='py-2 font-mono text-xs'>
                            {row.inventory_no}
                          </td>
                          <td>
                            <StatusBadge status={row.status} />
                          </td>
                          <td>{packageLabelMap.get(row.package_id) || '-'}</td>
                          <td>{formatUserLabel(row)}</td>
                          <td>
                            <div className='flex items-center gap-2'>
                              <ClipboardButton text={row.inventory_no} />
                              {canRefundInventory(row) && (
                                <Button
                                  variant='outline'
                                  size='sm'
                                  onClick={() =>
                                    void handleRefundInventory(row)
                                  }
                                >
                                  <RefreshCcw className='mr-2 h-4 w-4' />
                                  {t('Refund')}
                                </Button>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <TablePager
                  page={inventoryPage}
                  total={inventoryTotal}
                  onPageChange={setInventoryPage}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='ledger'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Ledger')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4 overflow-x-auto'>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      <th className='py-2'>{t('Time')}</th>
                      <th>{t('Type')}</th>
                      <th>{t('Source')}</th>
                      <th>{t('Delta')}</th>
                      <th>{t('Before')}</th>
                      <th>{t('After')}</th>
                      <th>{t('Description')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {ledger.length === 0 && <EmptyTableRow colSpan={7} />}
                    {ledger.map((row) => (
                      <tr key={row.id} className='border-b'>
                        <td className='py-2'>{formatTime(row.created_at)}</td>
                        <td>
                          {distributionLedgerEntryLabel(row.entry_type, t)}
                        </td>
                        <td>
                          {distributionSourceTypeLabel(row.source_type, t)}
                        </td>
                        <td>{formatMoney(row.delta)}</td>
                        <td>{formatMoney(row.balance_before)}</td>
                        <td>{formatMoney(row.balance_after)}</td>
                        <td>{row.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <TablePager
                  page={ledgerPage}
                  total={ledgerTotal}
                  onPageChange={setLedgerPage}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='profit'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Profit')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4 overflow-x-auto'>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      <th className='py-2'>{t('Order')}</th>
                      <th>{t('Child Agent')}</th>
                      <th>{t('Amount')}</th>
                      <th>{t('Status')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {profit.length === 0 && <EmptyTableRow colSpan={4} />}
                    {profit.map((row) => (
                      <tr key={row.id} className='border-b'>
                        <td className='py-2'>{row.order_id}</td>
                        <td>{row.child_agent_id}</td>
                        <td>{formatMoney(row.amount)}</td>
                        <td>
                          <StatusBadge status={row.status} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <TablePager
                  page={profitPage}
                  total={profitTotal}
                  onPageChange={setProfitPage}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='customers'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Customers')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4 overflow-x-auto'>
                <div className='grid gap-3 lg:grid-cols-3'>
                  <div className='space-y-2'>
                    <Label>{t('Promotion Code')}</Label>
                    <div className='flex gap-2'>
                      <Input
                        value={profile?.aff_code || ''}
                        readOnly
                        className='font-mono'
                      />
                      <ClipboardButton text={profile?.aff_code || ''} />
                    </div>
                  </div>
                  <div className='space-y-2'>
                    <Label>{t('Registration Link')}</Label>
                    <div className='flex gap-2'>
                      <Input
                        value={referralLink}
                        readOnly
                        className='font-mono'
                      />
                      <ClipboardButton text={referralLink} />
                    </div>
                  </div>
                  <div className='space-y-2'>
                    <Label>{t('Login Link')}</Label>
                    <div className='flex gap-2'>
                      <Input
                        value={referralLink}
                        readOnly
                        className='font-mono'
                      />
                      <ClipboardButton text={referralLink} />
                    </div>
                  </div>
                </div>
                <div className='flex flex-col gap-2 md:flex-row md:items-end'>
                  <div className='space-y-2 md:w-80'>
                    <Label>{t('Search customer')}</Label>
                    <Input
                      value={customerKeywordInput}
                      onChange={(event) =>
                        setCustomerKeywordInput(event.target.value)
                      }
                      onKeyDown={(event) => {
                        if (event.key === 'Enter') {
                          setCustomerPage(1)
                          setCustomerKeyword(customerKeywordInput)
                        }
                      }}
                    />
                  </div>
                  <Button
                    variant='outline'
                    onClick={() => {
                      setCustomerPage(1)
                      setCustomerKeyword(customerKeywordInput)
                    }}
                  >
                    <Search className='mr-2 h-4 w-4' />
                    {t('Search')}
                  </Button>
                </div>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      <th className='py-2'>{t('Customer')}</th>
                      <th>{t('Source')}</th>
                      <th>{t('Bound At')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {customers.length === 0 && <EmptyTableRow colSpan={3} />}
                    {customers.map((row) => (
                      <tr key={row.id} className='border-b'>
                        <td className='py-2'>{formatUserLabel(row)}</td>
                        <td>
                          {distributionSourceTypeLabel(row.source_type, t)}
                        </td>
                        <td>{formatTime(row.bound_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                <TablePager
                  page={customerPage}
                  total={customerTotal}
                  onPageChange={setCustomerPage}
                />
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='coupons' className='space-y-4'>
            <Card>
              <CardHeader className='flex flex-row items-center justify-between gap-3'>
                <CardTitle>{t('Coupons')}</CardTitle>
                <Button onClick={() => setCouponDialogOpen(true)}>
                  <Plus className='mr-2 h-4 w-4' />
                  {t('Apply Coupon')}
                </Button>
              </CardHeader>
              <CardContent className='space-y-4'>
                <div className='overflow-x-auto'>
                  <table className='w-full text-sm'>
                    <thead>
                      <tr className='border-b text-left'>
                        <th className='py-2'>{t('Code')}</th>
                        <th>{t('Amount')}</th>
                        <th>{t('Source')}</th>
                        <th>{t('Status')}</th>
                        <th>{t('Expires At')}</th>
                        <th>{t('Used At')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {coupons.length === 0 && <EmptyTableRow colSpan={6} />}
                      {coupons.map((row) => (
                        <tr key={row.id} className='border-b'>
                          <td className='py-2'>
                            <div className='flex items-center gap-2'>
                              <span className='font-mono text-xs'>
                                {row.code}
                              </span>
                              <ClipboardButton text={row.code} />
                            </div>
                          </td>
                          <td>{formatMoney(row.amount)}</td>
                          <td>
                            {distributionCouponSourceLabel(row.source, t)}
                          </td>
                          <td>
                            <StatusBadge status={row.status} />
                          </td>
                          <td>{formatTime(row.expires_at)}</td>
                          <td>{formatTime(row.used_at)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <TablePager
                  page={couponPage}
                  total={couponTotal}
                  onPageChange={setCouponPage}
                />
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        <Separator className='my-6' />

        <Dialog open={couponDialogOpen} onOpenChange={setCouponDialogOpen}>
          <DialogContent className='max-h-[calc(100vh-2rem)] overflow-hidden sm:max-w-lg'>
            <DialogHeader>
              <DialogTitle>{t('Apply Coupon')}</DialogTitle>
            </DialogHeader>
            <div className='flex max-h-[calc(100vh-12rem)] flex-col gap-4 overflow-y-auto'>
              <div className='space-y-2'>
                <Label>{t('Current Balance')}</Label>
                <Input value={formatMoney(profile?.agent?.balance)} readOnly />
              </div>
              <div className='space-y-2'>
                <Label>{t('Amount')}</Label>
                <Input
                  type='number'
                  min={0}
                  step='0.01'
                  placeholder='0.00'
                  value={couponAmount}
                  onChange={(event) =>
                    setCouponAmount(
                      normalizeCouponAmountInput(event.target.value)
                    )
                  }
                />
              </div>
              <p className='text-muted-foreground text-xs'>
                {t(
                  'The amount will be deducted from your agent balance. The coupon is valid for 3 days; if unused, it will be refunded to your balance automatically when it expires.'
                )}
              </p>
            </div>
            <DialogFooter>
              <Button
                variant='outline'
                onClick={() => setCouponDialogOpen(false)}
              >
                {t('Cancel')}
              </Button>
              <Button
                disabled={couponSubmitting}
                onClick={() => void handleApplyCoupon()}
              >
                {couponSubmitting ? t('Loading') : t('Apply')}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
