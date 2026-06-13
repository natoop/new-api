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
import { type ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { formatTimestampToDate } from '@/lib/format'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableColumnHeader } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import { TableId } from '@/components/table-id'
import {
  BUSINESS_LEAD_STATUSES,
  getCooperationTypeLabel,
} from '../constants'
import { type BusinessLead } from '../types'
import { DataTableRowActions } from './data-table-row-actions'

export function useBusinessLeadsColumns(): ColumnDef<BusinessLead>[] {
  const { t } = useTranslation()
  return [
    {
      accessorKey: 'id',
      meta: { label: t('ID'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('ID')} />
      ),
      cell: ({ row }) => (
        <TableId value={row.getValue('id') as number} className='w-[60px]' />
      ),
    },
    {
      accessorKey: 'company_name',
      meta: { label: t('Company Name'), mobileTitle: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Company Name')} />
      ),
      cell: ({ row }) => (
        <div className='max-w-[180px] truncate font-medium'>
          {row.getValue('company_name')}
        </div>
      ),
    },
    {
      accessorKey: 'contact_name',
      meta: { label: t('Contact Name') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Contact Name')} />
      ),
      cell: ({ row }) => (
        <div className='max-w-[140px] truncate'>
          {row.getValue('contact_name')}
        </div>
      ),
    },
    {
      accessorKey: 'contact_info',
      meta: { label: t('Contact Info') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Contact Info')} />
      ),
      cell: ({ row }) => (
        <div className='max-w-[180px] truncate font-mono text-sm'>
          {row.getValue('contact_info')}
        </div>
      ),
      enableSorting: false,
    },
    {
      accessorKey: 'cooperation_type',
      meta: { label: t('Cooperation Type'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Cooperation Type')} />
      ),
      cell: function CooperationTypeCell({ row }) {
        return (
          <StatusBadge
            label={getCooperationTypeLabel(
              t,
              row.getValue('cooperation_type') as string
            )}
            variant='neutral'
            copyable={false}
          />
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: 'requirements',
      meta: { label: t('Requirements'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Requirements')} />
      ),
      cell: ({ row }) => {
        const requirements = row.getValue('requirements') as string
        if (!requirements) {
          return <span className='text-muted-foreground text-sm'>-</span>
        }
        return (
          <Tooltip>
            <TooltipTrigger
              render={
                <div className='max-w-[200px] cursor-help truncate text-sm'>
                  {requirements}
                </div>
              }
            ></TooltipTrigger>
            <TooltipContent>
              <div className='max-w-[320px] text-xs whitespace-pre-wrap'>
                {requirements}
              </div>
            </TooltipContent>
          </Tooltip>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: 'status',
      meta: { label: t('Status'), mobileBadge: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Status')} />
      ),
      cell: ({ row }) => {
        const statusConfig = BUSINESS_LEAD_STATUSES[row.getValue('status') as string]
        if (!statusConfig) {
          return null
        }
        return (
          <StatusBadge
            label={t(statusConfig.labelKey)}
            variant={statusConfig.variant}
            copyable={false}
          />
        )
      },
      // Status filtering is performed server-side; rows arrive pre-filtered.
      filterFn: () => true,
    },
    {
      accessorKey: 'created_at',
      meta: { label: t('Submitted'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Submitted')} />
      ),
      cell: ({ row }) => (
        <div className='min-w-[140px] font-mono text-sm'>
          {formatTimestampToDate(row.getValue('created_at'))}
        </div>
      ),
    },
    {
      id: 'actions',
      cell: ({ row }) => <DataTableRowActions row={row} />,
    },
  ]
}
