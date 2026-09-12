package cex

import (
	"math/rand"

	"github.com/jimtang2/nux/lib/simulator"
)

func init() {
	mustAdd(
		ViewPortfolioAction{},
		ViewSpotPageAction{},
		ViewFuturesPageAction{},
		ViewEarnPageAction{},
		ConvertAction{},
		PlaceFuturesOrderAction{},
		AdjustLeverageAction{},
		CloseFuturesPositionAction{},
		CancelAllOrdersAction{},
		EarnSubscribeAction{},
		EarnRedeemAction{},
		ClaimRewardsAction{},
		LaunchpadParticipateAction{},
		DepositCryptoAction{},
		WithdrawCryptoAction{},
		DepositFiatAction{},
		WithdrawFiatAction{},
		P2PBuyAction{},
		P2PSellAction{},
		APIRequestAction{},
		Enable2FAAction{},
		Disable2FAAction{},
		Change2FAMethodAction{},
		AddWithdrawalAddressAction{},
		RemoveWithdrawalAddressAction{},
		ChangePasswordAction{},
		ResetPasswordAction{},
		KYCSubmitAction{},
		KYCApproveAction{},
		KYCRejectAction{},
		LoginAttemptAction{},
		NewDeviceLoginAction{},
		CountryChangeAction{},
		RiskFlagAction{},
		SupportTicketCreateAction{},
		SupportTicketResolveAction{},
		ReferralInviteAction{},
		PromoCodeAction{},
		EmailOpenAction{},
		EmailClickAction{},
		PushClickAction{},
	)
}

func mustAdd(actions ...simulator.Action) {
	for _, action := range actions {
		if err := simulator.AddAction(action); err != nil {
			panic(err)
		}
	}
}

func loggedIn(pl *simulator.Player) bool {
	return pl.StateBool("loggedIn")
}

func onboarded(pl *simulator.Player) bool {
	return pl.StateBool("onboarded")
}

func verified(pl *simulator.Player) bool {
	return pl.StateString("cex_kyc_status") == "approved"
}

func usdtBalance(pl *simulator.Player) int {
	return pl.StateInt("cex_balance_usdt")
}

func setUSDTBalance(pl *simulator.Player, amount int) {
	if amount < 0 {
		amount = 0
	}
	pl.SetState("cex_balance_usdt", amount)
}

func randomChoice(values []string) string {
	return values[rand.Intn(len(values))]
}

func randomAmount(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

func eventAttributes(pl *simulator.Player, values map[string]any) map[string]any {
	attrs := map[string]any{
		"cex_user_id": pl.State("user.id"),
	}
	for key, value := range values {
		attrs[key] = value
	}
	return attrs
}

func addCounter(pl *simulator.Player, key string, delta int) int {
	value := pl.StateInt(key) + delta
	if value < 0 {
		value = 0
	}
	pl.SetState(key, value)
	return value
}

func recordProductView(pl *simulator.Player, page string) string {
	symbol := randomChoice([]string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "XRPUSDT"})
	pl.SetState("cex_last_viewed_product", symbol)
	pl.SetState("cex_last_viewed_page", page)
	addCounter(pl, "cex_product_view_count", 1)
	return symbol
}

// ViewPortfolioAction represents a user reviewing holdings and account allocation.
type ViewPortfolioAction struct{}

func (ViewPortfolioAction) Name() string { return "cex_view_portfolio" }
func (ViewPortfolioAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (ViewPortfolioAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	tab := randomChoice([]string{"overview", "spot", "futures", "earn"})
	pl.SetState("cex_last_portfolio_tab", tab)
	addCounter(pl, "cex_portfolio_view_count", 1)
	return nil
}
func (ViewPortfolioAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_portfolio_tab": pl.StateString("cex_last_portfolio_tab"),
		"cex_balance_usdt":  usdtBalance(pl),
	})
}

type ViewSpotPageAction struct{}

func (ViewSpotPageAction) Name() string { return "cex_view_spot_page" }
func (ViewSpotPageAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (ViewSpotPageAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	recordProductView(pl, "spot")
	return nil
}
func (ViewSpotPageAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_page":   "spot",
		"cex_symbol": pl.StateString("cex_last_viewed_product"),
	})
}

type ViewFuturesPageAction struct{}

func (ViewFuturesPageAction) Name() string { return "cex_view_futures_page" }
func (ViewFuturesPageAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl)
}
func (ViewFuturesPageAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	recordProductView(pl, "futures")
	return nil
}
func (ViewFuturesPageAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_page":   "futures",
		"cex_symbol": pl.StateString("cex_last_viewed_product"),
	})
}

type ViewEarnPageAction struct{}

func (ViewEarnPageAction) Name() string { return "cex_view_earn_page" }
func (ViewEarnPageAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (ViewEarnPageAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_last_viewed_page", "earn")
	addCounter(pl, "cex_earn_page_view_count", 1)
	return nil
}
func (ViewEarnPageAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_page": "earn"})
}

// ConvertAction models a simple asset-conversion workflow rather than an order-book trade.
type ConvertAction struct{}

func (ConvertAction) Name() string { return "cex_convert" }
func (ConvertAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && usdtBalance(pl) >= 10
}
func (ConvertAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(10, minInt(balance, 250))
	asset := randomChoice([]string{"BTC", "ETH", "BNB", "SOL"})
	setUSDTBalance(pl, balance-amount)
	addCounter(pl, "cex_balance_"+asset, amount)
	pl.SetState("cex_last_convert_from", "USDT")
	pl.SetState("cex_last_convert_to", asset)
	pl.SetState("cex_last_convert_amount", amount)
	return nil
}
func (ConvertAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_from_asset":  pl.StateString("cex_last_convert_from"),
		"cex_to_asset":    pl.StateString("cex_last_convert_to"),
		"cex_amount_usdt": pl.StateInt("cex_last_convert_amount"),
	})
}

// PlaceFuturesOrderAction opens a synthetic leveraged futures position.
type PlaceFuturesOrderAction struct{}

func (PlaceFuturesOrderAction) Name() string { return "cex_place_futures_order" }
func (PlaceFuturesOrderAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl) && usdtBalance(pl) >= 25
}
func (PlaceFuturesOrderAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	margin := randomAmount(25, minInt(balance, 500))
	leverage := randomAmount(1, 20)
	symbol := randomChoice([]string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT"})
	side := randomChoice([]string{"LONG", "SHORT"})

	setUSDTBalance(pl, balance-margin)
	addCounter(pl, "cex_open_futures_positions", 1)
	pl.SetState("cex_last_futures_symbol", symbol)
	pl.SetState("cex_last_futures_side", side)
	pl.SetState("cex_last_futures_margin", margin)
	pl.SetState("cex_last_futures_leverage", leverage)
	return nil
}
func (PlaceFuturesOrderAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_symbol":   pl.StateString("cex_last_futures_symbol"),
		"cex_side":     pl.StateString("cex_last_futures_side"),
		"cex_margin":   pl.StateInt("cex_last_futures_margin"),
		"cex_leverage": pl.StateInt("cex_last_futures_leverage"),
	})
}

type AdjustLeverageAction struct{}

func (AdjustLeverageAction) Name() string { return "cex_adjust_leverage" }
func (AdjustLeverageAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_open_futures_positions") > 0
}
func (AdjustLeverageAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	oldLeverage := pl.StateInt("cex_last_futures_leverage")
	if oldLeverage < 1 {
		oldLeverage = 1
	}
	newLeverage := randomAmount(1, 20)
	pl.SetState("cex_last_futures_old_leverage", oldLeverage)
	pl.SetState("cex_last_futures_leverage", newLeverage)
	return nil
}
func (AdjustLeverageAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_symbol":       pl.StateString("cex_last_futures_symbol"),
		"cex_old_leverage": pl.StateInt("cex_last_futures_old_leverage"),
		"cex_new_leverage": pl.StateInt("cex_last_futures_leverage"),
	})
}

type CloseFuturesPositionAction struct{}

func (CloseFuturesPositionAction) Name() string { return "cex_close_futures_position" }
func (CloseFuturesPositionAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_open_futures_positions") > 0
}
func (CloseFuturesPositionAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	addCounter(pl, "cex_open_futures_positions", -1)
	margin := pl.StateInt("cex_last_futures_margin")
	pnl := randomAmount(-minInt(margin, 100), maxInt(10, margin/2))
	setUSDTBalance(pl, usdtBalance(pl)+margin+pnl)
	pl.SetState("cex_last_futures_pnl", pnl)
	return nil
}
func (CloseFuturesPositionAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_symbol": pl.StateString("cex_last_futures_symbol"),
		"cex_pnl":    pl.StateInt("cex_last_futures_pnl"),
	})
}

type CancelAllOrdersAction struct{}

func (CancelAllOrdersAction) Name() string { return "cex_cancel_all_orders" }
func (CancelAllOrdersAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_open_spot_orders") > 0
}
func (CancelAllOrdersAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	cancelled := pl.StateInt("cex_open_spot_orders")
	pl.SetState("cex_open_spot_orders", 0)
	pl.SetState("cex_last_cancelled_order_count", cancelled)
	return nil
}
func (CancelAllOrdersAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_cancelled_order_count": pl.StateInt("cex_last_cancelled_order_count"),
	})
}

// EarnSubscribeAction allocates USDT to an earn position.
type EarnSubscribeAction struct{}

func (EarnSubscribeAction) Name() string { return "cex_earn_subscribe" }
func (EarnSubscribeAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && usdtBalance(pl) >= 50
}
func (EarnSubscribeAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(50, minInt(balance, 1000))
	product := randomChoice([]string{"flexible_usdt", "locked_usdt", "eth_staking", "bnb_vault"})
	setUSDTBalance(pl, balance-amount)
	addCounter(pl, "cex_earn_balance_usdt", amount)
	pl.SetState("cex_last_earn_product", product)
	pl.SetState("cex_last_earn_amount", amount)
	return nil
}
func (EarnSubscribeAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_earn_product": pl.StateString("cex_last_earn_product"),
		"cex_amount_usdt":  pl.StateInt("cex_last_earn_amount"),
	})
}

type EarnRedeemAction struct{}

func (EarnRedeemAction) Name() string { return "cex_earn_redeem" }
func (EarnRedeemAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_earn_balance_usdt") > 0
}
func (EarnRedeemAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	earnBalance := pl.StateInt("cex_earn_balance_usdt")
	amount := randomAmount(1, earnBalance)
	pl.SetState("cex_earn_balance_usdt", earnBalance-amount)
	setUSDTBalance(pl, usdtBalance(pl)+amount)
	pl.SetState("cex_last_earn_redeem_amount", amount)
	return nil
}
func (EarnRedeemAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_earn_product":      pl.StateString("cex_last_earn_product"),
		"cex_amount_usdt":       pl.StateInt("cex_last_earn_redeem_amount"),
		"cex_earn_balance_usdt": pl.StateInt("cex_earn_balance_usdt"),
	})
}

type ClaimRewardsAction struct{}

func (ClaimRewardsAction) Name() string { return "cex_staking_claim_rewards" }
func (ClaimRewardsAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_pending_rewards_usdt") > 0
}
func (ClaimRewardsAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	rewards := pl.StateInt("cex_pending_rewards_usdt")
	pl.SetState("cex_pending_rewards_usdt", 0)
	setUSDTBalance(pl, usdtBalance(pl)+rewards)
	pl.SetState("cex_last_claimed_rewards_usdt", rewards)
	return nil
}
func (ClaimRewardsAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_rewards_usdt": pl.StateInt("cex_last_claimed_rewards_usdt")})
}

type LaunchpadParticipateAction struct{}

func (LaunchpadParticipateAction) Name() string { return "cex_launchpad_participate" }
func (LaunchpadParticipateAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl) && usdtBalance(pl) >= 100
}
func (LaunchpadParticipateAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(100, minInt(balance, 2000))
	project := randomChoice([]string{"project_orbit", "project_nova", "project_spark"})
	setUSDTBalance(pl, balance-amount)
	addCounter(pl, "cex_launchpad_commitment_usdt", amount)
	pl.SetState("cex_last_launchpad_project", project)
	pl.SetState("cex_last_launchpad_commitment", amount)
	return nil
}
func (LaunchpadParticipateAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_project":     pl.StateString("cex_last_launchpad_project"),
		"cex_amount_usdt": pl.StateInt("cex_last_launchpad_commitment"),
	})
}

// Funding actions.
type DepositCryptoAction struct{}

func (DepositCryptoAction) Name() string { return "cex_deposit_crypto" }
func (DepositCryptoAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (DepositCryptoAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	asset := randomChoice([]string{"USDT", "BTC", "ETH", "BNB"})
	amount := randomAmount(25, 2500)
	if asset == "USDT" {
		setUSDTBalance(pl, usdtBalance(pl)+amount)
	} else {
		addCounter(pl, "cex_balance_"+asset, amount)
	}
	pl.SetState("cex_last_deposit_asset", asset)
	pl.SetState("cex_last_deposit_amount", amount)
	pl.SetState("cex_last_deposit_network", randomChoice([]string{"ERC20", "BEP20", "TRC20", "BTC"}))
	return nil
}
func (DepositCryptoAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_asset":   pl.StateString("cex_last_deposit_asset"),
		"cex_amount":  pl.StateInt("cex_last_deposit_amount"),
		"cex_network": pl.StateString("cex_last_deposit_network"),
	})
}

type WithdrawCryptoAction struct{}

func (WithdrawCryptoAction) Name() string { return "cex_withdraw_crypto" }
func (WithdrawCryptoAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && usdtBalance(pl) >= 50
}
func (WithdrawCryptoAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(25, maxInt(25, balance/2))
	setUSDTBalance(pl, balance-amount)
	pl.SetState("cex_last_withdraw_asset", "USDT")
	pl.SetState("cex_last_withdraw_amount", amount)
	pl.SetState("cex_last_withdraw_network", randomChoice([]string{"ERC20", "BEP20", "TRC20"}))
	return nil
}
func (WithdrawCryptoAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_asset":   pl.StateString("cex_last_withdraw_asset"),
		"cex_amount":  pl.StateInt("cex_last_withdraw_amount"),
		"cex_network": pl.StateString("cex_last_withdraw_network"),
	})
}

type DepositFiatAction struct{}

func (DepositFiatAction) Name() string { return "cex_deposit_fiat" }
func (DepositFiatAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl)
}
func (DepositFiatAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	amount := randomAmount(50, 5000)
	setUSDTBalance(pl, usdtBalance(pl)+amount)
	pl.SetState("cex_last_fiat_currency", randomChoice([]string{"USD", "EUR", "GBP", "AUD"}))
	pl.SetState("cex_last_fiat_method", randomChoice([]string{"bank_transfer", "card", "p2p"}))
	pl.SetState("cex_last_fiat_amount", amount)
	return nil
}
func (DepositFiatAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_currency": pl.StateString("cex_last_fiat_currency"),
		"cex_method":   pl.StateString("cex_last_fiat_method"),
		"cex_amount":   pl.StateInt("cex_last_fiat_amount"),
	})
}

type WithdrawFiatAction struct{}

func (WithdrawFiatAction) Name() string { return "cex_withdraw_fiat" }
func (WithdrawFiatAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl) && usdtBalance(pl) >= 100
}
func (WithdrawFiatAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(50, maxInt(50, balance/2))
	setUSDTBalance(pl, balance-amount)
	pl.SetState("cex_last_fiat_currency", randomChoice([]string{"USD", "EUR", "GBP", "AUD"}))
	pl.SetState("cex_last_fiat_method", "bank_transfer")
	pl.SetState("cex_last_fiat_amount", amount)
	return nil
}
func (WithdrawFiatAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_currency": pl.StateString("cex_last_fiat_currency"),
		"cex_method":   pl.StateString("cex_last_fiat_method"),
		"cex_amount":   pl.StateInt("cex_last_fiat_amount"),
	})
}

type P2PBuyAction struct{}

func (P2PBuyAction) Name() string { return "cex_p2p_buy" }
func (P2PBuyAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl)
}
func (P2PBuyAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	amount := randomAmount(20, 1500)
	setUSDTBalance(pl, usdtBalance(pl)+amount)
	pl.SetState("cex_last_p2p_amount", amount)
	pl.SetState("cex_last_p2p_currency", randomChoice([]string{"USD", "EUR", "CNY", "NGN"}))
	pl.SetState("cex_last_p2p_role", "buyer")
	return nil
}
func (P2PBuyAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_role":          pl.StateString("cex_last_p2p_role"),
		"cex_fiat_currency": pl.StateString("cex_last_p2p_currency"),
		"cex_amount":        pl.StateInt("cex_last_p2p_amount"),
	})
}

type P2PSellAction struct{}

func (P2PSellAction) Name() string { return "cex_p2p_sell" }
func (P2PSellAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl) && usdtBalance(pl) >= 20
}
func (P2PSellAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	balance := usdtBalance(pl)
	amount := randomAmount(20, maxInt(20, balance/2))
	setUSDTBalance(pl, balance-amount)
	pl.SetState("cex_last_p2p_amount", amount)
	pl.SetState("cex_last_p2p_currency", randomChoice([]string{"USD", "EUR", "CNY", "NGN"}))
	pl.SetState("cex_last_p2p_role", "seller")
	return nil
}
func (P2PSellAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_role":          pl.StateString("cex_last_p2p_role"),
		"cex_fiat_currency": pl.StateString("cex_last_p2p_currency"),
		"cex_amount":        pl.StateInt("cex_last_p2p_amount"),
	})
}

// APIRequestAction represents a successful authenticated API request from an API-enabled user.
type APIRequestAction struct{}

func (APIRequestAction) Name() string { return "cex_api_request" }
func (APIRequestAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl) && pl.StateBool("cex_api_enabled")
}
func (APIRequestAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_last_api_endpoint", randomChoice([]string{
		"/api/v3/account", "/api/v3/order", "/api/v3/openOrders", "/fapi/v2/positionRisk",
	}))
	pl.SetState("cex_last_api_method", randomChoice([]string{"GET", "POST", "DELETE"}))
	pl.SetState("cex_last_api_status", randomChoiceInt([]int{200, 200, 200, 201, 400, 429}))
	addCounter(pl, "cex_api_request_count", 1)
	return nil
}
func (APIRequestAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_endpoint":    pl.StateString("cex_last_api_endpoint"),
		"cex_http_method": pl.StateString("cex_last_api_method"),
		"cex_status_code": pl.StateInt("cex_last_api_status"),
	})
}

// Security and account-management actions.
type Enable2FAAction struct{}

func (Enable2FAAction) Name() string { return "cex_enable_2fa" }
func (Enable2FAAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && !pl.StateBool("cex_2fa_enabled")
}
func (Enable2FAAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_2fa_enabled", true)
	pl.SetState("cex_2fa_method", randomChoice([]string{"totp", "sms", "hardware_key"}))
	return nil
}
func (Enable2FAAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_2fa_method": pl.StateString("cex_2fa_method")})
}

type Disable2FAAction struct{}

func (Disable2FAAction) Name() string { return "cex_disable_2fa" }
func (Disable2FAAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateBool("cex_2fa_enabled")
}
func (Disable2FAAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_2fa_enabled", false)
	return nil
}
func (Disable2FAAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_2fa_method": pl.StateString("cex_2fa_method")})
}

type Change2FAMethodAction struct{}

func (Change2FAMethodAction) Name() string { return "cex_change_2fa_method" }
func (Change2FAMethodAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateBool("cex_2fa_enabled")
}
func (Change2FAMethodAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	oldMethod := pl.StateString("cex_2fa_method")
	newMethod := randomChoice([]string{"totp", "sms", "hardware_key"})
	pl.SetState("cex_last_2fa_method", oldMethod)
	pl.SetState("cex_2fa_method", newMethod)
	return nil
}
func (Change2FAMethodAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_old_2fa_method": pl.StateString("cex_last_2fa_method"),
		"cex_new_2fa_method": pl.StateString("cex_2fa_method"),
	})
}

type AddWithdrawalAddressAction struct{}

func (AddWithdrawalAddressAction) Name() string { return "cex_add_withdrawal_address" }
func (AddWithdrawalAddressAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl)
}
func (AddWithdrawalAddressAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	count := addCounter(pl, "cex_withdrawal_address_count", 1)
	pl.SetState("cex_last_withdrawal_address_asset", randomChoice([]string{"BTC", "ETH", "USDT", "BNB"}))
	pl.SetState("cex_last_withdrawal_address_count", count)
	return nil
}
func (AddWithdrawalAddressAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_asset":                    pl.StateString("cex_last_withdrawal_address_asset"),
		"cex_withdrawal_address_count": pl.StateInt("cex_last_withdrawal_address_count"),
	})
}

type RemoveWithdrawalAddressAction struct{}

func (RemoveWithdrawalAddressAction) Name() string { return "cex_remove_withdrawal_address" }
func (RemoveWithdrawalAddressAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && pl.StateInt("cex_withdrawal_address_count") > 0
}
func (RemoveWithdrawalAddressAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	count := addCounter(pl, "cex_withdrawal_address_count", -1)
	pl.SetState("cex_last_withdrawal_address_count", count)
	return nil
}
func (RemoveWithdrawalAddressAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_withdrawal_address_count": pl.StateInt("cex_last_withdrawal_address_count")})
}

type ChangePasswordAction struct{}

func (ChangePasswordAction) Name() string { return "cex_change_password" }
func (ChangePasswordAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (ChangePasswordAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	addCounter(pl, "cex_password_change_count", 1)
	pl.SetState("cex_last_security_event", "password_changed")
	return nil
}
func (ChangePasswordAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_security_event": pl.StateString("cex_last_security_event")})
}

type ResetPasswordAction struct{}

func (ResetPasswordAction) Name() string { return "cex_reset_password" }
func (ResetPasswordAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl)
}
func (ResetPasswordAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	addCounter(pl, "cex_password_reset_count", 1)
	pl.SetState("cex_last_security_event", "password_reset")
	return nil
}
func (ResetPasswordAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_security_event": pl.StateString("cex_last_security_event")})
}

// KYC actions model a simplified compliance lifecycle.
type KYCSubmitAction struct{}

func (KYCSubmitAction) Name() string { return "cex_kyc_submit" }
func (KYCSubmitAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	status := pl.StateString("cex_kyc_status")
	return loggedIn(pl) && (status == "" || status == "not_started" || status == "rejected")
}
func (KYCSubmitAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_kyc_status", "pending")
	pl.SetState("cex_last_kyc_document", randomChoice([]string{"passport", "national_id", "drivers_license"}))
	return nil
}
func (KYCSubmitAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_kyc_status":   pl.StateString("cex_kyc_status"),
		"cex_kyc_document": pl.StateString("cex_last_kyc_document"),
	})
}

type KYCApproveAction struct{}

func (KYCApproveAction) Name() string { return "cex_kyc_approved" }
func (KYCApproveAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return pl.StateString("cex_kyc_status") == "pending"
}
func (KYCApproveAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_kyc_status", "approved")
	pl.SetState("cex_kyc_tier", randomChoice([]string{"basic", "intermediate", "advanced"}))
	return nil
}
func (KYCApproveAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_kyc_status": pl.StateString("cex_kyc_status"),
		"cex_kyc_tier":   pl.StateString("cex_kyc_tier"),
	})
}

type KYCRejectAction struct{}

func (KYCRejectAction) Name() string { return "cex_kyc_rejected" }
func (KYCRejectAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return pl.StateString("cex_kyc_status") == "pending"
}
func (KYCRejectAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_kyc_status", "rejected")
	pl.SetState("cex_last_kyc_rejection_reason", randomChoice([]string{"document_unreadable", "identity_mismatch", "unsupported_region"}))
	return nil
}
func (KYCRejectAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_kyc_status":           pl.StateString("cex_kyc_status"),
		"cex_kyc_rejection_reason": pl.StateString("cex_last_kyc_rejection_reason"),
	})
}

// LoginAttemptAction is intentionally allowed while logged out so failed-auth patterns can be generated.
type LoginAttemptAction struct{}

func (LoginAttemptAction) Name() string { return "cex_login_attempt" }
func (LoginAttemptAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl) && !loggedIn(pl)
}
func (LoginAttemptAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	success := rand.Intn(100) < 90
	pl.SetState("cex_last_login_attempt_success", success)
	pl.SetState("cex_last_login_country", randomChoice([]string{"US", "DE", "GB", "BR", "IN", "NG"}))
	if success {
		addCounter(pl, "cex_successful_login_attempt_count", 1)
	} else {
		addCounter(pl, "cex_failed_login_attempt_count", 1)
	}
	return nil
}
func (LoginAttemptAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_success":    pl.StateBool("cex_last_login_attempt_success"),
		"cex_ip_country": pl.StateString("cex_last_login_country"),
	})
}

type NewDeviceLoginAction struct{}

func (NewDeviceLoginAction) Name() string { return "cex_new_device_login" }
func (NewDeviceLoginAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (NewDeviceLoginAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	count := addCounter(pl, "cex_known_device_count", 1)
	pl.SetState("cex_last_device_type", randomChoice([]string{"web", "ios", "android", "desktop"}))
	pl.SetState("cex_last_known_device_count", count)
	return nil
}
func (NewDeviceLoginAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_device_type":        pl.StateString("cex_last_device_type"),
		"cex_known_device_count": pl.StateInt("cex_last_known_device_count"),
	})
}

type CountryChangeAction struct{}

func (CountryChangeAction) Name() string { return "cex_ip_country_change" }
func (CountryChangeAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (CountryChangeAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	oldCountry := pl.StateString("cex_home_country")
	if oldCountry == "" {
		oldCountry = "US"
	}
	newCountry := randomChoice([]string{"US", "DE", "GB", "BR", "IN", "NG"})
	pl.SetState("cex_last_ip_country", newCountry)
	pl.SetState("cex_last_previous_ip_country", oldCountry)
	return nil
}
func (CountryChangeAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_previous_ip_country": pl.StateString("cex_last_previous_ip_country"),
		"cex_ip_country":          pl.StateString("cex_last_ip_country"),
	})
}

type RiskFlagAction struct{}

func (RiskFlagAction) Name() string { return "cex_risk_flag_triggered" }
func (RiskFlagAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl)
}
func (RiskFlagAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	flag := randomChoice([]string{"unusual_withdrawal", "rapid_device_change", "failed_login_burst", "velocity_limit"})
	pl.SetState("cex_last_risk_flag", flag)
	addCounter(pl, "cex_risk_flag_count", 1)
	return nil
}
func (RiskFlagAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_risk_flag":       pl.StateString("cex_last_risk_flag"),
		"cex_risk_flag_count": pl.StateInt("cex_risk_flag_count"),
	})
}

// Support, marketing, and referral actions.
type SupportTicketCreateAction struct{}

func (SupportTicketCreateAction) Name() string { return "cex_support_ticket_created" }
func (SupportTicketCreateAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (SupportTicketCreateAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	category := randomChoice([]string{"deposit", "withdrawal", "trading", "security", "kyc"})
	addCounter(pl, "cex_open_support_ticket_count", 1)
	pl.SetState("cex_last_support_category", category)
	return nil
}
func (SupportTicketCreateAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_ticket_category":   pl.StateString("cex_last_support_category"),
		"cex_open_ticket_count": pl.StateInt("cex_open_support_ticket_count"),
	})
}

type SupportTicketResolveAction struct{}

func (SupportTicketResolveAction) Name() string { return "cex_support_ticket_resolved" }
func (SupportTicketResolveAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return pl.StateInt("cex_open_support_ticket_count") > 0
}
func (SupportTicketResolveAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	count := addCounter(pl, "cex_open_support_ticket_count", -1)
	pl.SetState("cex_last_support_resolution", randomChoice([]string{"resolved", "refunded", "escalated"}))
	pl.SetState("cex_last_open_ticket_count", count)
	return nil
}
func (SupportTicketResolveAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{
		"cex_ticket_category": pl.StateString("cex_last_support_category"),
		"cex_resolution":      pl.StateString("cex_last_support_resolution"),
	})
}

type ReferralInviteAction struct{}

func (ReferralInviteAction) Name() string { return "cex_referral_invite_sent" }
func (ReferralInviteAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl) && verified(pl)
}
func (ReferralInviteAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	count := addCounter(pl, "cex_referral_invite_count", 1)
	pl.SetState("cex_last_referral_invite_count", count)
	return nil
}
func (ReferralInviteAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_referral_invite_count": pl.StateInt("cex_last_referral_invite_count")})
}

type PromoCodeAction struct{}

func (PromoCodeAction) Name() string { return "cex_promo_code_applied" }
func (PromoCodeAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return loggedIn(pl)
}
func (PromoCodeAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	code := randomChoice([]string{"WELCOME10", "ZERO_FEE", "FUTURES20", "EARNBOOST"})
	pl.SetState("cex_last_promo_code", code)
	addCounter(pl, "cex_promo_code_count", 1)
	return nil
}
func (PromoCodeAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_promo_code": pl.StateString("cex_last_promo_code")})
}

type EmailOpenAction struct{}

func (EmailOpenAction) Name() string { return "cex_email_opened" }
func (EmailOpenAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl)
}
func (EmailOpenAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	campaign := randomChoice([]string{"market_update", "security_alert", "earn_campaign", "fee_promotion"})
	pl.SetState("cex_last_email_campaign", campaign)
	addCounter(pl, "cex_email_open_count", 1)
	return nil
}
func (EmailOpenAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_campaign": pl.StateString("cex_last_email_campaign")})
}

type EmailClickAction struct{}

func (EmailClickAction) Name() string { return "cex_email_clicked" }
func (EmailClickAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl) && pl.StateInt("cex_email_open_count") > 0
}
func (EmailClickAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	addCounter(pl, "cex_email_click_count", 1)
	return nil
}
func (EmailClickAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_campaign": pl.StateString("cex_last_email_campaign")})
}

type PushClickAction struct{}

func (PushClickAction) Name() string { return "cex_push_notification_clicked" }
func (PushClickAction) Condition(pl *simulator.Player, s *simulator.Simulator) bool {
	return onboarded(pl) && pl.StateBool("cex_push_enabled")
}
func (PushClickAction) Exec(pl *simulator.Player, s *simulator.Simulator) error {
	pl.SetState("cex_last_push_campaign", randomChoice([]string{"price_alert", "liquidation_alert", "new_listing", "earn_campaign"}))
	addCounter(pl, "cex_push_click_count", 1)
	return nil
}
func (PushClickAction) Attributes(pl *simulator.Player, s *simulator.Simulator) map[string]any {
	return eventAttributes(pl, map[string]any{"cex_campaign": pl.StateString("cex_last_push_campaign")})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func randomChoiceInt(values []int) int {
	return values[rand.Intn(len(values))]
}
