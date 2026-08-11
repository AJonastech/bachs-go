package bachs

import (
	"context"
	"net/url"
)

// Payment method type values, used as the PaymentMethod query parameter when
// listing rails.
const (
	PaymentMethodCard         = "CARD"
	PaymentMethodCrypto       = "CRYPTO"
	PaymentMethodBankTransfer = "BANK_TRANSFER"
	PaymentMethodMobileMoney  = "MOBILE_MONEY"
)

// PaymentMethod is a payment method available to your customers at checkout.
type PaymentMethod struct {
	// ID is the stable identifier, e.g. "card".
	ID string `json:"id"`
	// DisplayName is the human-readable name shown at checkout.
	DisplayName string `json:"display_name"`
	// Icon is a URL or key for the method's icon.
	Icon string `json:"icon"`
	// Description explains the method to customers.
	Description string `json:"description"`
	// Type is the broad category, e.g. "CARD" or "MOBILE_MONEY".
	Type string `json:"type"`
	// EnabledByDefault reports whether the method is on unless configured off.
	EnabledByDefault bool `json:"enabled_by_default"`
	// Currencies lists the currency codes this method supports.
	Currencies []string `json:"currencies"`
}

// PaymentRail is one concrete rail (network or provider) that can process a
// payment method in a given currency, such as a specific bank or mobile-money
// operator.
type PaymentRail struct {
	// ID is the rail identifier.
	ID string `json:"id"`
	// Name is the display name, or nil if unnamed.
	Name *string `json:"name"`
	// Active reports whether the rail is currently usable, or nil if unknown.
	Active *bool `json:"active"`
}

// PaymentRails is the set of rails available for a payment method and currency.
type PaymentRails struct {
	PaymentMethod string        `json:"payment_method"`
	Currency      string        `json:"currency"`
	CountryCode   *string       `json:"country_code"`
	Rails         []PaymentRail `json:"rails"`
}

// PaymentRailsParams selects which rails to list. PaymentMethod and Currency
// are required; CountryCode optionally narrows the results.
type PaymentRailsParams struct {
	// PaymentMethod is one of the PaymentMethod* constants. Required.
	PaymentMethod string
	// Currency is the currency code, e.g. "NGN" or "USDT_TRC20". Required.
	Currency string
	// CountryCode optionally filters rails by ISO country code, e.g. "NG".
	CountryCode string
}

func (p *PaymentRailsParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.PaymentMethod != "" {
		v.Set("payment_method", p.PaymentMethod)
	}
	if p.Currency != "" {
		v.Set("currency", p.Currency)
	}
	if p.CountryCode != "" {
		v.Set("country_code", p.CountryCode)
	}
	return v
}

// PaymentMethodService accesses the /v1/payment-methods endpoints. Reach it
// through Client.PaymentMethods.
type PaymentMethodService struct {
	core *client
}

// List returns the payment methods available on your account.
func (s *PaymentMethodService) List(ctx context.Context, opts ...RequestOption) ([]PaymentMethod, error) {
	var out struct {
		PaymentMethods []PaymentMethod `json:"payment_methods"`
	}
	if err := s.core.do(ctx, "GET", "/v1/payment-methods", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.PaymentMethods, nil
}

// Rails lists the concrete rails available for a payment method and currency.
func (s *PaymentMethodService) Rails(ctx context.Context, params *PaymentRailsParams, opts ...RequestOption) (*PaymentRails, error) {
	var out PaymentRails
	if err := s.core.do(ctx, "GET", "/v1/payment-methods/rails", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
