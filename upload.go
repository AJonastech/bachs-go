package bachs

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

// FileUpload is a file to upload through a multipart/form-data endpoint such as
// DisputeService.UploadDocument. Provide the file contents via Reader and a
// FileName; ContentType is optional (the field's part uses it when set).
type FileUpload struct {
	// FileName is the name recorded for the uploaded file, e.g. "receipt.pdf".
	FileName string
	// Reader supplies the file contents.
	Reader io.Reader
	// ContentType optionally sets the part's Content-Type, e.g. "application/pdf".
	ContentType string
}

// buildMultipart encodes a multipart/form-data body with the file under the
// given form field name, preceded by any extra text fields. It returns the
// encoded bytes and the Content-Type header value (which includes the generated
// boundary). Text fields with an empty value are skipped.
func buildMultipart(field string, file FileUpload, fields map[string]string) ([]byte, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	for name, value := range fields {
		if value == "" {
			continue
		}
		if err := w.WriteField(name, value); err != nil {
			return nil, "", fmt.Errorf("bachs: building upload: %w", err)
		}
	}

	var (
		part io.Writer
		err  error
	)
	if file.ContentType != "" {
		// CreatePart lets us set a Content-Type on the file part; CreateFormFile
		// hard-codes application/octet-stream.
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, field, file.FileName))
		h.Set("Content-Type", file.ContentType)
		part, err = w.CreatePart(h)
	} else {
		part, err = w.CreateFormFile(field, file.FileName)
	}
	if err != nil {
		return nil, "", fmt.Errorf("bachs: building upload: %w", err)
	}
	if _, err := io.Copy(part, file.Reader); err != nil {
		return nil, "", fmt.Errorf("bachs: reading upload contents: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, "", fmt.Errorf("bachs: finalizing upload: %w", err)
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}
