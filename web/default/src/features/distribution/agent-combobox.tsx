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
import { useEffect, useMemo, useState } from 'react'
import { Check, ChevronsUpDown } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { adminSearchAgents } from './api'
import type { DistributionAgent } from './types'

type AgentComboboxProps = {
  value?: string
  onValueChange: (value: string) => void
  placeholder?: string
  includeEmpty?: boolean
  emptyLabel?: string
  selectedAgent?: DistributionAgent
  onAgentSelected?: (agent: DistributionAgent | null) => void
}

export function formatAgentLabel(agent?: DistributionAgent) {
  if (!agent) return '-'
  const username = agent.username?.trim()
  const displayName = agent.display_name?.trim()
  const name = agent.name?.trim()
  if (username && displayName) return `${displayName} (${username})`
  return username || displayName || name || '-'
}

export function AgentCombobox({
  value,
  onValueChange,
  placeholder,
  includeEmpty,
  emptyLabel,
  selectedAgent,
  onAgentSelected,
}: AgentComboboxProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [agents, setAgents] = useState<DistributionAgent[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!open) return
    const timer = window.setTimeout(async () => {
      setLoading(true)
      try {
        const res = await adminSearchAgents({
          p: 1,
          page_size: 20,
          keyword: search,
        })
        if (res.success) {
          setAgents(res.data?.items || [])
        }
      } finally {
        setLoading(false)
      }
    }, 220)
    return () => window.clearTimeout(timer)
  }, [open, search])

  const selected = useMemo(() => {
    if (!value || value === '0') return undefined
    return selectedAgent || agents.find((agent) => String(agent.id) === value)
  }, [agents, selectedAgent, value])

  const handleSelect = (agent: DistributionAgent | null) => {
    onValueChange(agent ? String(agent.id) : '0')
    onAgentSelected?.(agent)
    setOpen(false)
    setSearch('')
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            type='button'
            variant='outline'
            role='combobox'
            aria-expanded={open}
            className='h-9 w-full justify-between gap-2 px-3 text-left font-normal'
          />
        }
      >
        <span className='min-w-0 flex-1 truncate'>
          {value === '0'
            ? emptyLabel || t('No parent agent')
            : selected
              ? formatAgentLabel(selected)
              : placeholder || t('Select agent')}
        </span>
        <ChevronsUpDown className='h-4 w-4 shrink-0 opacity-50' />
      </PopoverTrigger>
      <PopoverContent className='w-[var(--anchor-width)] p-0'>
        <Command shouldFilter={false}>
          <CommandInput
            value={search}
            onValueChange={setSearch}
            placeholder={t('Search agents by username...')}
          />
          <CommandList className='max-h-[320px]'>
            <CommandEmpty>
              {loading ? t('Loading') : t('No agent found.')}
            </CommandEmpty>
            <CommandGroup>
              {includeEmpty && (
                <CommandItem
                  value='0'
                  onSelect={() => handleSelect(null)}
                  className='gap-2'
                >
                  <Check
                    className={cn(
                      'h-4 w-4',
                      value === '0' ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  {emptyLabel || t('No parent agent')}
                </CommandItem>
              )}
              {agents.map((agent) => (
                <CommandItem
                  key={agent.id}
                  value={`${agent.username || ''} ${agent.display_name || ''} ${agent.name}`}
                  onSelect={() => handleSelect(agent)}
                  className='gap-2'
                >
                  <Check
                    className={cn(
                      'h-4 w-4',
                      String(agent.id) === value ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <span className='min-w-0'>
                    <span className='block truncate'>
                      {formatAgentLabel(agent)}
                    </span>
                    {agent.name && agent.name !== formatAgentLabel(agent) ? (
                      <span className='text-muted-foreground block truncate text-xs'>
                        {agent.name}
                      </span>
                    ) : null}
                  </span>
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
