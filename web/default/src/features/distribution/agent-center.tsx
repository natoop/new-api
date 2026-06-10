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
import { Copy, Handshake, Plus, RefreshCcw } from 'lucide-react'
import { toast } from 'sonner'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { SectionPageLayout } from '@/components/layout'
import { generateAffiliateLink } from '@/features/wallet/lib'
import { CopyButton } from '@/components/copy-button'
import {
  acceptAgentInvitation,
  assignAgentInventory,
  createAgentInvitation,
  getAgentCustomers,
  getAgentInventory,
  getAgentInvitations,
  getAgentLedger,
  getAgentPackages,
  getAgentProfile,
  getAgentProfit,
  getAgentPromoCodes,
  purchaseAgentPackage,
  saveAgentPromoCode,
} from './api'
import { DateOnlyField, DateRangeField } from './date-fields'
import type {
  DistributionCustomerOwnership,
  DistributionInventory,
  DistributionInvitation,
  DistributionLedger,
  DistributionPackage,
  DistributionProfile,
  DistributionProfit,
  DistributionPromoCode,
} from './types'

function formatTime(value?: number) {
  if (!value) return '-'
  return new Date(value * 1000).toLocaleString()
}

function formatMoney(value?: number) {
  return ((value ?? 0) / 100).toFixed(2)
}

function getPurchaseMessageKey(message?: string) {
  const normalized = (message || '').toLowerCase()
  if (normalized.includes('insufficient balance')) return 'Insufficient balance'
  if (normalized.includes('agent profile not found')) {
    return 'Agent profile not found'
  }
  if (normalized.includes('package not found')) return 'Package not found'
  if (normalized.includes('record not found')) return 'Data not found'
  return 'Purchase failed'
}

function StatusBadge({ status }: { status?: string }) {
  return <Badge variant='secondary'>{status || '-'}</Badge>
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
        className='py-8 text-center text-sm text-muted-foreground'
      >
        {t('No data')}
      </td>
    </tr>
  )
}

export function AgentCenter() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState('packages')
  const [profile, setProfile] = useState<DistributionProfile | null>(null)
  const [packages, setPackages] = useState<DistributionPackage[]>([])
  const [inventory, setInventory] = useState<DistributionInventory[]>([])
  const [ledger, setLedger] = useState<DistributionLedger[]>([])
  const [profit, setProfit] = useState<DistributionProfit[]>([])
  const [customers, setCustomers] = useState<DistributionCustomerOwnership[]>(
    []
  )
  const [invitations, setInvitations] = useState<DistributionInvitation[]>([])
  const [promoCodes, setPromoCodes] = useState<DistributionPromoCode[]>([])
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteLevel, setInviteLevel] = useState('1')
  const [inviteExpiresAt, setInviteExpiresAt] = useState('')
  const [promoCode, setPromoCode] = useState('')
  const [promoDiscountType, setPromoDiscountType] = useState('percent')
  const [promoDiscountValue, setPromoDiscountValue] = useState('0')
  const [promoMaxRedemptions, setPromoMaxRedemptions] = useState('0')
  const [promoStartsAt, setPromoStartsAt] = useState('')
  const [promoExpiresAt, setPromoExpiresAt] = useState('')
  const [assignInventoryId, setAssignInventoryId] = useState('')
  const [assignCustomerUserId, setAssignCustomerUserId] = useState('')
  const [loading, setLoading] = useState(false)
  const [loadedTabs, setLoadedTabs] = useState<Record<string, boolean>>({
    packages: true,
    inventory: false,
    ledger: false,
    profit: false,
    customers: false,
    invitations: false,
    promo: false,
  })

  const refresh = useCallback(async (tab: string) => {
    setLoading(true)
    try {
      const profileRes = await getAgentProfile().catch(() => null)
      if (profileRes?.success) setProfile(profileRes.data)

      if (tab === 'packages') {
        const packageRes = await getAgentPackages().catch(() => null)
        if (packageRes?.success) setPackages(packageRes.data || [])
      }
      if (tab === 'inventory') {
        const inventoryRes = await getAgentInventory().catch(() => null)
        if (inventoryRes?.success) setInventory(inventoryRes.data || [])
      }
      if (tab === 'ledger') {
        const ledgerRes = await getAgentLedger().catch(() => null)
        if (ledgerRes?.success) setLedger(ledgerRes.data || [])
      }
      if (tab === 'profit') {
        const profitRes = await getAgentProfit().catch(() => null)
        if (profitRes?.success) setProfit(profitRes.data || [])
      }
      if (tab === 'customers') {
        const customerRes = await getAgentCustomers().catch(() => null)
        if (customerRes?.success) setCustomers(customerRes.data || [])
      }
      if (tab === 'invitations') {
        const invitationRes = await getAgentInvitations().catch(() => null)
        if (invitationRes?.success) setInvitations(invitationRes.data || [])
      }
      if (tab === 'promo') {
        const promoRes = await getAgentPromoCodes().catch(() => null)
        if (promoRes?.success) setPromoCodes(promoRes.data || [])
      }
      setLoadedTabs((current) => ({ ...current, [tab]: true }))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void refresh('packages')
  }, [refresh])

  useEffect(() => {
    if (!loadedTabs[activeTab]) {
      void refresh(activeTab)
    }
  }, [activeTab, loadedTabs, refresh])

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
        value: loadedTabs.customers ? String(customers.length) : '-',
      },
      {
        title: t('Invitation Count'),
        value: loadedTabs.invitations ? String(invitations.length) : '-',
      },
    ],
    [customers.length, invitations.length, loadedTabs, profile, t]
  )
  const referralLink = useMemo(
    () => (profile?.aff_code ? generateAffiliateLink(profile.aff_code) : ''),
    [profile?.aff_code]
  )
  const packageLabelMap = useMemo(
    () => new Map(packages.map((pkg) => [pkg.id, pkg.name])),
    [packages]
  )

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Agent Center')}</SectionPageLayout.Title>
      <SectionPageLayout.Actions>
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
          onValueChange={(value) => setActiveTab(value)}
          className='mt-6'
        >
          <TabsList className='grid w-full grid-cols-7'>
            <TabsTrigger value='packages'>{t('Packages')}</TabsTrigger>
            <TabsTrigger value='inventory'>{t('Inventory')}</TabsTrigger>
            <TabsTrigger value='ledger'>{t('Ledger')}</TabsTrigger>
            <TabsTrigger value='profit'>{t('Profit')}</TabsTrigger>
            <TabsTrigger value='customers'>{t('Customers')}</TabsTrigger>
            <TabsTrigger value='invitations'>{t('Invitations')}</TabsTrigger>
            <TabsTrigger value='promo'>{t('Promo Codes')}</TabsTrigger>
          </TabsList>

          <TabsContent value='packages' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Available Packages')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4'>
                {packages.length === 0 && (
                  <div className='rounded-md border border-dashed py-8 text-center text-sm text-muted-foreground'>
                    {t('No data')}
                  </div>
                )}
                {packages.map((pkg) => (
                  <div
                    key={pkg.id}
                    className='flex flex-wrap items-center justify-between gap-3 rounded-md border p-3'
                  >
                    <div className='space-y-1'>
                      <div className='font-medium'>{pkg.name}</div>
                      <div className='text-sm text-muted-foreground'>
                        {pkg.sku} · {pkg.description || '-'}
                      </div>
                      <div className='text-xs text-muted-foreground'>
                        {t('Agent Price')}: {formatMoney(pkg.agent_price)} ·{' '}
                        {t('Retail Price')}: {formatMoney(pkg.retail_price)} ·{' '}
                        {t('Credit')}: {pkg.credit_amount}
                      </div>
                    </div>
                    <Button
                      onClick={async () => {
                        const result = await purchaseAgentPackage(pkg.id)
                        if (result.success) {
                          toast.success(t('Purchased'))
                          setLoadedTabs((current) => ({
                            ...current,
                            inventory: false,
                            ledger: false,
                          }))
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
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='inventory' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Inventory')}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4'>
                <div className='grid gap-3 md:grid-cols-3'>
                  <div className='space-y-2'>
                    <Label htmlFor='inventory-id'>{t('Inventory ID')}</Label>
                    <Input
                      id='inventory-id'
                      value={assignInventoryId}
                      onChange={(e) => setAssignInventoryId(e.target.value)}
                    />
                  </div>
                  <div className='space-y-2'>
                    <Label htmlFor='customer-user-id'>
                      {t('Customer User ID')}
                    </Label>
                    <Input
                      id='customer-user-id'
                      value={assignCustomerUserId}
                      onChange={(e) => setAssignCustomerUserId(e.target.value)}
                    />
                  </div>
                  <div className='flex items-end'>
                    <Button
                      className='w-full'
                      onClick={async () => {
                        await assignAgentInventory(
                          Number(assignInventoryId),
                          Number(assignCustomerUserId)
                        )
                    toast.success(t('Assigned'))
                        await refresh('inventory')
                      }}
                    >
                      {t('Assign')}
                    </Button>
                  </div>
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
                      {inventory.length === 0 && <EmptyTableRow colSpan={6} />}
                      {inventory.map((row) => (
                        <tr key={row.id} className='border-b'>
                          <td className='font-mono text-xs'>{row.inventory_no}</td>
                          <td>
                            <StatusBadge status={row.status} />
                          </td>
                          <td>{packageLabelMap.get(row.package_id) || '-'}</td>
                          <td>{row.assigned_to || '-'}</td>
                          <td>
                            <ClipboardButton text={row.inventory_no} />
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='ledger'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Ledger')}</CardTitle>
              </CardHeader>
              <CardContent className='overflow-x-auto'>
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
                        <td>{row.entry_type}</td>
                        <td>{row.source_type}</td>
                        <td>{formatMoney(row.delta)}</td>
                        <td>{formatMoney(row.balance_before)}</td>
                        <td>{formatMoney(row.balance_after)}</td>
                        <td>{row.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='profit'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Profit')}</CardTitle>
              </CardHeader>
              <CardContent className='overflow-x-auto'>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      {/* <th className='py-2'>{t('No')}</th> */}
                      <th>{t('Order')}</th>
                      <th>{t('Child Agent')}</th>
                      <th>{t('Amount')}</th>
                      <th>{t('Status')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {profit.length === 0 && <EmptyTableRow colSpan={5} />}
                    {profit.map((row) => (
                      <tr key={row.id} className='border-b'>
                        {/* <td className='py-2 font-mono text-xs'>{row.profit_no}</td> */}
                        <td>{row.order_id}</td>
                        <td>{row.child_agent_id}</td>
                        <td>{formatMoney(row.amount)}</td>
                        <td>
                          <StatusBadge status={row.status} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
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
                      <Input value={referralLink} readOnly className='font-mono' />
                      <ClipboardButton text={referralLink} />
                    </div>
                  </div>
                  <div className='space-y-2'>
                    <Label>{t('Login Link')}</Label>
                    <div className='flex gap-2'>
                      <Input value={referralLink} readOnly className='font-mono' />
                      <ClipboardButton text={referralLink} />
                    </div>
                  </div>
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
                        <td className='py-2'>{row.customer_user_id}</td>
                        <td>{row.source_type}</td>
                        <td>{formatTime(row.bound_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='invitations' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Create Invitation')}</CardTitle>
              </CardHeader>
              <CardContent className='grid gap-3 md:grid-cols-4'>
                <div className='space-y-2'>
                  <Label htmlFor='invitee-email'>{t('Invitee Email')}</Label>
                  <Input
                    id='invitee-email'
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                  />
                </div>
                <div className='space-y-2'>
                  <Label htmlFor='invite-level'>{t('Level')}</Label>
                  <Input
                    id='invite-level'
                    value={inviteLevel}
                    onChange={(e) => setInviteLevel(e.target.value)}
                  />
                </div>
                <DateOnlyField
                  label='Expires At'
                  value={inviteExpiresAt}
                  onChange={setInviteExpiresAt}
                  endOfDay
                />
                <div className='flex items-end'>
                  <Button
                    className='w-full'
                    onClick={async () => {
                      await createAgentInvitation({
                        invitee_email: inviteEmail,
                        level: Number(inviteLevel),
                        expires_at: Number(inviteExpiresAt),
                      })
                      toast.success(t('Created'))
                      await refresh('invitations')
                    }}
                  >
                    <Handshake className='mr-2 h-4 w-4' />
                    {t('Create')}
                  </Button>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>{t('Invitations')}</CardTitle>
              </CardHeader>
              <CardContent className='overflow-x-auto'>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      <th className='py-2'>{t('No')}</th>
                      <th>{t('Email')}</th>
                      <th>{t('Status')}</th>
                      <th>{t('Expire')}</th>
                      <th>{t('Action')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {invitations.length === 0 && (
                      <EmptyTableRow colSpan={5} />
                    )}
                    {invitations.map((row) => (
                      <tr key={row.id} className='border-b'>
                        <td className='py-2 font-mono text-xs'>
                          {row.invitation_no}
                        </td>
                        <td>{row.invitee_email || '-'}</td>
                        <td>
                          <StatusBadge status={row.status} />
                        </td>
                        <td>{formatTime(row.expires_at)}</td>
                        <td>
                          <Button
                            variant='outline'
                            size='sm'
                            onClick={async () => {
                              await acceptAgentInvitation(row.invitation_no)
                              toast.success(t('Accepted'))
                              await refresh('invitations')
                            }}
                          >
                            {t('Accept')}
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value='promo' className='space-y-4'>
            <Card>
              <CardHeader>
                <CardTitle>{t('Promo Codes')}</CardTitle>
              </CardHeader>
              <CardContent className='grid gap-3 md:grid-cols-3'>
                <div className='space-y-2'>
                  <Label>{t('Code')}</Label>
                  <Input
                    value={promoCode}
                    onChange={(e) => setPromoCode(e.target.value)}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Discount Type')}</Label>
                  <Input
                    value={promoDiscountType}
                    onChange={(e) => setPromoDiscountType(e.target.value)}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Discount Value')}</Label>
                  <Input
                    value={promoDiscountValue}
                    onChange={(e) => setPromoDiscountValue(e.target.value)}
                  />
                </div>
                <div className='space-y-2'>
                  <Label>{t('Max Redemptions')}</Label>
                  <Input
                    value={promoMaxRedemptions}
                    onChange={(e) => setPromoMaxRedemptions(e.target.value)}
                  />
                </div>
                <div className='md:col-span-2'>
                  <DateRangeField
                    startValue={promoStartsAt}
                    endValue={promoExpiresAt}
                    onStartChange={setPromoStartsAt}
                    onEndChange={setPromoExpiresAt}
                  />
                </div>
                <div className='flex items-end'>
                  <Button
                    className='w-full'
                    onClick={async () => {
                      await saveAgentPromoCode({
                        code: promoCode,
                        discount_type: promoDiscountType,
                        discount_value: Number(promoDiscountValue),
                        max_redemptions: Number(promoMaxRedemptions),
                        starts_at: Number(promoStartsAt),
                        expires_at: Number(promoExpiresAt),
                      })
                      toast.success(t('Saved'))
                      await refresh('promo')
                    }}
                  >
                    {t('Save')}
                  </Button>
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <CardTitle>{t('Promo Code List')}</CardTitle>
              </CardHeader>
              <CardContent className='overflow-x-auto'>
                <table className='w-full text-sm'>
                  <thead>
                    <tr className='border-b text-left'>
                      <th className='py-2'>{t('Code')}</th>
                      <th>{t('Status')}</th>
                      <th>{t('Type')}</th>
                      <th>{t('Value')}</th>
                      <th>{t('Used')}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {promoCodes.length === 0 && <EmptyTableRow colSpan={5} />}
                    {promoCodes.map((row) => (
                      <tr key={row.id} className='border-b'>
                        <td className='py-2 font-mono text-xs'>{row.code}</td>
                        <td>
                          <StatusBadge status={row.status} />
                        </td>
                        <td>{row.discount_type}</td>
                        <td>{row.discount_value}</td>
                        <td>
                          {row.used_count}/{row.max_redemptions}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
        <Separator className='my-6' />
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
