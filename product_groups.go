package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// ProductGroup bundles related products under a single named group. It is the
// object returned by the product-group create, retrieve, update, and list
// endpoints.
type ProductGroup struct {
	// ID is the unique identifier.
	ID string `json:"id"`
	// OrganizationID is the organization that owns the group.
	OrganizationID string `json:"organization_id"`
	// Name is the display name of the group.
	Name string `json:"name"`
	// Products are the products in the group, in order.
	Products []Product `json:"products"`
	// CreatedAt is when the group was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the group was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// ProductGroupCreateParams are the parameters for ProductGroupService.Create.
// Both fields are required.
type ProductGroupCreateParams struct {
	// Name is the display name of the group. Required.
	Name string `json:"name"`
	// ProductIDs are the products to include, by ID. Required.
	ProductIDs []string `json:"product_ids"`
}

// ProductGroupUpdateParams are the parameters for ProductGroupService.Update.
// Only the fields you set are changed. Setting ProductIDs replaces the group's
// membership.
type ProductGroupUpdateParams struct {
	Name *string `json:"name,omitempty"`
	// ProductIDs, when set, replaces the group's products with this exact list.
	ProductIDs []string `json:"product_ids,omitempty"`
}

// ProductGroupListParams filters and pages ProductGroupService.List. Product
// groups are paginated by cursor.
type ProductGroupListParams struct {
	// Limit is the page size.
	Limit int
	// Cursor is a next_cursor from a previous page. Empty starts at the first
	// page.
	Cursor string
}

func (p *ProductGroupListParams) query() url.Values {
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
	return v
}

// ProductGroupGetParams are optional parameters for ProductGroupService.Get.
type ProductGroupGetParams struct {
	// IncludeArchived includes archived products in the returned group.
	IncludeArchived bool
}

func (p *ProductGroupGetParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.IncludeArchived {
		v.Set("include_archived", "true")
	}
	return v
}

// ProductGroupService accesses the /v1/product-groups endpoints. Reach it
// through Client.ProductGroups.
type ProductGroupService struct {
	core *client
}

// Create creates a product group. Requires the products:write scope.
func (s *ProductGroupService) Create(ctx context.Context, params *ProductGroupCreateParams, opts ...RequestOption) (*ProductGroup, error) {
	var out ProductGroup
	if err := s.core.do(ctx, "POST", "/v1/product-groups", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single product group by ID. Pass params to include archived
// products. Requires the products:read scope.
func (s *ProductGroupService) Get(ctx context.Context, id string, params *ProductGroupGetParams, opts ...RequestOption) (*ProductGroup, error) {
	var out ProductGroup
	if err := s.core.do(ctx, "GET", "/v1/product-groups/"+url.PathEscape(id), params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update changes the fields set on params and returns the updated group.
// Requires the products:write scope.
func (s *ProductGroupService) Update(ctx context.Context, id string, params *ProductGroupUpdateParams, opts ...RequestOption) (*ProductGroup, error) {
	var out ProductGroup
	if err := s.core.do(ctx, "PATCH", "/v1/product-groups/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a product group. The products themselves are not deleted.
// Requires the products:write scope.
func (s *ProductGroupService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	return s.core.do(ctx, "DELETE", "/v1/product-groups/"+url.PathEscape(id), nil, nil, nil, opts)
}

// List returns one page of product groups. Requires the products:read scope.
// Use All to iterate across every page automatically.
func (s *ProductGroupService) List(ctx context.Context, params *ProductGroupListParams, opts ...RequestOption) (*Page[ProductGroup], error) {
	var out Page[ProductGroup]
	if err := s.core.do(ctx, "GET", "/v1/product-groups", params.query(), nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// All iterates over every product group matching params, following the response
// cursor.
func (s *ProductGroupService) All(ctx context.Context, params *ProductGroupListParams, opts ...RequestOption) iter.Seq2[ProductGroup, error] {
	var p ProductGroupListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[ProductGroup], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore || m.NextCursor == nil {
				return false
			}
			p.Cursor = *m.NextCursor
			return true
		},
	)
}
