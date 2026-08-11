package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Customer groups a buyer's payments, subscriptions, and saved payment methods
// under one record. It is the object returned by the customer create, retrieve,
// update, and list endpoints.
type Customer struct {
	// ID is the unique identifier, prefixed with "cust_". (Wire field:
	// customer_id.)
	ID string `json:"customer_id"`
	// Email is the customer's email address. Nil when not set.
	Email *string `json:"email"`
	// Name is the customer's full name. Nil when not set.
	Name *string `json:"name"`
	// PhoneNumber is in E.164 format, e.g. "+2348012345678". Nil when not set.
	// Not populated on list responses.
	PhoneNumber *string `json:"phone_number"`
	// Metadata is your own key/value data attached to the customer.
	Metadata Metadata `json:"metadata"`
	// BillingAddress is the customer's billing address, or nil if none is set.
	// Not populated on list responses.
	BillingAddress *Address `json:"billing_address"`
	// CreatedAt is when the customer was created.
	CreatedAt *time.Time `json:"created_at"`
	// UpdatedAt is when the customer was last updated. Not populated on list
	// responses.
	UpdatedAt *time.Time `json:"updated_at"`
}

// CustomerCreateParams are the parameters for CustomerService.Create. Only
// Email is required.
type CustomerCreateParams struct {
	// Email identifies the customer and receives receipts. Required.
	Email string `json:"email"`
	// Name is the customer's full name. Optional.
	Name *string `json:"name,omitempty"`
	// PhoneNumber in E.164 format, e.g. "+2348012345678". Optional.
	PhoneNumber *string `json:"phone_number,omitempty"`
	// Metadata is arbitrary key/value data, returned unchanged. Optional.
	Metadata Metadata `json:"metadata,omitempty"`
	// BillingAddress is optional on create. When supplied, Line1 and Country
	// are required. Omit it (leave nil) to create the customer without one.
	BillingAddress *Address `json:"billing_address,omitempty"`
}

// CustomerUpdateParams are the parameters for CustomerService.Update. Only the
// fields you set are changed; unset fields are left untouched.
type CustomerUpdateParams struct {
	Email       *string  `json:"email,omitempty"`
	Name        *string  `json:"name,omitempty"`
	PhoneNumber *string  `json:"phone_number,omitempty"`
	Metadata    Metadata `json:"metadata,omitempty"`
	// BillingAddress is tri-state, because the API replaces the address rather
	// than merging it:
	//
	//   - leave it as the zero value to leave the stored address untouched;
	//   - bachs.Null[bachs.Address]() to clear it;
	//   - bachs.Set(addr) to replace it in full — every component you omit from
	//     addr becomes null, so repeat the fields you want to keep.
	//
	// When you Set an address, Line1 and Country are required.
	BillingAddress Opt[Address] `json:"billing_address,omitzero"`
}

// CustomerListParams filters and pages CustomerService.List. Customers are
// paginated by offset. A zero field is omitted, letting the API apply its
// default (limit 50, offset 0).
type CustomerListParams struct {
	// Limit is the page size, 1–100. Defaults to 50 server-side.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// Search filters by customer email or name.
	Search string
}

// query encodes the params as URL query values. A nil receiver yields empty
// values.
func (p *CustomerListParams) query() url.Values {
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
	if p.Search != "" {
		v.Set("search", p.Search)
	}
	return v
}

// CustomerService accesses the /v1/customers endpoints. Reach it through
// Client.Customers.
type CustomerService struct {
	core *client
}

// Create creates a customer. Requires the customers:write scope.
func (s *CustomerService) Create(ctx context.Context, params *CustomerCreateParams, opts ...RequestOption) (*Customer, error) {
	var out Customer
	if err := s.core.do(ctx, "POST", "/v1/customers", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single customer by ID. Requires the customers:read scope.
func (s *CustomerService) Get(ctx context.Context, id string, opts ...RequestOption) (*Customer, error) {
	var out Customer
	if err := s.core.do(ctx, "GET", "/v1/customers/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the fields set on params and returns the updated customer.
// Requires the customers:write scope. See CustomerUpdateParams.BillingAddress
// for the address replace-vs-clear semantics.
func (s *CustomerService) Update(ctx context.Context, id string, params *CustomerUpdateParams, opts ...RequestOption) (*Customer, error) {
	var out Customer
	if err := s.core.do(ctx, "PATCH", "/v1/customers/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns one page of customers, most recent first. Requires the
// customers:read scope. Use All to iterate across every page automatically.
func (s *CustomerService) List(ctx context.Context, params *CustomerListParams, opts ...RequestOption) (*Page[Customer], error) {
	var out Page[Customer]
	if err := s.core.do(ctx, "GET", "/v1/customers", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// All iterates over every customer matching params, fetching pages as needed by
// advancing the offset. Range over it with Go 1.23+ iterators:
//
//	for cust, err := range client.Customers.All(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Println(cust.ID)
//	}
func (s *CustomerService) All(ctx context.Context, params *CustomerListParams, opts ...RequestOption) iter.Seq2[Customer, error] {
	// Copy params so advancing the offset never mutates the caller's value.
	var p CustomerListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Customer], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
