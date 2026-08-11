package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Transfer status values.
const (
	TransferStatusPending = "pending"
	TransferStatusPaid    = "paid"
)

// Transfer moves funds to a destination such as a connected account. It is the
// object returned by the transfer create, retrieve, and list endpoints.
type Transfer struct {
	// ID is the unique identifier.
	ID string `json:"id"`
	// Source is the account the funds moved from.
	Source string `json:"source"`
	// Destination is the account the funds moved to.
	Destination string `json:"destination"`
	// Amount is the transferred amount, as a decimal string.
	Amount Money `json:"amount"`
	// Currency is the ISO currency code.
	Currency string `json:"currency"`
	// Status is TransferStatusPending or TransferStatusPaid.
	Status string `json:"status"`
	// Description is an optional description.
	Description *string `json:"description"`
	// Metadata is your own key/value data.
	Metadata Metadata `json:"metadata"`
	// TransferGroup is an optional label linking related transfers.
	TransferGroup *string `json:"transfer_group"`
	// CreatedAt is when the transfer was created.
	CreatedAt time.Time `json:"created_at"`
}

// TransferCreateParams are the parameters for TransferService.Create.
// Destination, Amount, and Currency are required.
type TransferCreateParams struct {
	// Destination is the account to transfer to. Required.
	Destination string `json:"destination"`
	// Amount is the amount to transfer, as a decimal string. Required.
	Amount Money `json:"amount"`
	// Currency is the ISO currency code. Required.
	Currency string `json:"currency"`
	// Description is an optional description.
	Description *string `json:"description,omitempty"`
	// Metadata is optional key/value data.
	Metadata Metadata `json:"metadata,omitempty"`
	// TransferGroup optionally labels related transfers.
	TransferGroup *string `json:"transfer_group,omitempty"`
}

// TransferListParams filters and pages TransferService.List. Transfers are
// paginated by offset.
type TransferListParams struct {
	// Limit is the page size.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// ConnectedAccountID limits results to transfers involving that connected
	// account (acct_...).
	ConnectedAccountID string
}

func (p *TransferListParams) query() url.Values {
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
	if p.ConnectedAccountID != "" {
		v.Set("connected_account_id", p.ConnectedAccountID)
	}
	return v
}

// TransferService accesses the /v1/transfers endpoints. Reach it through
// Client.Transfers.
type TransferService struct {
	core *client
}

// Create moves funds to a destination account. Requires the transfers:write
// scope.
func (s *TransferService) Create(ctx context.Context, params *TransferCreateParams, opts ...RequestOption) (*Transfer, error) {
	var out Transfer
	if err := s.core.do(ctx, "POST", "/v1/transfers", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single transfer by ID. Requires the transfers:read scope.
func (s *TransferService) Get(ctx context.Context, id string, opts ...RequestOption) (*Transfer, error) {
	var out Transfer
	if err := s.core.do(ctx, "GET", "/v1/transfers/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of transfers. The transfers endpoint reports total,
// limit, and offset rather than a cursor, so the returned page's HasMore is
// derived from offset + returned < total. Use All to iterate across every page
// automatically.
func (s *TransferService) List(ctx context.Context, params *TransferListParams, opts ...RequestOption) (*Page[Transfer], error) {
	var wire struct {
		Items  []Transfer `json:"items"`
		Total  int        `json:"total"`
		Limit  int        `json:"limit"`
		Offset int        `json:"offset"`
	}
	if err := s.core.do(ctx, "GET", "/v1/transfers", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	returned := len(wire.Items)
	return &Page[Transfer]{
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

// All iterates over every transfer matching params, advancing by offset.
func (s *TransferService) All(ctx context.Context, params *TransferListParams, opts ...RequestOption) iter.Seq2[Transfer, error] {
	var p TransferListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Transfer], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
