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
import { z } from 'zod'

// ============================================================================
// Business lead form schema — mirrors POST /api/business/lead body
// ============================================================================

export const businessLeadFormSchema = z.object({
  company_name: z.string().min(1, 'Please enter your company name'),
  contact_name: z.string().min(1, 'Please enter a contact name'),
  contact_info: z.string().min(1, 'Please enter your contact details'),
  cooperation_type: z.enum(['api_wholesale', 'reseller', 'integration', 'other']),
  requirements: z
    .string()
    .min(5, 'Please describe your needs in at least 5 characters'),
})

export type BusinessLeadFormValues = z.infer<typeof businessLeadFormSchema>

// ============================================================================
// Cooperation type options — value matches API contract, labelKey goes through t()
// ============================================================================

export const COOPERATION_TYPE_OPTIONS = [
  { value: 'api_wholesale', labelKey: 'API wholesale & volume pricing' },
  { value: 'reseller', labelKey: 'Reseller & channel partnership' },
  { value: 'integration', labelKey: 'Technical integration' },
  { value: 'other', labelKey: 'Other' },
] as const
