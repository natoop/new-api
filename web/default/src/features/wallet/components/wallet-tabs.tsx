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
import { Crown, Receipt, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs'
import type { TopupInfo } from '../types'
import { PayAsYouGoTab } from './pay-as-you-go-tab'
import { PurchasePlansTab } from './purchase-plans-tab'

interface WalletTabsProps {
  topupInfo: TopupInfo | null
  userQuota?: number
  onPurchaseSuccess?: () => void | Promise<void>
  onRedeemed?: () => void | Promise<void>
  onOpenBilling?: () => void
}

export function WalletTabs({
  topupInfo,
  userQuota,
  onPurchaseSuccess,
  onRedeemed,
  onOpenBilling,
}: WalletTabsProps) {
  const { t } = useTranslation()

  return (
    <Tabs defaultValue='plans' className='gap-3 sm:gap-4'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <TabsList className='h-9'>
          <TabsTrigger value='plans' className='gap-1.5 px-3'>
            <Crown className='h-3.5 w-3.5' />
            {t('Buy Plans')}
          </TabsTrigger>
          <TabsTrigger value='topup' className='gap-1.5 px-3'>
            <WalletCards className='h-3.5 w-3.5' />
            {t('Pay as you go')}
          </TabsTrigger>
        </TabsList>
        {onOpenBilling && (
          <Button
            variant='outline'
            size='sm'
            onClick={onOpenBilling}
            className='gap-2'
          >
            <Receipt className='h-4 w-4' />
            {t('Order History')}
          </Button>
        )}
      </div>

      <TabsContent value='plans'>
        <PurchasePlansTab
          topupInfo={topupInfo}
          userQuota={userQuota}
          onPurchaseSuccess={onPurchaseSuccess}
        />
      </TabsContent>

      <TabsContent value='topup'>
        <PayAsYouGoTab topupInfo={topupInfo} onPaid={onRedeemed} />
      </TabsContent>
    </Tabs>
  )
}
