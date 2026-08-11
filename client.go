package bachs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// client is the shared, unexported core used by every resource service. It
// holds configuration and performs the actual HTTP work.
type client struct {
	apiKey         string
	baseURL        string
	userAgent      string
	httpClient     *http.Client
	maxRetries     int
	baseBackoff    time.Duration
	defaultHeaders http.Header
}

// baseURLForKey derives the environment base URL from the API key prefix.
// Production keys (sk_live_) target production; sandbox keys (sk_sandbox_)
// target the sandbox. Anything else defaults to production and can be
// overridden with WithBaseURL.
func baseURLForKey(apiKey string) string {
	switch {
	case strings.HasPrefix(apiKey, "sk_sandbox_"):
		return SandboxBaseURL
	default:
		return ProductionBaseURL
	}
}

// do executes an API request and decodes the result.
//
//   - method/path: the HTTP verb and an already-interpolated path such as
//     "/v1/customers/cust_123".
//   - query: URL query parameters, or nil.
//   - body: a value marshaled to a JSON request body, or nil for none.
//   - out: a pointer the successful (2xx) JSON body is decoded into, or nil to
//     discard it (e.g. 204 responses).
//
// On a non-2xx status it returns an *APIError. It retries 429 and 5xx
// responses per the client's retry policy.
func (c *client) do(
	ctx context.Context,
	method, path string,
	query url.Values,
	body, out any,
	opts []RequestOption,
) error {
	ro := resolveRequestOptions(opts)

	var bodyBytes []byte
	if !isNilBody(body) {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("bachs: encoding request body: %w", err)
		}
		bodyBytes = b
	}

	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	retryable := isRetryableMethod(method, ro.idempotencyKey != "")

	var lastErr error
	for attempt := 0; ; attempt++ {
		req, err := c.newRequest(ctx, method, endpoint, bodyBytes, ro)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Transport-level failure (timeout, connection reset). Retry if the
			// method is safe to replay and we have attempts left.
			lastErr = fmt.Errorf("bachs: request failed: %w", err)
			if retryable && attempt < c.maxRetries {
				if werr := sleepWithContext(ctx, backoff(attempt, 0, c.baseBackoff)); werr != nil {
					return werr
				}
				continue
			}
			return lastErr
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("bachs: reading response body: %w", err)
			if retryable && attempt < c.maxRetries {
				if werr := sleepWithContext(ctx, backoff(attempt, 0, c.baseBackoff)); werr != nil {
					return werr
				}
				continue
			}
			return lastErr
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return decodeSuccess(resp.StatusCode, respBody, out)
		}

		apiErr := parseAPIError(resp, respBody)
		if isRetryableStatus(resp.StatusCode) && retryable && attempt < c.maxRetries {
			if werr := sleepWithContext(ctx, backoff(attempt, apiErr.RetryAfter, c.baseBackoff)); werr != nil {
				return werr
			}
			lastErr = apiErr
			continue
		}
		return apiErr
	}
}

// newRequest builds a single *http.Request, applying default and per-request
// headers, authentication, and the idempotency key.
func (c *client) newRequest(
	ctx context.Context,
	method, endpoint string,
	bodyBytes []byte,
	ro requestOptions,
) (*http.Request, error) {
	var reader io.Reader
	if bodyBytes != nil {
		reader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, fmt.Errorf("bachs: building request: %w", err)
	}

	// Client-wide default headers first, so per-request headers can override.
	for key, values := range c.defaultHeaders {
		for _, v := range values {
			req.Header.Set(key, v)
		}
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if ro.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", ro.idempotencyKey)
	}
	for key, values := range ro.headers {
		for _, v := range values {
			req.Header.Set(key, v)
		}
	}

	return req, nil
}

// decodeSuccess decodes a 2xx body into out. A 204 or empty body with a non-nil
// out is a no-op; out stays at its zero value.
// isNilBody reports whether body should be treated as "no request body". It
// catches both an untyped nil and a typed nil pointer (e.g. a (*XParams)(nil)
// passed through the any parameter), which a plain body == nil check misses and
// which would otherwise marshal to the literal "null".
func isNilBody(body any) bool {
	if body == nil {
		return true
	}
	switch v := reflect.ValueOf(body); v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func decodeSuccess(status int, respBody []byte, out any) error {
	if out == nil || status == http.StatusNoContent || len(bytes.TrimSpace(respBody)) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("bachs: decoding response (status %d): %w", status, err)
	}
	return nil
}

// parseAPIError builds an *APIError from a non-2xx response.
func parseAPIError(resp *http.Response, respBody []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-Id"),
		Raw:        respBody,
	}

	var env errorEnvelope
	if err := json.Unmarshal(respBody, &env); err == nil {
		apiErr.Detail = env.Detail
		apiErr.ErrorCode = env.ErrorCode
		apiErr.DocURL = env.DocURL
		apiErr.Fields = env.Errors
	}
	if apiErr.Detail == "" {
		apiErr.Detail = strings.TrimSpace(string(respBody))
		if apiErr.Detail == "" {
			apiErr.Detail = http.StatusText(resp.StatusCode)
		}
	}

	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(ra)); err == nil {
			apiErr.RetryAfter = secs
		}
	}

	return apiErr
}

// isRetryableMethod reports whether a method is safe to replay. Reads are
// always safe; mutations are only safe when an idempotency key guarantees the
// server applies them once.
func isRetryableMethod(method string, hasIdempotencyKey bool) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		return true
	case http.MethodPost, http.MethodPatch, http.MethodPut:
		return hasIdempotencyKey
	default:
		return false
	}
}

// isRetryableStatus reports whether an HTTP status warrants a retry.
func isRetryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

// backoff returns how long to wait before the next attempt. It honors a
// server-provided Retry-After (in seconds) when present, otherwise uses
// exponential backoff with jitter starting from base.
func backoff(attempt, retryAfterSecs int, base time.Duration) time.Duration {
	if retryAfterSecs > 0 {
		return time.Duration(retryAfterSecs) * time.Second
	}
	if base <= 0 {
		base = 500 * time.Millisecond
	}
	d := base << attempt
	if max := 8 * time.Second; d > max {
		d = max
	}
	// Full jitter in [d/2, d].
	half := d / 2
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

// sleepWithContext waits for d or until ctx is done, whichever comes first.
func sleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
