package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
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
	DistributionOrderStatusRefunded      = "refunded"
	DistributionInventoryStatusAvailable = "available"
	DistributionInventoryStatusReserved  = "reserved"
	DistributionInventoryStatusAssigned  = "assigned"
	DistributionInventoryStatusRedeemed  = "redeemed"
	DistributionInventoryStatusVoided    = "voided"
	DistributionInventoryStatusRefunded  = "refunded"
	DistributionLedgerEntryCredit        = "credit"
	DistributionLedgerEntryDebit         = "debit"
	DistributionSourceAdjust             = "adjust"
	DistributionSourcePurchase           = "purchase"
	DistributionSourceInventory          = "inventory"
	DistributionSourceInvitation         = "invitation"
	DistributionSourceProfit             = "profit"
	DistributionSourceRefund             = "refund"
	DistributionLogStatusPosted          = "posted"
	DistributionLogStatusRefunded        = "refunded"
	DistributionCustomerEventBind        = "bind"
	DistributionCustomerEventPurchase    = "purchase"
	DistributionCustomerEventAssign      = "assign"
	DistributionPriceTargetCustomer      = "customer"
	DistributionPriceTargetLevel         = "level"
	DistributionPriceTypeDiscount        = "discount"
	DistributionPriceTypeFixed           = "fixed"
	DistributionInvitationStatusPending  = "pending"
	DistributionInvitationStatusAccepted = "accepted"
	DistributionInvitationStatusExpired  = "expired"
	DistributionInvitationStatusRevoked  = "revoked"
	DistributionAgentLevelPrimary        = 1
	DistributionAgentLevelSecondary      = 2
	DistributionDiscountTypePercent      = "percent"
	DistributionDiscountTypeAmount       = "amount"
)

var (
	ErrDistributionInvalidIdempotencyKey = errors.New("invalid idempotency_key")
	ErrDistributionInvalidBPS            = errors.New("invalid bps")
	ErrDistributionInvalidAmount         = errors.New("invalid amount")
	ErrDistributionInvalidStatus         = errors.New("invalid status")
	distributionIdempotencyKeyRegexp     = regexp.MustCompile(DistributionIdempotencyKeyPattern)
	distributionSnowflakeMu              sync.Mutex
	distributionSnowflakeLastMilli       int64
	distributionSnowflakeSequence        int64
)

type DistributionPriceConfigRule struct {
	TargetType     string
	CustomerUserId int
	AgentLevel     int
	PriceType      string
	PriceValue     float64
	Status         string
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

func CalcCommission(amount float64, bps int) (float64, error) {
	if amount < 0 {
		return 0, ErrDistributionInvalidAmount
	}
	if bps < 0 || bps > DistributionMaxBPS {
		return 0, ErrDistributionInvalidBPS
	}
	return normalizeDistributionMoneyAmount(amount * float64(bps) / float64(DistributionMaxBPS)), nil
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
	return strconv.FormatInt(nextDistributionSnowflakeID(), 10)
}

func nextDistributionSnowflakeID() int64 {
	const (
		epochMilli   = int64(1704067200000)
		workerID     = int64(1)
		sequenceMask = int64(4095)
	)
	distributionSnowflakeMu.Lock()
	defer distributionSnowflakeMu.Unlock()

	nowMilli := time.Now().UnixMilli()
	if nowMilli < distributionSnowflakeLastMilli {
		nowMilli = distributionSnowflakeLastMilli
	}
	if nowMilli == distributionSnowflakeLastMilli {
		distributionSnowflakeSequence = (distributionSnowflakeSequence + 1) & sequenceMask
		if distributionSnowflakeSequence == 0 {
			for nowMilli <= distributionSnowflakeLastMilli {
				nowMilli = time.Now().UnixMilli()
			}
		}
	} else {
		distributionSnowflakeSequence = 0
	}
	distributionSnowflakeLastMilli = nowMilli
	return ((nowMilli - epochMilli) << 22) | (workerID << 12) | distributionSnowflakeSequence
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

func CanApplyDelta(balance float64, delta float64) bool {
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

func applyDistributionPriceConfig(basePrice float64, priceType string, priceValue float64) float64 {
	switch priceType {
	case DistributionPriceTypeFixed:
		if priceValue >= 0 {
			return normalizeDistributionMoneyAmount(priceValue)
		}
	case DistributionPriceTypeDiscount:
		if priceValue >= 1 && priceValue <= 10 {
			return normalizeDistributionMoneyAmount(basePrice * priceValue / 10)
		}
	}
	return basePrice
}

func ResolveDistributionAgentPrice(packageDefaultPrice float64, agentLevel int, configs []DistributionPriceConfigRule) float64 {
	for _, cfg := range configs {
		if cfg.TargetType == DistributionPriceTargetLevel && cfg.AgentLevel == agentLevel && cfg.Status == DistributionStatusEnabled {
			return applyDistributionPriceConfig(packageDefaultPrice, cfg.PriceType, cfg.PriceValue)
		}
	}
	return packageDefaultPrice
}

func ResolveDistributionCustomerPrice(packageDefaultPrice float64, customerUserID int, configs []DistributionPriceConfigRule) float64 {
	for _, cfg := range configs {
		if cfg.TargetType == DistributionPriceTargetCustomer && cfg.CustomerUserId == customerUserID && cfg.Status == DistributionStatusEnabled {
			return applyDistributionPriceConfig(packageDefaultPrice, cfg.PriceType, cfg.PriceValue)
		}
	}
	return packageDefaultPrice
}

func ValidateDistributionPromoDiscount(discountType string, value float64) error {
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
