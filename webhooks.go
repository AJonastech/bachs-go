package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Webhook event types that an endpoint can subscribe to.
const (
	EventCollectionSucceeded  = "collection.succeeded"
	EventCollectionFailed     = "collection.failed"
	EventCollectionUnderpaid  = "collection.underpaid"
	EventCheckoutCompleted    = "checkout.completed"
	EventCheckoutExpired      = "checkout.expired"
	EventPayoutCreated        = "payout.created"
	EventPayoutPaid           = "payout.paid"
	EventPayoutFailed         = "payout.failed"
	EventRefundCreated        = "refund.created"
	EventRefundPaid           = "refund.paid"
	EventRefundFailed         = "refund.failed"
	EventConversionCompleted  = "conversion.completed"
	EventConversionFailed     = "conversion.failed"
	EventCustomerCreated      = "customer.created"
	EventCustomerUpdated      = "customer.updated"
	EventDisputeCreated       = "dispute.created"
	EventDisputeUpdated       = "dispute.updated"
	EventSubscriptionCreated  = "customer.subscription.created"
	EventSubscriptionUpdated  = "customer.subscription.updated"
	EventSubscriptionDeleted  = "customer.subscription.deleted"
	EventInvoiceCreated       = "invoice.created"
	EventInvoicePaid          = "invoice.paid"
	EventInvoicePaymentFailed = "invoice.payment_failed"
)

// WebhookEndpoint is a subscription that delivers events to a URL.
type WebhookEndpoint struct {
	EndpointID string    `json:"endpoint_id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Enabled    bool      `json:"enabled"`
	EventTypes []string  `json:"event_types"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// WebhookEndpointWithSecret is a newly created endpoint together with its
// signing secret. The secret is used to verify delivered event signatures and
// is returned only here and from RotateSecret / Secret.
type WebhookEndpointWithSecret struct {
	WebhookEndpoint
	SigningSecret string `json:"signing_secret"`
}

// WebhookEndpointSecret is an endpoint's current signing secret, as returned by
// WebhookService.Secret.
type WebhookEndpointSecret struct {
	EndpointID string    `json:"endpoint_id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Enabled    bool      `json:"enabled"`
	EventTypes []string  `json:"event_types"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Secret     string    `json:"secret"`
}

// WebhookEventAttempt is one delivery attempt of an event to an endpoint.
type WebhookEventAttempt struct {
	AttemptID       string    `json:"attempt_id"`
	AttemptNo       int       `json:"attempt_no"`
	Status          string    `json:"status"`
	CallbackURL     *string   `json:"callback_url"`
	HTTPStatus      *int      `json:"http_status"`
	ResponseSnippet *string   `json:"response_snippet"`
	LastError       *string   `json:"last_error"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WebhookEvent is a single webhook event with its payload and delivery attempts.
type WebhookEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	EntityType *string   `json:"entity_type"`
	EntityID   *string   `json:"entity_id"`
	CreatedAt  time.Time `json:"created_at"`
	// Payload is the event body delivered to endpoints.
	Payload  map[string]any        `json:"payload"`
	Attempts []WebhookEventAttempt `json:"attempts"`
}

// WebhookEventListItem is the condensed event shape from the account-wide event
// list.
type WebhookEventListItem struct {
	EventID       string     `json:"event_id"`
	EventType     string     `json:"event_type"`
	EntityType    *string    `json:"entity_type"`
	EntityID      *string    `json:"entity_id"`
	CreatedAt     time.Time  `json:"created_at"`
	Attempts      int        `json:"attempts"`
	Success       int        `json:"success"`
	Failed        int        `json:"failed"`
	LastAttemptAt *time.Time `json:"last_attempt_at"`
}

// WebhookEndpointEventListItem is the condensed event shape from a single
// endpoint's event list. It carries per-endpoint delivery detail.
type WebhookEndpointEventListItem struct {
	EventID               string     `json:"event_id"`
	EventType             string     `json:"event_type"`
	EntityID              *string    `json:"entity_id"`
	Attempts              int        `json:"attempts"`
	Success               int        `json:"success"`
	Failed                int        `json:"failed"`
	LastAttemptStatus     *string    `json:"last_attempt_status"`
	LastAttemptHTTPStatus *int       `json:"last_attempt_http_status"`
	LastAttemptAt         *time.Time `json:"last_attempt_at"`
}

// WebhookMetricsDataPoint is one bucket of endpoint delivery metrics.
type WebhookMetricsDataPoint struct {
	Date    string `json:"date"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
}

// WebhookMetrics is an endpoint's delivery metrics over a period.
type WebhookMetrics struct {
	Total  string                    `json:"total"`
	Period string                    `json:"period"`
	Data   []WebhookMetricsDataPoint `json:"data"`
}

// ResendResult is returned by WebhookService.ResendEvent.
type ResendResult struct {
	Status    string `json:"status"`
	AttemptID string `json:"attempt_id"`
}

// ReplayResult is returned by WebhookService.Replay.
type ReplayResult struct {
	Success   bool   `json:"success"`
	EventID   string `json:"event_id"`
	AttemptID string `json:"attempt_id"`
	AttemptNo int    `json:"attempt_no"`
	EventType string `json:"event_type"`
}

// WebhookEndpointCreateParams are the parameters for
// WebhookService.CreateEndpoint. All fields are required. Use the Event*
// constants for EventTypes.
type WebhookEndpointCreateParams struct {
	Name       string   `json:"name"`
	URL        string   `json:"url"`
	EventTypes []string `json:"event_types"`
}

// WebhookEndpointUpdateParams are the parameters for
// WebhookService.UpdateEndpoint. Only the fields you set are changed.
type WebhookEndpointUpdateParams struct {
	Name       *string  `json:"name,omitempty"`
	URL        *string  `json:"url,omitempty"`
	EventTypes []string `json:"event_types,omitempty"`
}

// WebhookReplayParams selects the event to replay. Provide exactly one of
// EventID, ChargeID, or Reference.
type WebhookReplayParams struct {
	EventID   string `json:"event_id,omitempty"`
	ChargeID  string `json:"charge_id,omitempty"`
	Reference string `json:"reference,omitempty"`
}

// WebhookEventListParams pages the webhook event lists. Events are paginated by
// offset.
type WebhookEventListParams struct {
	Limit  int
	Offset int
}

func (p *WebhookEventListParams) query() url.Values {
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
	return v
}

// WebhookMetricsParams filters WebhookService.EndpointMetrics.
type WebhookMetricsParams struct {
	// Period groups the data, e.g. "day".
	Period string
	// DateFrom and DateTo bound the range (ISO-8601).
	DateFrom string
	DateTo   string
}

func (p *WebhookMetricsParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Period != "" {
		v.Set("period", p.Period)
	}
	if p.DateFrom != "" {
		v.Set("date_from", p.DateFrom)
	}
	if p.DateTo != "" {
		v.Set("date_to", p.DateTo)
	}
	return v
}

// WebhookService accesses the /v1/webhooks endpoints: endpoint management,
// signing secrets, delivered events, and replay. Reach it through
// Client.Webhooks.
type WebhookService struct {
	core *client
}

// CreateEndpoint registers a webhook endpoint. The response includes the
// signing secret, which is your only chance to read it in full alongside
// creation; store it to verify event signatures.
func (s *WebhookService) CreateEndpoint(ctx context.Context, params *WebhookEndpointCreateParams, opts ...RequestOption) (*WebhookEndpointWithSecret, error) {
	var out WebhookEndpointWithSecret
	if err := s.core.do(ctx, "POST", "/v1/webhooks/endpoints", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEndpoints returns all webhook endpoints. This endpoint is not paginated.
func (s *WebhookService) ListEndpoints(ctx context.Context, opts ...RequestOption) ([]WebhookEndpoint, error) {
	var out []WebhookEndpoint
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// GetEndpoint retrieves a single webhook endpoint by ID.
func (s *WebhookService) GetEndpoint(ctx context.Context, id string, opts ...RequestOption) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEndpoint changes the fields set on params and returns the updated
// endpoint.
func (s *WebhookService) UpdateEndpoint(ctx context.Context, id string, params *WebhookEndpointUpdateParams, opts ...RequestOption) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	if err := s.core.do(ctx, "PATCH", "/v1/webhooks/endpoints/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEndpoint removes a webhook endpoint.
func (s *WebhookService) DeleteEndpoint(ctx context.Context, id string, opts ...RequestOption) error {
	return s.core.do(ctx, "DELETE", "/v1/webhooks/endpoints/"+url.PathEscape(id), nil, nil, nil, opts)
}

// Secret retrieves an endpoint's current signing secret.
func (s *WebhookService) Secret(ctx context.Context, endpointID string, opts ...RequestOption) (*WebhookEndpointSecret, error) {
	var out WebhookEndpointSecret
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/secret", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// RotateSecret generates a new signing secret for an endpoint and returns the
// endpoint. Any consumers verifying signatures must be updated to the new
// secret.
func (s *WebhookService) RotateSecret(ctx context.Context, endpointID string, opts ...RequestOption) (*WebhookEndpoint, error) {
	var out WebhookEndpoint
	if err := s.core.do(ctx, "POST", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/rotate-secret", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// EndpointMetrics retrieves delivery metrics for an endpoint over a period.
func (s *WebhookService) EndpointMetrics(ctx context.Context, endpointID string, params *WebhookMetricsParams, opts ...RequestOption) (*WebhookMetrics, error) {
	var out WebhookMetrics
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/metrics", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEvents returns one page of account-wide webhook events. Use AllEvents to
// iterate across every page automatically.
func (s *WebhookService) ListEvents(ctx context.Context, params *WebhookEventListParams, opts ...RequestOption) (*Page[WebhookEventListItem], error) {
	var wire struct {
		Items  []WebhookEventListItem `json:"items"`
		Total  int                    `json:"total"`
		Limit  int                    `json:"limit"`
		Offset int                    `json:"offset"`
	}
	if err := s.core.do(ctx, "GET", "/v1/webhooks/events", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	returned := len(wire.Items)
	return &Page[WebhookEventListItem]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Limit:    wire.Limit,
			Offset:   wire.Offset,
			Returned: returned,
			HasMore:  wire.Offset+returned < wire.Total,
		},
	}, nil
}

// AllEvents iterates over every account-wide webhook event, advancing by offset.
func (s *WebhookService) AllEvents(ctx context.Context, params *WebhookEventListParams, opts ...RequestOption) iter.Seq2[WebhookEventListItem, error] {
	var p WebhookEventListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[WebhookEventListItem], error) { return s.ListEvents(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}

// GetEvent retrieves a single webhook event (with payload and attempts) by ID.
func (s *WebhookService) GetEvent(ctx context.Context, eventID string, opts ...RequestOption) (*WebhookEvent, error) {
	var out WebhookEvent
	if err := s.core.do(ctx, "GET", "/v1/webhooks/events/"+url.PathEscape(eventID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEndpointEvents returns one page of events delivered to a specific
// endpoint. Use AllEndpointEvents to iterate across every page automatically.
func (s *WebhookService) ListEndpointEvents(ctx context.Context, endpointID string, params *WebhookEventListParams, opts ...RequestOption) (*Page[WebhookEndpointEventListItem], error) {
	var wire struct {
		Items  []WebhookEndpointEventListItem `json:"items"`
		Total  int                            `json:"total"`
		Limit  int                            `json:"limit"`
		Offset int                            `json:"offset"`
	}
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	returned := len(wire.Items)
	return &Page[WebhookEndpointEventListItem]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Limit:    wire.Limit,
			Offset:   wire.Offset,
			Returned: returned,
			HasMore:  wire.Offset+returned < wire.Total,
		},
	}, nil
}

// AllEndpointEvents iterates over every event delivered to an endpoint,
// advancing by offset.
func (s *WebhookService) AllEndpointEvents(ctx context.Context, endpointID string, params *WebhookEventListParams, opts ...RequestOption) iter.Seq2[WebhookEndpointEventListItem, error] {
	var p WebhookEventListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[WebhookEndpointEventListItem], error) {
			return s.ListEndpointEvents(ctx, endpointID, &p, opts...)
		},
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}

// GetEndpointEvent retrieves a single event delivered to a specific endpoint.
func (s *WebhookService) GetEndpointEvent(ctx context.Context, endpointID, eventID string, opts ...RequestOption) (*WebhookEvent, error) {
	var out WebhookEvent
	if err := s.core.do(ctx, "GET", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events/"+url.PathEscape(eventID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ResendEvent redelivers an event to a specific endpoint.
func (s *WebhookService) ResendEvent(ctx context.Context, endpointID, eventID string, opts ...RequestOption) (*ResendResult, error) {
	var out ResendResult
	if err := s.core.do(ctx, "POST", "/v1/webhooks/endpoints/"+url.PathEscape(endpointID)+"/events/"+url.PathEscape(eventID)+"/resend", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Replay re-emits a webhook event selected by event ID, charge ID, or
// reference.
func (s *WebhookService) Replay(ctx context.Context, params *WebhookReplayParams, opts ...RequestOption) (*ReplayResult, error) {
	var out ReplayResult
	if err := s.core.do(ctx, "POST", "/v1/webhooks/replay", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
