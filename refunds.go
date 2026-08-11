package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Refund status values, as reported by Refund.Status.
const (
	RefundStatusProcessing = "processing"
	RefundStatusSuccess    = "success"
	RefundStatusFailed     = "failed"
)

// Fee bearer values decide who absorbs the refund fee.
const (
	// FeeBearerOrg means your organization absorbs the refund fee.
	FeeBearerOrg = "org"
	// FeeBearerCustomer means the fee is deducted from the customer's refund.
	FeeBearerCustomer = "customer"
)

// Refund is money returned to a customer for a charge. It is the object
// returned by the refund create, retrieve, and list endpoints.
type Refund struct {
	// RefundID is the unique identifier.
	RefundID string `json:"refund_id"`
	// ChargeID is the charge that was refunded.
	ChargeID string `json:"charge_id"`
	// Reference is your own reference for the refund.
	Reference string `json:"reference"`
	// Status is one of the RefundStatus* constants.
	Status string `json:"status"`
	// RequestedAmount is the amount requested for refund.
	RequestedAmount Money `json:"requested_amount"`
	// RefundedAmount is the amount actually refunded, or nil while processing.
	RefundedAmount *Money `json:"refunded_amount"`
	// RefundFeeAmount is the fee charged for the refund.
	RefundFeeAmount Money `json:"refund_fee_amount"`
	// FeeBearer is FeeBearerOrg or FeeBearerCustomer.
	FeeBearer string `json:"fee_bearer"`
	// Reason is the optional refund reason.
	Reason      *string    `json:"reason"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// RefundCreateParams are the parameters for RefundService.Create. ChargeID and
// Reference are required.
type RefundCreateParams struct {
	// ChargeID is the charge to refund. Required.
	ChargeID string `json:"charge_id"`
	// Reference is your own unique reference for the refund. Required.
	Reference string `json:"reference"`
	// RefundAddress is the destination for a crypto refund, when applicable.
	RefundAddress *string `json:"refund_address,omitempty"`
	// Amount is the amount to refund; omit for a full refund.
	Amount *Money `json:"amount,omitempty"`
	// FeeBearer decides who absorbs the fee: FeeBearerOrg or FeeBearerCustomer.
	FeeBearer *string `json:"fee_bearer,omitempty"`
	// Reason is an optional human-readable reason.
	Reason *string `json:"reason,omitempty"`
	// IdempotencyKey is an optional body-level dedupe key. Prefer the
	// WithIdempotencyKey request option, which sets the standard header.
	IdempotencyKey *string `json:"idempotency_key,omitempty"`
	// SimulatedOutcome forces a "success" or "failed" result in the sandbox; it
	// is ignored in production.
	SimulatedOutcome *string `json:"simulated_outcome,omitempty"`
}

// RefundListParams filters and pages RefundService.List. Refunds are paginated
// by offset.
type RefundListParams struct {
	// Limit is the page size, 1–100. Defaults to 50 server-side.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// Status filters by refund status. The API expects upper-case values:
	// "PROCESSING", "SUCCESS", or "FAILED".
	Status string
}

func (p *RefundListParams) query() url.Values {
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
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	return v
}

// RefundService accesses the /v1/refunds endpoints. Reach it through
// Client.Refunds.
type RefundService struct {
	core *client
}

// Create issues a refund against a charge. Omit Amount for a full refund.
// Requires the refunds:write scope.
func (s *RefundService) Create(ctx context.Context, params *RefundCreateParams, opts ...RequestOption) (*Refund, error) {
	var out Refund
	if err := s.core.do(ctx, "POST", "/v1/refunds", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single refund by ID. Requires the refunds:read scope.
func (s *RefundService) Get(ctx context.Context, id string, opts ...RequestOption) (*Refund, error) {
	var out Refund
	if err := s.core.do(ctx, "GET", "/v1/refunds/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetByCharge retrieves the refund associated with a charge (payment) ID.
// Requires the refunds:read scope.
func (s *RefundService) GetByCharge(ctx context.Context, paymentID string, opts ...RequestOption) (*Refund, error) {
	var out Refund
	if err := s.core.do(ctx, "GET", "/v1/refunds/by-charge/"+url.PathEscape(paymentID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of refunds. The refunds endpoint reports only a total
// count rather than a pagination cursor, so the returned page's metadata
// carries Total, Offset, Returned, and a derived HasMore. Use All to iterate
// across every page automatically.
func (s *RefundService) List(ctx context.Context, params *RefundListParams, opts ...RequestOption) (*Page[Refund], error) {
	// The wire response is {total, items} rather than the standard envelope.
	var wire struct {
		Total int      `json:"total"`
		Items []Refund `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/refunds", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}

	var offset, limit int
	if params != nil {
		offset, limit = params.Offset, params.Limit
	}
	returned := len(wire.Items)
	return &Page[Refund]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Returned: returned,
			Offset:   offset,
			Limit:    limit,
			HasMore:  offset+returned < wire.Total,
		},
	}, nil
}

// All iterates over every refund matching params, advancing by offset.
func (s *RefundService) All(ctx context.Context, params *RefundListParams, opts ...RequestOption) iter.Seq2[Refund, error] {
	var p RefundListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Refund], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
