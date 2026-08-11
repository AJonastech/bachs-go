package bachs

import (
	"context"
	"net/url"
	"time"
)

// Upload scope values for UploadService.Create.
const (
	// UploadScopeGeneral is the default scope for an upload.
	UploadScopeGeneral = "general"
	// UploadScopeProductMedia marks an upload as product imagery.
	UploadScopeProductMedia = "product-media"
)

// Upload is a stored file, as returned by the utilities upload endpoints. Use
// its UploadID where an API accepts an uploaded file (e.g. product media).
type Upload struct {
	UploadID      string  `json:"upload_id"`
	Provider      string  `json:"provider"`
	FileName      string  `json:"file_name"`
	MimeType      string  `json:"mime_type"`
	FileSizeBytes int     `json:"file_size_bytes"`
	URL           *string `json:"url"`
	// LinkedResourceType and LinkedResourceID identify the resource this upload
	// is attached to, if any.
	LinkedResourceType *string   `json:"linked_resource_type"`
	LinkedResourceID   *string   `json:"linked_resource_id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UploadService accesses the /v1/utilities/uploads endpoints for general file
// storage. Reach it through Client.Uploads.
type UploadService struct {
	core *client
}

// Create uploads a file (up to 20 MB) and returns its stored details. Pass a
// scope such as UploadScopeProductMedia to group the upload, or an empty string
// for the default ("general").
func (s *UploadService) Create(ctx context.Context, file FileUpload, scope string, opts ...RequestOption) (*Upload, error) {
	body, contentType, err := buildMultipart("file", file, map[string]string{"scope": scope})
	if err != nil {
		return nil, err
	}
	var out Upload
	if err := s.core.doMultipart(ctx, "POST", "/v1/utilities/uploads", body, contentType, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get retrieves a single upload by ID.
func (s *UploadService) Get(ctx context.Context, id string, opts ...RequestOption) (*Upload, error) {
	var out Upload
	if err := s.core.do(ctx, "GET", "/v1/utilities/uploads/"+url.PathEscape(id), nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes an upload. It returns nil on success.
func (s *UploadService) Delete(ctx context.Context, id string, opts ...RequestOption) error {
	var out struct {
		UploadID string `json:"upload_id"`
		Deleted  bool   `json:"deleted"`
	}
	return s.core.do(ctx, "DELETE", "/v1/utilities/uploads/"+url.PathEscape(id), nil, nil, &out, opts)
}
