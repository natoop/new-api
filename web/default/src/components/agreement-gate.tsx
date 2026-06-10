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
import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useAuthStore } from '@/stores/auth-store'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Markdown } from '@/components/ui/markdown'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { logout } from '@/features/auth/api'
import {
  consentAgreement,
  getAgreementStatus,
  getUserAgreement,
} from '@/features/legal/api'

const STATUS_QUERY_KEY = ['user-agreement-status']

/**
 * Blocking consent gate shown after sign-in when the admin requires users to
 * (re-)sign the service agreement. The dialog cannot be dismissed by clicking
 * outside or pressing Escape — the user must either agree or sign out.
 */
export function AgreementGate() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { auth } = useAuthStore()
  const user = auth.user
  const [accepted, setAccepted] = useState(false)
  const [isSigningOut, setIsSigningOut] = useState(false)

  const statusQuery = useQuery({
    queryKey: STATUS_QUERY_KEY,
    queryFn: getAgreementStatus,
    enabled: Boolean(user),
    staleTime: 60 * 1000,
  })

  const status = statusQuery.data?.data
  const blocked = Boolean(
    user && statusQuery.data?.success && status?.required && !status?.agreed
  )

  const agreementQuery = useQuery({
    queryKey: ['user-agreement'],
    queryFn: getUserAgreement,
    enabled: blocked,
    staleTime: 10 * 60 * 1000,
  })

  const consentMutation = useMutation({
    mutationFn: () => consentAgreement(status?.version ?? ''),
    onSuccess: (res) => {
      if (res.success) {
        toast.success(t('Agreement signed successfully'))
      }
      // On business failure the api interceptor already toasts the message;
      // refetch so a stale version gets refreshed either way.
      queryClient.invalidateQueries({ queryKey: STATUS_QUERY_KEY })
    },
    onError: () => {
      // e.g. 400 version mismatch — message toast handled by the api
      // interceptor; refetch the latest required version.
      queryClient.invalidateQueries({ queryKey: STATUS_QUERY_KEY })
    },
  })

  const handleDecline = async () => {
    setIsSigningOut(true)
    try {
      await logout()
    } catch {
      /* empty */
    }
    auth.reset()
    try {
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem('uid')
      }
    } catch {
      /* empty */
    }
    if (typeof window !== 'undefined') {
      window.location.href = '/sign-in'
    }
  }

  if (!blocked) return null

  const content = agreementQuery.data?.data?.trim() ?? ''
  const isBusy = consentMutation.isPending || isSigningOut

  return (
    // Controlled open with a no-op onOpenChange: Escape / outside clicks
    // request a close via onOpenChange, which we ignore — the dialog stays up.
    <Dialog open onOpenChange={() => {}}>
      <DialogContent
        showCloseButton={false}
        className='rounded-xl sm:max-w-2xl'
      >
        <DialogHeader>
          <DialogTitle>{t('Service Agreement and Terms of Use')}</DialogTitle>
          <DialogDescription>
            {t(
              'Please read the following agreement carefully. You must agree before continuing to use the console.'
            )}
          </DialogDescription>
        </DialogHeader>

        <ScrollArea className='bg-muted/30 max-h-[50vh] overflow-hidden rounded-lg border'>
          <div className='p-4'>
            {agreementQuery.isLoading ? (
              <div className='flex flex-col gap-3'>
                <Skeleton className='h-4 w-[60%]' />
                <Skeleton className='h-4 w-full' />
                <Skeleton className='h-4 w-[90%]' />
                <Skeleton className='h-4 w-[80%]' />
              </div>
            ) : content ? (
              <Markdown>{content}</Markdown>
            ) : (
              <p className='text-muted-foreground text-sm'>
                {t(
                  'The administrator has not configured a user agreement yet.'
                )}
              </p>
            )}
          </div>
        </ScrollArea>

        <div className='flex items-start gap-3'>
          <Checkbox
            id='agreement-gate-consent'
            checked={accepted}
            onCheckedChange={(value) => setAccepted(value === true)}
            className='mt-0.5'
            disabled={isBusy}
          />
          <Label
            htmlFor='agreement-gate-consent'
            className='text-sm leading-5 font-normal'
          >
            {t('I have read and agree to the above agreement')}
          </Label>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={handleDecline} disabled={isBusy}>
            {t('Decline and Sign Out')}
          </Button>
          <Button
            onClick={() => consentMutation.mutate()}
            disabled={!accepted || isBusy}
          >
            {t('Agree and Continue')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
