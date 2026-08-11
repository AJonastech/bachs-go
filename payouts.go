package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Payout (withdrawal) status values, reported by Payout.Status.
const (
	PayoutStatusPendingSubmission = "pending_submission"
	PayoutStatusPendingCollection = "pending_collection"
	PayoutStatusProcessing        = "processing"
	PayoutStatusManualReview      = "manual_review"
	PayoutStatusSuccessful        = "successful"
	PayoutStatusFailed            = "failed"
	PayoutStatusCancelled         = "cancelled"
	PayoutStatusExpired           = "expired"
	PayoutStatusReconciled        = "reconciled"
)

// Payout destination types.
const (
	PayoutDestinationBankAccount  = "bank_account"
	PayoutDestinationMobileMoney  = "mobile_money"
	PayoutDestinationCryptoWallet = "crypto_wallet"
)

// PayoutQuote is a locked exchange rate for a cross-currency payout, valid until
// ExpiresAt. Pass its QuoteID to PayoutService.CreateWithdrawal.
type PayoutQuote struct {
	QuoteID      string    `json:"quote_id"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	FromAmount   Money     `json:"from_amount"`
	ToAmount     Money     `json:"to_amount"`
	ExchangeRate string    `json:"exchange_rate"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// PayoutDestination is a saved withdrawal target: a bank account, mobile money
// wallet, or crypto wallet.
type PayoutDestination struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	// Env is "test" or "live".
	Env string `json:"env"`
	// DestinationType is one of the PayoutDestination* constants.
	DestinationType string    `json:"destination_type"`
	Currency        string    `json:"currency"`
	Label           *string   `json:"label"`
	AccountNumber   *string   `json:"account_number"`
	AccountName     *string   `json:"account_name"`
	BankCode        *string   `json:"bank_code"`
	BankName        *string   `json:"bank_name"`
	PhoneNumber     *string   `json:"phone_number"`
	MobileProvider  *string   `json:"mobile_provider"`
	WalletAddress   *string   `json:"wallet_address"`
	Network         *string   `json:"network"`
	IsActive        bool      `json:"is_active"`
	Metadata        Metadata  `json:"metadata"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Payout is a withdrawal of funds to a destination, as returned by
// PayoutService.Get and List.
type Payout struct {
	WithdrawalID   string  `json:"withdrawal_id"`
	OrganizationID string  `json:"organization_id"`
	Reference      *string `json:"reference"`
	Amount         Money   `json:"amount"`
	Currency       string  `json:"currency"`
	// Status is one of the PayoutStatus* constants.
	Status             string     `json:"status"`
	FromCurrency       *string    `json:"from_currency"`
	ToCurrency         *string    `json:"to_currency"`
	FromAmount         *Money     `json:"from_amount"`
	ToAmount           *Money     `json:"to_amount"`
	ExchangeRate       *string    `json:"exchange_rate"`
	Destination        *string    `json:"destination"`
	QuoteID            *string    `json:"quote_id"`
	PayoutMethod       *string    `json:"payout_method"`
	ProviderReference  *string    `json:"provider_reference"`
	Metadata           Metadata   `json:"metadata"`
	RejectionReason    *string    `json:"rejection_reason"`
	ParentWithdrawalID *string    `json:"parent_withdrawal_id"`
	BatchIndex         *int       `json:"batch_index"`
	BatchTotal         *int       `json:"batch_total"`
	CreatedAt          time.Time  `json:"created_at"`
	ApprovedAt         *time.Time `json:"approved_at"`
	CompletedAt        *time.Time `json:"completed_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

// WithdrawalCreated is the compact result of creating a withdrawal.
type WithdrawalCreated struct {
	WithdrawalID      string  `json:"withdrawal_id"`
	Status            string  `json:"status"`
	ProviderReference *string `json:"provider_reference"`
}

// Bank is a bank in the payout bank list.
type Bank struct {
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Code          string  `json:"code"`
	NIBSSBankCode *string `json:"nibss_bank_code"`
	Country       string  `json:"country"`
}

// bankListResponse is the wire envelope for ListBanks.
type bankListResponse struct {
	Status  bool    `json:"status"`
	Message string  `json:"message"`
	Data    []Bank  `json:"data"`
	Error   *string `json:"error"`
}

// ResolvedBankAccount is the result of resolving a bank account number to its
// account holder.
type ResolvedBankAccount struct {
	Status  bool           `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
	Error   *string        `json:"error"`
}

// PayoutQuoteParams are the parameters for PayoutService.Quote. FromCurrency,
// ToCurrency, and Amount are required.
type PayoutQuoteParams struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       Money  `json:"amount"`
	// PayoutMethod optionally scopes the quote to a specific payout method.
	PayoutMethod *string `json:"payout_method,omitempty"`
}

// PayoutDestinationParams are the parameters for creating or updating a payout
// destination. DestinationType and Currency are always required; the remaining
// fields depend on the destination type (bank fields for bank_account, phone
// fields for mobile_money, wallet fields for crypto_wallet).
type PayoutDestinationParams struct {
	// DestinationType is one of the PayoutDestination* constants. Required.
	DestinationType string `json:"destination_type"`
	// Currency is the destination currency. Required.
	Currency       string   `json:"currency"`
	Label          string   `json:"label,omitempty"`
	AccountNumber  string   `json:"account_number,omitempty"`
	AccountName    string   `json:"account_name,omitempty"`
	BankCode       string   `json:"bank_code,omitempty"`
	BankName       string   `json:"bank_name,omitempty"`
	PhoneNumber    string   `json:"phone_number,omitempty"`
	MobileProvider string   `json:"mobile_provider,omitempty"`
	WalletAddress  string   `json:"wallet_address,omitempty"`
	Network        string   `json:"network,omitempty"`
	Metadata       Metadata `json:"metadata,omitempty"`
}

// WithdrawalCreateParams are the parameters for PayoutService.CreateWithdrawal.
// FromCurrency, ToCurrency, Amount, PaymentMethod, Reference, and Email are
// required. Supply either a saved PayoutDestinationID or the inline destination
// fields for the chosen payment method.
type WithdrawalCreateParams struct {
	FromCurrency  string `json:"from_currency"`
	ToCurrency    string `json:"to_currency"`
	Amount        Money  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
	Reference     string `json:"reference"`
	Email         string `json:"email"`
	// QuoteID executes against a quote from Quote.
	QuoteID *string `json:"quote_id,omitempty"`
	// IdempotencyKey is a body-level idempotency key specific to withdrawals.
	// This is separate from the Idempotency-Key header set by
	// WithIdempotencyKey.
	IdempotencyKey *string  `json:"idempotency_key,omitempty"`
	PaymentRail    *string  `json:"payment_rail,omitempty"`
	Metadata       Metadata `json:"metadata,omitempty"`
	// PayoutDestinationID references a saved destination instead of inline fields.
	PayoutDestinationID *string `json:"payout_destination_id,omitempty"`
	AccountNumber       *string `json:"account_number,omitempty"`
	BankCode            *string `json:"bank_code,omitempty"`
	PhoneNumber         *string `json:"phone_number,omitempty"`
	WalletAddress       *string `json:"wallet_address,omitempty"`
	Network             *string `json:"network,omitempty"`
	Memo                *string `json:"memo,omitempty"`
	Description         *string `json:"description,omitempty"`
}

// PayoutListParams filters and pages PayoutService.List. Payouts are paginated
// by offset.
type PayoutListParams struct {
	Limit  int
	Offset int
	// StatusFilter, when set, returns only payouts with that exact status. Note
	// the API's list filter accepts a workflow-level status such as "completed"
	// or "pending".
	StatusFilter string
}

func (p *PayoutListParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		v.Set("offset", strconv.Itoa(p.Offset))
	}
	if p.StatusFilter != "" {
		v.Set("status_filter", p.StatusFilter)
	}
	return v
}

// PayoutService accesses the /v1/payouts endpoints: withdrawals, saved
// destinations, quotes, and the supporting bank lookups. Reach it through
// Client.Payouts.
type PayoutService struct {
	core *client
}

// SupportedCurrencies returns the currencies available for the given payout
// method.
func (s *PayoutService) SupportedCurrencies(ctx context.Context, method string, opts ...RequestOption) ([]string, error) {
	var out struct {
		Method     string   `json:"method"`
		Currencies []string `json:"currencies"`
	}
	q := url.Values{"method": {method}}
	if err := s.core.do(ctx, "GET", "/v1/payouts/supported-currencies", q, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Currencies, nil
}

// ListBanks returns the banks available for payouts, optionally filtered by ISO
// country code (e.g. "NG"). Pass an empty countryCode for the default.
func (s *PayoutService) ListBanks(ctx context.Context, countryCode string, opts ...RequestOption) ([]Bank, error) {
	var q url.Values
	if countryCode != "" {
		q = url.Values{"country_code": {countryCode}}
	}
	var out bankListResponse
	if err := s.core.do(ctx, "GET", "/v1/payouts/banks", q, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// ResolveAccount looks up the account holder for a bank code and account number.
func (s *PayoutService) ResolveAccount(ctx context.Context, bankCode, accountNumber string, opts ...RequestOption) (*ResolvedBankAccount, error) {
	body := map[string]string{"bank_code": bankCode, "account_number": accountNumber}
	var out ResolvedBankAccount
	if err := s.core.do(ctx, "POST", "/v1/payouts/resolve-account", nil, body, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Quote locks an exchange rate for a cross-currency payout.
func (s *PayoutService) Quote(ctx context.Context, params *PayoutQuoteParams, opts ...RequestOption) (*PayoutQuote, error) {
	var out PayoutQuote
	if err := s.core.do(ctx, "POST", "/v1/payouts/quotes", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDestinations returns all saved payout destinations. This endpoint is not
// paginated.
func (s *PayoutService) ListDestinations(ctx context.Context, opts ...RequestOption) ([]PayoutDestination, error) {
	var out struct {
		Destinations []PayoutDestination `json:"destinations"`
		Total        int                 `json:"total"`
	}
	if err := s.core.do(ctx, "GET", "/v1/payouts/destinations", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Destinations, nil
}

// CreateDestination saves a new payout destination.
func (s *PayoutService) CreateDestination(ctx context.Context, params *PayoutDestinationParams, opts ...RequestOption) (*PayoutDestination, error) {
	var out PayoutDestination
	if err := s.core.do(ctx, "POST", "/v1/payouts/destinations", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDestination replaces a saved payout destination. The full destination
// (including DestinationType and Currency) must be supplied.
func (s *PayoutService) UpdateDestination(ctx context.Context, id string, params *PayoutDestinationParams, opts ...RequestOption) (*PayoutDestination, error) {
	var out PayoutDestination
	if err := s.core.do(ctx, "PUT", "/v1/payouts/destinations/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDestination removes a saved payout destination.
func (s *PayoutService) DeleteDestination(ctx context.Context, id string, opts ...RequestOption) error {
	return s.core.do(ctx, "DELETE", "/v1/payouts/destinations/"+url.PathEscape(id), nil, nil, nil, opts)
}

// CreateWithdrawal initiates a withdrawal to a destination. Requires the
// payouts:write scope.
func (s *PayoutService) CreateWithdrawal(ctx context.Context, params *WithdrawalCreateParams, opts ...RequestOption) (*WithdrawalCreated, error) {
	var out WithdrawalCreated
	if err := s.core.do(ctx, "POST", "/v1/payouts/withdrawals", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single payout by its withdrawal ID.
func (s *PayoutService) Get(ctx context.Context, withdrawalID string, opts ...RequestOption) (*Payout, error) {
	var out Payout
	if err := s.core.do(ctx, "GET", "/v1/payouts/"+url.PathEscape(withdrawalID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of payouts. The endpoint reports only a total, so the
// returned page's HasMore is derived from offset + returned < total. Use All to
// iterate across every page automatically.
func (s *PayoutService) List(ctx context.Context, params *PayoutListParams, opts ...RequestOption) (*Page[Payout], error) {
	var wire struct {
		Total int      `json:"total"`
		Items []Payout `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/payouts", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	offset := 0
	if params != nil {
		offset = params.Offset
	}
	returned := len(wire.Items)
	return &Page[Payout]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Offset:   offset,
			Returned: returned,
			HasMore:  offset+returned < wire.Total,
		},
	}, nil
}

// All iterates over every payout matching params, advancing by offset.
func (s *PayoutService) All(ctx context.Context, params *PayoutListParams, opts ...RequestOption) iter.Seq2[Payout, error] {
	var p PayoutListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Payout], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
