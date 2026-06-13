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
import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog'
import { type BusinessLead } from '../types'

export type BusinessLeadsDialogType = 'view' | 'delete'

type BusinessLeadsContextType = {
  open: BusinessLeadsDialogType | null
  setOpen: (str: BusinessLeadsDialogType | null) => void
  currentRow: BusinessLead | null
  setCurrentRow: React.Dispatch<React.SetStateAction<BusinessLead | null>>
}

const BusinessLeadsContext =
  React.createContext<BusinessLeadsContextType | null>(null)

export function BusinessLeadsProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const [open, setOpen] = useDialogState<BusinessLeadsDialogType>(null)
  const [currentRow, setCurrentRow] = useState<BusinessLead | null>(null)

  return (
    <BusinessLeadsContext
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
      }}
    >
      {children}
    </BusinessLeadsContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useBusinessLeads = () => {
  const ctx = React.useContext(BusinessLeadsContext)

  if (!ctx) {
    throw new Error(
      'useBusinessLeads has to be used within <BusinessLeadsProvider>'
    )
  }

  return ctx
}
