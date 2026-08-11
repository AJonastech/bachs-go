package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Product status values.
const (
	// ProductStatusActive means the product is live and usable in checkouts and
	// subscriptions.
	ProductStatusActive = "active"
	// ProductStatusArchived means the product is retired: kept for reference but
	// not available for new purchases.
	ProductStatusArchived = "archived"
)

// Product is an item you sell, with its pricing. Products are sold through
// checkout sessions and subscriptions. It is the object returned by the product
// create, retrieve, update, list, archive, and unarchive endpoints.
type Product struct {
	// ID is the unique identifier, prefixed with "prod_".
	ID string `json:"id"`
	// OrganizationID is the organization that owns the product.
	OrganizationID string `json:"organization_id"`
	// Name is the display name shown to customers at checkout.
	Name string `json:"name"`
	// Description is an optional description. Nil when not set.
	Description *string `json:"description"`
	// Price is the primary price, in the product's default currency.
	Price Price `json:"price"`
	// Status is ProductStatusActive or ProductStatusArchived.
	Status string `json:"status"`
	// Metadata is your own key/value data, returned unchanged.
	Metadata Metadata `json:"metadata"`
	// Media holds images attached to the product; empty when none are set.
	Media []MediaItem `json:"media"`
	// ActorID identifies the user or key that created the product.
	ActorID string `json:"actor_id"`
	// CreatedAt is when the product was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the product was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// ArchivedAt is when the product was archived, or nil while active.
	ArchivedAt *time.Time `json:"archived_at"`
	// BillingCycle, when non-nil, makes the product recurring on that cadence;
	// nil for a one-time product.
	BillingCycle *Cadence `json:"billing_cycle"`
	// TrialPeriod is a free trial before the first charge. Only valid on a
	// recurring product. Nil when none.
	TrialPeriod *TrialPeriod `json:"trial_period"`
	// Prices holds all configured prices, one per currency.
	Prices []ProductPrice `json:"prices"`
	// TotalPayments is the running count of completed payments for this product.
	TotalPayments int `json:"total_payments"`
	// TotalAmount is the running total collected, as a decimal string.
	TotalAmount Money `json:"total_amount"`
}

// ProductPrice is one currency's price on a product, as returned in
// Product.Prices.
type ProductPrice struct {
	Currency      string `json:"currency"`
	Amount        Money  `json:"amount"`
	MinimumAmount *Money `json:"minimum_amount"`
	MaximumAmount *Money `json:"maximum_amount"`
	IsDefault     bool   `json:"is_default"`
}

// ProductCreateParams are the parameters for ProductService.Create. Name and
// Price are required. Add a BillingCycle to make the product recurring, or omit
// it for a one-time product.
type ProductCreateParams struct {
	Name         string       `json:"name"`
	Description  *string      `json:"description,omitempty"`
	Price        PriceParams  `json:"price"`
	Metadata     Metadata     `json:"metadata,omitempty"`
	BillingCycle *Cadence     `json:"billing_cycle,omitempty"`
	TrialPeriod  *TrialPeriod `json:"trial_period,omitempty"`
}

// ProductUpdateParams are the parameters for ProductService.Update. Only the
// fields you set are changed. A billing cycle is immutable once set, so a
// recurring product's interval cannot be changed — create a new product for a
// different cadence.
type ProductUpdateParams struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
	// Media is an ordered list of upload IDs that replaces existing media.
	Media        []string           `json:"media,omitempty"`
	Price        *UpdatePriceParams `json:"price,omitempty"`
	BillingCycle *Cadence           `json:"billing_cycle,omitempty"`
	TrialPeriod  *TrialPeriod       `json:"trial_period,omitempty"`
}

// ProductListParams filters and pages ProductService.List. Products are
// paginated by cursor.
type ProductListParams struct {
	// Limit is the page size, 1–100. Defaults to 20 server-side.
	Limit int
	// Cursor is a next_cursor from a previous page. Empty starts at the first
	// page.
	Cursor string
	// IncludeArchived includes archived products in the results.
	IncludeArchived bool
}

// query encodes the params as URL query values. A nil receiver yields empty
// values.
func (p *ProductListParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Cursor != "" {
		v.Set("cursor", p.Cursor)
	}
	if p.IncludeArchived {
		v.Set("include_archived", "true")
	}
	return v
}

// ProductService accesses the /v1/products endpoints. Reach it through
// Client.Products.
type ProductService struct {
	core *client
}

// Create creates a product with its pricing. Requires the products:write scope.
func (s *ProductService) Create(ctx context.Context, params *ProductCreateParams, opts ...RequestOption) (*Product, error) {
	var out Product
	if err := s.core.do(ctx, "POST", "/v1/products", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single product by ID. Requires the products:read scope.
func (s *ProductService) Get(ctx context.Context, id string, opts ...RequestOption) (*Product, error) {
	var out Product
	if err := s.core.do(ctx, "GET", "/v1/products/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the fields set on params and returns the updated product.
// Requires the products:write scope.
func (s *ProductService) Update(ctx context.Context, id string, params *ProductUpdateParams, opts ...RequestOption) (*Product, error) {
	var out Product
	if err := s.core.do(ctx, "PATCH", "/v1/products/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of products, most recent first. Archived products are
// excluded unless params.IncludeArchived is set. Requires the products:read
// scope. Use All to iterate across every page automatically.
func (s *ProductService) List(ctx context.Context, params *ProductListParams, opts ...RequestOption) (*Page[Product], error) {
	var out Page[Product]
	if err := s.core.do(ctx, "GET", "/v1/products", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// All iterates over every product matching params, fetching pages as needed by
// following the response cursor. Range over it with Go 1.23+ iterators:
//
//	for prod, err := range client.Products.All(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Println(prod.ID)
//	}
func (s *ProductService) All(ctx context.Context, params *ProductListParams, opts ...RequestOption) iter.Seq2[Product, error] {
	// Copy params so advancing the cursor never mutates the caller's value.
	var p ProductListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Product], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore || m.NextCursor == nil {
				return false
			}
			p.Cursor = *m.NextCursor
			return true
		},
	)
}

// Archive archives a product so it can no longer be used in new checkouts or
// subscriptions; existing subscriptions keep billing. It is idempotent and
// returns the updated product. Requires the products:write scope.
func (s *ProductService) Archive(ctx context.Context, id string, opts ...RequestOption) (*Product, error) {
	var out Product
	if err := s.core.do(ctx, "POST", "/v1/products/"+url.PathEscape(id)+"/archive", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Unarchive restores an archived product to active status. It is idempotent and
// returns the updated product. Requires the products:write scope.
func (s *ProductService) Unarchive(ctx context.Context, id string, opts ...RequestOption) (*Product, error) {
	var out Product
	if err := s.core.do(ctx, "POST", "/v1/products/"+url.PathEscape(id)+"/unarchive", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
