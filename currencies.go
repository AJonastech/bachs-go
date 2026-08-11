package bachs

import "context"

// SupportedCurrencies lists the currencies the API supports, split by kind.
type SupportedCurrencies struct {
	// Fiat holds supported fiat currency codes, e.g. "USD", "NGN".
	Fiat []string `json:"fiat"`
	// Crypto holds supported crypto currency codes, e.g. "USDT_TRC20".
	Crypto []string `json:"crypto"`
}

// CurrencyService accesses the /v1/currencies endpoints. Reach it through
// Client.Currencies.
type CurrencyService struct {
	core *client
}

// Supported lists all currencies available for charging customers.
func (s *CurrencyService) Supported(ctx context.Context, opts ...RequestOption) (*SupportedCurrencies, error) {
	var out SupportedCurrencies
	if err := s.core.do(ctx, "GET", "/v1/currencies/supported", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutSupported lists the currencies available for paying out funds, which
// may differ from the set available for charging.
func (s *CurrencyService) PayoutSupported(ctx context.Context, opts ...RequestOption) (*SupportedCurrencies, error) {
	var out SupportedCurrencies
	if err := s.core.do(ctx, "GET", "/v1/currencies/payout-supported", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
