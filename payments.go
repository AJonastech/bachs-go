package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Payment status values, as reported by Payment.Status and the status filter on
// PaymentListParams.
const (
	PaymentStatusCreated           = "created"
	PaymentStatusProcessing        = "processing"
	PaymentStatusSucceeded         = "succeeded"
	PaymentStatusAccepted          = "accepted"
	PaymentStatusFailed            = "failed"
	PaymentStatusExpired           = "expired"
	PaymentStatusCancelled         = "cancelled"
	PaymentStatusRefunded          = "refunded"
	PaymentStatusPartiallyRefunded = "partially_refunded"
	PaymentStatusUnderpaid         = "underpaid"
	PaymentStatusOverpaid          = "overpaid"
)

// Billing reason values describe why a payment was created.
const (
	BillingReasonPurchase           = "purchase"
	BillingReasonSubscriptionCreate = "subscription_create"
	BillingReasonSubscriptionCycle  = "subscription_cycle"
	BillingReasonSubscriptionUpdate = "subscription_update"
)

// PaymentStatusEvent is one entry in a payment or charge's status history.
type PaymentStatusEvent struct {
	Status            string    `json:"status"`
	OccurredAt        time.Time `json:"occurred_at"`
	ProviderReference *string   `json:"provider_reference"`
	Reason            *string   `json:"reason"`
}

// PaymentCustomer is the buyer summary embedded in a payment.
type PaymentCustomer struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

// PaymentLineItem is a single product line on a payment.
type PaymentLineItem struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	UnitAmount  Money  `json:"unit_amount"`
	Currency    string `json:"currency"`
	LineTotal   Money  `json:"line_total"`
}

// PaymentInvoiceInfo links a payment to the invoice that generated it.
type PaymentInvoiceInfo struct {
	InvoiceID      string    `json:"invoice_id"`
	Number         *string   `json:"number"`
	SubscriptionID *string   `json:"subscription_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	// Kind is "cycle" or "proration".
	Kind string `json:"kind"`
}

// Payment is a single payment record with its full detail, as returned by
// PaymentService.Get.
type Payment struct {
	// PaymentID is the unique identifier.
	PaymentID string `json:"payment_id"`
	// Reference is your own reference for the payment, if any.
	Reference *string `json:"reference"`
	// BillingReason is why the payment was created; see the BillingReason*
	// constants.
	BillingReason string `json:"billing_reason"`
	// CheckoutID is the checkout that produced this payment, if any.
	CheckoutID *string `json:"checkout_id"`
	// Status is one of the PaymentStatus* constants.
	Status string `json:"status"`
	// IsRefundable reports whether the payment can still be refunded.
	IsRefundable *bool `json:"is_refundable"`
	// Amount is the total amount, as a decimal string.
	Amount Money `json:"amount"`
	// AmountPaid and AmountRemaining track partial payments.
	AmountPaid      *Money `json:"amount_paid"`
	AmountRemaining *Money `json:"amount_remaining"`
	// Currency is the ISO currency code.
	Currency string `json:"currency"`
	// FeeUSD is the Bachs fee in USD, if known.
	FeeUSD *Money `json:"fee_usd"`
	// MerchantBearsCost reports whether you absorb the processing cost.
	MerchantBearsCost *bool   `json:"merchant_bears_cost"`
	PaymentMethod     *string `json:"payment_method"`
	Channel           *string `json:"channel"`
	Narration         *string `json:"narration"`
	// Meta is your own key/value data attached to the payment.
	Meta    Metadata `json:"meta"`
	Message *string  `json:"message"`
	// Customer is the buyer summary, or nil if unavailable.
	Customer *PaymentCustomer `json:"customer"`
	// LineItems are the products purchased.
	LineItems []PaymentLineItem `json:"line_items"`
	// SubscriptionID links the payment to a subscription, if any.
	SubscriptionID *string `json:"subscription_id"`
	// Invoice links the payment to its invoice, if any.
	Invoice *PaymentInvoiceInfo `json:"invoice"`
	// Refunds holds the IDs of refunds issued against this payment.
	Refunds []string `json:"refunds"`
	// StatusHistory is the ordered list of status transitions.
	StatusHistory []PaymentStatusEvent `json:"status_history"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	CompletedAt   *time.Time           `json:"completed_at"`
}

// PaymentListItem is the condensed payment shape returned by list endpoints.
type PaymentListItem struct {
	ID                 *string    `json:"id"`
	Reference          *string    `json:"reference"`
	Status             string     `json:"status"`
	IsRefundable       *bool      `json:"is_refundable"`
	Amount             Money      `json:"amount"`
	CustomerName       string     `json:"customer_name"`
	CustomerEmail      string     `json:"customer_email"`
	AmountPaid         *Money     `json:"amount_paid"`
	AmountRemaining    *Money     `json:"amount_remaining"`
	SettlementAmount   *Money     `json:"settlement_amount"`
	SettlementCurrency *string    `json:"settlement_currency"`
	Fee                *Money     `json:"fee"`
	VAT                *Money     `json:"vat"`
	Currency           string     `json:"currency"`
	Meta               Metadata   `json:"meta"`
	TransactionDate    *time.Time `json:"transaction_date"`
	CompletedAt        *time.Time `json:"completed_at"`
}

// ChargeStatus is the status detail of a single charge attempt, as returned by
// PaymentService.Charge.
type ChargeStatus struct {
	ChargeID           string `json:"charge_id"`
	OrganizationID     string `json:"organization_id"`
	CustomerID         string `json:"customer_id"`
	Amount             Money  `json:"amount"`
	Currency           string `json:"currency"`
	SettlementCurrency string `json:"settlement_currency"`
	SettlementAmount   Money  `json:"settlement_amount"`
	// Status is an upper-case charge status such as "COMPLETED" or "FAILED".
	Status        string               `json:"status"`
	Metadata      Metadata             `json:"metadata"`
	StatusHistory []PaymentStatusEvent `json:"status_history"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
}

// PaymentListParams filters and pages PaymentService.List. Payments are
// paginated by offset.
type PaymentListParams struct {
	// Limit is the page size, 1–100. Defaults to 50 server-side.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// StatusFilter, when set, returns only payments with that exact status; use
	// a PaymentStatus* constant.
	StatusFilter string
}

func (p *PaymentListParams) query() url.Values {
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

// PaymentService accesses the /v1/payments endpoints. Reach it through
// Client.Payments. Payments are created through checkout and subscriptions, so
// this service is read-only.
type PaymentService struct {
	core *client
}

// List returns one page of payments, most recent first. Use All to iterate
// across every page automatically.
func (s *PaymentService) List(ctx context.Context, params *PaymentListParams, opts ...RequestOption) (*Page[PaymentListItem], error) {
	var out Page[PaymentListItem]
	if err := s.core.do(ctx, "GET", "/v1/payments", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// All iterates over every payment matching params, advancing by offset.
func (s *PaymentService) All(ctx context.Context, params *PaymentListParams, opts ...RequestOption) iter.Seq2[PaymentListItem, error] {
	var p PaymentListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[PaymentListItem], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}

// Get retrieves a single payment by ID.
func (s *PaymentService) Get(ctx context.Context, id string, opts ...RequestOption) (*Payment, error) {
	var out Payment
	if err := s.core.do(ctx, "GET", "/v1/payments/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Charge retrieves the status of a single charge by its charge ID.
func (s *PaymentService) Charge(ctx context.Context, chargeID string, opts ...RequestOption) (*ChargeStatus, error) {
	var out ChargeStatus
	if err := s.core.do(ctx, "GET", "/v1/payments/charges/"+url.PathEscape(chargeID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
