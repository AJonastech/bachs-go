package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Connect capability names, reported by ConnectedAccountCapability.Name and
// TaskCapabilityGroup.CapabilityName.
const (
	CapabilityPayouts     = "payouts"
	CapabilityTransfers   = "transfers"
	CapabilityConversions = "conversions"
	CapabilityConnect     = "connect"
)

// Connect capability status values, reported by ConnectedAccountCapability.Status.
const (
	CapabilityStatusActive      = "active"
	CapabilityStatusPending     = "pending"
	CapabilityStatusRestricted  = "restricted"
	CapabilityStatusUnrequested = "unrequested"
	CapabilityStatusUnsupported = "unsupported"
)

// Account link types, used for AccountLinkCreateParams.Type.
const (
	AccountLinkOnboarding = "onboarding"
	AccountLinkUpdate     = "update"
)

// CapabilityStatusDetail explains why a capability has its current status.
type CapabilityStatusDetail struct {
	Code       string  `json:"code"`
	Resolution *string `json:"resolution"`
	Message    *string `json:"message"`
}

// ConnectedAccountCapability is one capability of a connected account and its
// current status.
type ConnectedAccountCapability struct {
	// Name is one of the Capability* constants.
	Name string `json:"name"`
	// Status is one of the CapabilityStatus* constants.
	Status        string                   `json:"status"`
	Requested     bool                     `json:"requested"`
	StatusDetails []CapabilityStatusDetail `json:"status_details"`
}

// AccountLink is a time-limited URL that walks a connected account through
// onboarding or updating its details, as returned by
// ConnectedAccountService.CreateAccountLink.
type AccountLink struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Account string `json:"account"`
	// Type is AccountLinkOnboarding or AccountLinkUpdate.
	Type    string    `json:"type"`
	Created time.Time `json:"created"`
	// ExpiresAt is when the link stops working; generate a fresh one after.
	ExpiresAt time.Time `json:"expires_at"`
	URL       string    `json:"url"`
	// PreviousLinkSuperseded reports whether creating this link invalidated an
	// earlier one.
	PreviousLinkSuperseded bool `json:"previous_link_superseded"`
}

// TaskFieldReference points a task field at a related resource (a bank account
// or a person).
type TaskFieldReference struct {
	// Type is "account" or "person".
	Type     string  `json:"type"`
	Resource *string `json:"resource"`
	Label    *string `json:"label"`
}

// TaskFieldItem is one onboarding field in a checklist or capability group.
type TaskFieldItem struct {
	FieldKey string  `json:"field_key"`
	Label    string  `json:"label"`
	Group    *string `json:"group"`
	// State is one of currently_due, eventually_due, pending_verification,
	// pending_review, satisfied, or past_due.
	State       string              `json:"state"`
	Provided    bool                `json:"provided"`
	ErrorReason *string             `json:"error_reason"`
	Reference   *TaskFieldReference `json:"reference"`
}

// TaskCapabilityGroup groups the onboarding fields required for one capability.
type TaskCapabilityGroup struct {
	// CapabilityName is one of the Capability* constants.
	CapabilityName string  `json:"capability_name"`
	Description    *string `json:"description"`
	Category       *string `json:"category"`
	// State is "requested", "pending_review", or "enabled".
	State     string          `json:"state"`
	Satisfied bool            `json:"satisfied"`
	Fields    []TaskFieldItem `json:"fields"`
}

// TaskChecklist summarizes what a connected account still owes to complete
// onboarding, as returned by ConnectedAccountService.Checklist and SubmitTasks.
type TaskChecklist struct {
	OrganizationID string `json:"organization_id"`
	// EntityType is "company" or "individual" once known.
	EntityType     *string `json:"entity_type"`
	Country        *string `json:"country"`
	CurrentlyDue   int     `json:"currently_due"`
	PendingReview  int     `json:"pending_review"`
	InVerification int     `json:"in_verification"`
	NeedsAttention int     `json:"needs_attention"`
	// SetupStatus is one of the SetupStatus* constants.
	SetupStatus  string                `json:"setup_status"`
	Checklist    []TaskFieldItem       `json:"checklist"`
	Capabilities []TaskCapabilityGroup `json:"capabilities"`
}

// Task is a single onboarding task, as returned by
// ConnectedAccountService.Tasks.
type Task struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	// Type is "form_field", "document", "action", or "edit_section".
	Type string `json:"type"`
	// Status is "open", "in_review", "completed", or "rejected".
	Status   string  `json:"status"`
	FieldRef *string `json:"field_ref"`
	// DocumentType names the document expected, for document tasks.
	DocumentType *string `json:"document_type"`
	// Requirements and ResponseContract describe what the task expects and the
	// shape of an acceptable response.
	Requirements      map[string]any `json:"requirements"`
	ResponseContract  map[string]any `json:"response_contract"`
	DueDate           *time.Time     `json:"due_date"`
	ImpactsCapability *string        `json:"impacts_capability"`
	SectionKey        *string        `json:"section_key"`
	PastDue           bool           `json:"past_due"`
	RejectionReason   *string        `json:"rejection_reason"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// TaskValueItem is one submitted onboarding value.
type TaskValueItem struct {
	Field    string  `json:"field"`
	Label    string  `json:"label"`
	Group    *string `json:"group"`
	Provided bool    `json:"provided"`
	// Sensitive marks values (e.g. tax IDs) that are masked in Display.
	Sensitive bool `json:"sensitive"`
	// Value is the raw submitted value; nil when not provided or withheld.
	Value         any            `json:"value"`
	Display       *string        `json:"display"`
	ReferenceData map[string]any `json:"reference_data"`
}

// TaskValues is the set of values a connected account has submitted, as
// returned by ConnectedAccountService.TaskValues.
type TaskValues struct {
	OrganizationID string          `json:"organization_id"`
	EntityType     *string         `json:"entity_type"`
	Values         []TaskValueItem `json:"values"`
	// Persons holds submitted person records, if any.
	Persons []map[string]any `json:"persons"`
}

// TaskBank is a bank available for a connected account's onboarding, as
// returned by ConnectedAccountService.ListBanks.
type TaskBank struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// ResolvedTaskBankAccount is the result of resolving a bank account during
// onboarding, as returned by ConnectedAccountService.ResolveBankAccount.
type ResolvedTaskBankAccount struct {
	Resolved      bool    `json:"resolved"`
	AccountName   *string `json:"account_name"`
	AccountNumber *string `json:"account_number"`
	Message       *string `json:"message"`
}

// ReusableIdentity reports whether a previously verified person can be reused
// for this connected account, as returned by
// ConnectedAccountService.ReusableIdentity.
type ReusableIdentity struct {
	Available          bool     `json:"available"`
	PersonPublicID     *string  `json:"person_public_id"`
	FirstName          *string  `json:"first_name"`
	LastName           *string  `json:"last_name"`
	Country            *string  `json:"country"`
	VerificationStatus *string  `json:"verification_status"`
	UsedBy             []string `json:"used_by"`
}

// ReusableIdentityApplyResult is returned by
// ConnectedAccountService.ApplyReusableIdentity.
type ReusableIdentityApplyResult struct {
	Applied            bool   `json:"applied"`
	VerificationStatus string `json:"verification_status"`
}

// ConnectedAccountUpdateParams are the parameters for
// ConnectedAccountService.Update. Capabilities maps capability names (the
// Capability* constants) to a request such as {"requested": true}.
type ConnectedAccountUpdateParams struct {
	Capabilities map[string]any `json:"capabilities"`
}

// AccountLinkCreateParams are the parameters for
// ConnectedAccountService.CreateAccountLink. Type, RefreshURL, and ReturnURL
// are required.
type AccountLinkCreateParams struct {
	// Type is AccountLinkOnboarding or AccountLinkUpdate.
	Type string `json:"type"`
	// RefreshURL is where the account is sent if the link expires or is retried.
	RefreshURL string `json:"refresh_url"`
	// ReturnURL is where the account is sent on completion.
	ReturnURL string `json:"return_url"`
	// CollectionOptions optionally tunes what the hosted flow collects.
	CollectionOptions map[string]any `json:"collection_options,omitempty"`
}

// SubmitTasksParams are the parameters for ConnectedAccountService.SubmitTasks.
// Fields maps field keys to their values. Set Draft to save progress without
// submitting for review.
type SubmitTasksParams struct {
	Fields map[string]any `json:"fields,omitempty"`
	Draft  bool           `json:"draft,omitempty"`
}

// ResolveBankAccountParams are the parameters for
// ConnectedAccountService.ResolveBankAccount. AccountNumber and BankCode are
// required.
type ResolveBankAccountParams struct {
	AccountNumber string  `json:"account_number"`
	BankCode      string  `json:"bank_code"`
	Country       *string `json:"country,omitempty"`
}

// ApplyReusableIdentityParams are the parameters for
// ConnectedAccountService.ApplyReusableIdentity.
type ApplyReusableIdentityParams struct {
	PersonPublicID string `json:"person_public_id"`
}

// TaskListParams filters and pages ConnectedAccountService.Tasks. Tasks are
// paginated by offset.
type TaskListParams struct {
	// Status, when set, returns only tasks in these states (comma-separated).
	Status string
	// Limit is the page size.
	Limit int
	// Offset is the record offset to start from.
	Offset int
}

func (p *TaskListParams) query() url.Values {
	v := url.Values{}
	if p == nil {
		return v
	}
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if p.Offset > 0 {
		v.Set("offset", strconv.Itoa(p.Offset))
	}
	return v
}

// ConnectedAccountService accesses the per-account /v1/connected-accounts
// endpoints: account details, capability requests, hosted onboarding links, the
// requirements/tasks workflow, and account document uploads. Create and list
// connected accounts through Client.Organizations. Reach this service through
// Client.ConnectedAccounts.
type ConnectedAccountService struct {
	core *client
}

// Get retrieves a connected account by ID.
func (s *ConnectedAccountService) Get(ctx context.Context, id string, opts ...RequestOption) (*Organization, error) {
	var out Organization
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update requests capabilities on a connected account and returns the updated
// account.
func (s *ConnectedAccountService) Update(ctx context.Context, id string, params *ConnectedAccountUpdateParams, opts ...RequestOption) (*Organization, error) {
	var out Organization
	if err := s.core.do(ctx, "PATCH", "/v1/connected-accounts/"+url.PathEscape(id), nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAccountLink creates a hosted onboarding or update link for a connected
// account. The link is time-limited; check AccountLink.ExpiresAt.
func (s *ConnectedAccountService) CreateAccountLink(ctx context.Context, id string, params *AccountLinkCreateParams, opts ...RequestOption) (*AccountLink, error) {
	var out AccountLink
	if err := s.core.do(ctx, "POST", "/v1/connected-accounts/"+url.PathEscape(id)+"/account-links", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Capabilities lists a connected account's capabilities and their statuses.
func (s *ConnectedAccountService) Capabilities(ctx context.Context, id string, opts ...RequestOption) ([]ConnectedAccountCapability, error) {
	var out struct {
		Items []ConnectedAccountCapability `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/capabilities", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Items, nil
}

// Checklist retrieves the onboarding task checklist for a connected account.
func (s *ConnectedAccountService) Checklist(ctx context.Context, id string, opts ...RequestOption) (*TaskChecklist, error) {
	var out TaskChecklist
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/checklist", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Tasks returns one page of onboarding tasks for a connected account. The
// endpoint reports only a total, so the returned page's HasMore is derived from
// offset + returned < total. Use AllTasks to iterate across every page.
func (s *ConnectedAccountService) Tasks(ctx context.Context, id string, params *TaskListParams, opts ...RequestOption) (*Page[Task], error) {
	var wire struct {
		Total int    `json:"total"`
		Items []Task `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/tasks", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	offset := 0
	if params != nil {
		offset = params.Offset
	}
	returned := len(wire.Items)
	return &Page[Task]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Offset:   offset,
			Returned: returned,
			HasMore:  offset+returned < wire.Total,
		},
	}, nil
}

// AllTasks iterates over every onboarding task for a connected account,
// advancing by offset.
func (s *ConnectedAccountService) AllTasks(ctx context.Context, id string, params *TaskListParams, opts ...RequestOption) iter.Seq2[Task, error] {
	var p TaskListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[Task], error) { return s.Tasks(ctx, id, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}

// TaskValues retrieves the values a connected account has submitted so far.
func (s *ConnectedAccountService) TaskValues(ctx context.Context, id string, opts ...RequestOption) (*TaskValues, error) {
	var out TaskValues
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/values", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubmitTasks submits onboarding field values for a connected account and
// returns the refreshed checklist. Set params.Draft to save without submitting
// for review.
func (s *ConnectedAccountService) SubmitTasks(ctx context.Context, id string, params *SubmitTasksParams, opts ...RequestOption) (*TaskChecklist, error) {
	var out TaskChecklist
	if err := s.core.do(ctx, "POST", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/submit", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListBanks lists the banks available for a connected account's onboarding. If
// country is empty, the account's own country is used.
func (s *ConnectedAccountService) ListBanks(ctx context.Context, id, country string, opts ...RequestOption) ([]TaskBank, error) {
	q := url.Values{}
	if country != "" {
		q.Set("country", country)
	}
	var out struct {
		Country string     `json:"country"`
		Banks   []TaskBank `json:"banks"`
	}
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/banks", q, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Banks, nil
}

// ListMobileMoneyProviders lists the mobile money providers available for a
// connected account's onboarding. If country is empty, the account's own
// country is used.
func (s *ConnectedAccountService) ListMobileMoneyProviders(ctx context.Context, id, country string, opts ...RequestOption) ([]string, error) {
	q := url.Values{}
	if country != "" {
		q.Set("country", country)
	}
	var out struct {
		Country   string   `json:"country"`
		Providers []string `json:"providers"`
	}
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/momo", q, nil, &out, opts); err != nil {
		return nil, err
	}
	return out.Providers, nil
}

// ResolveBankAccount looks up the account holder name for a bank account during
// onboarding, so it can be confirmed before submission.
func (s *ConnectedAccountService) ResolveBankAccount(ctx context.Context, id string, params *ResolveBankAccountParams, opts ...RequestOption) (*ResolvedTaskBankAccount, error) {
	var out ResolvedTaskBankAccount
	if err := s.core.do(ctx, "POST", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/accounts/resolve", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ReusableIdentity reports whether an already-verified person can be reused for
// this connected account, avoiding re-verification.
func (s *ConnectedAccountService) ReusableIdentity(ctx context.Context, id string, opts ...RequestOption) (*ReusableIdentity, error) {
	var out ReusableIdentity
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/reusable-identity", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// ApplyReusableIdentity applies a previously verified person to this connected
// account.
func (s *ConnectedAccountService) ApplyReusableIdentity(ctx context.Context, id string, params *ApplyReusableIdentityParams, opts ...RequestOption) (*ReusableIdentityApplyResult, error) {
	var out ReusableIdentityApplyResult
	if err := s.core.do(ctx, "POST", "/v1/connected-accounts/"+url.PathEscape(id)+"/requirements/reusable-identity/apply", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadDocument uploads an onboarding document (up to 20 MB) for a connected
// account and returns its stored details. Pass a scope such as
// "company-documents" to group the file, or an empty string for the default
// ("general").
func (s *ConnectedAccountService) UploadDocument(ctx context.Context, id string, file FileUpload, scope string, opts ...RequestOption) (*Upload, error) {
	body, contentType, err := buildMultipart("file", file, map[string]string{"scope": scope})
	if err != nil {
		return nil, err
	}
	var out Upload
	if err := s.core.doMultipart(ctx, "POST", "/v1/connected-accounts/"+url.PathEscape(id)+"/uploads", body, contentType, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUpload retrieves a single uploaded account document by ID.
func (s *ConnectedAccountService) GetUpload(ctx context.Context, id, uploadID string, opts ...RequestOption) (*Upload, error) {
	var out Upload
	if err := s.core.do(ctx, "GET", "/v1/connected-accounts/"+url.PathEscape(id)+"/uploads/"+url.PathEscape(uploadID), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
