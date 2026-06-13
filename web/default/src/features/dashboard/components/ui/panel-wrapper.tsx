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
import { type ReactNode } from 'react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Skeleton } from '@/components/ui/skeleton'

interface PanelWrapperProps {
  title: ReactNode
  description?: ReactNode
  icon?: ReactNode
  loading?: boolean
  empty?: boolean
  emptyMessage?: string
  height?: string
  className?: string
  contentClassName?: string
  headerActions?: ReactNode
  children?: ReactNode
}

function PanelHeader(props: {
  title: ReactNode
  description?: ReactNode
  icon?: ReactNode
  actions?: ReactNode
}) {
  const heading = (
    <div className='flex min-w-0 items-center gap-3'>
      {props.icon != null && (
        <span className='bg-primary/10 text-primary ring-primary/15 flex size-8 shrink-0 items-center justify-center rounded-lg ring-1 ring-inset'>
          {props.icon}
        </span>
      )}
      <div className='flex min-w-0 flex-col gap-0.5'>
        <div className='truncate text-sm font-semibold'>{props.title}</div>
        {props.description != null && (
          <div className='text-muted-foreground truncate text-xs'>
            {props.description}
          </div>
        )}
      </div>
    </div>
  )

  return (
    <div className='from-muted/30 border-b bg-gradient-to-b to-transparent px-4 py-3 sm:px-5'>
      {props.actions != null ? (
        <div className='flex items-center justify-between gap-2'>
          {heading}
          <div className='shrink-0'>{props.actions}</div>
        </div>
      ) : (
        heading
      )}
    </div>
  )
}

export function PanelWrapper(props: PanelWrapperProps) {
  const { t } = useTranslation()
  const resolvedEmptyMessage = props.emptyMessage ?? t('No data available')
  const height = props.height ?? 'h-64'
  const frameClassName = cn(
    'glass-card overflow-hidden rounded-2xl',
    props.className
  )

  if (props.loading) {
    return (
      <div className={frameClassName}>
        <PanelHeader
          title={props.title}
          description={props.description}
          icon={props.icon}
        />
        <div className={cn('p-4 sm:p-5', props.contentClassName)}>
          <Skeleton className={`w-full ${height}`} />
        </div>
      </div>
    )
  }

  if (props.empty) {
    return (
      <div className={frameClassName}>
        <PanelHeader
          title={props.title}
          description={props.description}
          icon={props.icon}
        />
        <div
          className={cn(
            'text-muted-foreground flex items-center justify-center px-4 text-sm',
            height,
            props.contentClassName
          )}
        >
          {resolvedEmptyMessage}
        </div>
      </div>
    )
  }

  return (
    <div className={frameClassName}>
      <PanelHeader
        title={props.title}
        description={props.description}
        icon={props.icon}
        actions={props.headerActions}
      />
      <div className={cn('p-4 sm:p-5', props.contentClassName)}>
        {props.children}
      </div>
    </div>
  )
}
