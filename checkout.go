package bachs

import (
	"context"
	"net/url"
	"time"
)

// Checkout session status values, reported by CheckoutSession.Status and
// Checkout.Status.
const (
	CheckoutStatusOpen      = "OPEN"
	CheckoutStatusCompleted = "COMPLETED"
	CheckoutStatusExpired   = "EXPIRED"
	CheckoutStatusCancelled = "CANCELLED"
)

// Checkout source types, reported by Checkout.SourceType.
const (
	CheckoutSourceAPI             = "API"
	CheckoutSourceCheckoutSession = "CHECKOUT_SESSION"
	CheckoutSourcePaymentLink     = "PAYMENT_LINK"
)

// Checkout session modes, reported by CheckoutSession.SessionMode.
const (
	// CheckoutModeCart is a session built from an explicit product cart.
	CheckoutModeCart = "CART"
	// CheckoutModeSelection is a session where the buyer selects what to pay.
	CheckoutModeSelection = "SELECTION"
)

// CheckoutSessionCreated is the compact result of creating a checkout session.
// Redirect the buyer to CheckoutURL to complete payment, then retrieve the full
// session with CheckoutService.GetSession.
type CheckoutSessionCreated struct {
	CheckoutID  string    `json:"checkout_id"`
	CheckoutURL string    `json:"checkout_url"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// CheckoutRecurring describes the recurring cadence of a checkout session, when
// the session sets up a subscription.
type CheckoutRecurring struct {
	// Interval is one of the Interval* constants.
	Interval string `json:"interval"`
	// IntervalCount is how many intervals between charges; defaults to 1.
	IntervalCount int `json:"interval_count"`
}

// CheckoutCustomer is the buyer summary embedded in a retrieved checkout
// session.
type CheckoutCustomer struct {
	ID    *string `json:"id"`
	Email string  `json:"email"`
	Name  *string `json:"name"`
}

// ResolvedProductItem is a fully priced line in a retrieved checkout session.
type ResolvedProductItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	UnitAmount  Money  `json:"unit_amount"`
	Currency    string `json:"currency"`
	// PriceType is one of the PriceType* constants.
	PriceType     string `json:"price_type"`
	MinimumAmount *Money `json:"minimum_amount"`
	MaximumAmount *Money `json:"maximum_amount"`
	LineTotal     Money  `json:"line_total"`
}

// CheckoutSession is the full state of a hosted checkout session, as returned
// by CheckoutService.GetSession.
type CheckoutSession struct {
	CheckoutID string `json:"checkout_id"`
	// Status is one of the CheckoutStatus* constants.
	Status string `json:"status"`
	// Recurring is set when the session establishes a subscription.
	Recurring *CheckoutRecurring `json:"recurring"`
	// PaymentStatus is the underlying payment's status, or nil before payment.
	PaymentStatus *string `json:"payment_status"`
	SourceType    *string `json:"source_type"`
	Amount        Money   `json:"amount"`
	Currency      string  `json:"currency"`
	Reference     *string `json:"reference"`
	// Charge is the resulting payment once the session completes, or nil.
	Charge          *Payment              `json:"charge"`
	PaymentMethod   *string               `json:"payment_method"`
	Customer        CheckoutCustomer      `json:"customer"`
	SuccessURL      *string               `json:"success_url"`
	CancelURL       *string               `json:"cancel_url"`
	Products        []ResolvedProductItem `json:"products"`
	BillingCurrency *string               `json:"billing_currency"`
	// SessionMode is CheckoutModeCart or CheckoutModeSelection.
	SessionMode *string    `json:"session_mode"`
	Metadata    Metadata   `json:"metadata"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Checkout is a completed or in-progress checkout record retrieved by ID,
// regardless of how it was initiated (API, checkout session, or payment link).
type Checkout struct {
	CheckoutID string `json:"checkout_id"`
	// Status is one of the CheckoutStatus* constants.
	Status string `json:"status"`
	// SourceType is one of the CheckoutSource* constants.
	SourceType string  `json:"source_type"`
	Amount     Money   `json:"amount"`
	Currency   string  `json:"currency"`
	Reference  *string `json:"reference"`
	// Charge is the resulting payment, or nil if none yet.
	Charge        *Payment   `json:"charge"`
	CustomerEmail string     `json:"customer_email"`
	CustomerName  *string    `json:"customer_name"`
	CustomerID    *string    `json:"customer_id"`
	SuccessURL    *string    `json:"success_url"`
	CancelURL     *string    `json:"cancel_url"`
	Metadata      Metadata   `json:"metadata"`
	CreatedAt     time.Time  `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// CheckoutCustomerParams identifies the buyer for a checkout session. Provide
// exactly one of CustomerID (an existing customer) or the new-customer fields
// (Email and Name required).
type CheckoutCustomerParams struct {
	// CustomerID references an existing customer. When set, leave the
	// new-customer fields empty.
	CustomerID string `json:"customer_id,omitempty"`
	// Email and Name create a new customer inline. Both required when CustomerID
	// is empty.
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	// PhoneNumber is optional for a new customer.
	PhoneNumber *string `json:"phone_number,omitempty"`
}

// AdhocPriceParams overrides pricing for a single cart line at checkout.
type AdhocPriceParams struct {
	// PriceType is one of the PriceType* constants.
	PriceType     string `json:"price_type,omitempty"`
	Amount        *Money `json:"amount,omitempty"`
	PresetAmount  *Money `json:"preset_amount,omitempty"`
	MinimumAmount *Money `json:"minimum_amount,omitempty"`
	MaximumAmount *Money `json:"maximum_amount,omitempty"`
}

// CheckoutProductItem is one line in a checkout session's product cart.
type CheckoutProductItem struct {
	// ProductID is the product to sell. Required.
	ProductID string `json:"product_id"`
	// Quantity defaults to 1 when omitted.
	Quantity int `json:"quantity,omitempty"`
	// Amount overrides the unit amount for a custom-priced product.
	Amount *Money `json:"amount,omitempty"`
	// Pricing overrides the line's pricing in full.
	Pricing *AdhocPriceParams `json:"pricing,omitempty"`
}

// CheckoutPricing is a merchant-defined amount for a SELECTION-mode session,
// where the buyer pays an amount you specify rather than a fixed product cart.
type CheckoutPricing struct {
	// Currency is the pricing currency. Required.
	Currency string `json:"currency"`
	Amount   *Money `json:"amount,omitempty"`
	// PriceType is one of the PriceType* constants.
	PriceType     string `json:"price_type,omitempty"`
	PresetAmount  *Money `json:"preset_amount,omitempty"`
	MinimumAmount *Money `json:"minimum_amount,omitempty"`
	MaximumAmount *Money `json:"maximum_amount,omitempty"`
	// CurrencyOptions holds amounts in additional currencies, keyed by code.
	CurrencyOptions map[string]any `json:"currency_options,omitempty"`
}

// CheckoutSessionCreateParams are the parameters for
// CheckoutService.CreateSession. Customer and SuccessURL are required. Supply
// either ProductCart (CART mode) or Pricing (SELECTION mode).
type CheckoutSessionCreateParams struct {
	// Customer identifies or creates the buyer. Required.
	Customer CheckoutCustomerParams `json:"customer"`
	// SuccessURL is where the buyer is sent after paying. Required.
	SuccessURL string `json:"success_url"`
	// CancelURL is where the buyer is sent if they cancel.
	CancelURL *string `json:"cancel_url,omitempty"`
	// ReturnURL is an alternative post-checkout redirect.
	ReturnURL *string `json:"return_url,omitempty"`
	// BillingCurrency overrides the currency the buyer is charged in.
	BillingCurrency *string `json:"billing_currency,omitempty"`
	// AllowedPaymentMethodTypes restricts the payment methods offered; use the
	// PaymentMethod* constants.
	AllowedPaymentMethodTypes []string `json:"allowed_payment_method_types,omitempty"`
	// ProductCart lists the products to sell (CART mode).
	ProductCart []CheckoutProductItem `json:"product_cart,omitempty"`
	// Pricing sets a merchant-defined amount (SELECTION mode).
	Pricing *CheckoutPricing `json:"pricing,omitempty"`
	// Reference is your own reference for the session.
	Reference *string `json:"reference,omitempty"`
	// Metadata is optional key/value data.
	Metadata Metadata `json:"metadata,omitempty"`
	// ExpiresInMinutes sets how long the session stays open.
	ExpiresInMinutes *int `json:"expires_in_minutes,omitempty"`
}

// PortalSession is a short-lived link to the customer self-service portal,
// created by CustomerService.CreatePortalSession.
type PortalSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// CheckoutService accesses the /v1/checkout-sessions and /v1/checkouts
// endpoints. Reach it through Client.Checkout.
type CheckoutService struct {
	core *client
}

// CreateSession creates a hosted checkout session and returns its ID and URL.
// Redirect the buyer to the returned CheckoutURL. Requires the checkout:write
// scope.
func (s *CheckoutService) CreateSession(ctx context.Context, params *CheckoutSessionCreateParams, opts ...RequestOption) (*CheckoutSessionCreated, error) {
	var out CheckoutSessionCreated
	if err := s.core.do(ctx, "POST", "/v1/checkout-sessions", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSession retrieves the full state of a checkout session by ID. Requires the
// checkout:read scope.
func (s *CheckoutService) GetSession(ctx context.Context, id string, opts ...RequestOption) (*CheckoutSession, error) {
	var out CheckoutSession
	if err := s.core.do(ctx, "GET", "/v1/checkout-sessions/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a checkout record by ID, regardless of how it was initiated.
// Requires the checkout:read scope.
func (s *CheckoutService) Get(ctx context.Context, id string, opts ...RequestOption) (*Checkout, error) {
	var out Checkout
	if err := s.core.do(ctx, "GET", "/v1/checkouts/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePortalSession creates a customer self-service portal session for the
// given customer and returns a short-lived URL to redirect them to. Requires
// the customers:write scope.
func (s *CustomerService) CreatePortalSession(ctx context.Context, customerID string, opts ...RequestOption) (*PortalSession, error) {
	var out PortalSession
	if err := s.core.do(ctx, "POST", "/v1/customers/"+url.PathEscape(customerID)+"/portal-sessions", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
