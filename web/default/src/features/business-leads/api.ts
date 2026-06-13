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
import { api } from '@/lib/api'
import type {
  ApiResponse,
  ListBusinessLeadsParams,
  ListBusinessLeadsResponse,
} from './types'

// ============================================================================
// Business Lead Management (Admin)
// ============================================================================

// Get paginated business leads list. Supports optional status filter and
// keyword search (company_name / contact_name).
export async function listBusinessLeads(
  params: ListBusinessLeadsParams = {}
): Promise<ListBusinessLeadsResponse> {
  const { page = 1, page_size = 20, status, keyword } = params
  const query = new URLSearchParams({
    page: String(page),
    page_size: String(page_size),
  })
  if (status) {
    query.set('status', status)
  }
  if (keyword) {
    query.set('keyword', keyword)
  }
  const res = await api.get(`/api/business-lead/?${query.toString()}`)
  return res.data
}

// Update a business lead status (pending | contacted | archived).
export async function updateBusinessLeadStatus(
  id: number,
  status: string
): Promise<ApiResponse<{ id: number; status: string }>> {
  const res = await api.put(`/api/business-lead/${id}/status`, { status })
  return res.data
}

// Delete a business lead.
export async function deleteBusinessLead(
  id: number
): Promise<ApiResponse<{ id: number }>> {
  const res = await api.delete(`/api/business-lead/${id}`)
  return res.data
}
