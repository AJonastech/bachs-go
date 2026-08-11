package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Subscription status values.
const (
	SubscriptionStatusTrialing = "trialing"
	SubscriptionStatusActive   = "active"
	SubscriptionStatusPastDue  = "past_due"
	SubscriptionStatusUnpaid   = "unpaid"
	SubscriptionStatusCanceled = "canceled"
	SubscriptionStatusPaused   = "paused"
)

// Proration behavior values control how a subscription change is billed.
const (
	// ProrationInvoiceNow bills the prorated difference immediately.
	ProrationInvoiceNow = "invoice_now"
	// ProrationNextCycle rolls the change into the next billing cycle.
	ProrationNextCycle = "next_cycle"
	// ProrationNone applies the change without proration.
	ProrationNone = "none"
)

// SubscriptionCatalogProduct is the product snapshot embedded in a subscription
// or subscription item.
type SubscriptionCatalogProduct struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  *string      `json:"description"`
	Status       string       `json:"status"`
	BillingCycle *Cadence     `json:"billing_cycle"`
	TrialPeriod  *TrialPeriod `json:"trial_period"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// SubscriptionItemPrice is the price snapshot on a subscription item.
type SubscriptionItemPrice struct {
	ID           string         `json:"id"`
	ProductID    string         `json:"product_id"`
	PriceType    string         `json:"price_type"`
	Currency     string         `json:"currency"`
	UnitAmount   Money          `json:"unit_amount"`
	BillingCycle *Cadence       `json:"billing_cycle"`
	TrialPeriod  *TrialPeriod   `json:"trial_period"`
	SeatTiers    map[string]any `json:"seat_tiers"`
	IsArchived   bool           `json:"is_archived"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// SubscriptionItem is a single line on a subscription.
type SubscriptionItem struct {
	ID                 string                      `json:"id"`
	Status             string                      `json:"status"`
	Quantity           int                         `json:"quantity"`
	Recurring          bool                        `json:"recurring"`
	PriceType          string                      `json:"price_type"`
	UnitAmount         Money                       `json:"unit_amount"`
	Currency           string                      `json:"currency"`
	PreviouslyBilledAt *time.Time                  `json:"previously_billed_at"`
	NextBilledAt       *time.Time                  `json:"next_billed_at"`
	Price              *SubscriptionItemPrice      `json:"price"`
	Product            *SubscriptionCatalogProduct `json:"product"`
	CreatedAt          time.Time                   `json:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at"`
}

// Subscription is a recurring billing arrangement between you and a customer.
// It is the object returned by the subscription retrieve, update, list, and
// cancel endpoints.
type Subscription struct {
	// ID is the unique identifier, prefixed with "sub_".
	ID string `json:"id"`
	// Customer is the subscriber.
	Customer Customer `json:"customer"`
	// PaymentMethodID is the payment method billed, or nil if none is attached.
	PaymentMethodID *string `json:"payment_method_id"`
	// Status is one of the SubscriptionStatus* constants.
	Status string `json:"status"`
	// CollectionMethod is how the subscription is collected (e.g. charge
	// automatically).
	CollectionMethod string `json:"collection_method"`
	// Currency is the ISO currency code billed.
	Currency string `json:"currency"`
	// Amount is the recurring amount, as a decimal string.
	Amount Money `json:"amount"`
	// BillingCycle is the recurrence cadence.
	BillingCycle Cadence `json:"billing_cycle"`
	// Quantity is the subscribed quantity.
	Quantity int `json:"quantity"`
	// CurrentPeriodStart and CurrentPeriodEnd bound the current billing period.
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	// PreviouslyBilledAt and NextBilledAt are the surrounding billing instants.
	PreviouslyBilledAt *time.Time `json:"previously_billed_at"`
	NextBilledAt       *time.Time `json:"next_billed_at"`
	// TrialEnd is when the trial ends, or nil if there is no trial.
	TrialEnd *time.Time `json:"trial_end"`
	// CancelAtPeriodEnd reports whether the subscription will end at period end.
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
	// CanceledAt is when the subscription was canceled, or nil.
	CanceledAt *time.Time `json:"canceled_at"`
	CreatedAt  time.Time  `json:"created_at"`
	// Product is the primary product snapshot, or nil.
	Product *SubscriptionCatalogProduct `json:"product"`
	// Items are the subscription's line items.
	Items []SubscriptionItem `json:"items"`
	// Metadata is your own key/value data.
	Metadata Metadata `json:"metadata"`
}

// SubscriptionUpdateParams are the parameters for SubscriptionService.Update.
// Only the fields you set are changed.
type SubscriptionUpdateParams struct {
	// ProductID switches the subscription to a different product.
	ProductID *string `json:"product_id,omitempty"`
	// TrialEnd sets or extends the trial end.
	TrialEnd *time.Time `json:"trial_end,omitempty"`
	// PaymentMethodID changes the payment method billed.
	PaymentMethodID *string `json:"payment_method_id,omitempty"`
	// Metadata replaces the stored metadata.
	Metadata Metadata `json:"metadata,omitempty"`
	// ProrationBehavior controls how the change is billed; use a Proration*
	// constant.
	ProrationBehavior *string `json:"proration_behavior,omitempty"`
}

// SubscriptionCancelParams are the parameters for SubscriptionService.Cancel.
// The zero value cancels immediately with no reason.
type SubscriptionCancelParams struct {
	// CancelAtPeriodEnd cancels at the end of the current period when true, or
	// immediately when false.
	CancelAtPeriodEnd bool `json:"cancel_at_period_end"`
	// Reason is an optional free-text note (max 255 characters).
	Reason *string `json:"reason,omitempty"`
}

// SubscriptionListParams filters and pages SubscriptionService.List.
// Subscriptions are paginated by offset.
type SubscriptionListParams struct {
	// Limit is the page size. Defaults server-side.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// CustomerID limits results to one customer (cust_...).
	CustomerID string
	// Status filters by subscription status; use a SubscriptionStatus* constant.
	Status string
}

func (p *SubscriptionListParams) query() url.Values {
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
	if p.CustomerID != "" {
		v.Set("customer_id", p.CustomerID)
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	return v
}

// SubscriptionService accesses the /v1/subscriptions endpoints. Reach it
// through Client.Subscriptions. Subscriptions are created through checkout, so
// this service does not expose a create method.
type SubscriptionService struct {
	core *client
}

// Get retrieves a single subscription by ID. Requires the subscriptions:read
// scope.
func (s *SubscriptionService) Get(ctx context.Context, id string, opts ...RequestOption) (*Subscription, error) {
	var out Subscription
	if err := s.core.do(ctx, "GET", "/v1/subscriptions/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the fields set on params and returns the updated subscription.
// Requires the subscriptions:write scope.
func (s *SubscriptionService) Update(ctx context.Context, id string, params *SubscriptionUpdateParams, opts ...RequestOption) (*Subscription, error) {
	var out Subscription
	if err := s.core.do(ctx, "PATCH", "/v1/subscriptions/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Cancel cancels a subscription, either immediately or at the end of the
// current period, and returns the updated subscription. Requires the
// subscriptions:write scope. Pass nil params to cancel immediately.
func (s *SubscriptionService) Cancel(ctx context.Context, id string, params *SubscriptionCancelParams, opts ...RequestOption) (*Subscription, error) {
	var out Subscription
	if err := s.core.do(ctx, "DELETE", "/v1/subscriptions/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of subscriptions. Use All to iterate across every page
// automatically.
func (s *SubscriptionService) List(ctx context.Context, params *SubscriptionListParams, opts ...RequestOption) (*Page[Subscription], error) {
	var out Page[Subscription]
	if err := s.core.do(ctx, "GET", "/v1/subscriptions", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// All iterates over every subscription matching params, advancing by offset.
func (s *SubscriptionService) All(ctx context.Context, params *SubscriptionListParams, opts ...RequestOption) iter.Seq2[Subscription, error] {
	var p SubscriptionListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Subscription], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
