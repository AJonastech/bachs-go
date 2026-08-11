package bachs

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

// TestDisputeUploadDocument_Multipart verifies the multipart/form-data upload
// path: the request carries a multipart body with the file part named "file",
// preserving the filename and content type, and the response decodes.
func TestDisputeUploadDocument_Multipart(t *testing.T) {
	var gotContentType, gotFileName, gotPartType, gotContents string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		mediaType, params, err := mime.ParseMediaType(gotContentType)
		if err != nil || mediaType != "multipart/form-data" {
			t.Errorf("Content-Type = %q, want multipart/form-data", gotContentType)
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("reading part: %v", err)
		}
		if part.FormName() != "file" {
			t.Errorf("form field = %q, want file", part.FormName())
		}
		gotFileName = part.FileName()
		gotPartType = part.Header.Get("Content-Type")
		b, _ := io.ReadAll(part)
		gotContents = string(b)

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"document_id":"doc_1","file_name":"receipt.pdf","mime_type":"application/pdf","file_size_bytes":7,"storage_provider":"s3","uploaded_at":"2026-01-01T00:00:00Z"}`)
	})

	doc, err := c.Disputes.UploadDocument(context.Background(), FileUpload{
		FileName:    "receipt.pdf",
		Reader:      strings.NewReader("receipt"),
		ContentType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("UploadDocument: %v", err)
	}
	if gotFileName != "receipt.pdf" {
		t.Errorf("filename = %q, want receipt.pdf", gotFileName)
	}
	if gotPartType != "application/pdf" {
		t.Errorf("part Content-Type = %q, want application/pdf", gotPartType)
	}
	if gotContents != "receipt" {
		t.Errorf("contents = %q, want receipt", gotContents)
	}
	if doc.DocumentID != "doc_1" {
		t.Errorf("DocumentID = %q, want doc_1", doc.DocumentID)
	}
}

// TestDisputeList_SynthesizedPagination checks that the {total,items} envelope
// is adapted into a Page whose HasMore reflects offset + returned < total.
func TestDisputeList_SynthesizedPagination(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("status"); got != DisputeStatusNeedsResponse {
			t.Errorf("status query = %q", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"total":3,"items":[{"dispute_id":"dp_1"},{"dispute_id":"dp_2"}]}`)
	})

	page, err := c.Disputes.List(context.Background(), &DisputeListParams{
		Status: DisputeStatusNeedsResponse,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(page.Items))
	}
	if !page.Pagination.HasMore {
		t.Error("HasMore = false, want true (0+2 < 3)")
	}
	if page.Pagination.Total != 3 {
		t.Errorf("Total = %d, want 3", page.Pagination.Total)
	}
}
