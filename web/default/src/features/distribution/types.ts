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
export interface ApiResponse<T = unknown> {
  success: boolean
  message: string
  data: T
}

export interface DistributionAgent {
  id: number
  user_id: number
  name: string
  status: string
  balance: number
  commission_bps: number
  parent_agent_id: number
  level: number
  contact: string
  remark: string
  created_at: number
  updated_at: number
  username?: string
  display_name?: string
  email?: string
}

export interface DistributionPackage {
  id: number
  subscription_plan_id: number
  subscription_title: string
  subscription_subtitle: string
  name: string
  sku: string
  description: string
  status: string
  agent_price: number
  retail_price: number
  secondary_agent_price: number
  credit_amount: number
  sort_order: number
}

export interface DistributionInventory {
  id: number
  agent_id: number
  order_id: number
  package_id: number
  status: string
  credit_amount: number
  retail_price: number
  inventory_no: string
  assigned_to: number
  created_at: number
  username?: string
  display_name?: string
  email?: string
}

export interface DistributionLedger {
  id: number
  ledger_no: string
  agent_id: number
  entry_type: string
  source_type: string
  source_no: string
  delta: number
  balance_before: number
  balance_after: number
  description: string
  created_at: number
}

export interface DistributionProfit {
  id: number
  profit_no: string
  agent_id: number
  child_agent_id: number
  order_id: number
  amount: number
  parent_cost: number
  secondary_price: number
  status: string
  created_at: number
}

export interface DistributionInvitation {
  id: number
  invitation_no: string
  invitee_user_id: number
  invitee_email: string
  parent_agent_id: number
  level: number
  status: string
  accepted_agent_id: number
  expires_at: number
  accepted_at: number
}

export interface DistributionPromoCode {
  id: number
  agent_id: number
  package_id: number
  package_name?: string
  code: string
  status: string
  discount_type: string
  discount_value: number
  max_redemptions: number
  used_count: number
  starts_at: number
  expires_at: number
}

export interface DistributionCustomerOwnership {
  id: number
  customer_user_id: number
  agent_id: number
  source_type: string
  source_no: string
  bound_at: number
  username?: string
  display_name?: string
  email?: string
}

export interface DistributionAttributionLog {
  id: number
  customer_user_id: number
  agent_id: number
  source_type: string
  source_no: string
  event_type: string
  message: string
  created_at: number
}

export interface DistributionPriceConfig {
  id: number
  package_id: number
  target_type: string
  customer_user_id: number
  agent_level: number
  price_type: string
  price_value: number
  status: string
  remark: string
}

export interface DistributionGiftRule {
  id: number
  name: string
  package_id: number
  gift_package_id: number
  trigger_quantity: number
  gift_quantity: number
  starts_at: number
  expires_at: number
  status: string
}

export interface DistributionOpsAuthorization {
  id: number
  user_id: number
  status: string
  granted_by_user_id: number
  revoked_by_user_id: number
  granted_at: number
  revoked_at: number
  remark: string
}

export interface DistributionOrderRecord {
  id: number
  order_no: string
  buyer_username?: string
  buyer_display_name?: string
  buyer_email?: string
  subscription_plan_id: number
  subscription_title?: string
  package_name?: string
  original_amount: number
  discount_amount: number
  paid_amount: number
  commission_amount: number
  status: string
  paid_at?: number
  fulfilled_at?: number
  created_at: number
}

export interface DistributionProfile {
  agent: DistributionAgent
  available_inventory: number
  aff_code: string
}

export interface PaginatedData<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface DistributionUserOption {
  id: number
  username: string
  display_name: string
  email?: string
  role: number
  aff_code?: string
}

export interface DistributionInventoryPackageOption {
  package_id: number
  package_name: string
}

export interface DistributionSubscriptionPlan {
  id: number
  title: string
  subtitle?: string
  price_amount: number
  currency: string
  enabled: boolean
  total_amount: number
}

export interface DistributionSubscriptionPlanRecord {
  plan: DistributionSubscriptionPlan
}
