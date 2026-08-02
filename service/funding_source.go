package service

import (
	"github.com/QuantumNous/new-api/model"
)

// ---------------------------------------------------------------------------
// WalletFunding — 钱包资金来源
// ---------------------------------------------------------------------------

// WalletFunding 封装单次请求对用户钱包额度的预扣、结算与退款。
type WalletFunding struct {
	userId   int
	consumed int // 实际预扣的用户额度
}

// Source 返回资金来源标识，恒为 "wallet"。
func (w *WalletFunding) Source() string { return BillingSourceWallet }

// PreConsume 从用户钱包预扣 amount 额度。
func (w *WalletFunding) PreConsume(amount int) error {
	if amount <= 0 {
		return nil
	}
	if err := model.DecreaseUserQuota(w.userId, amount, false); err != nil {
		return err
	}
	w.consumed = amount
	return nil
}

// Settle 根据差额调整钱包额度（正数补扣，负数退还）。
func (w *WalletFunding) Settle(delta int) error {
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return model.DecreaseUserQuota(w.userId, delta, false)
	}
	return model.IncreaseUserQuota(w.userId, -delta, false)
}

// Refund 退还全部预扣额度。
func (w *WalletFunding) Refund() error {
	if w.consumed <= 0 {
		return nil
	}
	// IncreaseUserQuota 是 quota += N 的非幂等操作，不能重试，否则会多退额度。
	return model.IncreaseUserQuota(w.userId, w.consumed, false)
}
