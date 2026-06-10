package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	DistributionMaxBPS                   = 10000
	DistributionIdempotencyKeyPattern    = `^[A-Za-z0-9._:-]+$`
	DistributionIdempotencyKeyMinLen     = 8
	DistributionIdempotencyKeyMaxLen     = 128
	DistributionStatusEnabled            = "enabled"
	DistributionStatusDisabled           = "disabled"
	DistributionOrderStatusPending       = "pending"
	DistributionOrderStatusPaid          = "paid"
	DistributionOrderStatusFulfilled     = "fulfilled"
	DistributionOrderStatusCancelled     = "cancelled"
	DistributionInventoryStatusAvailable = "available"
	DistributionInventoryStatusReserved  = "reserved"
	DistributionInventoryStatusAssigned  = "assigned"
	DistributionInventoryStatusRedeemed  = "redeemed"
	DistributionInventoryStatusVoided    = "voided"
	DistributionLedgerEntryCredit        = "credit"
	DistributionLedgerEntryDebit         = "debit"
	DistributionSourceAdjust             = "adjust"
	DistributionSourcePurchase           = "purchase"
	DistributionSourceInventory          = "inventory"
	DistributionSourceInvitation         = "invitation"
	DistributionSourceProfit             = "profit"
	DistributionSourceRefund             = "refund"
	DistributionLogStatusPosted          = "posted"
	DistributionCustomerEventBind        = "bind"
	DistributionCustomerEventPurchase    = "purchase"
	DistributionCustomerEventAssign      = "assign"
	DistributionPriceScopeGlobal         = "global"
	DistributionPriceScopeLevel          = "level"
	DistributionPriceScopeAgent          = "agent"
	DistributionInvitationStatusPending  = "pending"
	DistributionInvitationStatusAccepted = "accepted"
	DistributionInvitationStatusExpired  = "expired"
	DistributionInvitationStatusRevoked  = "revoked"
	DistributionDiscountTypePercent      = "percent"
	DistributionDiscountTypeAmount       = "amount"
)

var (
	ErrDistributionInvalidIdempotencyKey = errors.New("invalid idempotency_key")
	ErrDistributionInvalidBPS            = errors.New("invalid bps")
	ErrDistributionInvalidAmount         = errors.New("invalid amount")
	ErrDistributionInvalidStatus         = errors.New("invalid status")
	distributionIdempotencyKeyRegexp     = regexp.MustCompile(DistributionIdempotencyKeyPattern)
)

type DistributionPriceConfigRule struct {
	ScopeType string
	AgentId   int
	Level     int
	UnitPrice int
	Status    string
}

func distributionStableHex(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(sum[:20])
}

func NormalizeIdempotencyKey(raw string) (string, error) {
	key := strings.TrimSpace(raw)
	if len(key) < DistributionIdempotencyKeyMinLen || len(key) > DistributionIdempotencyKeyMaxLen {
		return "", ErrDistributionInvalidIdempotencyKey
	}
	if !distributionIdempotencyKeyRegexp.MatchString(key) {
		return "", ErrDistributionInvalidIdempotencyKey
	}
	return key, nil
}

func CalcCommission(amountCent int, bps int) (int, error) {
	if amountCent < 0 {
		return 0, ErrDistributionInvalidAmount
	}
	if bps < 0 || bps > DistributionMaxBPS {
		return 0, ErrDistributionInvalidBPS
	}
	return int(int64(amountCent) * int64(bps) / int64(DistributionMaxBPS)), nil
}

func ValidateDistributionStatus(status string) error {
	switch strings.TrimSpace(status) {
	case DistributionStatusEnabled, DistributionStatusDisabled:
		return nil
	default:
		return ErrDistributionInvalidStatus
	}
}

func BuildPurchaseOrderNo(userID int, packageID int, key string) string {
	return "distribution_order_idem_" + distributionStableHex(fmt.Sprintf("purchase:%d:%d:%s", userID, packageID, key))
}

func BuildBalanceRef(agentID int, key string) string {
	return "distribution_balance_" + distributionStableHex(fmt.Sprintf("balance:%d:%s", agentID, key))
}

func BuildLedgerNo(agentID int, sourceType string, sourceNo string) string {
	return "distribution_ledger_" + distributionStableHex(fmt.Sprintf("ledger:%d:%s:%s", agentID, sourceType, sourceNo))
}

func BuildProfitNo(orderID int, parentAgentID int) string {
	return "distribution_profit_" + distributionStableHex(fmt.Sprintf("profit:%d:%d", orderID, parentAgentID))
}

func BuildInvitationNo(parentAgentID int, inviteeEmail string, key string) string {
	return "distribution_invitation_" + distributionStableHex(fmt.Sprintf("invite:%d:%s:%s", parentAgentID, strings.TrimSpace(strings.ToLower(inviteeEmail)), key))
}

func CanApplyDelta(balance int, delta int) bool {
	return balance+delta >= 0
}

func CanTransitionOrder(from string, to string) bool {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == to {
		return true
	}
	switch from {
	case DistributionOrderStatusPending:
		return to == DistributionOrderStatusPaid || to == DistributionOrderStatusCancelled
	case DistributionOrderStatusPaid:
		return to == DistributionOrderStatusFulfilled || to == DistributionOrderStatusCancelled
	default:
		return false
	}
}

func CanAssignInventory(status string, assignedTo int) bool {
	status = strings.TrimSpace(status)
	return assignedTo == 0 && (status == DistributionInventoryStatusAvailable || status == DistributionInventoryStatusReserved)
}

func ResolveDistributionAgentPrice(packageDefaultPrice int, agentID int, agentLevel int, configs []DistributionPriceConfigRule) int {
	for _, cfg := range configs {
		if cfg.ScopeType == DistributionPriceScopeAgent && cfg.AgentId == agentID && cfg.Status == DistributionStatusEnabled && cfg.UnitPrice >= 0 {
			return cfg.UnitPrice
		}
	}
	for _, cfg := range configs {
		if cfg.ScopeType == DistributionPriceScopeLevel && cfg.Level == agentLevel && cfg.Status == DistributionStatusEnabled && cfg.UnitPrice >= 0 {
			return cfg.UnitPrice
		}
	}
	for _, cfg := range configs {
		if cfg.ScopeType == DistributionPriceScopeGlobal && cfg.Status == DistributionStatusEnabled && cfg.UnitPrice >= 0 {
			return cfg.UnitPrice
		}
	}
	return packageDefaultPrice
}

func ValidateDistributionPromoDiscount(discountType string, value int) error {
	switch strings.TrimSpace(discountType) {
	case DistributionDiscountTypePercent:
		if value < 0 || value > DistributionMaxBPS {
			return ErrDistributionInvalidAmount
		}
	case DistributionDiscountTypeAmount:
		if value < 0 {
			return ErrDistributionInvalidAmount
		}
	default:
		return fmt.Errorf("invalid discount_type")
	}
	return nil
}

func ValidateDistributionTimeWindow(startsAt int64, expiresAt int64) error {
	if startsAt > 0 && expiresAt > 0 && startsAt >= expiresAt {
		return fmt.Errorf("starts_at must be earlier than expires_at")
	}
	return nil
}
