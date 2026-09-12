package cex

import (
	"math/rand"

	"github.com/jimtang2/nux/lib/simulator"
)

func init() {
	// Already registered in your example; keep or move all here for clarity.
	simulator.AddAction(ViewProductAction{})
	simulator.AddAction(DepositAction{})
	simulator.AddAction(PlaceSpotOrderAction{})
	simulator.AddAction(CancelSpotOrderAction{})
	simulator.AddAction(WithdrawAction{})
	simulator.AddAction(ViewBalancesAction{})
}

// ---------- ViewProduct ----------

type ViewProductAction struct{}

func (ac ViewProductAction) Name() string { return "cex_view_product" }

func (ac ViewProductAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return pl.StateBool("loggedIn")
}

func (ac ViewProductAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	// Choose a symbol to "view"
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT"}
	sym := symbols[rand.Intn(len(symbols))]
	pl.SetState("cex_last_viewed_product", sym)
	return nil
}

func (ac ViewProductAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.product":        pl.StateString("cex_last_viewed_product"),
		"cex.action_subtype": "view_product",
	}
}

// ---------- Deposit ----------

type DepositAction struct{}

func (ac DepositAction) Name() string { return "cex_deposit" }

func (ac DepositAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	// Must be logged in and onboarded
	return pl.StateBool("onboarded") && pl.StateBool("loggedIn")
}

func (ac DepositAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	// Simulate depositing some USDT
	amount := 100 + rand.Intn(900) // 100–999 USDT
	cur := pl.StateInt("cex_balance_USDT")
	pl.SetState("cex_balance_USDT", cur+amount)

	// Optionally record last deposit
	pl.SetState("cex_last_deposit_asset", "USDT")
	pl.SetState("cex_last_deposit_amount", amount)
	return nil
}

func (ac DepositAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.asset":          pl.StateString("cex_last_deposit_asset"),
		"cex.amount":         pl.StateInt("cex_last_deposit_amount"),
		"cex.action_subtype": "deposit",
	}
}

// ---------- PlaceSpotOrder ----------

type PlaceSpotOrderAction struct{}

func (ac PlaceSpotOrderAction) Name() string { return "cex_place_spot_order" }

func (ac PlaceSpotOrderAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	if !pl.StateBool("loggedIn") {
		return false
	}
	// Need some USDT balance to place an order
	bal := pl.StateInt("cex_balance_USDT")
	return bal > 0
}

func (ac PlaceSpotOrderAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	// Simple model: buy BTCUSDT with some USDT
	symbol := "BTCUSDT"
	usdtSpend := 10 + rand.Intn(90) // 10–99 USDT
	bal := pl.StateInt("cex_balance_USDT")
	if usdtSpend > bal {
		usdtSpend = bal
	}

	// Deduct USDT
	pl.SetState("cex_balance_USDT", bal-usdtSpend)

	// Credit a synthetic BTC amount (just for state consistency; not price-accurate)
	btcAmt := usdtSpend / 1000 // rough placeholder
	curBTC := pl.StateInt("cex_balance_BTC")
	pl.SetState("cex_balance_BTC", curBTC+btcAmt)

	// Track an "open order" in a very simple way: count of open orders
	open := pl.StateInt("cex_open_spot_orders")
	pl.SetState("cex_open_spot_orders", open+1)

	// Store last order info
	pl.SetState("cex_last_order_symbol", symbol)
	pl.SetState("cex_last_order_side", "BUY")
	pl.SetState("cex_last_order_amount_usdt", usdtSpend)

	return nil
}

func (ac PlaceSpotOrderAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.symbol":         pl.StateString("cex_last_order_symbol"),
		"cex.side":           pl.StateString("cex_last_order_side"),
		"cex.amount_usdt":    pl.StateInt("cex_last_order_amount_usdt"),
		"cex.action_subtype": "place_spot_order",
	}
}

// ---------- CancelSpotOrder ----------

type CancelSpotOrderAction struct{}

func (ac CancelSpotOrderAction) Name() string { return "cex_cancel_spot_order" }

func (ac CancelSpotOrderAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	if !pl.StateBool("loggedIn") {
		return false
	}
	return pl.StateInt("cex_open_spot_orders") > 0
}

func (ac CancelSpotOrderAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	open := pl.StateInt("cex_open_spot_orders")
	if open <= 0 {
		return nil
	}
	pl.SetState("cex_open_spot_orders", open-1)

	// For simplicity, assume cancellation returns the USDT used
	// (In reality you'd track per-order amounts; this is a coarse model.)
	refund := 10 + rand.Intn(40)
	bal := pl.StateInt("cex_balance_USDT")
	pl.SetState("cex_balance_USDT", bal+refund)

	pl.SetState("cex_last_cancel_refund_usdt", refund)
	return nil
}

func (ac CancelSpotOrderAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.refund_usdt":    pl.StateInt("cex_last_cancel_refund_usdt"),
		"cex.action_subtype": "cancel_spot_order",
	}
}

// ---------- Withdraw ----------

type WithdrawAction struct{}

func (ac WithdrawAction) Name() string { return "cex_withdraw" }

func (ac WithdrawAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	if !pl.StateBool("loggedIn") {
		return false
	}
	// Must have some USDT balance to withdraw
	return pl.StateInt("cex_balance_USDT") > 50
}

func (ac WithdrawAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	bal := pl.StateInt("cex_balance_USDT")
	amount := 20 + rand.Intn(bal-20) // withdraw 20..bal-1
	if amount < 20 {
		amount = bal
	}
	pl.SetState("cex_balance_USDT", bal-amount)
	pl.SetState("cex_last_withdraw_asset", "USDT")
	pl.SetState("cex_last_withdraw_amount", amount)
	return nil
}

func (ac WithdrawAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.asset":          "USDT",
		"cex.amount":         pl.StateInt("cex_last_withdraw_amount"),
		"cex.action_subtype": "withdraw",
	}
}

// ---------- ViewBalances ----------

type ViewBalancesAction struct{}

func (ac ViewBalancesAction) Name() string { return "cex_view_balances" }

func (ac ViewBalancesAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return pl.StateBool("loggedIn")
}

func (ac ViewBalancesAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	// No state change required; just a "view" action.
	// Optionally, you could update a "last_viewed_balances_at" turn.
	return nil
}

func (ac ViewBalancesAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return map[string]any{
		"user.id":            pl.State("user.id"),
		"cex.balance_USDT":   pl.StateInt("cex_balance_USDT"),
		"cex.balance_BTC":    pl.StateInt("cex_balance_BTC"),
		"cex.action_subtype": "view_balances",
	}
}
