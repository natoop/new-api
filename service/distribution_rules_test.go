package service

import (
	"strings"
	"testing"
)

func TestNormalizeIdempotencyKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid trimmed", input: "  abc12345  ", want: "abc12345"},
		{name: "too short", input: "abc1234", wantErr: true},
		{name: "too long", input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", wantErr: true},
		{name: "invalid chars", input: "abc12345@", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeIdempotencyKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil with value %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalcCommission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		amount  int
		bps     int
		want    int
		wantErr bool
	}{
		{name: "simple", amount: 10000, bps: 2500, want: 2500},
		{name: "round down", amount: 999, bps: 333, want: 33},
		{name: "zero rate", amount: 10000, bps: 0, want: 0},
		{name: "max rate", amount: 10000, bps: 10000, want: 10000},
		{name: "negative amount", amount: -1, bps: 100, wantErr: true},
		{name: "negative bps", amount: 100, bps: -1, wantErr: true},
		{name: "over max bps", amount: 100, bps: 10001, wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := CalcCommission(tt.amount, tt.bps)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil with value %d", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateDistributionStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "enabled", status: "enabled"},
		{name: "disabled", status: " disabled "},
		{name: "empty", status: "", wantErr: true},
		{name: "unknown", status: "archived", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateDistributionStatus(tt.status)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDistributionStableBusinessNumbers(t *testing.T) {
	t.Parallel()

	orderNo := BuildPurchaseOrderNo(12, 34, "idem-key-1")
	if orderNo != BuildPurchaseOrderNo(12, 34, "idem-key-1") {
		t.Fatalf("purchase order number should be stable")
	}
	if orderNo == BuildPurchaseOrderNo(12, 35, "idem-key-1") {
		t.Fatalf("purchase order number should include package id")
	}
	if !strings.HasPrefix(orderNo, "distribution_order_idem_") {
		t.Fatalf("unexpected purchase order prefix: %s", orderNo)
	}

	refNo := BuildBalanceRef(12, "idem-key-1")
	if refNo != BuildBalanceRef(12, "idem-key-1") {
		t.Fatalf("balance reference should be stable")
	}
	if !strings.HasPrefix(refNo, "distribution_balance_") {
		t.Fatalf("unexpected balance reference prefix: %s", refNo)
	}

	profitNo := BuildProfitNo(55, 7)
	if profitNo != BuildProfitNo(55, 7) {
		t.Fatalf("profit number should be stable")
	}
	if !strings.HasPrefix(profitNo, "distribution_profit_") {
		t.Fatalf("unexpected profit prefix: %s", profitNo)
	}

	invitationNo := BuildInvitationNo(7, " Invitee@Example.COM ", "idem-key-1")
	if invitationNo != BuildInvitationNo(7, "invitee@example.com", "idem-key-1") {
		t.Fatalf("invitation number should normalize email")
	}
	if !strings.HasPrefix(invitationNo, "distribution_invitation_") {
		t.Fatalf("unexpected invitation prefix: %s", invitationNo)
	}
}

func TestCanApplyDelta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		balance int
		delta   int
		want    bool
	}{
		{balance: 100, delta: -100, want: true},
		{balance: 100, delta: -101, want: false},
		{balance: 0, delta: 1, want: true},
		{balance: 0, delta: 0, want: true},
	}

	for _, tt := range tests {
		if got := CanApplyDelta(tt.balance, tt.delta); got != tt.want {
			t.Fatalf("CanApplyDelta(%d, %d) = %v, want %v", tt.balance, tt.delta, got, tt.want)
		}
	}
}

func TestCanTransitionOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from string
		to   string
		want bool
	}{
		{from: "pending", to: "paid", want: true},
		{from: "pending", to: "cancelled", want: true},
		{from: "paid", to: "fulfilled", want: true},
		{from: "paid", to: "cancelled", want: true},
		{from: "fulfilled", to: "fulfilled", want: true},
		{from: "fulfilled", to: "cancelled", want: false},
		{from: "cancelled", to: "paid", want: false},
	}

	for _, tt := range tests {
		if got := CanTransitionOrder(tt.from, tt.to); got != tt.want {
			t.Fatalf("CanTransitionOrder(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestCanAssignInventory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status     string
		assignedTo int
		want       bool
	}{
		{status: "available", assignedTo: 0, want: true},
		{status: "reserved", assignedTo: 0, want: true},
		{status: "assigned", assignedTo: 0, want: false},
		{status: "available", assignedTo: 10, want: false},
	}

	for _, tt := range tests {
		if got := CanAssignInventory(tt.status, tt.assignedTo); got != tt.want {
			t.Fatalf("CanAssignInventory(%q, %d) = %v, want %v", tt.status, tt.assignedTo, got, tt.want)
		}
	}
}

func TestResolveDistributionAgentPrice(t *testing.T) {
	t.Parallel()

	configs := []DistributionPriceConfigRule{
		{ScopeType: DistributionPriceScopeGlobal, UnitPrice: 300, Status: DistributionStatusEnabled},
		{ScopeType: DistributionPriceScopeLevel, Level: 2, UnitPrice: 200, Status: DistributionStatusEnabled},
		{ScopeType: DistributionPriceScopeAgent, AgentId: 9, UnitPrice: 100, Status: DistributionStatusEnabled},
	}

	if got := ResolveDistributionAgentPrice(500, 9, 2, configs); got != 100 {
		t.Fatalf("agent scope should win, got %d", got)
	}
	if got := ResolveDistributionAgentPrice(500, 8, 2, configs); got != 200 {
		t.Fatalf("level scope should win, got %d", got)
	}
	if got := ResolveDistributionAgentPrice(500, 8, 5, configs); got != 300 {
		t.Fatalf("global scope should win, got %d", got)
	}
	if got := ResolveDistributionAgentPrice(500, 8, 5, nil); got != 500 {
		t.Fatalf("fallback should use package default, got %d", got)
	}
}

func TestValidateDistributionPromoDiscount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		discountType string
		value        int
		wantErr      bool
	}{
		{discountType: "percent", value: 10000},
		{discountType: "percent", value: 10001, wantErr: true},
		{discountType: "amount", value: 0},
		{discountType: "amount", value: -1, wantErr: true},
		{discountType: "fixed", value: 1, wantErr: true},
	}

	for _, tt := range tests {
		err := ValidateDistributionPromoDiscount(tt.discountType, tt.value)
		if tt.wantErr && err == nil {
			t.Fatalf("expected error for %s %d", tt.discountType, tt.value)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("unexpected error for %s %d: %v", tt.discountType, tt.value, err)
		}
	}
}

func TestValidateDistributionTimeWindow(t *testing.T) {
	t.Parallel()

	if err := ValidateDistributionTimeWindow(10, 20); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateDistributionTimeWindow(20, 10); err == nil {
		t.Fatalf("expected invalid time window")
	}
}
