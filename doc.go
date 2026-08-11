// Package bachs is a Go client for the Bachs API — payments and billing
// infrastructure for internet businesses.
//
// # Getting started
//
// Construct a client with your secret key and reach each resource through its
// service field:
//
//	client, err := bachs.NewClient("sk_sandbox_...")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	cust, err := client.Customers.Create(ctx, &bachs.CustomerCreateParams{
//	    Email: "jane@example.com",
//	    Name:  bachs.String("Jane Doe"),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(cust.ID) // cust_...
//
// Every method takes a [context.Context] first and a variadic tail of
// [RequestOption] values last, so a call is always
// (ctx, [id,] [params,] opts...).
//
// # Environments
//
// The base URL is selected automatically from the key prefix: sk_sandbox_ keys
// target the sandbox ([SandboxBaseURL]) and sk_live_ keys target production
// ([ProductionBaseURL]). Going live is a key swap — no code change. Point the
// client somewhere else (a proxy or a mock) with [WithBaseURL].
//
// # Errors
//
// Any non-2xx response is returned as an [*APIError] carrying the HTTP status,
// the stable ErrorCode to branch on, the human-readable Detail, any per-field
// validation Fields, and the RequestID for support. Test for common cases with
// the helpers rather than comparing strings:
//
//	_, err := client.Customers.Get(ctx, id)
//	switch {
//	case bachs.IsNotFound(err):
//	    // no such customer
//	case bachs.IsRateLimited(err):
//	    // backed off automatically already; still failing
//	case err != nil:
//	    var apiErr *bachs.APIError
//	    if errors.As(err, &apiErr) {
//	        log.Printf("bachs %s: %s (request %s)", apiErr.ErrorCode, apiErr.Detail, apiErr.RequestID)
//	    }
//	}
//
// # Idempotency and retries
//
// The client retries idempotent requests (GET, HEAD, DELETE) on 429 and 5xx
// responses, honoring any Retry-After header with exponential backoff and
// jitter, up to [WithMaxRetries] attempts. Writes (POST, PATCH, PUT) are only
// retried when you attach an idempotency key, which makes the retry safe:
//
//	cust, err := client.Customers.Create(ctx, params,
//	    bachs.WithIdempotencyKey("order-42"))
//
// # Pagination
//
// Every List method returns a single [Page]. To iterate across every page
// without managing cursors or offsets yourself, range over the matching All
// method, which yields one item at a time and fetches pages as needed:
//
//	for prod, err := range client.Products.All(ctx, nil) {
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Println(prod.ID)
//	}
//
// Use [Collect] to gather an iterator into a slice when you want them all in
// memory.
//
// # Field conventions
//
// Money is a decimal string such as "29.00" (never minor units); see [Money].
// Timestamps are UTC [time.Time]. Response fields that may be absent are
// pointers (*string, *time.Time); build request pointers with the [String],
// [Int], [Bool], and [Ptr] helpers. Fields that must distinguish "leave
// unchanged" from "set to null" from "set to a value" — such as a customer's
// billing address on update — use the tri-state [Opt] type via [Set] and
// [Null].
package bachs
