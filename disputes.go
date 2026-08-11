package bachs

import (
	"context"
	"iter"
	"net/url"
	"strconv"
	"time"
)

// Dispute status values, reported by Dispute.Status and the status filter on
// DisputeListParams.
const (
	DisputeStatusNeedsResponse = "needs_response"
	DisputeStatusUnderReview   = "under_review"
	DisputeStatusWon           = "won"
	DisputeStatusLost          = "lost"
	DisputeStatusClosed        = "closed"
)

// DisputeEvidence is the evidence assembled to contest a dispute. Every field
// is optional; attachment fields hold document IDs from
// DisputeService.UploadDocument.
type DisputeEvidence struct {
	AccessActivityLog                 *string `json:"access_activity_log"`
	BillingAddress                    *string `json:"billing_address"`
	CancellationPolicyAttachmentID    *string `json:"cancellation_policy_attachment_id"`
	CancellationPolicyDisclosure      *string `json:"cancellation_policy_disclosure"`
	CustomerCommunicationAttachmentID *string `json:"customer_communication_attachment_id"`
	CustomerEmailAddress              *string `json:"customer_email_address"`
	CustomerName                      *string `json:"customer_name"`
	Notes                             *string `json:"notes"`
	ProductDescription                *string `json:"product_description"`
	RefundPolicyAttachmentID          *string `json:"refund_policy_attachment_id"`
	RefundPolicyDisclosure            *string `json:"refund_policy_disclosure"`
	RefundRefusalExplanation          *string `json:"refund_refusal_explanation"`
	ServiceDate                       *string `json:"service_date"`
	UncategorizedAttachmentID         *string `json:"uncategorized_attachment_id"`
}

// DisputeSubmission records one submission of a dispute's evidence to the
// network.
type DisputeSubmission struct {
	SubmissionID    string     `json:"submission_id"`
	Status          string     `json:"status"`
	TriggerSource   string     `json:"trigger_source"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	FailedAt        *time.Time `json:"failed_at"`
	AttemptSequence int        `json:"attempt_sequence"`
}

// Dispute is the full state of a dispute (chargeback), as returned by
// DisputeService.Get.
type Dispute struct {
	DisputeID string  `json:"dispute_id"`
	ChargeID  *string `json:"charge_id"`
	Amount    Money   `json:"amount"`
	Currency  string  `json:"currency"`
	// Status is one of the DisputeStatus* constants.
	Status string `json:"status"`
	// IsResponseEditable reports whether evidence can still be updated or
	// submitted.
	IsResponseEditable bool    `json:"is_response_editable"`
	Reason             *string `json:"reason"`
	// ResponseDeadlineAt is when a response is due, or nil if not applicable.
	ResponseDeadlineAt *time.Time      `json:"response_deadline_at"`
	Evidence           DisputeEvidence `json:"evidence"`
	// LatestSubmission is the most recent evidence submission, or nil if none.
	LatestSubmission *DisputeSubmission `json:"latest_submission"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// DisputeSummary is the condensed dispute shape returned by list endpoints.
type DisputeSummary struct {
	DisputeID          string     `json:"dispute_id"`
	ChargeID           *string    `json:"charge_id"`
	Amount             Money      `json:"amount"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	IsResponseEditable bool       `json:"is_response_editable"`
	Reason             *string    `json:"reason"`
	ResponseDeadlineAt *time.Time `json:"response_deadline_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// DisputeDocument is the result of uploading a dispute evidence document. Use
// its DocumentID as an attachment ID in DisputeEvidenceUpdateParams.
type DisputeDocument struct {
	DocumentID      string    `json:"document_id"`
	FileName        string    `json:"file_name"`
	MimeType        string    `json:"mime_type"`
	FileSizeBytes   int       `json:"file_size_bytes"`
	StorageProvider string    `json:"storage_provider"`
	UploadedAt      time.Time `json:"uploaded_at"`
}

// DisputeEvidenceUpdateResult is returned by DisputeService.UpdateEvidence.
type DisputeEvidenceUpdateResult struct {
	DisputeID          string    `json:"dispute_id"`
	Status             string    `json:"status"`
	IsResponseEditable bool      `json:"is_response_editable"`
	EvidenceUpdatedAt  time.Time `json:"evidence_updated_at"`
}

// DisputeSubmitResult is returned by DisputeService.Submit.
type DisputeSubmitResult struct {
	DisputeID          string `json:"dispute_id"`
	Status             string `json:"status"`
	IsResponseEditable bool   `json:"is_response_editable"`
	Submission         struct {
		SubmissionID     string     `json:"submission_id"`
		SubmissionStatus string     `json:"submission_status"`
		TriggerSource    string     `json:"trigger_source"`
		SubmittedAt      *time.Time `json:"submitted_at"`
	} `json:"submission"`
}

// DisputeEvidenceUpdateParams are the parameters for
// DisputeService.UpdateEvidence. Every field is optional; only the fields you
// set are changed. Attachment fields take document IDs from UploadDocument.
type DisputeEvidenceUpdateParams struct {
	AccessActivityLog                 *string `json:"access_activity_log,omitempty"`
	BillingAddress                    *string `json:"billing_address,omitempty"`
	CancellationPolicyAttachmentID    *string `json:"cancellation_policy_attachment_id,omitempty"`
	CancellationPolicyDisclosure      *string `json:"cancellation_policy_disclosure,omitempty"`
	CustomerCommunicationAttachmentID *string `json:"customer_communication_attachment_id,omitempty"`
	CustomerEmailAddress              *string `json:"customer_email_address,omitempty"`
	CustomerName                      *string `json:"customer_name,omitempty"`
	Notes                             *string `json:"notes,omitempty"`
	ProductDescription                *string `json:"product_description,omitempty"`
	RefundPolicyAttachmentID          *string `json:"refund_policy_attachment_id,omitempty"`
	RefundPolicyDisclosure            *string `json:"refund_policy_disclosure,omitempty"`
	RefundRefusalExplanation          *string `json:"refund_refusal_explanation,omitempty"`
	ServiceDate                       *string `json:"service_date,omitempty"`
	UncategorizedAttachmentID         *string `json:"uncategorized_attachment_id,omitempty"`
}

// DisputeListParams filters and pages DisputeService.List. Disputes are
// paginated by offset.
type DisputeListParams struct {
	// Limit is the page size, 1–100.
	Limit int
	// Offset is the record offset to start from.
	Offset int
	// Status, when set, returns only disputes with that status; use a
	// DisputeStatus* constant.
	Status string
	// FromDate and ToDate bound results by creation time (ISO-8601).
	FromDate string
	ToDate   string
}

func (p *DisputeListParams) query() url.Values {
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
	if p.Status != "" {
		v.Set("status", p.Status)
	}
	if p.FromDate != "" {
		v.Set("from_date", p.FromDate)
	}
	if p.ToDate != "" {
		v.Set("to_date", p.ToDate)
	}
	return v
}

// DisputeService accesses the /v1/disputes endpoints. Reach it through
// Client.Disputes.
type DisputeService struct {
	core *client
}

// List returns one page of disputes. The endpoint reports only a total, so the
// returned page's HasMore is derived from offset + returned < total. Use All to
// iterate across every page automatically.
func (s *DisputeService) List(ctx context.Context, params *DisputeListParams, opts ...RequestOption) (*Page[DisputeSummary], error) {
	var wire struct {
		Total int              `json:"total"`
		Items []DisputeSummary `json:"items"`
	}
	if err := s.core.do(ctx, "GET", "/v1/disputes", params.query(), nil, &wire, opts); err != nil {
		return nil, err
	}
	offset := 0
	if params != nil {
		offset = params.Offset
	}
	returned := len(wire.Items)
	return &Page[DisputeSummary]{
		Items: wire.Items,
		Pagination: ListMeta{
			Total:    wire.Total,
			Offset:   offset,
			Returned: returned,
			HasMore:  offset+returned < wire.Total,
		},
	}, nil
}

// All iterates over every dispute matching params, advancing by offset.
func (s *DisputeService) All(ctx context.Context, params *DisputeListParams, opts ...RequestOption) iter.Seq2[DisputeSummary, error] {
	var p DisputeListParams
	if params != nil {
		p = *params
	}
	return paginate(
		func() (*Page[DisputeSummary], error) { return s.List(ctx, &p, opts...) },
		func(m ListMeta) bool {
			if !m.HasMore {
				return false
			}
			p.Offset = m.Offset + m.Returned
			return true
		},
	)
}

// Get retrieves a single dispute by ID, including its evidence and latest
// submission.
func (s *DisputeService) Get(ctx context.Context, id string, opts ...RequestOption) (*Dispute, error) {
	var out Dispute
	if err := s.core.do(ctx, "GET", "/v1/disputes/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEvidence sets or changes the evidence on a dispute. Only the fields set
// on params are changed. Evidence can be updated only while the dispute's
// IsResponseEditable is true.
func (s *DisputeService) UpdateEvidence(ctx context.Context, id string, params *DisputeEvidenceUpdateParams, opts ...RequestOption) (*DisputeEvidenceUpdateResult, error) {
	var out DisputeEvidenceUpdateResult
	if err := s.core.do(ctx, "PATCH", "/v1/disputes/"+url.PathEscape(id)+"/evidence", nil, params, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Submit submits the dispute's assembled evidence to the network for review.
func (s *DisputeService) Submit(ctx context.Context, id string, opts ...RequestOption) (*DisputeSubmitResult, error) {
	var out DisputeSubmitResult
	if err := s.core.do(ctx, "POST", "/v1/disputes/"+url.PathEscape(id)+"/submit", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadDocument uploads an evidence document (PDF, JPEG, PNG, or GIF, up to
// 10 MB) and returns its details. Use the returned DocumentID as an attachment
// ID in DisputeEvidenceUpdateParams.
func (s *DisputeService) UploadDocument(ctx context.Context, file FileUpload, opts ...RequestOption) (*DisputeDocument, error) {
	body, contentType, err := buildMultipart("file", file, nil)
	if err != nil {
		return nil, err
	}
	var out DisputeDocument
	if err := s.core.doMultipart(ctx, "POST", "/v1/disputes/uploads", body, contentType, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}
