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
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { getSelf } from '@/lib/api'
import { SectionPageLayout } from '@/components/layout'
import { AffiliateRewardsCard } from './components/affiliate-rewards-card'
import { AgentBanner } from './components/agent-banner'
import { BillingHistoryDialog } from './components/dialogs/billing-history-dialog'
import { TransferDialog } from './components/dialogs/transfer-dialog'
import { MySubscriptionCard } from './components/my-subscription-card'
import { RedemptionCard } from './components/redemption-card'
import { WalletStatsCard } from './components/wallet-stats-card'
import { WalletTabs } from './components/wallet-tabs'
import { useTopupInfo, useAffiliate } from './hooks'
import type { UserWalletData } from './types'

interface WalletProps {
  initialShowHistory?: boolean
}

export function Wallet(props: WalletProps) {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserWalletData | null>(null)
  const [userLoading, setUserLoading] = useState(true)
  const [transferDialogOpen, setTransferDialogOpen] = useState(false)
  const [billingDialogOpen, setBillingDialogOpen] = useState(false)
  const [walletTab, setWalletTab] = useState<'topup' | 'plans'>('topup')
  const [subscriptionRefreshKey, setSubscriptionRefreshKey] = useState(0)

  const { topupInfo } = useTopupInfo()
  const {
    affiliateLink,
    loading: affiliateLoading,
    transferQuota,
    transferring,
  } = useAffiliate()

  // Fetch and refresh user data
  const fetchUser = useCallback(async () => {
    try {
      setUserLoading(true)
      const response = await getSelf()
      if (response.success && response.data) {
        setUser(response.data as UserWalletData)
      }
    } catch (error) {
      // eslint-disable-next-line no-console
      console.error('Failed to fetch user data:', error)
    } finally {
      setUserLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchUser()
  }, [fetchUser])

  useEffect(() => {
    if (props.initialShowHistory) {
      setBillingDialogOpen(true)
      window.history.replaceState({}, '', window.location.pathname)
    }
  }, [props.initialShowHistory])

  // Handle transfer
  const handleTransfer = async (amount: number) => {
    const success = await transferQuota(amount)
    if (success) {
      await fetchUser()
    }
    return success
  }

  const handlePurchaseSuccess = async () => {
    await fetchUser()
    setSubscriptionRefreshKey((key) => key + 1)
  }

  const goTopupTab = () => setWalletTab('topup')
  const goPlansTab = () => setWalletTab('plans')

  return (
    <>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Wallet')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='mx-auto flex w-full max-w-7xl flex-col gap-4 sm:gap-5'>
            <WalletStatsCard user={user} loading={userLoading} />

            <div className='grid gap-4 lg:grid-cols-[minmax(0,1fr)_360px] xl:grid-cols-[minmax(0,1fr)_400px]'>
              <div className='min-w-0 space-y-4'>
                <WalletTabs
                  value={walletTab}
                  onValueChange={(value) =>
                    setWalletTab(value as 'topup' | 'plans')
                  }
                  topupInfo={topupInfo}
                  userQuota={user?.quota}
                  onPurchaseSuccess={handlePurchaseSuccess}
                  onRedeemed={fetchUser}
                  onOpenBilling={() => setBillingDialogOpen(true)}
                  onGoTopup={goTopupTab}
                />

                <RedemptionCard
                  enabled={topupInfo?.enable_redemption !== false}
                  topupLink={topupInfo?.topup_link}
                  onRedeemed={async () => {
                    await fetchUser()
                    setSubscriptionRefreshKey((key) => key + 1)
                  }}
                />
              </div>

              <div className='min-w-0 space-y-4'>
                <MySubscriptionCard
                  refreshKey={subscriptionRefreshKey}
                  onGoPlans={goPlansTab}
                />

                <AgentBanner />

                <AffiliateRewardsCard
                  user={user}
                  affiliateLink={affiliateLink}
                  onTransfer={() => setTransferDialogOpen(true)}
                  complianceConfirmed={
                    topupInfo?.payment_compliance_confirmed !== false
                  }
                  loading={affiliateLoading}
                />
              </div>
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <TransferDialog
        open={transferDialogOpen}
        onOpenChange={setTransferDialogOpen}
        onConfirm={handleTransfer}
        availableQuota={user?.aff_quota ?? 0}
        transferring={transferring}
      />

      <BillingHistoryDialog
        open={billingDialogOpen}
        onOpenChange={setBillingDialogOpen}
      />
    </>
  )
}
