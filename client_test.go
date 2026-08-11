package bachs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient starts an httptest server with handler and returns a Client
// pointed at it, plus the server (which the caller must Close). Retries use a
// tiny backoff so retry tests stay fast.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	c, err := NewClient("sk_sandbox_test", WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	c.core.baseBackoff = time.Millisecond
	return c, srv
}

// decodeJSON reads r as a JSON object into a map, failing the test on error.
func decodeJSON(t *testing.T, r io.Reader) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("decode body: %v (body: %s)", err, body)
	}
	return m
}

// TestIsNilBody covers the typed-nil detection used to decide whether a request
// carries a body. A (*T)(nil) passed through an any parameter is not == nil, so
// without this it would marshal to the literal "null".
func TestIsNilBody(t *testing.T) {
	type params struct{ A int }
	var nilPtr *params
	var nilMap map[string]any
	cases := []struct {
		name string
		body any
		want bool
	}{
		{"untyped nil", nil, true},
		{"typed nil pointer", nilPtr, true},
		{"nil map", nilMap, true},
		{"non-nil pointer", &params{A: 1}, false},
		{"struct value", params{A: 1}, false},
		{"map value", map[string]any{"a": 1}, false},
	}
	for _, tc := range cases {
		if got := isNilBody(tc.body); got != tc.want {
			t.Errorf("isNilBody(%s) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestNewClient_RequiresKey(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Fatal("expected an error for an empty API key, got nil")
	}
}

func TestBaseURLFromKeyPrefix(t *testing.T) {
	cases := map[string]string{
		"sk_sandbox_abc": SandboxBaseURL,
		"sk_live_abc":    ProductionBaseURL,
		"weird_key":      ProductionBaseURL,
	}
	for key, want := range cases {
		if got := baseURLForKey(key); got != want {
			t.Errorf("baseURLForKey(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestWithBaseURLOverride(t *testing.T) {
	c, err := NewClient("sk_live_abc", WithBaseURL("https://proxy.example.com/"))
	if err != nil {
		t.Fatal(err)
	}
	if c.core.baseURL != "https://proxy.example.com" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.core.baseURL)
	}
}

func TestAuthAndDefaultHeaders(t *testing.T) {
	var gotAuth, gotAccept, gotUA, gotContentType string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotUA = r.Header.Get("User-Agent")
		gotContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"customer_id":"cust_1"}`)
	})

	_, err := c.Customers.Create(context.Background(), &CustomerCreateParams{Email: "a@b.com"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if gotAuth != "Bearer sk_sandbox_test" {
		t.Errorf("Authorization = %q", gotAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("Accept = %q", gotAccept)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q", gotContentType)
	}
	if gotUA == "" {
		t.Error("User-Agent was not set")
	}
}

func TestErrorDecoding_Validation(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req_123")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{
			"detail": "Missing required field(s): email",
			"error_code": "VALIDATION_ERROR",
			"doc_url": "https://docs.bachs.io/x",
			"errors": [{"field":"email","message":"Field required","type":"missing"}]
		}`)
	})

	_, err := c.Customers.Create(context.Background(), &CustomerCreateParams{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !IsValidation(err) {
		t.Errorf("IsValidation = false for %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %T", err)
	}
	if apiErr.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req_123" {
		t.Errorf("RequestID = %q, want req_123", apiErr.RequestID)
	}
	if len(apiErr.Fields) != 1 || apiErr.Fields[0].Field != "email" {
		t.Errorf("Fields = %+v, want one entry for email", apiErr.Fields)
	}
}

func TestErrorDecoding_NotFound(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"detail":"Customer not found","error_code":"NOT_FOUND"}`)
	})

	_, err := c.Customers.Get(context.Background(), "cust_missing")
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false for %v", err)
	}
}

func TestRetryOn429ThenSuccess(t *testing.T) {
	var attempts int
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"detail":"slow down","error_code":"TOO_MANY_REQUESTS"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"customer_id":"cust_1"}`)
	})

	cust, err := c.Customers.Get(context.Background(), "cust_1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2 (one retry)", attempts)
	}
	if cust.ID != "cust_1" {
		t.Errorf("ID = %q", cust.ID)
	}
}

func TestNoRetryOnPOSTWithoutIdempotencyKey(t *testing.T) {
	var attempts int
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"detail":"boom","error_code":"INTERNAL_SERVER_ERROR"}`)
	})

	_, err := c.Customers.Create(context.Background(), &CustomerCreateParams{Email: "a@b.com"})
	if err == nil {
		t.Fatal("expected an error")
	}
	if attempts != 1 {
		t.Errorf("attempts = %d, want 1 (POST without idempotency key must not retry)", attempts)
	}
}

func TestRetryOnPOSTWithIdempotencyKey(t *testing.T) {
	var attempts int
	var gotKey string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		gotKey = r.Header.Get("Idempotency-Key")
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = io.WriteString(w, `{"detail":"upstream","error_code":"BAD_GATEWAY"}`)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"customer_id":"cust_1"}`)
	})

	_, err := c.Customers.Create(
		context.Background(),
		&CustomerCreateParams{Email: "a@b.com"},
		WithIdempotencyKey("order_42"),
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2 (POST with idempotency key should retry)", attempts)
	}
	if gotKey != "order_42" {
		t.Errorf("Idempotency-Key = %q, want order_42", gotKey)
	}
}

func TestContextCancellationStopsRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		cancel() // cancel before the retry sleep completes
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `{"detail":"down","error_code":"SERVICE_UNAVAILABLE"}`)
	})
	c.core.baseBackoff = time.Hour // ensure the sleep would block if not cancelled

	_, err := c.Customers.Get(ctx, "cust_1")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
