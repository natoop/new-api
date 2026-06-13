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
// Business Lead Schema & Types — mirrors model/business_lead.go
// ============================================================================

export const businessLeadSchema = z.object({
  id: z.number(),
  company_name: z.string(),
  contact_name: z.string(),
  contact_info: z.string(),
  cooperation_type: z.string(),
  requirements: z.string(),
  status: z.string(), // pending | contacted | archived
  created_at: z.number(), // unix seconds
})

export type BusinessLead = z.infer<typeof businessLeadSchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface ListBusinessLeadsParams {
  page?: number
  page_size?: number
  status?: string
  keyword?: string
}

export interface ListBusinessLeadsResponse {
  success: boolean
  message?: string
  data?: {
    items: BusinessLead[]
    total: number
    page: number
    page_size: number
  }
}
