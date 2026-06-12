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

type Translate = (key: string) => string

const statusLabels: Record<string, string> = {
  accepted: 'Accepted',
  assigned: 'Assigned',
  available: 'Available',
  cancelled: 'Cancelled',
  disabled: 'Disabled',
  enabled: 'Enabled',
  expired: 'Expired',
  fulfilled: 'Fulfilled',
  granted: 'Granted',
  paid: 'Paid',
  pending: 'Pending',
  posted: 'Posted',
  redeemed: 'Redeemed',
  refunded: 'Refunded',
  reserved: 'Reserved',
  revoked: 'Revoked',
  voided: 'Voided',
}

export const distributionOrderStatuses = [
  'pending',
  'paid',
  'fulfilled',
  'refunded',
  'cancelled',
] as const

export const distributionOrderTypes = [
  'original',
  'inventory',
  'redeem',
] as const

const orderStatusLabels: Record<string, string> = {
  cancelled: 'Cancelled',
  fulfilled: 'Fulfilled',
  paid: 'Paid',
  pending: 'Pending',
  refunded: 'Refunded',
}

const orderTypeLabels: Record<string, string> = {
  inventory: 'Agent Inventory Order',
  original: 'Original Order',
  redeem: 'Redemption Order',
}

const scopeLabels: Record<string, string> = {
  agent: 'Agent',
  global: 'Global',
  level: 'Level',
}

const priceTargetLabels: Record<string, string> = {
  customer: 'Customer',
  level: 'Agent Level',
}

const priceTypeLabels: Record<string, string> = {
  discount: 'Discount',
  fixed: 'Price',
}

const ledgerEntryLabels: Record<string, string> = {
  credit: 'Credit',
  debit: 'Debit',
}

const sourceTypeLabels: Record<string, string> = {
  adjust: 'Adjustment',
  assign: 'Assign',
  bind: 'Bind',
  compensation: 'Compensation',
  invitation: 'Invitation',
  order: 'Order',
  profit: 'Profit',
  promo: 'Promotion',
  purchase: 'Purchase',
  redeem: 'Redeem',
  refund: 'Refund',
  sale: 'Sale',
}

const discountTypeLabels: Record<string, string> = {
  amount: 'Fixed Amount',
  percent: 'Percent',
}

export function distributionStatusLabel(
  status: string | undefined,
  t: Translate
) {
  if (!status) return '-'
  return t(statusLabels[status] || status)
}

export function distributionOrderStatusLabel(
  status: string | undefined,
  t: Translate
) {
  if (!status) return '-'
  return t(orderStatusLabels[status] || status)
}

export function distributionOrderTypeLabel(
  type: string | undefined,
  t: Translate
) {
  if (!type) return '-'
  return t(orderTypeLabels[type] || type)
}

export function distributionScopeLabel(
  scope: string | undefined,
  t: Translate
) {
  if (!scope) return '-'
  return t(scopeLabels[scope] || scope)
}

export function distributionPriceTargetLabel(
  target: string | undefined,
  t: Translate
) {
  if (!target) return '-'
  return t(priceTargetLabels[target] || target)
}

export function distributionPriceTypeLabel(
  type: string | undefined,
  t: Translate
) {
  if (!type) return '-'
  return t(priceTypeLabels[type] || type)
}

export function distributionLedgerEntryLabel(
  type: string | undefined,
  t: Translate
) {
  if (!type) return '-'
  return t(ledgerEntryLabels[type] || type)
}

export function distributionSourceTypeLabel(
  type: string | undefined,
  t: Translate
) {
  if (!type) return '-'
  return t(sourceTypeLabels[type] || type)
}

export function distributionDiscountTypeLabel(
  type: string | undefined,
  t: Translate
) {
  if (!type) return '-'
  return t(discountTypeLabels[type] || type)
}
