package bachs

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// Version is the SDK version, reported in the User-Agent header.
const Version = "0.1.0"

// Environment base URLs. The correct one is selected automatically from the API
// key prefix; override with WithBaseURL.
const (
	ProductionBaseURL = "https://api.bachs.io"
	SandboxBaseURL    = "https://sandbox-api.bachs.io"
)

// defaultTimeout is the request timeout of the default HTTP client. Override
// the whole client with WithHTTPClient to change it.
const defaultTimeout = 60 * time.Second

// defaultMaxRetries is how many times a request is retried on 429/5xx by
// default.
const defaultMaxRetries = 2

// Client is the entry point to the Bachs API. Construct one with NewClient and
// reach each resource through its service field (e.g. client.Customers).
//
// A Client is safe for concurrent use by multiple goroutines.
type Client struct {
	// core is the shared HTTP engine used by every resource service.
	core *client

	// Customers accesses the /v1/customers endpoints.
	Customers *CustomerService

	// Products accesses the /v1/products endpoints.
	Products *ProductService

	// Accounts accesses the /v1/accounts endpoints.
	Accounts *AccountService

	// Currencies accesses the /v1/currencies endpoints.
	Currencies *CurrencyService

	// PaymentMethods accesses the /v1/payment-methods endpoints.
	PaymentMethods *PaymentMethodService

	// Payments accesses the /v1/payments endpoints.
	Payments *PaymentService

	// Refunds accesses the /v1/refunds endpoints.
	Refunds *RefundService

	// Subscriptions accesses the /v1/subscriptions endpoints.
	Subscriptions *SubscriptionService

	// Transfers accesses the /v1/transfers endpoints.
	Transfers *TransferService
}

// NewClient creates a Client authenticated with the given secret key. The key
// is required; its prefix selects the environment base URL (sk_sandbox_ →
// sandbox, sk_live_ → production) unless overridden with WithBaseURL.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("bachs: api key is required")
	}

	core := &client{
		apiKey:         apiKey,
		baseURL:        baseURLForKey(apiKey),
		userAgent:      defaultUserAgent(),
		httpClient:     &http.Client{Timeout: defaultTimeout},
		maxRetries:     defaultMaxRetries,
		defaultHeaders: http.Header{},
	}
	for _, opt := range opts {
		opt(core)
	}

	c := &Client{core: core}
	c.Customers = &CustomerService{core: core}
	c.Products = &ProductService{core: core}
	c.Accounts = &AccountService{core: core}
	c.Currencies = &CurrencyService{core: core}
	c.PaymentMethods = &PaymentMethodService{core: core}
	c.Payments = &PaymentService{core: core}
	c.Refunds = &RefundService{core: core}
	c.Subscriptions = &SubscriptionService{core: core}
	c.Transfers = &TransferService{core: core}
	return c, nil
}

// defaultUserAgent identifies the SDK and Go runtime, e.g.
// "bachs-go/0.1.0 (go1.25.6)".
func defaultUserAgent() string {
	return fmt.Sprintf("bachs-go/%s (%s)", Version, runtime.Version())
}
