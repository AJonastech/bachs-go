package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Conversion status values.
const (
	ConversionStatusPending   = "pending"
	ConversionStatusCompleted = "completed"
	ConversionStatusFailed    = "failed"
)

// ConversionQuote is a locked exchange rate for a currency conversion, valid
// until ExpiresAt. Pass its QuoteID to ConversionService.Create to execute at
// the quoted rate.
type ConversionQuote struct {
	QuoteID      string    `json:"quote_id"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	FromAmount   Money     `json:"from_amount"`
	ToAmount     Money     `json:"to_amount"`
	ExchangeRate string    `json:"exchange_rate"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Conversion is an executed currency conversion. It is the object returned by
// the conversion execute, retrieve, and list endpoints.
type Conversion struct {
	ConversionID string `json:"conversion_id"`
	// Status is one of the ConversionStatus* constants.
	Status       string    `json:"status"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	FromAmount   Money     `json:"from_amount"`
	ToAmount     Money     `json:"to_amount"`
	ExchangeRate string    `json:"exchange_rate"`
	CreatedAt    time.Time `json:"created_at"`
	// QuoteID is the quote this conversion executed against, if any.
	QuoteID *string `json:"quote_id"`
	// Metadata is your own key/value data.
	Metadata Metadata `json:"metadata"`
}

// ConversionQuoteParams are the parameters for ConversionService.Quote. All
// three fields are required.
type ConversionQuoteParams struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       Money  `json:"amount"`
}

// ConversionCreateParams are the parameters for ConversionService.Create.
// Execute a previously fetched quote by passing its QuoteID. All fields are
// required.
type ConversionCreateParams struct {
	FromCurrency string `json:"from_currency"`
	ToCurrency   string `json:"to_currency"`
	Amount       Money  `json:"amount"`
	// QuoteID is the quote to execute, from ConversionService.Quote.
	QuoteID string `json:"quote_id"`
}

// ConversionListParams filters and pages ConversionService.List. Conversions
// are paginated by offset.
type ConversionListParams struct {
	Limit        int
	Offset       int
	FromCurrency string
	ToCurrency   string
	Status       string
	// StartDate and EndDate bound the results by creation date; the API accepts
	// ISO-8601 date strings.
	StartDate string
	EndDate   string
}

func (p *ConversionListParams) query() url.Values {
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
	if p.FromCurrency != "" {
		v.Set("from_currency", p.FromCurrency)
	}
	if p.ToCurrency != "" {
		v.Set("to_currency", p.ToCurrency)
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.StartDate != "" {
		v.Set("start_date", p.StartDate)
	}
	if p.EndDate != "" {
		v.Set("end_date", p.EndDate)
	}
	return v
}

// ConversionService accesses the /v1/conversions endpoints. Reach it through
// Client.Conversions. The usual flow is Quote to lock a rate, then Create to
// execute it.
type ConversionService struct {
	core *client
}

// Quote locks an exchange rate for a conversion. Requires the conversions:write
// scope.
func (s *ConversionService) Quote(ctx context.Context, params *ConversionQuoteParams, opts ...RequestOption) (*ConversionQuote, error) {
	var out ConversionQuote
	if err := s.core.do(ctx, "POST", "/v1/conversions/quotes", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create executes a conversion, typically against a quote from Quote. Requires
// the conversions:write scope.
func (s *ConversionService) Create(ctx context.Context, params *ConversionCreateParams, opts ...RequestOption) (*Conversion, error) {
	var out Conversion
	if err := s.core.do(ctx, "POST", "/v1/conversions", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single conversion by ID. Requires the conversions:read scope.
func (s *ConversionService) Get(ctx context.Context, id string, opts ...RequestOption) (*Conversion, error) {
	var out Conversion
	if err := s.core.do(ctx, "GET", "/v1/conversions/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of conversions. The endpoint reports total, limit, and
// offset rather than a cursor, so the returned page's HasMore is derived from
// offset + returned < total. Use All to iterate across every page.
func (s *ConversionService) List(ctx context.Context, params *ConversionListParams, opts ...RequestOption) (*Page[Conversion], error) {
	var wire struct {
		Total  int          `json:"total"`
		Limit  int          `json:"limit"`
		Offset int          `json:"offset"`
		Items  []Conversion `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/conversions", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	returned := len(wire.Items)
	return &Page[Conversion]{
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

// All iterates over every conversion matching params, advancing by offset.
func (s *ConversionService) All(ctx context.Context, params *ConversionListParams, opts ...RequestOption) iter.Seq2[Conversion, error] {
	var p ConversionListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Conversion], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
