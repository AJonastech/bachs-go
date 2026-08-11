package bachs

import (
	"net/http"
	"strings"
)

// Option configures a Client at construction time. Pass options to NewClient.
type Option func(*client)

// WithBaseURL overrides the API base URL. By default the base URL is derived
// from the API key prefix (sk_live_ → production, sk_sandbox_ → sandbox). Use
// this to target a proxy or a mock server in tests. A trailing slash is
// trimmed.
func WithBaseURL(url string) Option {
	return func(c *client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithHTTPClient sets the underlying *http.Client, letting you control
// timeouts, transports, and proxies. Defaults to a client with a 60s timeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *client) {
		if hc != nil {
			c.httpClient = hc
		}
	}
}

// WithHTTPHeader adds a header sent on every request from the client, such as a
// custom tracing header. Authorization, Content-Type, Accept, and User-Agent
// are managed by the SDK and should be set via the dedicated options instead.
func WithHTTPHeader(key, value string) Option {
	return func(c *client) {
		c.defaultHeaders.Set(key, value)
	}
}

// WithMaxRetries sets how many times a failed request is retried on 429 and 5xx
// responses. The default is 2 (three attempts total). Pass 0 to disable
// retries.
func WithMaxRetries(n int) Option {
	return func(c *client) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

// WithUserAgent overrides the default User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// RequestOption configures a single API call. Every resource method accepts a
// variadic tail of RequestOptions.
type RequestOption func(*requestOptions)

// requestOptions holds per-call settings resolved from the RequestOptions.
type requestOptions struct {
	idempotencyKey string
	headers        http.Header
}

// WithIdempotencyKey sets the Idempotency-Key header on a POST or PATCH so a
// retried request is applied at most once. Bachs caches the first successful
// response for 24 hours and returns it for any later call with the same key.
// Use a value tied to the business operation, e.g. "order_ORD-12345".
//
// Supplying a key also makes the SDK's automatic retries apply to the mutating
// request, since the retry is then safe.
func WithIdempotencyKey(key string) RequestOption {
	return func(o *requestOptions) {
		o.idempotencyKey = key
	}
}

// WithRequestHeader sets an extra header on a single request. It overrides any
// client-level header of the same name for that call.
func WithRequestHeader(key, value string) RequestOption {
	return func(o *requestOptions) {
		if o.headers == nil {
			o.headers = http.Header{}
		}
		o.headers.Set(key, value)
	}
}

// resolveRequestOptions applies opts onto a fresh requestOptions.
func resolveRequestOptions(opts []RequestOption) requestOptions {
	var ro requestOptions
	for _, opt := range opts {
		opt(&ro)
	}
	return ro
}
