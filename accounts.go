package bachs

import "context"

// AccountBalance is the available and pending balance in a single currency.
type AccountBalance struct {
	// Currency is the ISO currency code of this balance.
	Currency string `json:"currency"`
	// AvailableBalance is the amount available to pay out or spend now.
	AvailableBalance Money `json:"available_balance"`
	// PendingBalance is the amount not yet settled and available.
	PendingBalance Money `json:"pending_balance"`
}

// AccountBalances is your account's balances across every currency, as returned
// by AccountService.Balances.
type AccountBalances struct {
	// AccountID is the account these balances belong to.
	AccountID string `json:"account_id"`
	// Balances holds one entry per currency you hold funds in.
	Balances []AccountBalance `json:"balances"`
	// TotalBalanceUSD is the sum of all balances converted to USD.
	TotalBalanceUSD Money `json:"total_balance_usd"`
	// PendingSettlementsByDay lists upcoming settlements grouped by day; empty
	// when none are pending. Its entries are loosely typed by the API.
	PendingSettlementsByDay []map[string]any `json:"pending_settlements_by_day"`
}

// AccountService accesses the /v1/accounts endpoints. Reach it through
// Client.Accounts.
type AccountService struct {
	core *client
}

// Balances retrieves the current balances for your account across all
// currencies.
func (s *AccountService) Balances(ctx context.Context, opts ...RequestOption) (*AccountBalances, error) {
	var out AccountBalances
	if err := s.core.do(ctx, "GET", "/v1/accounts/balances", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
