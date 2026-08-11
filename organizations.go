package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Fee handling values, reported by Organization.FeeHandling: who absorbs
// processing fees.
const (
	FeeHandlingOrgPays      = "org_pays_fee"
	FeeHandlingCustomerPays = "customer_pays_fee"
)

// Requirement setup status values, reported by AccountRequirements.SetupStatus.
const (
	SetupStatusIncomplete     = "incomplete"
	SetupStatusAwaitingReview = "awaiting_review"
	SetupStatusComplete       = "complete"
)

// Controller describes fee responsibility for a connected account.
type Controller struct {
	// Fees.Payer is "account" when the connected account pays its own fees.
	Fees struct {
		Payer string `json:"payer"`
	} `json:"fees"`
}

// RequirementError is a validation problem on a specific onboarding field.
type RequirementError struct {
	Field  string  `json:"field"`
	Code   *string `json:"code"`
	Reason *string `json:"reason"`
}

// RequirementEntry is one onboarding field and its current state.
type RequirementEntry struct {
	Field string `json:"field"`
	// Status is one of currently_due, eventually_due, past_due, or
	// pending_verification.
	Status string `json:"status"`
	// RestrictsCapabilities lists the capabilities blocked until this field is
	// satisfied.
	RestrictsCapabilities []string `json:"restricts_capabilities"`
	Errors                []struct {
		Code   *string `json:"code"`
		Reason *string `json:"reason"`
	} `json:"errors"`
}

// AccountRequirements is the onboarding checklist for a connected account: the
// fields due now, later, or under review, plus any errors.
type AccountRequirements struct {
	// SetupStatus is one of the SetupStatus* constants.
	SetupStatus         string             `json:"setup_status"`
	CurrentlyDue        []string           `json:"currently_due"`
	EventuallyDue       []string           `json:"eventually_due"`
	PastDue             []string           `json:"past_due"`
	PendingVerification []string           `json:"pending_verification"`
	Errors              []RequirementError `json:"errors"`
	Entries             []RequirementEntry `json:"entries"`
}

// Organization is a Bachs account. Your own organization is returned by
// OrganizationService.Me; connected accounts you manage are also Organizations,
// returned by the connected-account endpoints.
type Organization struct {
	ID                   string  `json:"id"`
	Name                 *string `json:"name"`
	OwnerUserID          string  `json:"owner_user_id"`
	ParentOrganizationID *string `json:"parent_organization_id"`
	Country              *string `json:"country"`
	// FeeHandling is one of the FeeHandling* constants.
	FeeHandling string `json:"fee_handling"`
	// EnabledPaymentMethods maps payment method types to their configuration.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods"`
	AdaptivePricing       bool           `json:"adaptive_pricing"`
	BalanceCurrencies     []string       `json:"balance_currencies"`
	Website               *string        `json:"website"`
	PhoneNumber           *string        `json:"phone_number"`
	CompanyName           *string        `json:"company_name"`
	// EnabledCapabilities lists the capabilities granted to the account.
	EnabledCapabilities []string `json:"enabled_capabilities"`
	// Capabilities maps capability names to their status.
	Capabilities map[string]any `json:"capabilities"`
	// Requirements is the onboarding checklist, for connected accounts.
	Requirements              *AccountRequirements `json:"requirements"`
	FieldsNeedingResubmission *int                 `json:"fields_needing_resubmission"`
	SandboxOrgID              *string              `json:"sandbox_org_id"`
	LiveOrgID                 *string              `json:"live_org_id"`
	IsActive                  bool                 `json:"is_active"`
	CreatedAt                 time.Time            `json:"created_at"`
	UpdatedAt                 time.Time            `json:"updated_at"`
	Controller                *Controller          `json:"controller"`
}

// Checkout settings fee preference values.
const (
	FeePreferenceCustomerPays = "customer_pays"
	FeePreferenceOrgPays      = "org_pays"
)

// CheckoutSettings is an organization's checkout configuration, returned by
// OrganizationService.CheckoutSettings.
type CheckoutSettings struct {
	OrganizationID string `json:"organization_id"`
	// EnabledPaymentMethods maps payment method types to whether they are on.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods"`
	// FeePreference is FeePreferenceCustomerPays or FeePreferenceOrgPays.
	FeePreference string `json:"fee_preference"`
	// AvailableCurrencies describes the currencies available at checkout.
	AvailableCurrencies map[string]any `json:"available_currencies"`
}

// CheckoutSettingsUpdateResult is returned by
// OrganizationService.UpdateCheckoutSettings.
type CheckoutSettingsUpdateResult struct {
	OrganizationID        string         `json:"organization_id"`
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods"`
	FeePreference         string         `json:"fee_preference"`
	Message               string         `json:"message"`
}

// CheckoutSettingsUpdateParams updates an organization's checkout settings. Set
// only the fields you want to change.
type CheckoutSettingsUpdateParams struct {
	// EnabledPaymentMethods maps payment method types to whether they are on.
	EnabledPaymentMethods map[string]any `json:"enabled_payment_methods,omitempty"`
	// FeePreference is FeePreferenceCustomerPays or FeePreferenceOrgPays.
	FeePreference string `json:"fee_preference,omitempty"`
}

// ConnectedAccountCreateParams are the parameters for
// OrganizationService.CreateConnectedAccount. Only ContactEmail is required.
type ConnectedAccountCreateParams struct {
	// ContactEmail is the account's contact email. Required.
	ContactEmail string  `json:"contact_email"`
	DisplayName  *string `json:"display_name,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Country      *string `json:"country,omitempty"`
	// EntityType is "company", "individual", or "business".
	EntityType *string `json:"entity_type,omitempty"`
	// Capabilities requests specific capabilities for the account.
	Capabilities map[string]any `json:"capabilities,omitempty"`
	// Controller sets fee responsibility.
	Controller *Controller `json:"controller,omitempty"`
}

// ConnectedAccountListParams pages OrganizationService.ListConnectedAccounts.
// Connected accounts are paginated by offset.
type ConnectedAccountListParams struct {
	Limit  int
	Offset int
}

func (p *ConnectedAccountListParams) query() url.Values {
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

// OrganizationService accesses the /v1/organizations endpoints, including your
// own organization, its checkout settings, and the connected accounts it owns.
// Reach it through Client.Organizations.
type OrganizationService struct {
	core *client
}

// Me retrieves the organization that owns the API key in use.
func (s *OrganizationService) Me(ctx context.Context, opts ...RequestOption) (*Organization, error) {
	var out Organization
	if err := s.core.do(ctx, "GET", "/v1/organizations/me", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves an organization by ID.
func (s *OrganizationService) Get(ctx context.Context, id string, opts ...RequestOption) (*Organization, error) {
	var out Organization
	if err := s.core.do(ctx, "GET", "/v1/organizations/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckoutSettings retrieves the organization's checkout configuration.
func (s *OrganizationService) CheckoutSettings(ctx context.Context, opts ...RequestOption) (*CheckoutSettings, error) {
	var out CheckoutSettings
	if err := s.core.do(ctx, "GET", "/v1/organizations/checkout/settings", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateCheckoutSettings changes the organization's checkout configuration.
func (s *OrganizationService) UpdateCheckoutSettings(ctx context.Context, params *CheckoutSettingsUpdateParams, opts ...RequestOption) (*CheckoutSettingsUpdateResult, error) {
	var out CheckoutSettingsUpdateResult
	if err := s.core.do(ctx, "PUT", "/v1/organizations/checkout/settings", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateConnectedAccount creates a connected account under this organization.
// Requires the connect:write scope.
func (s *OrganizationService) CreateConnectedAccount(ctx context.Context, params *ConnectedAccountCreateParams, opts ...RequestOption) (*Organization, error) {
	var out Organization
	if err := s.core.do(ctx, "POST", "/v1/organizations/connected-accounts", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConnectedAccounts returns one page of the organization's connected
// accounts. The endpoint reports total, limit, and offset rather than a cursor,
// so the returned page's HasMore is derived from offset + returned < total. Use
// AllConnectedAccounts to iterate across every page.
func (s *OrganizationService) ListConnectedAccounts(ctx context.Context, params *ConnectedAccountListParams, opts ...RequestOption) (*Page[Organization], error) {
	var wire struct {
		Items  []Organization `json:"items"`
		Total  int            `json:"total"`
		Limit  int            `json:"limit"`
		Offset int            `json:"offset"`
	}
	if err := s.core.do(ctx, "GET", "/v1/organizations/connected-accounts", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	returned := len(wire.Items)
	return &Page[Organization]{
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

// AllConnectedAccounts iterates over every connected account, advancing by
// offset.
func (s *OrganizationService) AllConnectedAccounts(ctx context.Context, params *ConnectedAccountListParams, opts ...RequestOption) iter.Seq2[Organization, error] {
	var p ConnectedAccountListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Organization], error) { return s.ListConnectedAccounts(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}
