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
import { useState, type ComponentProps } from 'react'
import { Calendar as CalendarIcon } from 'lucide-react'
import { type DateRange } from 'react-day-picker'
import { enUS, fr, ja, ru, vi, zhCN } from 'react-day-picker/locale'
import { useTranslation } from 'react-i18next'
import dayjs from '@/lib/dayjs'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Label } from '@/components/ui/label'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { DatePicker } from '@/components/date-picker'

const minimumDate = new Date('1900-01-01')

const calendarLocales = {
  en: enUS,
  zh: zhCN,
  fr,
  ru,
  ja,
  vi,
} as const

export function unixSecondsToDate(value: string | number | undefined) {
  const seconds = Number(value)
  return seconds > 0 ? new Date(seconds * 1000) : undefined
}

export function dateToStartOfDayUnixSeconds(date: Date | undefined) {
  if (!date) return ''
  const nextDate = new Date(date)
  nextDate.setHours(0, 0, 0, 0)
  return String(Math.floor(nextDate.getTime() / 1000))
}

export function dateToEndOfDayUnixSeconds(date: Date | undefined) {
  if (!date) return ''
  const nextDate = new Date(date)
  nextDate.setHours(23, 59, 59, 999)
  return String(Math.floor(nextDate.getTime() / 1000))
}

function dateOnlyTime(date: Date) {
  const nextDate = new Date(date)
  nextDate.setHours(0, 0, 0, 0)
  return nextDate.getTime()
}

type DateOnlyFieldProps = {
  label: string
  value: string | number | undefined
  onChange: (value: string) => void
  endOfDay?: boolean
}

export function DateOnlyField({
  label,
  value,
  onChange,
  endOfDay = false,
}: DateOnlyFieldProps) {
  const { t } = useTranslation()

  return (
    <div className='space-y-2'>
      <Label>{t(label)}</Label>
      <DatePicker
        selected={unixSecondsToDate(value)}
        onSelect={(date) =>
          onChange(
            endOfDay
              ? dateToEndOfDayUnixSeconds(date)
              : dateToStartOfDayUnixSeconds(date)
          )
        }
        placeholder={t('Select date')}
        disabled={(date: Date) => date < minimumDate}
      />
    </div>
  )
}

type DateRangeFieldProps = {
  startLabel?: string
  endLabel?: string
  startValue: string | number | undefined
  endValue: string | number | undefined
  onStartChange: (value: string) => void
  onEndChange: (value: string) => void
  rangePicker?: boolean
  disabled?: ComponentProps<typeof Calendar>['disabled']
}

export function DateRangeField({
  startLabel = 'Starts At',
  endLabel = 'Expires At',
  startValue,
  endValue,
  onStartChange,
  onEndChange,
  rangePicker = false,
  disabled,
}: DateRangeFieldProps) {
  const { t, i18n } = useTranslation()
  const [open, setOpen] = useState(false)
  const startDate = unixSecondsToDate(startValue)
  const endDate = unixSecondsToDate(endValue)
  const calendarLocale =
    calendarLocales[i18n.language as keyof typeof calendarLocales] ?? enUS

  const updateStart = (date: Date | undefined) => {
    onStartChange(dateToStartOfDayUnixSeconds(date))
    if (date && endDate && dateOnlyTime(date) > dateOnlyTime(endDate)) {
      onEndChange(dateToEndOfDayUnixSeconds(date))
    }
  }

  const updateEnd = (date: Date | undefined) => {
    onEndChange(dateToEndOfDayUnixSeconds(date))
    if (date && startDate && dateOnlyTime(date) < dateOnlyTime(startDate)) {
      onStartChange(dateToStartOfDayUnixSeconds(date))
    }
  }

  if (rangePicker) {
    const selectedRange: DateRange | undefined =
      startDate || endDate ? { from: startDate, to: endDate } : undefined
    const rangeLabel =
      startDate && endDate
        ? `${dayjs(startDate).format('YYYY-MM-DD')} - ${dayjs(endDate).format('YYYY-MM-DD')}`
        : startDate
          ? dayjs(startDate).format('YYYY-MM-DD')
          : ''

    return (
      <div className='space-y-2'>
        <Label>{t('Time Range')}</Label>
        <Popover open={open} onOpenChange={setOpen}>
          <PopoverTrigger
            render={
              <Button
                variant='outline'
                data-empty={!rangeLabel}
                className='data-[empty=true]:text-muted-foreground w-full justify-start text-start font-normal'
              />
            }
          >
            {rangeLabel || <span>{t('Select date')}</span>}
            <CalendarIcon className='ms-auto h-4 w-4 opacity-50' />
          </PopoverTrigger>
          <PopoverContent className='w-auto p-0'>
            <Calendar
              mode='range'
              numberOfMonths={2}
              captionLayout='dropdown'
              selected={selectedRange}
              onSelect={(range) => {
                onStartChange(dateToStartOfDayUnixSeconds(range?.from))
                onEndChange(dateToEndOfDayUnixSeconds(range?.to))
                if (range?.from && range?.to) {
                  setOpen(false)
                }
              }}
              locale={calendarLocale}
              disabled={disabled ?? ((date: Date) => date < minimumDate)}
            />
          </PopoverContent>
        </Popover>
      </div>
    )
  }

  return (
    <div className='grid gap-3 md:grid-cols-2'>
      <div className='space-y-2'>
        <Label>{t(startLabel)}</Label>
        <DatePicker
          selected={startDate}
          onSelect={updateStart}
          placeholder={t('Select date')}
          disabled={(date: Date) => date < minimumDate}
        />
      </div>
      <div className='space-y-2'>
        <Label>{t(endLabel)}</Label>
        <DatePicker
          selected={endDate}
          onSelect={updateEnd}
          placeholder={t('Select date')}
          disabled={(date: Date) =>
            date < minimumDate ||
            (startDate ? dateOnlyTime(date) < dateOnlyTime(startDate) : false)
          }
        />
      </div>
    </div>
  )
}
