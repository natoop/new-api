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
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { submitBusinessLead } from '../api'
import {
  businessLeadFormSchema,
  COOPERATION_TYPE_OPTIONS,
  type BusinessLeadFormValues,
} from '../constants'

const DEFAULT_VALUES: BusinessLeadFormValues = {
  company_name: '',
  contact_name: '',
  contact_info: '',
  cooperation_type: 'api_wholesale',
  requirements: '',
}

export function CooperationForm() {
  const { t } = useTranslation()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<BusinessLeadFormValues>({
    resolver: zodResolver(businessLeadFormSchema),
    defaultValues: DEFAULT_VALUES,
  })

  async function onSubmit(values: BusinessLeadFormValues) {
    setIsSubmitting(true)
    try {
      const res = await submitBusinessLead(values)
      if (res?.success) {
        toast.success(t('Submitted! Our team will reach out soon.'))
        form.reset(DEFAULT_VALUES)
      }
      // Business failures are surfaced by the api interceptor toast.
    } catch (_error) {
      // Network/HTTP errors are handled by the global interceptor.
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='grid gap-4'>
        <FormField
          control={form.control}
          name='company_name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Company name')}</FormLabel>
              <FormControl>
                <Input placeholder={t('Your company or team name')} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='contact_name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Contact name')}</FormLabel>
              <FormControl>
                <Input placeholder={t('Who should we reach out to?')} {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='contact_info'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Contact details')}</FormLabel>
              <FormControl>
                <Input
                  placeholder={t('Phone, email, or WeChat')}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='cooperation_type'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Cooperation type')}</FormLabel>
              <Select
                items={COOPERATION_TYPE_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(o.labelKey),
                }))}
                value={field.value}
                onValueChange={field.onChange}
              >
                <FormControl>
                  <SelectTrigger className='w-full'>
                    <SelectValue placeholder={t('Select a cooperation type')} />
                  </SelectTrigger>
                </FormControl>
                <SelectContent alignItemWithTrigger={false}>
                  <SelectGroup>
                    {COOPERATION_TYPE_OPTIONS.map((o) => (
                      <SelectItem key={o.value} value={o.value}>
                        {t(o.labelKey)}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='requirements'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Your needs')}</FormLabel>
              <FormControl>
                <Textarea
                  rows={4}
                  placeholder={t(
                    'Tell us about your use case, expected volume, and timeline.'
                  )}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button
          type='submit'
          className='mt-1 w-full justify-center gap-2'
          disabled={isSubmitting}
        >
          {isSubmitting ? <Loader2 className='h-4 w-4 animate-spin' /> : null}
          {t('Submit')}
        </Button>
      </form>
    </Form>
  )
}
